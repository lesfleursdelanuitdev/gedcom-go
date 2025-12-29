package query

import (
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/badger/v4"
)

// debugHybrid is set via environment variable GEDCOM_DEBUG_HYBRID=1
var debugHybrid = os.Getenv("GEDCOM_DEBUG_HYBRID") == "1"

func debugLog(format string, args ...interface{}) {
	if debugHybrid {
		log.Printf("[HYBRID] "+format, args...)
	}
}

func debugBuildLog(format string, args ...interface{}) {
	if debugHybrid {
		log.Printf("[HYBRID_BUILD] "+format, args...)
	}
}

// hybridNodeLoader defines the interface for loading a specific node type from hybrid storage
type hybridNodeLoader struct {
	// getFromCache retrieves the node from cache if available
	getFromCache func(nodeID uint32) (GraphNode, bool)
	// getFromMemory retrieves the node from in-memory storage if available
	getFromMemory func(nodeID uint32) GraphNode
	// addToMemory adds the node to in-memory storage
	addToMemory func(nodeID uint32, node GraphNode)
	// typeName is used for error messages
	typeName string
}

// loadNodeFromHybrid is a generic helper that loads any node type from hybrid storage.
// It handles the common pattern: cache check -> SQLite lookup -> memory check -> BadgerDB load -> deserialize -> cache update
func (g *Graph) loadNodeFromHybrid(xrefID string, loader hybridNodeLoader) (GraphNode, error) {
	debugLog("loadNodeFromHybrid: xrefID=%s, type=%s", xrefID, loader.typeName)
	
	// Safety check: ensure queryHelpers is available (SQLite or PostgreSQL)
	if g.queryHelpers == nil && g.queryHelpersPostgres == nil {
		debugLog("loadNodeFromHybrid: queryHelpers and queryHelpersPostgres are both nil")
		return nil, nil
	}
	// Step 1: Check cache first
	if g.hybridCache != nil {
		if nodeID, found := g.hybridCache.GetXrefToID(xrefID); found {
			if cachedNode, ok := loader.getFromCache(nodeID); ok && cachedNode != nil {
				return cachedNode, nil
			}
		}
	}

	// Step 2: Get node ID from SQLite (with caching)
	// First check in-memory map (populated during build) as a fast path
	var nodeID uint32
	var err error
	g.mu.RLock()
	if id, exists := g.xrefToID[xrefID]; exists && id != 0 {
		nodeID = id
		g.mu.RUnlock()
		debugLog("loadNodeFromHybrid: found xrefToID in memory: %s -> %d", xrefID, nodeID)
	} else {
		g.mu.RUnlock()
		debugLog("loadNodeFromHybrid: xrefID not in memory map, querying database: %s", xrefID)
		// Not in memory, query database (SQLite or PostgreSQL)
		var queryHelper HybridQueryHelper
		if g.queryHelpersPostgres != nil {
			queryHelper = g.queryHelpersPostgres
		} else if g.queryHelpers != nil {
			queryHelper = g.queryHelpers
		} else {
			return nil, nil
		}

		if g.hybridCache != nil {
			if cachedID, found := g.hybridCache.GetXrefToID(xrefID); found {
				nodeID = cachedID
			} else {
				nodeID, err = queryHelper.FindByXref(xrefID)
				if err != nil {
					debugLog("loadNodeFromHybrid: FindByXref error: %v", err)
					// FindByXref returns (0, nil) when not found, so any error is a real error
					// But to match original behavior, return nil silently
					return nil, nil
				}
				if nodeID == 0 {
					debugLog("loadNodeFromHybrid: FindByXref returned 0 (not found): %s", xrefID)
					return nil, nil // Not found - return nil node and nil error (matches original behavior)
				}
				debugLog("loadNodeFromHybrid: found in database: %s -> %d", xrefID, nodeID)
				// Cache the mapping
				g.hybridCache.SetXrefToID(xrefID, nodeID)
			}
		} else {
			nodeID, err = queryHelper.FindByXref(xrefID)
			if err != nil {
				// FindByXref returns (0, nil) when not found, so any error is a real error
				// But to match original behavior, return nil silently
				return nil, nil
			}
			if nodeID == 0 {
				return nil, nil // Not found - return nil node and nil error (matches original behavior)
			}
		}
	}

	// Step 3: Check if already in memory
	g.mu.RLock()
	memNode := loader.getFromMemory(nodeID)
	g.mu.RUnlock()
	if memNode != nil {
		debugLog("loadNodeFromHybrid: node already in memory: %s (nodeID=%d, type=%T)", xrefID, nodeID, memNode)
		// Update cache
		if g.hybridCache != nil {
			g.hybridCache.SetNode(nodeID, memNode)
		}
		return memNode, nil
	}
	debugLog("loadNodeFromHybrid: node not in memory, loading from BadgerDB: %s (nodeID=%d)", xrefID, nodeID)

	// Step 4: Load from BadgerDB (works with both SQLite and PostgreSQL storage)
	var badgerDB *badger.DB
	if g.hybridStoragePostgres != nil {
		badgerDB = g.hybridStoragePostgres.BadgerDB()
	} else if g.hybridStorage != nil {
		badgerDB = g.hybridStorage.BadgerDB()
	} else {
		debugLog("loadNodeFromHybrid: hybridStorage and hybridStoragePostgres are both nil")
		return nil, nil
	}
	key := fmt.Sprintf("node:%d", nodeID)
	debugLog("loadNodeFromHybrid: loading from BadgerDB with key: %s", key)

	var nodeDataBytes []byte
	err = badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			nodeDataBytes = make([]byte, len(val))
			copy(nodeDataBytes, val)
			return nil
		})
	})
	if err != nil {
		// BadgerDB returns badger.ErrKeyNotFound when key doesn't exist
		// Return nil, nil to match original behavior (silently return nil when not found)
		if err == badger.ErrKeyNotFound {
			debugLog("loadNodeFromHybrid: key not found in BadgerDB: %s (xrefID=%s, nodeID=%d)", key, xrefID, nodeID)
			return nil, nil
		}
		// For other errors, return the error
		debugLog("loadNodeFromHybrid: BadgerDB error: %v", err)
		return nil, fmt.Errorf("failed to load %s %s from BadgerDB: %w", loader.typeName, xrefID, err)
	}
	debugLog("loadNodeFromHybrid: loaded %d bytes from BadgerDB for %s", len(nodeDataBytes), xrefID)

	// Step 5: Deserialize node
	node, err := DeserializeNode(nodeDataBytes, g)
	if err != nil || node == nil {
		debugLog("loadNodeFromHybrid: deserialization failed: err=%v, node=nil", err)
		// Return nil, nil to match original behavior (silently return nil when deserialization fails)
		return nil, nil
	}
	debugLog("loadNodeFromHybrid: successfully deserialized node: %s (nodeID=%d)", xrefID, nodeID)

	// Step 6: Add to in-memory cache
	g.mu.Lock()
	g.xrefToID[xrefID] = nodeID
	g.idToXref[nodeID] = xrefID
	g.nodes[nodeID] = node
	loader.addToMemory(nodeID, node)
	g.mu.Unlock()

	// Step 7: Update hybrid cache
	if g.hybridCache != nil {
		g.hybridCache.SetNode(nodeID, node)
		g.hybridCache.SetXrefToID(xrefID, nodeID)
	}

	// Step 8: Load edges
	g.loadEdgesFromHybrid(nodeID, node)

	debugLog("loadNodeFromHybrid: successfully loaded node: %s (nodeID=%d)", xrefID, nodeID)
	return node, nil
}

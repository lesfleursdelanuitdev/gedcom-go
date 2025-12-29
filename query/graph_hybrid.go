package query

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
)

// getIndividualFromHybrid loads an individual from hybrid storage
func (g *Graph) getIndividualFromHybrid(xrefID string) *IndividualNode {
	loader := hybridNodeLoader{
		getFromCache: func(nodeID uint32) (GraphNode, bool) {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if indiNode, ok := node.(*IndividualNode); ok {
					return indiNode, true
				}
			}
			return nil, false
		},
		getFromMemory: func(nodeID uint32) GraphNode {
			node, exists := g.individuals[nodeID]
			if debugHybrid {
				if exists && node != nil {
					log.Printf("[HYBRID] getFromMemory: found node in g.individuals[%d] = %T", nodeID, node)
				} else {
					log.Printf("[HYBRID] getFromMemory: g.individuals[%d] not found or nil (exists=%v, map has %d entries)", nodeID, exists, len(g.individuals))
				}
			}
			if !exists || node == nil {
				return nil // Return proper nil interface, not typed nil
			}
			return node
		},
		addToMemory: func(nodeID uint32, node GraphNode) {
			if indiNode, ok := node.(*IndividualNode); ok {
				g.individuals[nodeID] = indiNode
			}
		},
		typeName: "individual",
	}

	node, err := g.loadNodeFromHybrid(xrefID, loader)
	if err != nil {
		debugLog("getIndividualFromHybrid: loadNodeFromHybrid returned error: %v", err)
		return nil
	}
	if node == nil {
		debugLog("getIndividualFromHybrid: loadNodeFromHybrid returned nil node for %s", xrefID)
		return nil
	}

	indiNode, ok := node.(*IndividualNode)
	if !ok {
		debugLog("getIndividualFromHybrid: type assertion failed, got %T for %s", node, xrefID)
		return nil
	}

	debugLog("getIndividualFromHybrid: successfully loaded individual %s", xrefID)
	return indiNode
}

// getFamilyFromHybrid loads a family from hybrid storage
func (g *Graph) getFamilyFromHybrid(xrefID string) *FamilyNode {
	loader := hybridNodeLoader{
		getFromCache: func(nodeID uint32) (GraphNode, bool) {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if famNode, ok := node.(*FamilyNode); ok {
					return famNode, true
				}
			}
			return nil, false
		},
		getFromMemory: func(nodeID uint32) GraphNode {
			fam, exists := g.families[nodeID]
			if !exists || fam == nil {
				return nil // Return proper nil interface, not typed nil
			}
			return fam
		},
		addToMemory: func(nodeID uint32, node GraphNode) {
			if famNode, ok := node.(*FamilyNode); ok {
				g.families[nodeID] = famNode
			}
		},
		typeName: "family",
	}

	node, err := g.loadNodeFromHybrid(xrefID, loader)
	if err != nil || node == nil {
		return nil
	}

	famNode, ok := node.(*FamilyNode)
	if !ok {
		return nil
	}

	return famNode
}

// getNoteFromHybrid loads a note from hybrid storage
func (g *Graph) getNoteFromHybrid(xrefID string) *NoteNode {
	loader := hybridNodeLoader{
		getFromCache: func(nodeID uint32) (GraphNode, bool) {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if noteNode, ok := node.(*NoteNode); ok {
					return noteNode, true
				}
			}
			return nil, false
		},
		getFromMemory: func(nodeID uint32) GraphNode {
			note, exists := g.notes[nodeID]
			if !exists || note == nil {
				return nil // Return proper nil interface, not typed nil
			}
			return note
		},
		addToMemory: func(nodeID uint32, node GraphNode) {
			if noteNode, ok := node.(*NoteNode); ok {
				g.notes[nodeID] = noteNode
			}
		},
		typeName: "note",
	}

	node, err := g.loadNodeFromHybrid(xrefID, loader)
	if err != nil || node == nil {
		return nil
	}

	noteNode, ok := node.(*NoteNode)
	if !ok {
		return nil
	}

	return noteNode
}

// getSourceFromHybrid loads a source from hybrid storage
func (g *Graph) getSourceFromHybrid(xrefID string) *SourceNode {
	loader := hybridNodeLoader{
		getFromCache: func(nodeID uint32) (GraphNode, bool) {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if sourceNode, ok := node.(*SourceNode); ok {
					return sourceNode, true
				}
			}
			return nil, false
		},
		getFromMemory: func(nodeID uint32) GraphNode {
			source, exists := g.sources[nodeID]
			if !exists || source == nil {
				return nil // Return proper nil interface, not typed nil
			}
			return source
		},
		addToMemory: func(nodeID uint32, node GraphNode) {
			if sourceNode, ok := node.(*SourceNode); ok {
				g.sources[nodeID] = sourceNode
			}
		},
		typeName: "source",
	}

	node, err := g.loadNodeFromHybrid(xrefID, loader)
	if err != nil || node == nil {
		return nil
	}

	sourceNode, ok := node.(*SourceNode)
	if !ok {
		return nil
	}

	return sourceNode
}

// getRepositoryFromHybrid loads a repository from hybrid storage
func (g *Graph) getRepositoryFromHybrid(xrefID string) *RepositoryNode {
	loader := hybridNodeLoader{
		getFromCache: func(nodeID uint32) (GraphNode, bool) {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if repoNode, ok := node.(*RepositoryNode); ok {
					return repoNode, true
				}
			}
			return nil, false
		},
		getFromMemory: func(nodeID uint32) GraphNode {
			repo, exists := g.repositories[nodeID]
			if !exists || repo == nil {
				return nil // Return proper nil interface, not typed nil
			}
			return repo
		},
		addToMemory: func(nodeID uint32, node GraphNode) {
			if repoNode, ok := node.(*RepositoryNode); ok {
				g.repositories[nodeID] = repoNode
			}
		},
		typeName: "repository",
	}

	node, err := g.loadNodeFromHybrid(xrefID, loader)
	if err != nil || node == nil {
		return nil
	}

	repoNode, ok := node.(*RepositoryNode)
	if !ok {
		return nil
	}

	return repoNode
}

// getEventFromHybrid loads an event from hybrid storage
func (g *Graph) getEventFromHybrid(eventID string) *EventNode {
	loader := hybridNodeLoader{
		getFromCache: func(nodeID uint32) (GraphNode, bool) {
			if node, found := g.hybridCache.GetNode(nodeID); found {
				if eventNode, ok := node.(*EventNode); ok {
					return eventNode, true
				}
			}
			return nil, false
		},
		getFromMemory: func(nodeID uint32) GraphNode {
			event, exists := g.events[nodeID]
			if !exists || event == nil {
				return nil // Return proper nil interface, not typed nil
			}
			return event
		},
		addToMemory: func(nodeID uint32, node GraphNode) {
			if eventNode, ok := node.(*EventNode); ok {
				g.events[nodeID] = eventNode
			}
		},
		typeName: "event",
	}

	node, err := g.loadNodeFromHybrid(eventID, loader)
	if err != nil || node == nil {
		return nil
	}

	eventNode, ok := node.(*EventNode)
	if !ok {
		return nil
	}

	return eventNode
}

// loadEdgesFromHybrid loads edges for a node from BadgerDB (works with both SQLite and PostgreSQL storage)
func (g *Graph) loadEdgesFromHybrid(nodeID uint32, node GraphNode) {
	debugLog("loadEdgesFromHybrid: loading edges for nodeID=%d, node=%s", nodeID, node.ID())
	var badgerDB *badger.DB
	if g.hybridStoragePostgres != nil {
		badgerDB = g.hybridStoragePostgres.BadgerDB()
	} else if g.hybridStorage != nil {
		badgerDB = g.hybridStorage.BadgerDB()
	} else {
		debugLog("loadEdgesFromHybrid: hybridStorage and hybridStoragePostgres are both nil")
		return
	}
	key := fmt.Sprintf("edges:%d:out", nodeID)
	debugLog("loadEdgesFromHybrid: looking for edges with key: %s", key)

	err := badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			debugLog("loadEdgesFromHybrid: no edges found for key %s: %v", key, err)
			return err // No edges found
		}
		return item.Value(func(val []byte) error {
			var edges []EdgeData
			if err := deserialize(val, &edges); err != nil {
				debugLog("loadEdgesFromHybrid: failed to deserialize edges: %v", err)
				return err
			}
			debugLog("loadEdgesFromHybrid: deserialized %d edges for nodeID=%d", len(edges), nodeID)

			// Convert EdgeData to Edge objects and add to node
			for i, edgeData := range edges {
				debugLog("loadEdgesFromHybrid: processing edge %d/%d: type=%s, toID=%d", i+1, len(edges), edgeData.EdgeType, edgeData.ToID)
				// Get target node (use PostgreSQL or SQLite helpers)
				var toXref string
				var err error
				if g.queryHelpersPostgres != nil {
					toXref, err = g.queryHelpersPostgres.FindXrefByID(edgeData.ToID)
				} else if g.queryHelpers != nil {
					toXref, err = g.queryHelpers.FindXrefByID(edgeData.ToID)
				} else {
					debugLog("loadEdgesFromHybrid: no query helpers available")
					continue
				}
				if err != nil || toXref == "" {
					debugLog("loadEdgesFromHybrid: failed to find xref for toID=%d: err=%v, xref=%s", edgeData.ToID, err, toXref)
					continue
				}
				debugLog("loadEdgesFromHybrid: found toXref=%s for toID=%d", toXref, edgeData.ToID)

				// Get target node object based on edge type
				var toNode GraphNode
				switch edgeData.EdgeType {
				case EdgeTypeHUSB, EdgeTypeWIFE, EdgeTypeCHIL, EdgeTypeFAMC, EdgeTypeFAMS:
					// Family-related edges - target is individual
					toNode = g.GetIndividual(toXref)
					debugLog("loadEdgesFromHybrid: GetIndividual(%s) returned %v", toXref, toNode != nil)
				case EdgeTypeNOTE:
					// NOTE edge - target is note
					toNode = g.GetNote(toXref)
					debugLog("loadEdgesFromHybrid: GetNote(%s) returned %v", toXref, toNode != nil)
				case EdgeTypeSOUR:
					// SOUR edge - target is source
					toNode = g.GetSource(toXref)
					debugLog("loadEdgesFromHybrid: GetSource(%s) returned %v", toXref, toNode != nil)
				case EdgeTypeREPO:
					// REPO edge - target is repository
					toNode = g.GetRepository(toXref)
					debugLog("loadEdgesFromHybrid: GetRepository(%s) returned %v", toXref, toNode != nil)
				case EdgeTypeHasEvent:
					// has_event edge - target is event
					toNode = g.GetEvent(toXref)
					debugLog("loadEdgesFromHybrid: GetEvent(%s) returned %v", toXref, toNode != nil)
				default:
					// Other edge types - try GetNode
					toNode = g.GetNode(toXref)
					debugLog("loadEdgesFromHybrid: GetNode(%s) returned %v", toXref, toNode != nil)
				}

				// Check if toNode is actually nil (not just a typed nil)
				// In Go, a typed nil (*IndividualNode(nil)) assigned to an interface is not == nil
				// So we need to safely check if we can call methods on it
				var toNodeID string
				func() {
					defer func() {
						if r := recover(); r != nil {
							debugLog("loadEdgesFromHybrid: panic calling toNode.ID(): %v (typed nil)", r)
							toNodeID = "" // Mark as invalid
						}
					}()
					toNodeID = toNode.ID()
				}()

				if toNode == nil || toNodeID == "" {
					debugLog("loadEdgesFromHybrid: toNode is nil or has empty ID, skipping edge")
					continue
				}
				debugLog("loadEdgesFromHybrid: toNode is valid: %s (type=%T)", toNodeID, toNode)

				// Create edge
				var edge *Edge
				if edgeData.FamilyID != 0 {
					var famXref string
					if g.queryHelpersPostgres != nil {
						famXref, _ = g.queryHelpersPostgres.FindXrefByID(edgeData.FamilyID)
					} else if g.queryHelpers != nil {
						famXref, _ = g.queryHelpers.FindXrefByID(edgeData.FamilyID)
					}
					if famXref != "" {
						famNode := g.GetFamily(famXref)
						if famNode != nil {
							edgeID := fmt.Sprintf("%s_%s_%s", node.ID(), edgeData.EdgeType, toXref)
							edge = NewEdgeWithFamily(edgeID, node, toNode, edgeData.EdgeType, famNode)
							debugLog("loadEdgesFromHybrid: created edge with family: %s", edgeID)
						}
					}
				}
				if edge == nil {
					edgeID := fmt.Sprintf("%s_%s_%s", node.ID(), edgeData.EdgeType, toXref)
					edge = NewEdge(edgeID, node, toNode, edgeData.EdgeType)
					debugLog("loadEdgesFromHybrid: created edge: %s", edgeID)
				}

				// Verify edge was created correctly
				if edge == nil {
					debugLog("loadEdgesFromHybrid: failed to create edge")
					continue
				}
				if edge.To == nil {
					debugLog("loadEdgesFromHybrid: edge.To is nil, skipping")
					continue
				}
				// Verify edge.To has a valid ID (not a typed nil) - use the ID we already validated
				if toNodeID == "" {
					debugLog("loadEdgesFromHybrid: edge.To has invalid ID, skipping")
					continue
				}

				// Add edge to node using the interface methods
				// Note: We only add the forward edge here. Reverse edges will be added
				// when the target node's edges are loaded, to avoid circular dependencies
				// and ensure both nodes are fully initialized.
				node.AddOutEdge(edge)
				debugLog("loadEdgesFromHybrid: added edge to node: %s -> %s", node.ID(), toNodeID)
			}
			return nil
		})
	})
	if err != nil {
		// No edges or error - that's okay
		return
	}
}

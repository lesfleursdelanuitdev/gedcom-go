package query

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// BadgerDBStorage is an interface for storage types that have BadgerDB
type BadgerDBStorage interface {
	BadgerDB() *badger.DB
}

// buildGraphInBadgerDB stores graph structure in BadgerDB
// Works with both SQLite and PostgreSQL hybrid storage
func buildGraphInBadgerDB(storage BadgerDBStorage, tree *types.GedcomTree, graph *Graph) error {
	db := storage.BadgerDB()

	// Process all node types
	if err := processNodesForBadgerDB(db, tree, graph); err != nil {
		return err
	}

	// Process edges (after all nodes are stored)
	if err := buildEdgesInBadgerDB(db, tree, graph); err != nil {
		return fmt.Errorf("failed to build edges: %w", err)
	}

	// Update database indexes with relationship flags (has_children, has_spouse)
	// This needs to be done after edges are processed
	// Check which storage type we're using
	if graph.hybridStoragePostgres != nil {
		if err := updateRelationshipFlagsPostgreSQL(graph.hybridStoragePostgres, tree, graph); err != nil {
			return fmt.Errorf("failed to update relationship flags: %w", err)
		}
	} else if graph.hybridStorage != nil {
		if err := updateRelationshipFlags(graph.hybridStorage, tree, graph); err != nil {
			return fmt.Errorf("failed to update relationship flags: %w", err)
		}
	}

	return nil
}

// processNodesForBadgerDB processes all node types and stores them in BadgerDB
func processNodesForBadgerDB(db *badger.DB, tree *types.GedcomTree, graph *Graph) error {
	// Use batch write for better performance
	writeBatch := db.NewWriteBatch()
	defer writeBatch.Cancel()

	// Process individuals
	if err := processIndividualsForBadgerDB(writeBatch, tree, graph); err != nil {
		return err
	}

	// Process families
	if err := processFamiliesForBadgerDB(writeBatch, tree, graph); err != nil {
		return err
	}

	// Process notes
	if err := processNotesForBadgerDB(writeBatch, tree, graph); err != nil {
		return err
	}

	// Process sources
	if err := processSourcesForBadgerDB(writeBatch, tree, graph); err != nil {
		return err
	}

	// Process repositories
	if err := processRepositoriesForBadgerDB(writeBatch, tree, graph); err != nil {
		return err
	}

	// Process events
	if err := processEventsForBadgerDB(writeBatch, tree, graph); err != nil {
		return err
	}

	// Flush batch
	if err := writeBatch.Flush(); err != nil {
		return fmt.Errorf("failed to flush BadgerDB batch: %w", err)
	}

	return nil
}

// processIndividualsForBadgerDB processes individual records for BadgerDB
func processIndividualsForBadgerDB(writeBatch *badger.WriteBatch, tree *types.GedcomTree, graph *Graph) error {
	individuals := tree.GetAllIndividuals()

	for xrefID, record := range individuals {
		indiRecord, ok := record.(*types.IndividualRecord)
		if !ok {
			continue
		}

		// Get node ID (with locking)
		graph.mu.RLock()
		nodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if nodeID == 0 {
			debugLog("processIndividualsForBadgerDB: skipping %s (not in xrefToID mapping)", xrefID)
			continue // Skip if not in mapping
		}
		debugLog("processIndividualsForBadgerDB: storing %s (nodeID=%d)", xrefID, nodeID)

		// Create IndividualNode for serialization
		node := &IndividualNode{
			BaseNode: &BaseNode{
				xrefID:   xrefID,
				nodeType: NodeTypeIndividual,
				record:   indiRecord,
			},
			Individual: indiRecord,
		}

		// Serialize and store
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set node %s: %w", xrefID, err)
		}
		debugLog("processIndividualsForBadgerDB: stored %s in BadgerDB with key %s (%d bytes)", xrefID, key, len(data))
	}

	return nil
}

// processFamiliesForBadgerDB processes family records for BadgerDB
func processFamiliesForBadgerDB(writeBatch *badger.WriteBatch, tree *types.GedcomTree, graph *Graph) error {
	families := tree.GetAllFamilies()

	for xrefID, record := range families {
		famRecord, ok := record.(*types.FamilyRecord)
		if !ok {
			continue
		}

		// Get node ID (with locking)
		graph.mu.RLock()
		nodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if nodeID == 0 {
			continue
		}

		// Create FamilyNode for serialization
		node := &FamilyNode{
			BaseNode: &BaseNode{
				xrefID:   xrefID,
				nodeType: NodeTypeFamily,
				record:   famRecord,
			},
			Family: famRecord,
		}

		// Serialize and store
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize family node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set family node %s: %w", xrefID, err)
		}
	}

	return nil
}

// processNotesForBadgerDB processes note records for BadgerDB
func processNotesForBadgerDB(writeBatch *badger.WriteBatch, tree *types.GedcomTree, graph *Graph) error {
	notes := tree.GetAllNotes()

	for xrefID, record := range notes {
		noteRecord, ok := record.(*types.NoteRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		nodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if nodeID == 0 {
			continue
		}

		node := NewNoteNode(xrefID, noteRecord)
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize note node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set note node %s: %w", xrefID, err)
		}
	}

	return nil
}

// processSourcesForBadgerDB processes source records for BadgerDB
func processSourcesForBadgerDB(writeBatch *badger.WriteBatch, tree *types.GedcomTree, graph *Graph) error {
	sources := tree.GetAllSources()

	for xrefID, record := range sources {
		sourceRecord, ok := record.(*types.SourceRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		nodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if nodeID == 0 {
			continue
		}

		node := NewSourceNode(xrefID, sourceRecord)
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize source node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set source node %s: %w", xrefID, err)
		}
	}

	return nil
}

// processRepositoriesForBadgerDB processes repository records for BadgerDB
func processRepositoriesForBadgerDB(writeBatch *badger.WriteBatch, tree *types.GedcomTree, graph *Graph) error {
	repositories := tree.GetAllRepositories()

	for xrefID, record := range repositories {
		repoRecord, ok := record.(*types.RepositoryRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		nodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if nodeID == 0 {
			continue
		}

		node := NewRepositoryNode(xrefID, repoRecord)
		data, err := SerializeNode(node, graph)
		if err != nil {
			return fmt.Errorf("failed to serialize repository node %s: %w", xrefID, err)
		}

		key := fmt.Sprintf("node:%d", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set repository node %s: %w", xrefID, err)
		}
	}

	return nil
}

// processEventsForBadgerDB processes event nodes for BadgerDB
func processEventsForBadgerDB(writeBatch *badger.WriteBatch, tree *types.GedcomTree, graph *Graph) error {
	// Process individual events
	individualsForEvents := tree.GetAllIndividuals()
	for xrefID, record := range individualsForEvents {
		indi, ok := record.(*types.IndividualRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		indiNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if indiNodeID == 0 {
			continue
		}

		events := indi.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
			graph.mu.RLock()
			eventNodeID := graph.xrefToID[eventID]
			graph.mu.RUnlock()
			if eventNodeID == 0 {
				continue
			}

			eventNode := NewEventNode(eventID, eventType, eventData)
			// Owner will be set when node is loaded from BadgerDB
			// For now, we store the owner's xrefID in the event data if needed

			data, err := SerializeNode(eventNode, graph)
			if err != nil {
				return fmt.Errorf("failed to serialize event node %s: %w", eventID, err)
			}

			key := fmt.Sprintf("node:%d", eventNodeID)
			if err := writeBatch.Set([]byte(key), data); err != nil {
				return fmt.Errorf("failed to set event node %s: %w", eventID, err)
			}
		}
	}

	// Process family events
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		fam, ok := record.(*types.FamilyRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		famNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if famNodeID == 0 {
			continue
		}

		eventTypes := []string{"MARR", "DIV", "ANUL", "ENGA", "MARB", "MARC", "MARL", "MARS"}
		for _, eventType := range eventTypes {
			eventLines := fam.GetLines(eventType)
			for i := range eventLines {
				eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
				graph.mu.RLock()
				eventNodeID := graph.xrefToID[eventID]
				graph.mu.RUnlock()
				if eventNodeID == 0 {
					continue
				}

				// Create event data from line
				eventLine := eventLines[i]
				eventData := map[string]interface{}{
					"type":        eventType,
					"date":        eventLine.GetValue("DATE"),
					"place":       eventLine.GetValue("PLAC"),
					"description": eventLine.Value,
				}

				eventNode := NewEventNode(eventID, eventType, eventData)
				// Owner will be set when node is loaded from BadgerDB

				data, err := SerializeNode(eventNode, graph)
				if err != nil {
					return fmt.Errorf("failed to serialize event node %s: %w", eventID, err)
				}

				key := fmt.Sprintf("node:%d", eventNodeID)
				if err := writeBatch.Set([]byte(key), data); err != nil {
					return fmt.Errorf("failed to set event node %s: %w", eventID, err)
				}
			}
		}
	}

	return nil
}

// buildEdgesInBadgerDB stores edges in BadgerDB
func buildEdgesInBadgerDB(db *badger.DB, tree *types.GedcomTree, graph *Graph) error {
	// Process families to extract edges
	families := tree.GetAllFamilies()
	writeBatch := db.NewWriteBatch()
	defer writeBatch.Cancel()

	// Track edges per node for efficient storage
	nodeEdges := make(map[uint32][]EdgeData)

	// Process family relationships
	if err := processFamilyEdges(families, graph, nodeEdges); err != nil {
		return err
	}

	// Process individual NOTE and SOUR edges
	if err := processIndividualReferenceEdges(tree, graph, nodeEdges); err != nil {
		return err
	}

	// Process family NOTE and SOUR edges
	if err := processFamilyReferenceEdges(families, graph, nodeEdges); err != nil {
		return err
	}

	// Process note SOUR edges
	if err := processNoteSourceEdges(tree, graph, nodeEdges); err != nil {
		return err
	}

	// Process source REPO edges
	if err := processSourceRepositoryEdges(tree, graph, nodeEdges); err != nil {
		return err
	}

	// Process event edges
	if err := processEventEdges(tree, graph, nodeEdges); err != nil {
		return err
	}

	// Store edges in BadgerDB (grouped by node for efficient retrieval)
	for nodeID, edges := range nodeEdges {
		// Serialize edges
		data, err := serializeEdgeDataList(edges)
		if err != nil {
			return fmt.Errorf("failed to serialize edges for node %d: %w", nodeID, err)
		}

		// Store outgoing edges
		key := fmt.Sprintf("edges:%d:out", nodeID)
		if err := writeBatch.Set([]byte(key), data); err != nil {
			return fmt.Errorf("failed to set edges for node %d: %w", nodeID, err)
		}
	}

	// Flush batch
	if err := writeBatch.Flush(); err != nil {
		return fmt.Errorf("failed to flush edges batch: %w", err)
	}

	return nil
}

// processFamilyEdges processes family relationship edges
func processFamilyEdges(families map[string]types.Record, graph *Graph, nodeEdges map[uint32][]EdgeData) error {
	for xrefID, record := range families {
		famRecord, ok := record.(*types.FamilyRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		famNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if famNodeID == 0 {
			continue
		}

		// Get husband
		husbandXref := famRecord.GetHusband()
		if husbandXref != "" {
			graph.mu.RLock()
			husbandID := graph.xrefToID[husbandXref]
			graph.mu.RUnlock()
			if husbandID != 0 {
				// HUSB edge: Family -> Individual
				edgeData := EdgeData{
					FromID:     famNodeID,
					ToID:       husbandID,
					EdgeType:   EdgeTypeHUSB,
					FamilyID:   famNodeID,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)

				// FAMS edge: Individual -> Family (reverse)
				edgeData2 := EdgeData{
					FromID:     husbandID,
					ToID:       famNodeID,
					EdgeType:   EdgeTypeFAMS,
					FamilyID:   famNodeID,
					Direction:  DirectionBackward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[husbandID] = append(nodeEdges[husbandID], edgeData2)
			}
		}

		// Get wife
		wifeXref := famRecord.GetWife()
		if wifeXref != "" {
			graph.mu.RLock()
			wifeID := graph.xrefToID[wifeXref]
			graph.mu.RUnlock()
			if wifeID != 0 {
				// WIFE edge: Family -> Individual
				edgeData := EdgeData{
					FromID:     famNodeID,
					ToID:       wifeID,
					EdgeType:   EdgeTypeWIFE,
					FamilyID:   famNodeID,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)

				// FAMS edge: Individual -> Family (reverse)
				edgeData2 := EdgeData{
					FromID:     wifeID,
					ToID:       famNodeID,
					EdgeType:   EdgeTypeFAMS,
					FamilyID:   famNodeID,
					Direction:  DirectionBackward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[wifeID] = append(nodeEdges[wifeID], edgeData2)
			}
		}

		// Get children
		children := famRecord.GetChildren()
		for _, childXref := range children {
			graph.mu.RLock()
			childID := graph.xrefToID[childXref]
			graph.mu.RUnlock()
			if childID != 0 {
				// CHIL edge: Family -> Individual
				edgeData := EdgeData{
					FromID:     famNodeID,
					ToID:       childID,
					EdgeType:   EdgeTypeCHIL,
					FamilyID:   famNodeID,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)

				// FAMC edge: Individual -> Family (reverse)
				edgeData2 := EdgeData{
					FromID:     childID,
					ToID:       famNodeID,
					EdgeType:   EdgeTypeFAMC,
					FamilyID:   famNodeID,
					Direction:  DirectionBackward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[childID] = append(nodeEdges[childID], edgeData2)
			}
		}
	}

	return nil
}

// processIndividualReferenceEdges processes NOTE and SOUR edges for individuals
func processIndividualReferenceEdges(tree *types.GedcomTree, graph *Graph, nodeEdges map[uint32][]EdgeData) error {
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*types.IndividualRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		indiNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if indiNodeID == 0 {
			continue
		}

		// NOTE references
		notes := indi.GetNotes()
		for _, noteXref := range notes {
			graph.mu.RLock()
			noteNodeID := graph.xrefToID[noteXref]
			graph.mu.RUnlock()
			if noteNodeID != 0 {
				edgeData := EdgeData{
					FromID:     indiNodeID,
					ToID:       noteNodeID,
					EdgeType:   EdgeTypeNOTE,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[indiNodeID] = append(nodeEdges[indiNodeID], edgeData)
			}
		}

		// SOUR references
		sources := indi.GetSources()
		for _, sourceXref := range sources {
			graph.mu.RLock()
			sourceNodeID := graph.xrefToID[sourceXref]
			graph.mu.RUnlock()
			if sourceNodeID != 0 {
				edgeData := EdgeData{
					FromID:     indiNodeID,
					ToID:       sourceNodeID,
					EdgeType:   EdgeTypeSOUR,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[indiNodeID] = append(nodeEdges[indiNodeID], edgeData)
			}
		}
	}

	return nil
}

// processFamilyReferenceEdges processes NOTE and SOUR edges for families
func processFamilyReferenceEdges(families map[string]types.Record, graph *Graph, nodeEdges map[uint32][]EdgeData) error {
	for xrefID, record := range families {
		fam, ok := record.(*types.FamilyRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		famNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if famNodeID == 0 {
			continue
		}

		// NOTE references
		notes := fam.GetNotes()
		for _, noteXref := range notes {
			graph.mu.RLock()
			noteNodeID := graph.xrefToID[noteXref]
			graph.mu.RUnlock()
			if noteNodeID != 0 {
				edgeData := EdgeData{
					FromID:     famNodeID,
					ToID:       noteNodeID,
					EdgeType:   EdgeTypeNOTE,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)
			}
		}

		// SOUR references
		sources := fam.GetSources()
		for _, sourceXref := range sources {
			graph.mu.RLock()
			sourceNodeID := graph.xrefToID[sourceXref]
			graph.mu.RUnlock()
			if sourceNodeID != 0 {
				edgeData := EdgeData{
					FromID:     famNodeID,
					ToID:       sourceNodeID,
					EdgeType:   EdgeTypeSOUR,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)
			}
		}
	}

	return nil
}

// processNoteSourceEdges processes SOUR edges for notes
func processNoteSourceEdges(tree *types.GedcomTree, graph *Graph, nodeEdges map[uint32][]EdgeData) error {
	notes := tree.GetAllNotes()
	for xrefID, record := range notes {
		note, ok := record.(*types.NoteRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		noteNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if noteNodeID == 0 {
			continue
		}

		sources := note.GetValues("SOUR")
		for _, sourceXref := range sources {
			graph.mu.RLock()
			sourceNodeID := graph.xrefToID[sourceXref]
			graph.mu.RUnlock()
			if sourceNodeID != 0 {
				edgeData := EdgeData{
					FromID:     noteNodeID,
					ToID:       sourceNodeID,
					EdgeType:   EdgeTypeSOUR,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[noteNodeID] = append(nodeEdges[noteNodeID], edgeData)
			}
		}
	}

	return nil
}

// processSourceRepositoryEdges processes REPO edges for sources
func processSourceRepositoryEdges(tree *types.GedcomTree, graph *Graph, nodeEdges map[uint32][]EdgeData) error {
	sources := tree.GetAllSources()
	for xrefID, record := range sources {
		source, ok := record.(*types.SourceRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		sourceNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if sourceNodeID == 0 {
			continue
		}

		repos := source.GetValues("REPO")
		for _, repoXref := range repos {
			graph.mu.RLock()
			repoNodeID := graph.xrefToID[repoXref]
			graph.mu.RUnlock()
			if repoNodeID != 0 {
				edgeData := EdgeData{
					FromID:     sourceNodeID,
					ToID:       repoNodeID,
					EdgeType:   EdgeTypeREPO,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[sourceNodeID] = append(nodeEdges[sourceNodeID], edgeData)
			}
		}
	}

	return nil
}

// processEventEdges processes has_event edges for individuals and families
func processEventEdges(tree *types.GedcomTree, graph *Graph, nodeEdges map[uint32][]EdgeData) error {
	// Process individuals for has_event edges
	individuals := tree.GetAllIndividuals()
	for xrefID, record := range individuals {
		indi, ok := record.(*types.IndividualRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		indiNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if indiNodeID == 0 {
			continue
		}

		events := indi.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
			graph.mu.RLock()
			eventNodeID := graph.xrefToID[eventID]
			graph.mu.RUnlock()
			if eventNodeID != 0 {
				edgeData := EdgeData{
					FromID:     indiNodeID,
					ToID:       eventNodeID,
					EdgeType:   EdgeTypeHasEvent,
					Direction:  DirectionForward,
					Properties: make(map[string]interface{}),
				}
				nodeEdges[indiNodeID] = append(nodeEdges[indiNodeID], edgeData)
			}
		}
	}

	// Process families for has_event edges (MARR, DIV, etc.)
	families := tree.GetAllFamilies()
	for xrefID, record := range families {
		fam, ok := record.(*types.FamilyRecord)
		if !ok {
			continue
		}

		graph.mu.RLock()
		famNodeID := graph.xrefToID[xrefID]
		graph.mu.RUnlock()
		if famNodeID == 0 {
			continue
		}

		// Check for MARR, DIV events
		eventTypes := []string{"MARR", "DIV", "ANUL", "ENGA", "MARB", "MARC", "MARL", "MARS"}
		for _, eventType := range eventTypes {
			eventLines := fam.GetLines(eventType)
			for i := range eventLines {
				eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)
				graph.mu.RLock()
				eventNodeID := graph.xrefToID[eventID]
				graph.mu.RUnlock()
				if eventNodeID != 0 {
					edgeData := EdgeData{
						FromID:     famNodeID,
						ToID:       eventNodeID,
						EdgeType:   EdgeTypeHasEvent,
						Direction:  DirectionForward,
						Properties: make(map[string]interface{}),
					}
					nodeEdges[famNodeID] = append(nodeEdges[famNodeID], edgeData)
				}
			}
		}
	}

	return nil
}

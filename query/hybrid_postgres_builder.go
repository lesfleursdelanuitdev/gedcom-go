package query

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lesfleursdelanuitdev/ligneous-gedcom/types"
)

// buildGraphInPostgreSQL builds indexes in PostgreSQL
func buildGraphInPostgreSQL(storage *HybridStoragePostgres, tree *types.GedcomTree, graph *Graph) error {
	db := storage.PostgreSQL()
	fileID := storage.FileID()

	// Start transaction for batch inserts
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statements for batch inserts (with file_id)
	stmtNode, err := tx.Prepare(`
		INSERT INTO nodes (file_id, id, xref, type, name, name_lower, birth_date, birth_place, sex, 
		                   has_children, has_spouse, living, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare node statement: %w", err)
	}
	defer stmtNode.Close()

	stmtXref, err := tx.Prepare(`
		INSERT INTO xref_mapping (file_id, xref, node_id) VALUES ($1, $2, $3)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare xref statement: %w", err)
	}
	defer stmtXref.Close()

	now := time.Now().Unix()

	// Process all record types
	if err := processIndividualsForPostgreSQL(tree, graph, stmtNode, stmtXref, fileID, now); err != nil {
		return err
	}

	if err := processFamiliesForPostgreSQL(tree, graph, stmtNode, stmtXref, fileID, now); err != nil {
		return err
	}

	if err := processNotesForPostgreSQL(tree, graph, stmtNode, stmtXref, fileID, now); err != nil {
		return err
	}

	if err := processSourcesForPostgreSQL(tree, graph, stmtNode, stmtXref, fileID, now); err != nil {
		return err
	}

	if err := processRepositoriesForPostgreSQL(tree, graph, stmtNode, stmtXref, fileID, now); err != nil {
		return err
	}

	if err := processEventsForPostgreSQL(tree, graph, stmtNode, stmtXref, fileID, now); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// processIndividualsForPostgreSQL processes individual records for PostgreSQL
func processIndividualsForPostgreSQL(tree *types.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, fileID string, now int64) error {
	individuals := tree.GetAllIndividuals()

	for xrefID, record := range individuals {
		indiRecord, ok := record.(*types.IndividualRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

		// Extract indexed fields
		name := indiRecord.GetName()
		nameLower := toLower(name)
		birthDate := parseBirthDate(indiRecord)
		birthPlace := indiRecord.GetBirthPlace()
		sex := indiRecord.GetSex()

		// Determine boolean flags (will be updated after edges are processed)
		hasChildren := false // Will be updated later
		hasSpouse := false   // Will be updated later
		living := indiRecord.GetDeathDate() == ""

		// Insert into nodes table (with file_id)
		_, err := stmtNode.Exec(
			fileID, nodeID, xrefID, "individual", name, nameLower,
			birthDate, birthPlace, sex,
			boolToInt(hasChildren), boolToInt(hasSpouse), boolToInt(living),
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping (with file_id)
		_, err = stmtXref.Exec(fileID, xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert xref mapping %s: %w", xrefID, err)
		}
	}

	return nil
}

// processFamiliesForPostgreSQL processes family records for PostgreSQL
func processFamiliesForPostgreSQL(tree *types.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, fileID string, now int64) error {
	families := tree.GetAllFamilies()

	for xrefID, record := range families {
		_, ok := record.(*types.FamilyRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

		// Families don't have as many indexed fields
		_, err := stmtNode.Exec(
			fileID, nodeID, xrefID, "family", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert family node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping (with file_id)
		_, err = stmtXref.Exec(fileID, xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert family xref mapping %s: %w", xrefID, err)
		}
	}

	return nil
}

// processNotesForPostgreSQL processes note records for PostgreSQL
func processNotesForPostgreSQL(tree *types.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, fileID string, now int64) error {
	notes := tree.GetAllNotes()

	for xrefID, record := range notes {
		_, ok := record.(*types.NoteRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

		// Notes don't have indexed fields (for now)
		_, err := stmtNode.Exec(
			fileID, nodeID, xrefID, "note", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert note node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping (with file_id)
		_, err = stmtXref.Exec(fileID, xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert note xref mapping %s: %w", xrefID, err)
		}
	}

	return nil
}

// processSourcesForPostgreSQL processes source records for PostgreSQL
func processSourcesForPostgreSQL(tree *types.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, fileID string, now int64) error {
	sources := tree.GetAllSources()

	for xrefID, record := range sources {
		_, ok := record.(*types.SourceRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

		// Sources don't have indexed fields (for now)
		_, err := stmtNode.Exec(
			fileID, nodeID, xrefID, "source", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert source node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping (with file_id)
		_, err = stmtXref.Exec(fileID, xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert source xref mapping %s: %w", xrefID, err)
		}
	}

	return nil
}

// processRepositoriesForPostgreSQL processes repository records for PostgreSQL
func processRepositoriesForPostgreSQL(tree *types.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, fileID string, now int64) error {
	repositories := tree.GetAllRepositories()

	for xrefID, record := range repositories {
		_, ok := record.(*types.RepositoryRecord)
		if !ok {
			continue
		}

		// Get or create node ID (with locking)
		graph.mu.Lock()
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			nodeID = graph.nextID
			graph.nextID++
			graph.xrefToID[xrefID] = nodeID
			graph.idToXref[nodeID] = xrefID
		}
		graph.mu.Unlock()

		// Repositories don't have indexed fields (for now)
		_, err := stmtNode.Exec(
			fileID, nodeID, xrefID, "repository", "", "",
			nil, "", "",
			0, 0, 0,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert repository node %s: %w", xrefID, err)
		}

		// Insert into xref_mapping (with file_id)
		_, err = stmtXref.Exec(fileID, xrefID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to insert repository xref mapping %s: %w", xrefID, err)
		}
	}

	return nil
}

// processEventsForPostgreSQL processes event nodes for PostgreSQL
func processEventsForPostgreSQL(tree *types.GedcomTree, graph *Graph, stmtNode, stmtXref *sql.Stmt, fileID string, now int64) error {
	// Process individual events
	individualsForEvents := tree.GetAllIndividuals()
	for xrefID, record := range individualsForEvents {
		indi, ok := record.(*types.IndividualRecord)
		if !ok {
			continue
		}

		events := indi.GetEvents()
		for i, eventData := range events {
			eventType, ok := eventData["type"].(string)
			if !ok {
				continue
			}

			eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)

			// Get or create node ID (with locking)
			graph.mu.Lock()
			nodeID := graph.xrefToID[eventID]
			if nodeID == 0 {
				nodeID = graph.nextID
				graph.nextID++
				graph.xrefToID[eventID] = nodeID
				graph.idToXref[nodeID] = eventID
			}
			graph.mu.Unlock()

			// Events don't have indexed fields (for now)
			_, err := stmtNode.Exec(
				fileID, nodeID, eventID, "event", "", "",
				nil, "", "",
				0, 0, 0,
				now, now,
			)
			if err != nil {
				return fmt.Errorf("failed to insert event node %s: %w", eventID, err)
			}

			// Insert into xref_mapping (with file_id)
			_, err = stmtXref.Exec(fileID, eventID, nodeID)
			if err != nil {
				return fmt.Errorf("failed to insert event xref mapping %s: %w", eventID, err)
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

		// Check for MARR, DIV events
		eventTypes := []string{"MARR", "DIV", "ANUL", "ENGA", "MARB", "MARC", "MARL", "MARS"}
		for _, eventType := range eventTypes {
			eventLines := fam.GetLines(eventType)
			for i := range eventLines {
				eventID := fmt.Sprintf("%s_%s_%d", xrefID, eventType, i)

				// Get or create node ID
				graph.mu.Lock()
				nodeID := graph.xrefToID[eventID]
				if nodeID == 0 {
					nodeID = graph.nextID
					graph.nextID++
					graph.xrefToID[eventID] = nodeID
					graph.idToXref[nodeID] = eventID
				}
				graph.mu.Unlock()

				// Events don't have indexed fields (for now)
				_, err := stmtNode.Exec(
					fileID, nodeID, eventID, "event", "", "",
					nil, "", "",
					0, 0, 0,
					now, now,
				)
				if err != nil {
					return fmt.Errorf("failed to insert event node %s: %w", eventID, err)
				}

				// Insert into xref_mapping (with file_id)
				_, err = stmtXref.Exec(fileID, eventID, nodeID)
				if err != nil {
					return fmt.Errorf("failed to insert event xref mapping %s: %w", eventID, err)
				}
			}
		}
	}

	return nil
}

// updateRelationshipFlagsPostgreSQL updates has_children and has_spouse flags in PostgreSQL
func updateRelationshipFlagsPostgreSQL(storage *HybridStoragePostgres, tree *types.GedcomTree, graph *Graph) error {
	db := storage.PostgreSQL()
	fileID := storage.FileID()

	// Process families to determine relationships
	families := tree.GetAllFamilies()

	// Track which individuals have children/spouses
	hasChildren := make(map[uint32]bool)
	hasSpouse := make(map[uint32]bool)

	for _, record := range families {
		famRecord, ok := record.(*types.FamilyRecord)
		if !ok {
			continue
		}

		// Get family node ID
		famXref := famRecord.XrefID()
		famNodeID := graph.xrefToID[famXref]
		if famNodeID == 0 {
			continue
		}

		// Check husband
		husbandXref := famRecord.GetHusband()
		if husbandXref != "" {
			husbandID := graph.xrefToID[husbandXref]
			if husbandID != 0 {
				hasSpouse[husbandID] = true
			}
		}

		// Check wife
		wifeXref := famRecord.GetWife()
		if wifeXref != "" {
			wifeID := graph.xrefToID[wifeXref]
			if wifeID != 0 {
				hasSpouse[wifeID] = true
			}
		}

		// Check children
		children := famRecord.GetChildren()
		for _, childXref := range children {
			childID := graph.xrefToID[childXref]
			if childID != 0 {
				// Child has parents (but we're tracking if parents have children)
				// So we need to mark the parents
				if husbandID := graph.xrefToID[husbandXref]; husbandID != 0 {
					hasChildren[husbandID] = true
				}
				if wifeID := graph.xrefToID[wifeXref]; wifeID != 0 {
					hasChildren[wifeID] = true
				}
			}
		}
	}

	// Update PostgreSQL with relationship flags
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE nodes SET has_children = $1, has_spouse = $2 WHERE file_id = $3 AND id = $4")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	// Update all individuals
	individuals := tree.GetAllIndividuals()
	for xrefID := range individuals {
		nodeID := graph.xrefToID[xrefID]
		if nodeID == 0 {
			continue
		}

		hasChildrenVal := boolToInt(hasChildren[nodeID])
		hasSpouseVal := boolToInt(hasSpouse[nodeID])

		_, err := stmt.Exec(hasChildrenVal, hasSpouseVal, fileID, nodeID)
		if err != nil {
			return fmt.Errorf("failed to update node %d: %w", nodeID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}


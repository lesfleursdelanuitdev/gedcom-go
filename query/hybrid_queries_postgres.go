package query

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

// HybridQueryHelpersPostgres provides PostgreSQL query methods for hybrid storage
// All queries include file_id for multi-file support
type HybridQueryHelpersPostgres struct {
	db     *sql.DB
	fileID string

	// Prepared statements for performance (all include file_id)
	stmtFindByXref          *sql.Stmt
	stmtFindXrefByID        *sql.Stmt
	stmtFindByName          *sql.Stmt
	stmtFindByNameExact     *sql.Stmt
	stmtFindByNameStarts    *sql.Stmt
	stmtFindByBirthDate     *sql.Stmt
	stmtFindByBirthPlace    *sql.Stmt
	stmtFindBySex           *sql.Stmt
	stmtHasChildren         *sql.Stmt
	stmtHasSpouse           *sql.Stmt
	stmtIsLiving            *sql.Stmt
	stmtGetAllIndividualIDs  *sql.Stmt

	mu sync.Mutex
}

// NewHybridQueryHelpersPostgres creates a new PostgreSQL helper instance and prepares statements
func NewHybridQueryHelpersPostgres(db *sql.DB, fileID string) (*HybridQueryHelpersPostgres, error) {
	helpers := &HybridQueryHelpersPostgres{
		db:     db,
		fileID: fileID,
	}

	// Prepare statements for better performance (all include file_id)
	var err error
	helpers.stmtFindByXref, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND xref = $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindByXref: %w", err)
	}

	helpers.stmtFindXrefByID, err = db.Prepare("SELECT xref FROM nodes WHERE file_id = $1 AND id = $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindXrefByID: %w", err)
	}

	helpers.stmtFindByName, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND type = 'individual' AND name_lower LIKE $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindByName: %w", err)
	}

	helpers.stmtFindByNameExact, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND type = 'individual' AND name_lower = $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindByNameExact: %w", err)
	}

	helpers.stmtFindByNameStarts, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND type = 'individual' AND name_lower LIKE $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindByNameStarts: %w", err)
	}

	helpers.stmtFindByBirthDate, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND type = 'individual' AND birth_date IS NOT NULL AND birth_date >= $2 AND birth_date <= $3")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindByBirthDate: %w", err)
	}

	helpers.stmtFindByBirthPlace, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND type = 'individual' AND birth_place IS NOT NULL AND LOWER(birth_place) LIKE $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindByBirthPlace: %w", err)
	}

	helpers.stmtFindBySex, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND type = 'individual' AND sex = $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare FindBySex: %w", err)
	}

	helpers.stmtHasChildren, err = db.Prepare("SELECT has_children FROM nodes WHERE file_id = $1 AND id = $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare HasChildren: %w", err)
	}

	helpers.stmtHasSpouse, err = db.Prepare("SELECT has_spouse FROM nodes WHERE file_id = $1 AND id = $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare HasSpouse: %w", err)
	}

	helpers.stmtIsLiving, err = db.Prepare("SELECT living FROM nodes WHERE file_id = $1 AND id = $2")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare IsLiving: %w", err)
	}

	helpers.stmtGetAllIndividualIDs, err = db.Prepare("SELECT id FROM nodes WHERE file_id = $1 AND type = 'individual'")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare GetAllIndividualIDs: %w", err)
	}

	return helpers, nil
}

// Close closes all prepared statements
func (h *HybridQueryHelpersPostgres) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var errs []error
	if h.stmtFindByXref != nil {
		if err := h.stmtFindByXref.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtFindXrefByID != nil {
		if err := h.stmtFindXrefByID.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtFindByName != nil {
		if err := h.stmtFindByName.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtFindByNameExact != nil {
		if err := h.stmtFindByNameExact.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtFindByNameStarts != nil {
		if err := h.stmtFindByNameStarts.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtFindByBirthDate != nil {
		if err := h.stmtFindByBirthDate.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtFindByBirthPlace != nil {
		if err := h.stmtFindByBirthPlace.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtFindBySex != nil {
		if err := h.stmtFindBySex.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtHasChildren != nil {
		if err := h.stmtHasChildren.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtHasSpouse != nil {
		if err := h.stmtHasSpouse.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtIsLiving != nil {
		if err := h.stmtIsLiving.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if h.stmtGetAllIndividualIDs != nil {
		if err := h.stmtGetAllIndividualIDs.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing statements: %v", errs)
	}
	return nil
}

// FindByXref returns node ID for a given XREF
func (h *HybridQueryHelpersPostgres) FindByXref(xref string) (uint32, error) {
	var nodeID uint32
	err := h.stmtFindByXref.QueryRow(h.fileID, xref).Scan(&nodeID)
	if err == sql.ErrNoRows {
		return 0, nil // Not found
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query by xref: %w", err)
	}
	return nodeID, nil
}

// FindXrefByID returns XREF for a given node ID
func (h *HybridQueryHelpersPostgres) FindXrefByID(nodeID uint32) (string, error) {
	var xref string
	err := h.stmtFindXrefByID.QueryRow(h.fileID, nodeID).Scan(&xref)
	if err == sql.ErrNoRows {
		return "", nil // Not found
	}
	if err != nil {
		return "", fmt.Errorf("failed to query by id: %w", err)
	}
	return xref, nil
}

// FindByName finds node IDs by name (substring match, case-insensitive)
func (h *HybridQueryHelpersPostgres) FindByName(pattern string) ([]uint32, error) {
	patternLower := strings.ToLower(pattern)
	rows, err := h.stmtFindByName.Query(h.fileID, "%"+patternLower+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query by name: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uint32
	for rows.Next() {
		var nodeID uint32
		if err := rows.Scan(&nodeID); err != nil {
			return nil, fmt.Errorf("failed to scan node ID: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, rows.Err()
}

// FindByNameExact finds node IDs by exact name match (case-insensitive)
func (h *HybridQueryHelpersPostgres) FindByNameExact(name string) ([]uint32, error) {
	nameLower := strings.ToLower(name)
	rows, err := h.stmtFindByNameExact.Query(h.fileID, nameLower)
	if err != nil {
		return nil, fmt.Errorf("failed to query by exact name: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uint32
	for rows.Next() {
		var nodeID uint32
		if err := rows.Scan(&nodeID); err != nil {
			return nil, fmt.Errorf("failed to scan node ID: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, rows.Err()
}

// FindByNameStarts finds node IDs by name prefix (case-insensitive)
func (h *HybridQueryHelpersPostgres) FindByNameStarts(prefix string) ([]uint32, error) {
	prefixLower := strings.ToLower(prefix)
	rows, err := h.stmtFindByNameStarts.Query(h.fileID, prefixLower+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query by name prefix: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uint32
	for rows.Next() {
		var nodeID uint32
		if err := rows.Scan(&nodeID); err != nil {
			return nil, fmt.Errorf("failed to scan node ID: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, rows.Err()
}

// FindByBirthDate finds node IDs by birth date range
func (h *HybridQueryHelpersPostgres) FindByBirthDate(start, end time.Time) ([]uint32, error) {
	startUnix := start.Unix()
	endUnix := end.Unix()
	rows, err := h.stmtFindByBirthDate.Query(h.fileID, startUnix, endUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to query by birth date: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uint32
	for rows.Next() {
		var nodeID uint32
		if err := rows.Scan(&nodeID); err != nil {
			return nil, fmt.Errorf("failed to scan node ID: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, rows.Err()
}

// FindByBirthPlace finds node IDs by birth place (substring match, case-insensitive)
func (h *HybridQueryHelpersPostgres) FindByBirthPlace(place string) ([]uint32, error) {
	placeLower := strings.ToLower(place)
	rows, err := h.stmtFindByBirthPlace.Query(h.fileID, "%"+placeLower+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query by birth place: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uint32
	for rows.Next() {
		var nodeID uint32
		if err := rows.Scan(&nodeID); err != nil {
			return nil, fmt.Errorf("failed to scan node ID: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, rows.Err()
}

// FindBySex finds node IDs by sex
func (h *HybridQueryHelpersPostgres) FindBySex(sex string) ([]uint32, error) {
	sexUpper := strings.ToUpper(sex)
	rows, err := h.stmtFindBySex.Query(h.fileID, sexUpper)
	if err != nil {
		return nil, fmt.Errorf("failed to query by sex: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uint32
	for rows.Next() {
		var nodeID uint32
		if err := rows.Scan(&nodeID); err != nil {
			return nil, fmt.Errorf("failed to scan node ID: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, rows.Err()
}

// HasChildren checks if a node has children
func (h *HybridQueryHelpersPostgres) HasChildren(nodeID uint32) (bool, error) {
	var hasChildren int
	err := h.stmtHasChildren.QueryRow(h.fileID, nodeID).Scan(&hasChildren)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query has_children: %w", err)
	}
	return hasChildren == 1, nil
}

// HasSpouse checks if a node has a spouse
func (h *HybridQueryHelpersPostgres) HasSpouse(nodeID uint32) (bool, error) {
	var hasSpouse int
	err := h.stmtHasSpouse.QueryRow(h.fileID, nodeID).Scan(&hasSpouse)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query has_spouse: %w", err)
	}
	return hasSpouse == 1, nil
}

// IsLiving checks if a node is living
func (h *HybridQueryHelpersPostgres) IsLiving(nodeID uint32) (bool, error) {
	var living int
	err := h.stmtIsLiving.QueryRow(h.fileID, nodeID).Scan(&living)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query living: %w", err)
	}
	return living == 1, nil
}

// GetAllIndividualIDs returns all individual node IDs
func (h *HybridQueryHelpersPostgres) GetAllIndividualIDs() ([]uint32, error) {
	rows, err := h.stmtGetAllIndividualIDs.Query(h.fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query all individuals: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uint32
	for rows.Next() {
		var nodeID uint32
		if err := rows.Scan(&nodeID); err != nil {
			return nil, fmt.Errorf("failed to scan node ID: %w", err)
		}
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs, rows.Err()
}


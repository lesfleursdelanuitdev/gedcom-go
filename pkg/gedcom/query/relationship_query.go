package query

import (
	"time"
)

// RelationshipQuery represents a query for the relationship between two individuals.
type RelationshipQuery struct {
	fromXrefID string
	toXrefID   string
	graph      *Graph
}

// Execute calculates and returns the relationship result.
func (rq *RelationshipQuery) Execute() (*RelationshipResult, error) {
	// Record metrics if available
	start := time.Now()
	defer func() {
		if rq.graph.metrics != nil {
			duration := time.Since(start)
			rq.graph.metrics.RecordQuery(duration)
		}
	}()

	return rq.graph.CalculateRelationship(rq.fromXrefID, rq.toXrefID)
}

// GetRelationshipType returns the human-readable relationship type.
func (rr *RelationshipResult) GetRelationshipType() string {
	return rr.RelationshipType
}

// IsBloodRelation checks if this is a blood relation.
func (rr *RelationshipResult) IsBloodRelation() bool {
	return rr.Path != nil && (rr.Path.Type == PathTypeBlood || rr.Path.Type == PathTypeMixed)
}

// IsMaritalRelation checks if this is a marital relation.
func (rr *RelationshipResult) IsMaritalRelation() bool {
	return rr.Path != nil && (rr.Path.Type == PathTypeMarital || rr.Path.Type == PathTypeMixed)
}

package query

import (
	"fmt"
	"time"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// EventInfo represents information about an event.
type EventInfo struct {
	EventID    string
	EventType  string
	Date       string
	Place      string
	Description string
	Owner      GraphNode // Individual or Family that owns this event
}

// GetEventsForIndividual returns all events associated with an individual.
func (iq *IndividualQuery) GetEvents() ([]EventInfo, error) {
	node := iq.graph.GetIndividual(iq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("individual %s not found", iq.xrefID)
	}

	events := make([]EventInfo, 0)

	// Get events via has_event edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeHasEvent {
			if eventNode, ok := edge.To.(*EventNode); ok {
				eventInfo := EventInfo{
					EventID:     eventNode.EventID,
					EventType:   eventNode.EventType,
					Owner:       node,
				}
				
				// Extract date, place, description from event data
				if eventNode.EventData != nil {
					if date, ok := eventNode.EventData["date"].(string); ok {
						eventInfo.Date = date
					}
					if place, ok := eventNode.EventData["place"].(string); ok {
						eventInfo.Place = place
					}
					if desc, ok := eventNode.EventData["description"].(string); ok {
						eventInfo.Description = desc
					}
				}
				
				events = append(events, eventInfo)
			}
		}
	}

	// Also get events directly from the record (for backward compatibility)
	if node.Individual != nil {
		recordEvents := node.Individual.GetEvents()
		for _, event := range recordEvents {
			// Check if we already have this event from edges
			found := false
			for _, existingEvent := range events {
				if existingEvent.EventType == event["type"].(string) &&
					existingEvent.Date == event["date"].(string) {
					found = true
					break
				}
			}
			
			if !found {
				events = append(events, EventInfo{
					EventID:      fmt.Sprintf("%s_%s_%d", iq.xrefID, event["type"], len(events)),
					EventType:    event["type"].(string),
					Date:         event["date"].(string),
					Place:        event["place"].(string),
					Description:  event["description"].(string),
					Owner:        node,
				})
			}
		}
	}

	return events, nil
}

// GetEventsForFamily returns all events associated with a family.
func (fq *FamilyQuery) GetEvents() ([]EventInfo, error) {
	node := fq.graph.GetFamily(fq.xrefID)
	if node == nil {
		return nil, fmt.Errorf("family %s not found", fq.xrefID)
	}

	events := make([]EventInfo, 0)

	// Get events via has_event edges
	for _, edge := range node.OutEdges() {
		if edge.EdgeType == EdgeTypeHasEvent {
			if eventNode, ok := edge.To.(*EventNode); ok {
				eventInfo := EventInfo{
					EventID:     eventNode.EventID,
					EventType:   eventNode.EventType,
					Owner:       node,
				}
				
				// Extract date, place, description from event data
				if eventNode.EventData != nil {
					if date, ok := eventNode.EventData["date"].(string); ok {
						eventInfo.Date = date
					}
					if place, ok := eventNode.EventData["place"].(string); ok {
						eventInfo.Place = place
					}
					if desc, ok := eventNode.EventData["description"].(string); ok {
						eventInfo.Description = desc
					}
				}
				
				events = append(events, eventInfo)
			}
		}
	}

	// Also get events directly from the record (for backward compatibility)
	if node.Family != nil {
		recordEvents := node.Family.GetEvents()
		for _, event := range recordEvents {
			// Check if we already have this event from edges
			found := false
			for _, existingEvent := range events {
				if existingEvent.EventType == event["type"].(string) &&
					existingEvent.Date == event["date"].(string) {
					found = true
					break
				}
			}
			
			if !found {
				events = append(events, EventInfo{
					EventID:      fmt.Sprintf("%s_%s_%d", fq.xrefID, event["type"], len(events)),
					EventType:    event["type"].(string),
					Date:         event["date"].(string),
					Place:        event["place"].(string),
					Description:  event["description"].(string),
					Owner:        node,
				})
			}
		}
	}

	return events, nil
}

// GetRecordsForEvent returns all records (individuals, families) that have this event.
func (g *Graph) GetRecordsForEvent(eventID string) ([]GraphNode, error) {
	eventNode := g.GetEvent(eventID)
	if eventNode == nil {
		return nil, fmt.Errorf("event %s not found", eventID)
	}

	records := make([]GraphNode, 0)
	seen := make(map[string]bool)

	// Traverse in-edges to find all records that have this event
	for _, edge := range eventNode.InEdges() {
		if edge.From != nil {
			fromID := edge.From.ID()
			if !seen[fromID] {
				seen[fromID] = true
				records = append(records, edge.From)
			}
		}
	}

	return records, nil
}

// GetEventsOnDate returns all events that occur on a specific date.
// The date can be specified as year, month, day, or a combination.
func (g *Graph) GetEventsOnDate(year int, month int, day int) ([]EventInfo, error) {
	allEvents := make([]EventInfo, 0)
	
	// Get all individuals and families
	allIndividuals := g.GetAllIndividuals()
	allFamilies := g.GetAllFamilies()
	
	// Check individual events
	for _, indiNode := range allIndividuals {
		if indiNode.Individual != nil {
			events := indiNode.Individual.GetEvents()
			for _, event := range events {
				dateStr := event["date"].(string)
				if matchesDate(dateStr, year, month, day) {
					allEvents = append(allEvents, EventInfo{
						EventID:      fmt.Sprintf("%s_%s_%d", indiNode.ID(), event["type"], len(allEvents)),
						EventType:    event["type"].(string),
						Date:         dateStr,
						Place:        event["place"].(string),
						Description:  event["description"].(string),
						Owner:        indiNode,
					})
				}
			}
		}
	}
	
	// Check family events
	for _, famNode := range allFamilies {
		if famNode.Family != nil {
			events := famNode.Family.GetEvents()
			for _, event := range events {
				dateStr := event["date"].(string)
				if matchesDate(dateStr, year, month, day) {
					allEvents = append(allEvents, EventInfo{
						EventID:      fmt.Sprintf("%s_%s_%d", famNode.ID(), event["type"], len(allEvents)),
						EventType:    event["type"].(string),
						Date:         dateStr,
						Place:        event["place"].(string),
						Description:  event["description"].(string),
						Owner:        famNode,
					})
				}
			}
		}
	}
	
	return allEvents, nil
}

// GetEventsOnDateByType returns all events of a specific type that occur on a specific date.
func (g *Graph) GetEventsOnDateByType(eventType string, year int, month int, day int) ([]EventInfo, error) {
	allEvents, err := g.GetEventsOnDate(year, month, day)
	if err != nil {
		return nil, err
	}
	
	// Filter by event type
	filtered := make([]EventInfo, 0)
	for _, event := range allEvents {
		if event.EventType == eventType {
			filtered = append(filtered, event)
		}
	}
	
	return filtered, nil
}

// matchesDate checks if a date string matches the specified year/month/day.
// If a component is 0, it's not checked (e.g., month=0 means match any month).
func matchesDate(dateStr string, year, month, day int) bool {
	if dateStr == "" {
		return false
	}
	
	parsedDate, err := gedcom.ParseDate(dateStr)
	if err != nil || parsedDate == nil {
		return false
	}
	
	// For exact dates, check components
	if parsedDate.Type == gedcom.DateTypeExact {
		if year > 0 && parsedDate.Year != year {
			return false
		}
		if month > 0 && parsedDate.Month != month {
			return false
		}
		if day > 0 && parsedDate.Day != day {
			return false
		}
		return true
	}
	
	// For range dates, check if the target date falls within the range
	if parsedDate.Type == gedcom.DateTypeBetween || parsedDate.Type == gedcom.DateTypeFromTo {
		// Get earliest and latest dates from the range
		startTime := time.Date(parsedDate.StartYear, time.Month(parsedDate.StartMonth), parsedDate.StartDay, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(parsedDate.EndYear, time.Month(parsedDate.EndMonth), parsedDate.EndDay, 23, 59, 59, 999999999, time.UTC)
		
		// Create target time (use 1 for missing components)
		targetYear := year
		if targetYear == 0 {
			targetYear = startTime.Year()
		}
		targetMonth := month
		if targetMonth == 0 {
			targetMonth = int(startTime.Month())
		}
		targetDay := day
		if targetDay == 0 {
			targetDay = 1
		}
		
		targetTime := time.Date(targetYear, time.Month(targetMonth), targetDay, 0, 0, 0, 0, time.UTC)
		
		return !targetTime.Before(startTime) && !targetTime.After(endTime)
	}
	
	// For ABOUT, BEFORE, AFTER - use earliest date
	if parsedDate.Year > 0 {
		if year > 0 && parsedDate.Year != year {
			return false
		}
		if month > 0 && parsedDate.Month != month {
			return false
		}
		if day > 0 && parsedDate.Day != day {
			return false
		}
		return true
	}
	
	return false
}


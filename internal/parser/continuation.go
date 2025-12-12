package parser

import (
	"fmt"
	"strings"
)

// TagInfo holds information about the last processed tag for CONC/CONT validation
type TagInfo struct {
	Tag   string
	Level int
}

// ContinuationHandler handles CONC and CONT continuation lines
type ContinuationHandler struct {
	currentValue strings.Builder
	lastTag      *TagInfo
}

// NewContinuationHandler creates a new continuation handler
func NewContinuationHandler() *ContinuationHandler {
	return &ContinuationHandler{
		currentValue: strings.Builder{},
		lastTag:      nil,
	}
}

// HandleContinuation processes a CONC or CONT line.
//
// CONC (Concatenate): Appends value directly to previous line (no space, no newline)
// CONT (Continue): Appends value with newline to previous line
//
// Validation rules:
// - CONC/CONT cannot be subordinate to another CONC/CONT at a lower level
// - CONC/CONT must follow a line that can have continuation (typically has a value)
//
// Returns error if validation fails, nil otherwise.
func (ch *ContinuationHandler) HandleContinuation(tag string, level int, value string) error {
	if tag != "CONC" && tag != "CONT" {
		return fmt.Errorf("HandleContinuation called with non-continuation tag: %s", tag)
	}

	// Validate: CONC/CONT cannot be subordinate to another CONC/CONT
	if ch.lastTag != nil {
		if (ch.lastTag.Tag == "CONC" || ch.lastTag.Tag == "CONT") && ch.lastTag.Level < level {
			return fmt.Errorf("CONC or CONT cannot be subordinate to CONC or CONT at line level %d (parent level %d)", level, ch.lastTag.Level)
		}
	}

	// Accumulate value based on tag type
	if tag == "CONC" {
		// Direct concatenation (no space, no newline)
		ch.currentValue.WriteString(value)
	} else if tag == "CONT" {
		// Add newline then concatenate
		ch.currentValue.WriteString("\n")
		ch.currentValue.WriteString(value)
	}

	// Update last tag info
	ch.lastTag = &TagInfo{
		Tag:   tag,
		Level: level,
	}

	return nil
}

// HasAccumulatedValue returns true if there is accumulated continuation value
func (ch *ContinuationHandler) HasAccumulatedValue() bool {
	return ch.currentValue.Len() > 0
}

// GetAccumulatedValue returns the accumulated continuation value and resets it
func (ch *ContinuationHandler) GetAccumulatedValue() string {
	value := ch.currentValue.String()
	ch.currentValue.Reset()
	return value
}

// Reset clears the accumulated value and last tag info
func (ch *ContinuationHandler) Reset() {
	ch.currentValue.Reset()
	ch.lastTag = nil
}

// SetLastTag sets the last processed tag (for non-CONC/CONT lines)
func (ch *ContinuationHandler) SetLastTag(tag string, level int) {
	ch.lastTag = &TagInfo{
		Tag:   tag,
		Level: level,
	}
}

// GetLastTag returns the last processed tag info
func (ch *ContinuationHandler) GetLastTag() *TagInfo {
	return ch.lastTag
}


package parser

import (
	"fmt"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

// LineStack is a stack of GedcomLine pointers used for hierarchical parsing.
// It maintains the current parent chain as we parse through the GEDCOM file.
type LineStack struct {
	lines []*gedcom.GedcomLine
}

// NewLineStack creates a new empty LineStack.
func NewLineStack() *LineStack {
	return &LineStack{
		lines: make([]*gedcom.GedcomLine, 0),
	}
}

// Push adds a line to the top of the stack.
func (ls *LineStack) Push(line *gedcom.GedcomLine) {
	ls.lines = append(ls.lines, line)
}

// Pop removes and returns the top line from the stack.
// Returns nil if the stack is empty.
func (ls *LineStack) Pop() *gedcom.GedcomLine {
	if ls.IsEmpty() {
		return nil
	}
	top := ls.lines[len(ls.lines)-1]
	ls.lines = ls.lines[:len(ls.lines)-1]
	return top
}

// Peek returns the top line without removing it.
// Returns nil if the stack is empty.
func (ls *LineStack) Peek() *gedcom.GedcomLine {
	if ls.IsEmpty() {
		return nil
	}
	return ls.lines[len(ls.lines)-1]
}

// IsEmpty returns true if the stack is empty.
func (ls *LineStack) IsEmpty() bool {
	return len(ls.lines) == 0
}

// Size returns the number of lines in the stack.
func (ls *LineStack) Size() int {
	return len(ls.lines)
}

// Clear removes all lines from the stack.
func (ls *LineStack) Clear() {
	ls.lines = ls.lines[:0]
}

// FindParent finds the appropriate parent for a line at the given level.
// It pops lines from the stack until it finds a parent with a level less than
// the current level, or until the stack is empty.
//
// Returns:
//   - The parent line if found, nil if no parent found (orphaned line)
//   - The stack after finding the parent (with parent at top)
//
// Algorithm:
//   - While stack is not empty and top line's level >= current level:
//     - Pop from stack
//   - If stack is empty after popping:
//     - Return nil (orphaned line)
//   - Else:
//     - Return top of stack (the parent)
func (ls *LineStack) FindParent(currentLevel int) (*gedcom.GedcomLine, error) {
	// Pop until we find a parent with level < currentLevel
	for !ls.IsEmpty() {
		top := ls.Peek()
		if top.Level < currentLevel {
			// Found the parent
			return top, nil
		}
		// Top level is >= current level, pop it
		ls.Pop()
	}

	// Stack is empty, no parent found (orphaned line)
	return nil, fmt.Errorf("no parent found for level %d (orphaned line)", currentLevel)
}

// PopUntilLevel pops lines from the stack until the top line has a level
// less than the target level, or until the stack is empty.
// This is a helper method that can be used before FindParent.
// It pops all lines with level >= targetLevel, keeping only lines with level < targetLevel.
func (ls *LineStack) PopUntilLevel(targetLevel int) {
	for !ls.IsEmpty() {
		top := ls.Peek()
		if top.Level < targetLevel {
			// Found a line with level < target, stop popping
			return
		}
		// top.Level >= targetLevel, pop it
		ls.Pop()
	}
}

// GetAll returns all lines in the stack (from bottom to top).
// Useful for debugging.
func (ls *LineStack) GetAll() []*gedcom.GedcomLine {
	result := make([]*gedcom.GedcomLine, len(ls.lines))
	copy(result, ls.lines)
	return result
}

// String returns a string representation of the stack for debugging.
func (ls *LineStack) String() string {
	if ls.IsEmpty() {
		return "[]"
	}
	result := "["
	for i, line := range ls.lines {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%d:%s", line.Level, line.Tag)
	}
	result += "]"
	return result
}


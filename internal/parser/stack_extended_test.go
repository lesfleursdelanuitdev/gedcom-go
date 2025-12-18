package parser

import (
	"testing"

	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
)

func TestLineStack_GetAll_Extended(t *testing.T) {
	stack := NewLineStack()

	// Empty stack
	all := stack.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(all))
	}

	// Add lines
	line1 := gedcom.NewGedcomLine(0, "HEAD", "", "")
	line2 := gedcom.NewGedcomLine(1, "GEDC", "", "")
	line3 := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")

	stack.Push(line1)
	stack.Push(line2)
	stack.Push(line3)

	// Get all
	all = stack.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 items, got %d", len(all))
	}

	// Verify order (bottom to top)
	if all[0] != line1 {
		t.Error("Expected line1 at index 0")
	}
	if all[1] != line2 {
		t.Error("Expected line2 at index 1")
	}
	if all[2] != line3 {
		t.Error("Expected line3 at index 2")
	}

	// Verify it's a copy (modifying shouldn't affect stack)
	all[0] = nil
	if stack.Peek() != line3 {
		t.Error("Modifying GetAll result should not affect stack")
	}
}

func TestLineStack_String_Extended(t *testing.T) {
	stack := NewLineStack()

	// Empty stack
	str := stack.String()
	if str != "[]" {
		t.Errorf("Expected '[]', got '%s'", str)
	}

	// Add lines
	line1 := gedcom.NewGedcomLine(0, "HEAD", "", "")
	line2 := gedcom.NewGedcomLine(1, "GEDC", "", "")
	line3 := gedcom.NewGedcomLine(2, "VERS", "", "")

	stack.Push(line1)
	str = stack.String()
	expected := "[0:HEAD]"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}

	stack.Push(line2)
	stack.Push(line3)
	str = stack.String()
	expected = "[0:HEAD, 1:GEDC, 2:VERS]"
	if str != expected {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

func TestLineStack_Clear(t *testing.T) {
	stack := NewLineStack()

	// Add lines
	line1 := gedcom.NewGedcomLine(0, "HEAD", "", "")
	line2 := gedcom.NewGedcomLine(1, "GEDC", "", "")
	stack.Push(line1)
	stack.Push(line2)

	if stack.IsEmpty() {
		t.Error("Expected stack to not be empty")
	}
	if stack.Size() != 2 {
		t.Errorf("Expected size 2, got %d", stack.Size())
	}

	// Clear
	stack.Clear()

	if !stack.IsEmpty() {
		t.Error("Expected stack to be empty after Clear")
	}
	if stack.Size() != 0 {
		t.Errorf("Expected size 0, got %d", stack.Size())
	}
	if stack.Peek() != nil {
		t.Error("Expected Peek() to return nil after Clear")
	}
}

func TestLineStack_FindParent_EdgeCases(t *testing.T) {
	stack := NewLineStack()

	// Test with empty stack
	parent, err := stack.FindParent(0)
	if err == nil {
		t.Error("Expected error for empty stack")
	}
	if parent != nil {
		t.Error("Expected nil parent for empty stack")
	}

	// Test with level 0 (should find nothing as parent - level 0 has no parent)
	line0 := gedcom.NewGedcomLine(0, "HEAD", "", "")
	stack.Push(line0)

	parent, err = stack.FindParent(0)
	// FindParent pops while top.Level >= currentLevel
	// For level 0, it pops while top.Level >= 0, so it pops line0 (level 0)
	// Then stack is empty, so it returns error
	if err == nil {
		t.Error("Expected error for level 0 (no parent possible)")
	}
	if parent != nil {
		t.Error("Expected nil parent for level 0")
	}

	// Test with level 1 (should find level 0 as parent)
	// Reset stack
	stack.Clear()
	stack.Push(line0)
	
	parent, err = stack.FindParent(1)
	// FindParent pops while top.Level >= currentLevel (1)
	// line0.Level (0) < 1, so it doesn't pop, and returns line0 as parent
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if parent != line0 {
		t.Error("Expected line0 as parent for level 1")
	}
}

func TestLineStack_PopUntilLevel_EdgeCases(t *testing.T) {
	stack := NewLineStack()

	// Empty stack - should not panic
	stack.PopUntilLevel(0)
	if !stack.IsEmpty() {
		t.Error("Expected empty stack")
	}

	// Add lines
	line0 := gedcom.NewGedcomLine(0, "HEAD", "", "")
	line1 := gedcom.NewGedcomLine(1, "GEDC", "", "")
	line2 := gedcom.NewGedcomLine(2, "VERS", "", "")
	stack.Push(line0)
	stack.Push(line1)
	stack.Push(line2)

	// Pop until level 0
	// PopUntilLevel pops while top.Level >= targetLevel
	// For targetLevel 0, it pops while top.Level >= 0
	// So it pops line2 (2 >= 0), line1 (1 >= 0), line0 (0 >= 0)
	// Stack becomes empty
	stack.PopUntilLevel(0)
	if !stack.IsEmpty() {
		t.Errorf("Expected empty stack after PopUntilLevel(0), got size %d", stack.Size())
	}

	// Reset and test with level 1
	stack.Clear()
	stack.Push(line0)
	stack.Push(line1)
	stack.Push(line2)
	
	// Pop until level 1 (should keep level 0)
	// Pops while top.Level >= 1, so pops line2 (2 >= 1), line1 (1 >= 1)
	// Keeps line0 (0 < 1)
	stack.PopUntilLevel(1)
	if stack.Size() != 1 {
		t.Errorf("Expected size 1, got %d", stack.Size())
	}
	if stack.Peek() != line0 {
		t.Error("Expected line0 at top")
	}

	// Reset and test with level 3 (should keep all)
	stack.Clear()
	stack.Push(line0)
	stack.Push(line1)
	stack.Push(line2)
	
	// Pop until level 3
	// Pops while top.Level >= 3, so nothing pops (all levels < 3)
	stack.PopUntilLevel(3)
	if stack.Size() != 3 {
		t.Errorf("Expected size 3, got %d", stack.Size())
	}
}


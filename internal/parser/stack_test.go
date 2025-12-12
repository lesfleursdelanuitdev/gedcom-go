package parser

import (
	"testing"

	"github.com/yourorg/gedcom/pkg/gedcom"
)

func TestLineStack_BasicOperations(t *testing.T) {
	stack := NewLineStack()

	// Test empty stack
	if !stack.IsEmpty() {
		t.Error("Expected empty stack to be empty")
	}
	if stack.Size() != 0 {
		t.Errorf("Expected size 0, got %d", stack.Size())
	}
	if stack.Peek() != nil {
		t.Error("Expected Peek() on empty stack to return nil")
	}
	if stack.Pop() != nil {
		t.Error("Expected Pop() on empty stack to return nil")
	}

	// Create test lines
	line1 := gedcom.NewGedcomLine(0, "HEAD", "", "")
	line2 := gedcom.NewGedcomLine(1, "GEDC", "", "")
	line3 := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")

	// Test push
	stack.Push(line1)
	if stack.IsEmpty() {
		t.Error("Expected stack to not be empty after push")
	}
	if stack.Size() != 1 {
		t.Errorf("Expected size 1, got %d", stack.Size())
	}
	if stack.Peek() != line1 {
		t.Error("Expected Peek() to return line1")
	}

	// Test push more
	stack.Push(line2)
	stack.Push(line3)
	if stack.Size() != 3 {
		t.Errorf("Expected size 3, got %d", stack.Size())
	}
	if stack.Peek() != line3 {
		t.Error("Expected Peek() to return line3 (top)")
	}

	// Test pop
	popped := stack.Pop()
	if popped != line3 {
		t.Error("Expected Pop() to return line3")
	}
	if stack.Size() != 2 {
		t.Errorf("Expected size 2 after pop, got %d", stack.Size())
	}
	if stack.Peek() != line2 {
		t.Error("Expected Peek() to return line2 after pop")
	}

	// Test clear
	stack.Clear()
	if !stack.IsEmpty() {
		t.Error("Expected stack to be empty after Clear()")
	}
}

func TestLineStack_FindParent(t *testing.T) {
	tests := []struct {
		name        string
		setupStack  func() *LineStack
		currentLevel int
		wantParent  bool
		wantParentLevel int
		wantError   bool
	}{
		{
			name: "find parent at level 0 for level 1",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
				return stack
			},
			currentLevel: 1,
			wantParent:   true,
			wantParentLevel: 0,
			wantError:    false,
		},
		{
			name: "find parent at level 1 for level 2",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
				stack.Push(gedcom.NewGedcomLine(1, "GEDC", "", ""))
				return stack
			},
			currentLevel: 2,
			wantParent:   true,
			wantParentLevel: 1,
			wantError:    false,
		},
		{
			name: "pop until parent found (level decrease)",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
				stack.Push(gedcom.NewGedcomLine(1, "GEDC", "", ""))
				stack.Push(gedcom.NewGedcomLine(2, "VERS", "5.5.5", ""))
				return stack
			},
			currentLevel: 1, // Level decreases from 2 to 1
			wantParent:   true,
			wantParentLevel: 0, // Should pop GEDC and VERS, HEAD is parent
			wantError:    false,
		},
		{
			name: "same level (sibling)",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
				stack.Push(gedcom.NewGedcomLine(1, "GEDC", "", ""))
				return stack
			},
			currentLevel: 1, // Same level as top
			wantParent:   true,
			wantParentLevel: 0, // Should pop GEDC, HEAD is parent
			wantError:    false,
		},
		{
			name: "orphaned line (no parent)",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				// Empty stack
				return stack
			},
			currentLevel: 1,
			wantParent:   false,
			wantError:    true,
		},
		{
			name: "orphaned line after popping all",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(2, "VERS", "5.5.5", ""))
				return stack
			},
			currentLevel: 0, // Level 0, but stack only has level 2
			wantParent:   false,
			wantError:    true,
		},
		{
			name: "deep hierarchy",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "INDI", "", "@I1@"))
				stack.Push(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
				stack.Push(gedcom.NewGedcomLine(2, "GIVN", "John", ""))
				return stack
			},
			currentLevel: 2, // Same level, should pop GIVN, NAME is parent
			wantParent:   true,
			wantParentLevel: 1,
			wantError:    false,
		},
		{
			name: "level jump (skip levels)",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "INDI", "", "@I1@"))
				stack.Push(gedcom.NewGedcomLine(1, "NAME", "John /Doe/", ""))
				stack.Push(gedcom.NewGedcomLine(2, "GIVN", "John", ""))
				stack.Push(gedcom.NewGedcomLine(3, "NICK", "Johnny", ""))
				return stack
			},
			currentLevel: 1, // Jump from level 3 to level 1
			wantParent:   true,
			wantParentLevel: 0, // Should pop NAME, GIVN, NICK, INDI is parent
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := tt.setupStack()
			parent, err := stack.FindParent(tt.currentLevel)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if parent != nil {
					t.Errorf("Expected nil parent on error, got %v", parent)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.wantParent {
					if parent == nil {
						t.Error("Expected parent, got nil")
					} else {
						if parent.Level != tt.wantParentLevel {
							t.Errorf("Expected parent level %d, got %d", tt.wantParentLevel, parent.Level)
						}
					}
				} else {
					if parent != nil {
						t.Errorf("Expected nil parent, got %v", parent)
					}
				}
			}
		})
	}
}

func TestLineStack_PopUntilLevel(t *testing.T) {
	tests := []struct {
		name        string
		setupStack  func() *LineStack
		targetLevel int
		wantSize    int
		wantTopLevel int
	}{
		{
			name: "pop until level 0",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
				stack.Push(gedcom.NewGedcomLine(1, "GEDC", "", ""))
				stack.Push(gedcom.NewGedcomLine(2, "VERS", "5.5.5", ""))
				return stack
			},
			targetLevel: 0,
			wantSize:    0, // All lines have level >= 0, so all are popped
			wantTopLevel: -1, // Stack empty
		},
		{
			name: "pop until level 1",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
				stack.Push(gedcom.NewGedcomLine(1, "GEDC", "", ""))
				stack.Push(gedcom.NewGedcomLine(2, "VERS", "5.5.5", ""))
				return stack
			},
			targetLevel: 1,
			wantSize:    1, // Pop VERS(2) and GEDC(1), keep HEAD(0)
			wantTopLevel: 0,
		},
		{
			name: "no pop needed",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
				stack.Push(gedcom.NewGedcomLine(1, "GEDC", "", ""))
				return stack
			},
			targetLevel: 2,
			wantSize:    2,
			wantTopLevel: 1,
		},
		{
			name: "pop all",
			setupStack: func() *LineStack {
				stack := NewLineStack()
				stack.Push(gedcom.NewGedcomLine(2, "VERS", "5.5.5", ""))
				stack.Push(gedcom.NewGedcomLine(3, "FORM", "LINEAGE-LINKED", ""))
				return stack
			},
			targetLevel: 0,
			wantSize:    0,
			wantTopLevel: -1, // Stack empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := tt.setupStack()
			stack.PopUntilLevel(tt.targetLevel)

			if stack.Size() != tt.wantSize {
				t.Errorf("Expected size %d, got %d", tt.wantSize, stack.Size())
			}

			if tt.wantTopLevel == -1 {
				if !stack.IsEmpty() {
					t.Error("Expected stack to be empty")
				}
			} else {
				top := stack.Peek()
				if top == nil {
					t.Error("Expected non-empty stack")
				} else if top.Level != tt.wantTopLevel {
					t.Errorf("Expected top level %d, got %d", tt.wantTopLevel, top.Level)
				}
			}
		})
	}
}

func TestLineStack_RealWorldExample(t *testing.T) {
	// Simulate parsing a real GEDCOM structure
	stack := NewLineStack()

	// 0 HEAD
	head := gedcom.NewGedcomLine(0, "HEAD", "", "")
	stack.Push(head)
	if stack.Size() != 1 {
		t.Errorf("Expected size 1, got %d", stack.Size())
	}

	// 1 GEDC
	gedc := gedcom.NewGedcomLine(1, "GEDC", "", "")
	parent, err := stack.FindParent(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent != head {
		t.Error("Expected HEAD to be parent of GEDC")
	}
	stack.Push(gedc)

	// 2 VERS 5.5.5
	vers := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")
	parent, err = stack.FindParent(2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent != gedc {
		t.Error("Expected GEDC to be parent of VERS")
	}
	stack.Push(vers)

	// 1 CHAR UTF-8 (level decreases)
	char := gedcom.NewGedcomLine(1, "CHAR", "UTF-8", "")
	parent, err = stack.FindParent(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent != head {
		t.Error("Expected HEAD to be parent of CHAR (after popping GEDC and VERS)")
	}
	if stack.Size() != 1 {
		t.Errorf("Expected stack to have only HEAD after FindParent, got size %d", stack.Size())
	}
	stack.Push(char)

	// 0 @I1@ INDI (level 0, reset stack)
	indi := gedcom.NewGedcomLine(0, "INDI", "", "@I1@")
	parent, err = stack.FindParent(0)
	if err == nil {
		t.Error("Expected error for level 0 (no parent possible)")
	}
	// For level 0, we reset the stack
	stack.Clear()
	stack.Push(indi)

	// 1 NAME John /Doe/
	name := gedcom.NewGedcomLine(1, "NAME", "John /Doe/", "")
	parent, err = stack.FindParent(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent != indi {
		t.Error("Expected INDI to be parent of NAME")
	}
	stack.Push(name)

	// 2 GIVN John
	givn := gedcom.NewGedcomLine(2, "GIVN", "John", "")
	parent, err = stack.FindParent(2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent != name {
		t.Error("Expected NAME to be parent of GIVN")
	}
	stack.Push(givn)

	// 2 SURN Doe (same level as GIVN, sibling)
	surn := gedcom.NewGedcomLine(2, "SURN", "Doe", "")
	parent, err = stack.FindParent(2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if parent != name {
		t.Error("Expected NAME to be parent of SURN (after popping GIVN)")
	}
	stack.Push(surn)
}

func TestLineStack_GetAll(t *testing.T) {
	stack := NewLineStack()
	line1 := gedcom.NewGedcomLine(0, "HEAD", "", "")
	line2 := gedcom.NewGedcomLine(1, "GEDC", "", "")
	line3 := gedcom.NewGedcomLine(2, "VERS", "5.5.5", "")

	stack.Push(line1)
	stack.Push(line2)
	stack.Push(line3)

	all := stack.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(all))
	}
	if all[0] != line1 {
		t.Error("Expected first line to be line1")
	}
	if all[1] != line2 {
		t.Error("Expected second line to be line2")
	}
	if all[2] != line3 {
		t.Error("Expected third line to be line3")
	}

	// Modify returned slice should not affect stack
	all[0] = nil
	if stack.Peek() != line3 {
		t.Error("Modifying returned slice should not affect stack")
	}
}

func TestLineStack_String(t *testing.T) {
	stack := NewLineStack()
	if stack.String() != "[]" {
		t.Errorf("Expected empty stack string '[]', got %q", stack.String())
	}

	stack.Push(gedcom.NewGedcomLine(0, "HEAD", "", ""))
	stack.Push(gedcom.NewGedcomLine(1, "GEDC", "", ""))
	stack.Push(gedcom.NewGedcomLine(2, "VERS", "5.5.5", ""))

	result := stack.String()
	expected := "[0:HEAD, 1:GEDC, 2:VERS]"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}


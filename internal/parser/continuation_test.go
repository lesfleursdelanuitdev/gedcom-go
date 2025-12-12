package parser

import (
	"strings"
	"testing"
)

func TestContinuationHandler_HandleContinuation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *ContinuationHandler
		tag       string
		level     int
		value     string
		wantValue string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "simple CONC",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("NOTE", 1)
				ch.currentValue.WriteString("First part")
				return ch
			},
			tag:       "CONC",
			level:     1,
			value:     " second part",
			wantValue: "First part second part",
			wantErr:   false,
		},
		{
			name: "simple CONT",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("NOTE", 1)
				ch.currentValue.WriteString("First line")
				return ch
			},
			tag:       "CONT",
			level:     1,
			value:     "Second line",
			wantValue: "First line\nSecond line",
			wantErr:   false,
		},
		{
			name: "multiple CONC",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("NOTE", 1)
				ch.currentValue.WriteString("Start")
				return ch
			},
			tag:       "CONC",
			level:     1,
			value:     " middle",
			wantValue: "Start middle",
			wantErr:   false,
		},
		{
			name: "multiple CONT",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("NOTE", 1)
				ch.currentValue.WriteString("Line 1")
				return ch
			},
			tag:       "CONT",
			level:     1,
			value:     "Line 2",
			wantValue: "Line 1\nLine 2",
			wantErr:   false,
		},
		{
			name: "CONC subordinate to CONC (invalid)",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("CONC", 1)
				ch.currentValue.WriteString("First")
				return ch
			},
			tag:     "CONC",
			level:   2, // Higher level = subordinate
			value:   "Second",
			wantErr: true,
			errMsg:  "cannot be subordinate",
		},
		{
			name: "CONT subordinate to CONT (invalid)",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("CONT", 1)
				ch.currentValue.WriteString("First")
				return ch
			},
			tag:     "CONT",
			level:   2, // Higher level = subordinate
			value:   "Second",
			wantErr: true,
			errMsg:  "cannot be subordinate",
		},
		{
			name: "CONC at same level as previous CONC (valid)",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("CONC", 1)
				ch.currentValue.WriteString("First")
				return ch
			},
			tag:       "CONC",
			level:     1, // Same level = valid
			value:     "Second",
			wantValue: "FirstSecond",
			wantErr:   false,
		},
		{
			name: "CONC at lower level than previous CONC (valid)",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("CONC", 2)
				ch.currentValue.WriteString("First")
				return ch
			},
			tag:       "CONC",
			level:     1, // Lower level = valid (not subordinate)
			value:     "Second",
			wantValue: "FirstSecond",
			wantErr:   false,
		},
		{
			name: "mixed CONC and CONT",
			setup: func() *ContinuationHandler {
				ch := NewContinuationHandler()
				ch.SetLastTag("NOTE", 1)
				ch.currentValue.WriteString("Start")
				return ch
			},
			tag:       "CONC",
			level:     1,
			value:     " continued",
			wantValue: "Start continued",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setup()
			err := ch.HandleContinuation(tt.tag, tt.level, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Errorf("HandleContinuation() expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("HandleContinuation() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("HandleContinuation() unexpected error = %v", err)
					return
				}
				gotValue := ch.currentValue.String()
				if gotValue != tt.wantValue {
					t.Errorf("HandleContinuation() accumulated value = %q, want %q", gotValue, tt.wantValue)
				}
			}
		})
	}
}

func TestContinuationHandler_GetAccumulatedValue(t *testing.T) {
	ch := NewContinuationHandler()
	ch.SetLastTag("NOTE", 1)
	ch.currentValue.WriteString("Accumulated text")

	value := ch.GetAccumulatedValue()
	if value != "Accumulated text" {
		t.Errorf("GetAccumulatedValue() = %q, want %q", value, "Accumulated text")
	}

	// Should be reset after getting value
	if ch.HasAccumulatedValue() {
		t.Errorf("GetAccumulatedValue() should reset accumulated value")
	}
}

func TestContinuationHandler_HasAccumulatedValue(t *testing.T) {
	ch := NewContinuationHandler()

	if ch.HasAccumulatedValue() {
		t.Errorf("HasAccumulatedValue() = true for new handler, want false")
	}

	ch.currentValue.WriteString("Some text")
	if !ch.HasAccumulatedValue() {
		t.Errorf("HasAccumulatedValue() = false after writing, want true")
	}

	ch.Reset()
	if ch.HasAccumulatedValue() {
		t.Errorf("HasAccumulatedValue() = true after reset, want false")
	}
}

func TestContinuationHandler_Reset(t *testing.T) {
	ch := NewContinuationHandler()
	ch.SetLastTag("NOTE", 1)
	ch.currentValue.WriteString("Some text")

	ch.Reset()

	if ch.HasAccumulatedValue() {
		t.Errorf("Reset() did not clear accumulated value")
	}
	if ch.GetLastTag() != nil {
		t.Errorf("Reset() did not clear last tag")
	}
}

func TestContinuationHandler_SetLastTag(t *testing.T) {
	ch := NewContinuationHandler()
	ch.SetLastTag("NAME", 1)

	lastTag := ch.GetLastTag()
	if lastTag == nil {
		t.Fatalf("GetLastTag() = nil after SetLastTag")
	}
	if lastTag.Tag != "NAME" {
		t.Errorf("GetLastTag().Tag = %q, want %q", lastTag.Tag, "NAME")
	}
	if lastTag.Level != 1 {
		t.Errorf("GetLastTag().Level = %d, want %d", lastTag.Level, 1)
	}
}

func TestContinuationHandler_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name   string
		ops    []struct{ tag string; level int; value string }
		want   string
	}{
		{
			name: "simple note continuation",
			ops: []struct{ tag string; level int; value string }{
				{"NOTE", 1, "This is a note"},
				{"CONT", 1, "that continues"},
				{"CONT", 1, "on multiple lines"},
			},
			want: "This is a note\nthat continues\non multiple lines",
		},
		{
			name: "concatenated text",
			ops: []struct{ tag string; level int; value string }{
				{"NOTE", 1, "This is a long sentence"},
				{"CONC", 1, " that continues"},
				{"CONC", 1, " without breaks"},
			},
			want: "This is a long sentence that continues without breaks",
		},
		{
			name: "mixed continuation",
			ops: []struct{ tag string; level int; value string }{
				{"NOTE", 1, "Paragraph one"},
				{"CONT", 1, "Paragraph two"},
				{"CONC", 1, " continues"},
			},
			want: "Paragraph one\nParagraph two continues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := NewContinuationHandler()

			// Process first line (not a continuation)
			if len(tt.ops) == 0 {
				t.Fatal("test case has no operations")
			}
			firstOp := tt.ops[0]
			ch.SetLastTag(firstOp.tag, firstOp.level)
			ch.currentValue.WriteString(firstOp.value)

			// Process continuation lines
			for i := 1; i < len(tt.ops); i++ {
				op := tt.ops[i]
				err := ch.HandleContinuation(op.tag, op.level, op.value)
				if err != nil {
					t.Fatalf("HandleContinuation() error = %v", err)
				}
			}

			got := ch.GetAccumulatedValue()
			if got != tt.want {
				t.Errorf("accumulated value = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContinuationHandler_InvalidTag(t *testing.T) {
	ch := NewContinuationHandler()
	err := ch.HandleContinuation("NAME", 1, "value")
	if err == nil {
		t.Errorf("HandleContinuation() expected error for non-CONC/CONT tag")
	}
	if !strings.Contains(err.Error(), "non-continuation tag") {
		t.Errorf("HandleContinuation() error = %v, want error about non-continuation tag", err)
	}
}


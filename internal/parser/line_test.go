package parser

import (
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantLevel int
		wantTag   string
		wantValue string
		wantXref  string
		wantErr   bool
		errMsg    string
	}{
		// Valid cases - Level 0 records
		{
			name:      "level 0 HEAD no value",
			line:       "0 HEAD",
			wantLevel:  0,
			wantTag:    "HEAD",
			wantValue:  "",
			wantXref:   "",
			wantErr:    false,
		},
		{
			name:      "level 0 INDI with xref",
			line:       "0 @I1@ INDI",
			wantLevel:  0,
			wantTag:    "INDI",
			wantValue:  "",
			wantXref:   "@I1@",
			wantErr:    false,
		},
		{
			name:      "level 0 FAM with xref",
			line:       "0 @F1@ FAM",
			wantLevel:  0,
			wantTag:    "FAM",
			wantValue:  "",
			wantXref:   "@F1@",
			wantErr:    false,
		},
		{
			name:      "level 0 TRLR",
			line:       "0 TRLR",
			wantLevel:  0,
			wantTag:    "TRLR",
			wantValue:  "",
			wantXref:   "",
			wantErr:    false,
		},
		
		// Valid cases - Level 1
		{
			name:      "level 1 NAME with value",
			line:       "1 NAME John /Doe/",
			wantLevel:  1,
			wantTag:    "NAME",
			wantValue:  "John /Doe/",
			wantXref:   "",
			wantErr:    false,
		},
		{
			name:      "level 1 SEX with value",
			line:       "1 SEX M",
			wantLevel:  1,
			wantTag:    "SEX",
			wantValue:  "M",
			wantXref:   "",
			wantErr:    false,
		},
		{
			name:      "level 1 tag only",
			line:       "1 ADDR",
			wantLevel:  1,
			wantTag:    "ADDR",
			wantValue:  "",
			wantXref:   "",
			wantErr:    false,
		},
		
		// Valid cases - Level 2+
		{
			name:      "level 2 DATE with value",
			line:       "2 DATE 1 Jan 1900",
			wantLevel:  2,
			wantTag:    "DATE",
			wantValue:  "1 Jan 1900",
			wantXref:   "",
			wantErr:    false,
		},
		{
			name:      "level 2 PLAC with complex value",
			line:       "2 PLAC Weston, Madison, Connecticut, United States of America",
			wantLevel:  2,
			wantTag:    "PLAC",
			wantValue:  "Weston, Madison, Connecticut, United States of America",
			wantXref:   "",
			wantErr:    false,
		},
		{
			name:      "level 3 tag with value",
			line:       "3 PAGE Sec. 2, p. 45",
			wantLevel:  3,
			wantTag:    "PAGE",
			wantValue:  "Sec. 2, p. 45",
			wantXref:   "",
			wantErr:    false,
		},
		
		// Edge cases - XREF in different positions
		{
			name:      "level 0 NOTE with xref and value",
			line:       "0 @N1@ NOTE This is a note",
			wantLevel:  0,
			wantTag:    "NOTE",
			wantValue:  "This is a note",
			wantXref:   "@N1@",
			wantErr:    false,
		},
		{
			name:      "level 0 SOUR with xref",
			line:       "0 @S1@ SOUR",
			wantLevel:  0,
			wantTag:    "SOUR",
			wantValue:  "",
			wantXref:   "@S1@",
			wantErr:    false,
		},
		
		// Edge cases - Special characters in value
		{
			name:      "value with slashes (surname)",
			line:       "1 NAME Robert Eugene /Williams/",
			wantLevel:  1,
			wantTag:    "NAME",
			wantValue:  "Robert Eugene /Williams/",
			wantXref:   "",
			wantErr:    false,
		},
		{
			name:      "value with at symbol (not xref)",
			line:       "1 EMAIL user@example.com",
			wantLevel:  1,
			wantTag:    "EMAIL",
			wantValue:  "user@example.com",
			wantXref:   "",
			wantErr:    false,
		},
		
		// Error cases
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
			errMsg:  "empty line",
		},
		{
			name:    "whitespace only",
			line:    "   ",
			wantErr: true,
			errMsg:  "empty line",
		},
		{
			name:    "only level number",
			line:    "0",
			wantErr: true,
			errMsg:  "insufficient parts",
		},
		{
			name:    "invalid level (non-numeric)",
			line:    "X HEAD",
			wantErr: true,
			errMsg:  "invalid level",
		},
		{
			name:    "negative level",
			line:    "-1 HEAD",
			wantErr: true,
			errMsg:  "level cannot be negative",
		},
		{
			name:    "level with decimal",
			line:    "1.5 HEAD",
			wantErr: true,
			errMsg:  "invalid level",
		},
		{
			name:      "very high level (valid but unusual)",
			line:       "99 DEEP",
			wantLevel:  99,
			wantTag:    "DEEP",
			wantValue:  "",
			wantXref:   "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, tag, value, xref, err := ParseLine(tt.line)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseLine() expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseLine() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseLine() unexpected error = %v", err)
				return
			}
			
			if level != tt.wantLevel {
				t.Errorf("ParseLine() level = %d, want %d", level, tt.wantLevel)
			}
			if tag != tt.wantTag {
				t.Errorf("ParseLine() tag = %q, want %q", tag, tt.wantTag)
			}
			if value != tt.wantValue {
				t.Errorf("ParseLine() value = %q, want %q", value, tt.wantValue)
			}
			if xref != tt.wantXref {
				t.Errorf("ParseLine() xref = %q, want %q", xref, tt.wantXref)
			}
		})
	}
}

// TestParseLine_RealWorldExamples tests with examples from actual GEDCOM files
func TestParseLine_RealWorldExamples(t *testing.T) {
	realWorldExamples := []struct {
		name string
		line string
	}{
		{"header", "0 HEAD"},
		{"gedc version", "2 VERS 5.5.5"},
		{"character encoding", "1 CHAR UTF-8"},
		{"submitter reference", "1 SUBM @U1@"},
		{"individual with xref", "0 @I1@ INDI"},
		{"name with surname", "1 NAME Robert Eugene /Williams/"},
		{"surname component", "2 SURN Williams"},
		{"given name component", "2 GIVN Robert Eugene"},
		{"sex", "1 SEX M"},
		{"birth event", "1 BIRT"},
		{"birth date", "2 DATE 2 Oct 1822"},
		{"birth place", "2 PLAC Weston, Madison, Connecticut, United States of America"},
		{"source citation", "2 SOUR @S1@"},
		{"source page", "3 PAGE Sec. 2, p. 45"},
		{"death event", "1 DEAT"},
		{"death date", "2 DATE 14 Apr 1905"},
		{"family spouse link", "1 FAMS @F1@"},
		{"family child link", "1 FAMC @F1@"},
		{"family husband", "1 HUSB @I1@"},
		{"family wife", "1 WIFE @I2@"},
		{"family child", "1 CHIL @I3@"},
		{"marriage event", "1 MARR"},
		{"marriage date", "2 DATE Dec 1859"},
		{"trailer", "0 TRLR"},
	}
	
	for _, tt := range realWorldExamples {
		t.Run(tt.name, func(t *testing.T) {
			level, tag, value, xref, err := ParseLine(tt.line)
			if err != nil {
				t.Errorf("ParseLine(%q) unexpected error = %v", tt.line, err)
				return
			}
			
			// Basic validation
			if level < 0 {
				t.Errorf("ParseLine(%q) level = %d, want >= 0", tt.line, level)
			}
			if tag == "" {
				t.Errorf("ParseLine(%q) tag is empty", tt.line)
			}
			
			// Log parsed values for debugging
			t.Logf("Parsed: level=%d, tag=%q, value=%q, xref=%q", level, tag, value, xref)
		})
	}
}

// TestParseLine_EdgeCases tests edge cases and boundary conditions
func TestParseLine_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"leading whitespace", "  0 HEAD"},
		{"trailing whitespace", "0 HEAD  "},
		{"multiple spaces", "0    HEAD"},
		{"tab separator", "0\tHEAD"},
		{"value with leading space", "1 NAME  John"},
		{"value with trailing space", "1 NAME John "},
		{"empty value", "1 NOTE "},
		{"xref at start of value", "1 TEXT @notxref@ text"},
		{"xref-like in value", "1 NOTE See @reference@ in text"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, tag, value, xref, err := ParseLine(tt.line)
			if err != nil {
				t.Logf("ParseLine(%q) = error: %v (may be expected)", tt.line, err)
				return
			}
			t.Logf("Parsed: level=%d, tag=%q, value=%q, xref=%q", level, tag, value, xref)
		})
	}
}


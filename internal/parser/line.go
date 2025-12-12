package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseLine parses a single GEDCOM line into its components.
//
// GEDCOM line format: LEVEL [XREF_ID] TAG [VALUE]
//
// Examples:
//   - "0 HEAD"                    → (0, "HEAD", "", "", nil)
//   - "0 @I1@ INDI"              → (0, "INDI", "", "@I1@", nil)
//   - "1 NAME John /Doe/"        → (1, "NAME", "John /Doe/", "", nil)
//   - "2 DATE 1 Jan 1900"        → (2, "DATE", "1 Jan 1900", "", nil)
//
// Returns:
//   - level: The level number (0, 1, 2, etc.)
//   - tag: The tag name (HEAD, INDI, NAME, etc.)
//   - value: The value after the tag (empty if no value)
//   - xrefID: The cross-reference ID if present (empty if not)
//   - err: Error if line format is invalid
func ParseLine(line string) (level int, tag string, value string, xrefID string, err error) {
	// Trim whitespace
	line = strings.TrimSpace(line)
	
	// Check for empty line
	if line == "" {
		return 0, "", "", "", fmt.Errorf("empty line")
	}
	
	// Must have at least level and tag
	// First, split into max 3 parts: [level, tag/xref, rest]
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 2 {
		return 0, "", "", "", fmt.Errorf("line has insufficient parts: %q", line)
	}
	
	// Parse level (first part)
	level, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", "", fmt.Errorf("invalid level %q: %w", parts[0], err)
	}
	
	// Level must be non-negative
	if level < 0 {
		return 0, "", "", "", fmt.Errorf("level cannot be negative: %d", level)
	}
	
	// Check if second part is an xref
	hasXref := strings.HasPrefix(parts[1], "@") && strings.HasSuffix(parts[1], "@")
	
	if hasXref {
		// Format: level xref tag [value]
		// Examples: "0 @I1@ INDI" or "0 @N1@ NOTE This is a note"
		xrefID = parts[1]
		if len(parts) < 3 {
			return 0, "", "", "", fmt.Errorf("line with xref missing tag: %q", line)
		}
		
		// Split the rest (tag + value) by space, max 2 parts
		restParts := strings.SplitN(parts[2], " ", 2)
		tag = restParts[0]
		if len(restParts) == 2 {
			value = restParts[1]
		}
		return level, tag, value, xrefID, nil
	} else {
		// Format: level tag [value]
		// Examples: "0 HEAD" or "1 NAME John /Doe/"
		tag = parts[1]
		if len(parts) == 3 {
			value = parts[2]
		}
		return level, tag, value, "", nil
	}
}


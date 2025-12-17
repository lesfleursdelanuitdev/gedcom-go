package internal

import (
	"os"
	"strconv"

	"github.com/fatih/color"
)

var (
	// Color functions
	Success = color.New(color.FgGreen, color.Bold)
	Error   = color.New(color.FgRed, color.Bold)
	Warning = color.New(color.FgYellow, color.Bold)
	Info    = color.New(color.FgBlue, color.Bold)
	Hint    = color.New(color.FgCyan)

	// Data colors
	IndividualName = color.New(color.FgCyan, color.Bold)
	FamilyID       = color.New(color.FgMagenta, color.Bold)
	Date           = color.New(color.FgYellow)
	Place          = color.New(color.FgBlue)
)

// InitColor initializes color output based on environment and config
func InitColor(enableColor bool) {
	// Check NO_COLOR environment variable
	if noColor, _ := strconv.ParseBool(os.Getenv("NO_COLOR")); noColor {
		color.NoColor = true
		return
	}

	// Check if terminal supports color
	if !color.NoColor {
		color.NoColor = !enableColor
	}
}

// IsColorEnabled returns whether color output is enabled
func IsColorEnabled() bool {
	return !color.NoColor
}

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	Success.Printf(format, args...)
}

// PrintError prints an error message
func PrintError(format string, args ...interface{}) {
	Error.Printf(format, args...)
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	Warning.Printf(format, args...)
}

// PrintInfo prints an info message
func PrintInfo(format string, args ...interface{}) {
	Info.Printf(format, args...)
}

// PrintHint prints a hint message
func PrintHint(format string, args ...interface{}) {
	Hint.Printf(format, args...)
}

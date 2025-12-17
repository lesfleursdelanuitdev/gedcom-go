package internal

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

// ProgressBar wraps the progressbar library
type ProgressBar struct {
	bar *progressbar.ProgressBar
}

// NewProgressBar creates a new progress bar
func NewProgressBar(max int64, description string) *ProgressBar {
	if !ShouldShowProgress() {
		return &ProgressBar{bar: nil}
	}

	bar := progressbar.NewOptions64(
		max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(100),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			io.WriteString(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	return &ProgressBar{bar: bar}
}

// Add increments the progress bar
func (p *ProgressBar) Add(n int) {
	if p.bar != nil {
		p.bar.Add(n)
	}
}

// Set sets the progress bar value
func (p *ProgressBar) Set(n int) {
	if p.bar != nil {
		p.bar.Set(n)
	}
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() {
	if p.bar != nil {
		p.bar.Finish()
	}
}

// ShouldShowProgress determines if progress bars should be shown
func ShouldShowProgress() bool {
	// Check if quiet mode is enabled (via environment or config)
	// For now, always show unless explicitly disabled
	return true
}

// SetQuietMode sets whether to show progress bars
var quietMode = false

func SetQuietMode(quiet bool) {
	quietMode = quiet
}

func IsQuietMode() bool {
	return quietMode
}

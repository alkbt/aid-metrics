// Package reporter handles output generation for aid-metrics analysis results.
// This file implements progress reporting for console output.
package reporter

import (
	"fmt"
	"time"
	
	"github.com/schollz/progressbar/v3"
)

// ConsoleProgressReporter implements models.ProgressReporter using a terminal progress bar.
// It provides visual feedback during long-running operations like package discovery and analysis.
type ConsoleProgressReporter struct {
	bar *progressbar.ProgressBar
}

// NewConsoleProgressReporter creates a new progress reporter that outputs to the console.
// The progress bar shows the current operation description and progress percentage.
func NewConsoleProgressReporter() *ConsoleProgressReporter {
	return &ConsoleProgressReporter{}
}

// SetTotal initializes the progress bar with the given total value.
// For aid-metrics, this is typically set to 100 for a percentage-based display.
func (r *ConsoleProgressReporter) SetTotal(total int) {
	r.bar = progressbar.NewOptions(total,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowDescriptionAtLineEnd(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]█[reset]",
			SaucerHead:    "[green]█[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionThrottle(1*time.Second), // Update display at most once per second
	)
}

// Update sets the current progress and updates the description.
// This is thread-safe and can be called from multiple goroutines.
func (r *ConsoleProgressReporter) Update(current int, description string) {
	if r.bar == nil {
		return
	}
	r.bar.Describe(description)
	_ = r.bar.Set(current)
}

// Complete marks the progress as complete and cleans up the progress bar.
func (r *ConsoleProgressReporter) Complete() {
	if r.bar == nil {
		return
	}
	_ = r.bar.Finish()
	// Add newline after progress bar to separate from following output
	fmt.Println()
}
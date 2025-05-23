// Package models contains data structures and interfaces used throughout the aid-metrics tool.
// This file defines the progress reporting interface used to provide feedback during analysis.
package models

// ProgressReporter defines an interface for reporting progress during package analysis.
// Implementations can provide visual feedback through progress bars, spinners, or logs.
// The interface uses a fixed 0-100 scale for consistent progress representation.
type ProgressReporter interface {
	// SetTotal sets the total number of steps for the progress bar.
	// This should be called once at the beginning of the operation.
	// For aid-metrics, we use a fixed scale of 100.
	SetTotal(total int)

	// Update updates the current progress with a description of the current operation.
	// current should be between 0 and the total value set with SetTotal.
	// description should be a short, descriptive string of what's currently happening.
	//
	// Example:
	//   reporter.Update(25, "Loading package: github.com/user/project/pkg/analyzer")
	Update(current int, description string)

	// Complete marks the operation as complete.
	// This should be called when all operations are finished.
	// Implementations may use this to clean up resources or show a final message.
	Complete()
}
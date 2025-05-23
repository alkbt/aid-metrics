// Package analyzer provides functionality for analyzing Go modules and calculating design metrics.
// This file implements batch loading of packages with progress reporting.
package analyzer

import (
	"fmt"
	"strings"
	
	"github.com/alkbt/aid-metrics/pkg/models"
	"golang.org/x/tools/go/packages"
)

// BatchLoader handles loading packages in batches with progress reporting.
// This approach provides better memory usage and allows for progress feedback
// compared to loading all packages at once.
type BatchLoader struct {
	// batchSize controls how many packages are loaded in each batch
	batchSize int
	
	// config is the packages.Config used for loading
	config *packages.Config
	
	// progressReporter provides progress feedback during loading
	progressReporter models.ProgressReporter
	
	// totalPackages is the total number of packages to load
	totalPackages int
}

// NewBatchLoader creates a new BatchLoader with the given configuration.
//
// Parameters:
//   - batchSize: Number of packages to load in each batch (default: 100)
//   - config: The packages.Config to use for loading
//   - progressReporter: Optional progress reporter for feedback
//   - totalPackages: Total number of packages (used for progress calculation)
func NewBatchLoader(batchSize int, config *packages.Config, progressReporter models.ProgressReporter, totalPackages int) *BatchLoader {
	if batchSize <= 0 {
		batchSize = 100
	}
	
	return &BatchLoader{
		batchSize:        batchSize,
		config:           config,
		progressReporter: progressReporter,
		totalPackages:    totalPackages,
	}
}

// LoadPackages loads all packages in batches, reporting progress as it goes.
// The loading phase uses progress values 10-80 on the fixed 0-100 scale.
//
// This method:
//   1. Splits the package list into batches
//   2. Loads each batch using packages.Load
//   3. Reports progress after each batch
//   4. Collects all loaded packages and returns them
//
// Returns an error if any batch fails to load.
func (bl *BatchLoader) LoadPackages(packageInfos []PackageInfo) ([]*packages.Package, error) {
	var allPackages []*packages.Package
	packagesLoaded := 0
	
	// Calculate progress range (10-80 on our 0-100 scale)
	progressStart := 10
	progressEnd := 80
	progressRange := progressEnd - progressStart
	
	// Process packages in batches
	for i := 0; i < len(packageInfos); i += bl.batchSize {
		// Determine batch boundaries
		end := i + bl.batchSize
		if end > len(packageInfos) {
			end = len(packageInfos)
		}
		
		// Extract import paths for this batch
		batchPaths := make([]string, 0, end-i)
		for j := i; j < end; j++ {
			batchPaths = append(batchPaths, packageInfos[j].ImportPath)
		}
		
		// Report progress with current package being loaded
		if bl.progressReporter != nil && len(batchPaths) > 0 {
			progress := progressStart + (packagesLoaded * progressRange / bl.totalPackages)
			// Show only upper bound of loaded packages
			upperBound := packagesLoaded + len(batchPaths)
			description := fmt.Sprintf("Loading %d of %d packages", upperBound, bl.totalPackages)
			bl.progressReporter.Update(progress, description)
		}
		
		// Load this batch
		pkgs, err := packages.Load(bl.config, batchPaths...)
		if err != nil {
			return nil, fmt.Errorf("failed to load packages batch starting at %s: %w", batchPaths[0], err)
		}
		
		// Check for errors in loaded packages
		for _, pkg := range pkgs {
			if len(pkg.Errors) > 0 {
				// Log package errors but don't fail - some packages might have issues
				// This matches the behavior of the original implementation
				continue
			}
		}
		
		// Add to results
		allPackages = append(allPackages, pkgs...)
		packagesLoaded += len(pkgs)
		
		// Update progress after batch completes
		if bl.progressReporter != nil {
			progress := progressStart + (packagesLoaded * progressRange / bl.totalPackages)
			if progress > progressEnd {
				progress = progressEnd
			}
			bl.progressReporter.Update(progress, fmt.Sprintf("Loaded %d of %d packages", packagesLoaded, bl.totalPackages))
		}
	}
	
	return allPackages, nil
}

// shortenPackagePath creates a shorter, more readable version of a package path.
// For example: "github.com/cockroachdb/cockroach/build/bazelutil/staticcheckanalyzers/st1016"
// becomes: ".../staticcheckanalyzers/st1016"
func shortenPackagePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 3 {
		return path
	}
	// Show last 2 parts with ellipsis
	return ".../" + strings.Join(parts[len(parts)-2:], "/")
}
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alkbt/aid-metrics/pkg/analyzer"
	"github.com/alkbt/aid-metrics/pkg/models"
	"github.com/alkbt/aid-metrics/pkg/reporter"
)

func main() {
	// Parse command-line flags
	var format string
	var pattern string
	var progress bool
	var batchSize int

	flag.StringVar(&format, "format", "text", "Output format (text, csv, json)")
	flag.StringVar(&pattern, "pattern", "./...", "Package pattern to analyze (e.g., './...' or 'github.com/org/repo/pkg/...')")
	flag.BoolVar(&progress, "progress", false, "Show progress bar during analysis")
	flag.IntVar(&batchSize, "batch-size", 100, "Number of packages to load in each batch")
	flag.Parse()

	// Get module path
	args := flag.Args()
	modulePath := "."
	if len(args) > 0 {
		modulePath = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(modulePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	// Analyze module
	if !progress {
		fmt.Fprintf(os.Stderr, "Analyzing Go module at: %s\n", absPath)
	}
	
	// Create analyzer options with progress reporter if requested
	var metrics *models.ModuleMetrics
	if progress {
		opts := analyzer.AnalyzerOptions{
			ProgressReporter: reporter.NewConsoleProgressReporter(),
			BatchSize:        batchSize,
		}
		metrics, err = analyzer.AnalyzeModuleWithOptions(absPath, pattern, opts)
	} else {
		metrics, err = analyzer.AnalyzeModule(absPath, pattern)
	}
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to analyze module: %v\n", err)
		os.Exit(1)
	}

	// Generate report
	reportFormat := reporter.FormatType(format)
	if !progress {
		fmt.Fprintf(os.Stderr, "Generating %s report...\n", reportFormat)
	}
	r := reporter.NewReporter(metrics, reportFormat)
	if err := r.Generate(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate report: %v\n", err)
		os.Exit(1)
	}
}

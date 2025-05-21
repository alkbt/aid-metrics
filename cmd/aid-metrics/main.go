package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alkbt/aid-metrics/pkg/analyzer"
	"github.com/alkbt/aid-metrics/pkg/reporter"
)

func main() {
	// Parse command-line flags
	var format string
	var pattern string

	flag.StringVar(&format, "format", "text", "Output format (text, csv, json)")
	flag.StringVar(&pattern, "pattern", "./...", "Package pattern to analyze (e.g., './...' or 'github.com/org/repo/pkg/...')")
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
	fmt.Fprintf(os.Stderr, "Analyzing Go module at: %s\n", absPath)
	metrics, err := analyzer.AnalyzeModule(absPath, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to analyze module: %v\n", err)
		os.Exit(1)
	}

	// Generate report
	reportFormat := reporter.FormatType(format)
	fmt.Fprintf(os.Stderr, "Generating %s report...\n", reportFormat)
	r := reporter.NewReporter(metrics, reportFormat)
	if err := r.Generate(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate report: %v\n", err)
		os.Exit(1)
	}
}

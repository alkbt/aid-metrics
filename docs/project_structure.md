# aid-metrics: Package Design Metrics Analyzer

This document describes the structure of the aid-metrics project, a tool for analyzing Go modules to calculate Robert Martin's package design metrics.

## Project Overview

aid-metrics calculates the following metrics for Go packages:
- **Instability (I)**: How dependent a package is on other packages
- **Abstractness (A)**: The ratio of abstract types to all types
- **Distance from the main sequence (D)**: How far a package is from the optimal balance

## Project Structure

```
aid-metrics
├── cmd/                  # Command-line interface
│   └── aid-metrics/      # CLI implementation
│       └── main.go       # Entry point for the CLI tool
├── pkg/                  # Core library packages
│   ├── analyzer/         # Package analysis implementation
│   │   ├── analyzer.go   # Module analysis logic
│   │   ├── analyzer_test.go  # Tests for analyzer
│   │   ├── discovery.go  # Package discovery without loading
│   │   └── loader.go     # Batch loading for large projects
│   ├── models/           # Data models
│   │   ├── metrics.go    # Package metrics data structures
│   │   └── progress.go   # Progress reporting interface
│   └── reporter/         # Output reporting
│       ├── reporter.go   # Report generation in various formats
│       └── progress.go   # Console progress bar implementation
└── test/                 # Test utilities and fixtures
    └── testmodule/       # Test module for validating analysis
        ├── pkg1/         # Test package with nested subpackage
        │   └── pkg2/     # Nested test package
        ├── pkg3/         # Test package with dependencies
        ├── main.go       # Test module entry point
        └── go.mod        # Module definition for tests
```

## Component Details

### Command Line Interface

- **cmd/aid-metrics/main.go**: Entry point for the CLI tool
  - Parses command-line flags and arguments
  - Determines the module path to analyze
  - Invokes the analyzer with the specified pattern
  - Generates and outputs the report in the requested format
  - New flags:
    - `-progress`: Shows progress bar during analysis
    - `-batch-size`: Controls batch size for loading (default: 100)

### Core Library Packages

#### Analyzer

- **pkg/analyzer/analyzer.go**: The core analysis engine
  - Scans the Go module structure
  - Identifies dependencies between packages
  - Counts abstract and concrete types
  - Calculates instability, abstractness, and distance metrics
  - Key features:
    - Detects standard library imports reliably by checking module path from go.mod
    - Handles various module naming conventions (including those without dots)
    - Properly maps package import paths to friendly names in reports
    - Supports concurrent package analysis
    - Integrates with progress reporting for large projects

- **pkg/analyzer/analyzer_test.go**: Unit tests for the analyzer

- **pkg/analyzer/discovery.go**: Package discovery functionality
  - Performs fast filesystem traversal to find Go packages
  - Supports pattern matching (e.g., "./...", specific paths)
  - Reports progress during discovery phase
  - Returns package information without loading full AST

- **pkg/analyzer/loader.go**: Batch loading for memory efficiency
  - Loads packages in configurable batches
  - Reduces memory usage for large projects
  - Enables progress reporting during the loading phase
  - Maintains compatibility with original packages.Load behavior

#### Models

- **pkg/models/metrics.go**: Data structures for metrics
  - `PackageMetrics`: Stores metrics for a single package
    - Includes counts (Ca, Ce, Na, Nc) and calculated metrics (I, A, D)
  - `ModuleMetrics`: Collects metrics for all packages in a module

- **pkg/models/progress.go**: Progress reporting interface
  - `ProgressReporter`: Interface for progress updates
  - Simple 3-method API: SetTotal, Update, Complete
  - Enables pluggable progress implementations

#### Reporter

- **pkg/reporter/reporter.go**: Generates formatted reports
  - Supports multiple output formats (text, CSV, JSON)
  - Organizes metric data for presentation

- **pkg/reporter/progress.go**: Console progress bar implementation
  - `ConsoleProgressReporter`: Terminal progress bar using schollz/progressbar
  - Thread-safe for concurrent operations
  - Fixed 0-100 scale with phase-based progress

### Test Utilities

- **test/testmodule/**: A sample Go module for testing
  - Contains packages with different dependency relationships
  - Used to validate analyzer functionality
  - Includes a simple module name without dots to test edge cases

## Usage Modes

aid-metrics can be used in two ways:

1. **As a CLI tool**: Running the `aid-metrics` command to analyze modules and output reports
2. **As a library**: Importing the `analyzer` and `reporter` packages for programmatic use

## Key Concepts

### Standard Library Detection

The analyzer needs to differentiate between standard library packages, local module packages, and external dependencies. It does this by:

1. Reading the module path from go.mod
2. Comparing import paths against the module path
3. Using additional heuristics to identify standard library packages

### Package Path Resolution

For reporting purposes, the analyzer:
1. Displays package paths relative to their module
2. Handles different module naming conventions (GitHub-style and simple names)
3. Preserves the package hierarchy to show the logical structure 
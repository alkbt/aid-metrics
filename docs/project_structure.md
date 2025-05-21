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
│   │   └── analyzer_test.go  # Tests for analyzer
│   ├── models/           # Data models
│   │   └── metrics.go    # Package metrics data structures
│   └── reporter/         # Output reporting
│       └── reporter.go   # Report generation in various formats
└── test/                 # Test utilities and fixtures
```

## Component Details

### Command Line Interface

- **cmd/aid-metrics/main.go**: Entry point for the CLI tool
  - Parses command-line flags and arguments
  - Determines the module path to analyze
  - Invokes the analyzer with the specified pattern
  - Generates and outputs the report in the requested format

### Core Library Packages

#### Analyzer

- **pkg/analyzer/analyzer.go**: The core analysis engine
  - Scans the Go module structure
  - Identifies dependencies between packages
  - Counts abstract and concrete types
  - Calculates instability, abstractness, and distance metrics

- **pkg/analyzer/analyzer_test.go**: Unit tests for the analyzer

#### Models

- **pkg/models/metrics.go**: Data structures for metrics
  - `PackageMetrics`: Stores metrics for a single package
    - Includes counts (Ca, Ce, Na, Nc) and calculated metrics (I, A, D)
  - `ModuleMetrics`: Collects metrics for all packages in a module

#### Reporter

- **pkg/reporter/reporter.go**: Generates formatted reports
  - Supports multiple output formats (text, CSV, JSON)
  - Organizes metric data for presentation

## Usage Modes

aid-metrics can be used in two ways:

1. **As a CLI tool**: Running the `aid-metrics` command to analyze modules and output reports
2. **As a library**: Importing the `analyzer` and `reporter` packages for programmatic use 
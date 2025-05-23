# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **aid-metrics**, a Go code quality analysis tool that calculates Robert Martin's package design metrics:
- **Abstractness (A)**: Ratio of interfaces to total types
- **Instability (I)**: Measure of package dependencies (Ce/(Ca+Ce))
- **Distance from Main Sequence (D)**: Distance from optimal design balance

The tool helps assess architectural health, identify problematic packages, and track code quality over time.

## Build and Development Commands

```bash
# Build the CLI tool
go build ./cmd/aid-metrics

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run the tool locally
go run ./cmd/aid-metrics [path] -format=json -pattern="./..."

# Run with progress reporting (shows progress bar)
go run ./cmd/aid-metrics -progress

# Run with custom batch size for loading
go run ./cmd/aid-metrics -progress -batch-size=50

# Install the CLI globally
go install github.com/alkbt/aid-metrics/cmd/aid-metrics@latest
```

## Architecture

The codebase follows a clean architecture with clear separation of concerns:

### Core Components

1. **pkg/analyzer/** - The analysis engine that:
   - Parses Go AST to extract package structure
   - Builds dependency graphs (Ca/Ce coupling)
   - Counts abstract (interfaces) vs concrete types
   - Calculates the three key metrics
   - Handles standard library detection using go.mod

2. **pkg/models/** - Data structures for metrics:
   - `PackageMetrics`: Per-package metrics (Ca, Ce, Na, Nc, A, I, D)
   - `ModuleMetrics`: Collection of package metrics for a module

3. **pkg/reporter/** - Output generation supporting text, CSV, and JSON formats

4. **cmd/aid-metrics/** - CLI wrapper with flag parsing for format and pattern options

### Key Design Patterns

- **Dual-mode design**: Works as both CLI tool and importable library
- **Concurrent processing**: Analyzer supports parallel package analysis
- **Module-aware**: Properly handles Go modules including edge cases (modules without dots)
- **Standard library detection**: Uses go.mod to distinguish between stdlib, local, and external packages
- **Progress reporting**: Optional progress bar for large projects using fixed 0-100 scale
- **Batch loading**: Loads packages in configurable batches to balance memory usage and performance

## Testing Strategy

- Unit tests exist for all packages (e.g., `analyzer_test.go`)
- Test modules in `test/` directory provide realistic package structures
- Performance requirement: Must analyze 100-package module in under 15 seconds

## Important Implementation Details

1. **Type Counting**:
   - Interfaces count as abstract types (Na)
   - Structs and standalone functions count as concrete types (Nc)
   - Type aliases and other definitions are not counted

2. **Dependency Analysis**:
   - Ca (afferent coupling): Packages that depend on this package
   - Ce (efferent coupling): Packages that this package depends on
   - Only counts dependencies within the analyzed module

3. **Module Path Resolution**:
   - Reads module path from go.mod
   - Handles modules with and without dots in their names
   - Distinguishes between local, external, and standard library imports
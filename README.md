# AID-Metrics

A command-line tool and library for analyzing Go modules to calculate Robert Martin's package design metrics:
- Abstractness (A)
- Instability (I)
- Distance from the main sequence (D)

## Installation

```
go install github.com/alkbt/aid-metrics/cmd/aid-metrics@latest
```

## Usage

### As a CLI tool

```bash
# Analyze the current module
aid-metrics

# Analyze a specific module
aid-metrics /path/to/module

# Choose output format (text, csv, json)
aid-metrics -format=json

# Filter packages to analyze
aid-metrics -pattern="./pkg/..."

# Show progress bar during analysis (useful for large projects)
aid-metrics -progress

# Customize batch size for package loading (default: 100)
aid-metrics -progress -batch-size=50

# Combine flags for customized analysis
aid-metrics -progress -format=json -pattern="./pkg/..."
```

### Example Output

When running the tool, you'll see output similar to this:

```
Analyzing Go module at: /path/to/module
Generating text report...
MODULE: /path/to/module

PACKAGE          Ca  Ce  I     Na  Nc  A     D
-------          --  --  -     --  --  -     -
cmd/app          0   2   1.00  0   1   0.00  0.00
pkg/analyzer     1   2   0.67  0   5   0.00  0.33
pkg/models       2   0   0.00  0   2   0.00  1.00
pkg/reporter     1   1   0.50  0   4   0.00  0.50
```

Where:
- `PACKAGE`: Package path relative to the module
- `Ca`: Afferent Coupling (number of packages that depend on this package)
- `Ce`: Efferent Coupling (number of packages this package depends on)
- `I`: Instability (Ce / (Ca + Ce))
- `Na`: Number of abstract types (interfaces)
- `Nc`: Number of concrete types (structs + standalone functions)
- `A`: Abstractness (Na / Nc)
- `D`: Distance from the main sequence (|A + I - 1|)

### As a library

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/alkbt/aid-metrics/pkg/analyzer"
    "github.com/alkbt/aid-metrics/pkg/reporter"
)

func main() {
    // Basic usage
    metrics, err := analyzer.AnalyzeModule("/path/to/module", "./...")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // With progress reporting
    opts := analyzer.AnalyzerOptions{
        ProgressReporter: reporter.NewConsoleProgressReporter(),
        BatchSize:        50,
    }
    metrics, err = analyzer.AnalyzeModuleWithOptions("/path/to/module", "./...", opts)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // Generate a report
    r := reporter.NewReporter(metrics, reporter.FormatType("json"))
    if err := r.Generate(os.Stdout); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Metrics Explanation

### Instability (I)
- **Formula**: I = Ce / (Ca + Ce)
- **Range**: 0 (stable) to 1 (unstable)
- **Meaning**: How likely a package is to change. Higher instability indicates higher dependency on other packages.

### Abstractness (A)
- **Formula**: A = Na / Nc
- **Range**: 0 (concrete) to 1 (abstract)
- **Meaning**: The ratio of abstract types to all types in a package.
  - Na: Number of abstract types (interfaces)
  - Nc: Total number of concrete types (interfaces, structs) plus standalone functions
    - Only structs and standalone functions are counted as concrete types
    - Other type definitions (type aliases, etc.) are not counted

### Distance (D)
- **Formula**: D = |A + I - 1|
- **Range**: 0 (optimal) to 1 (problematic)
- **Meaning**: How far a package is from the "main sequence" (A + I = 1)
  - Packages with D=0 are either:
    - Stable and abstract (good for core functionality)
    - Unstable and concrete (good for application-specific code)
  - Packages with high D are either:
    - Stable and concrete ("pain") - hard to extend
    - Unstable and abstract ("waste") - over-engineered

## Documentation

See the [docs/](docs/) directory for:
- Implementation plans and feature documentation
- Architecture decisions
- Development guides

## License

MIT 
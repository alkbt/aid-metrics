# Go Code Metrics - Product Requirements Document

## 1. Introduction

Go Code Metrics is a command-line application that analyzes Go modules to calculate Robert Martin's package design metrics: instability, abstractness, and distance from the main sequence. These metrics help developers assess code quality, maintainability, and adherence to clean architecture principles. The project is designed to function both as a standalone CLI tool and as a reusable library that can be integrated into other Go projects.

## 2. Product Overview

Go Code Metrics scans Go module source files, analyzes package dependencies and structures, and generates reports on key architectural metrics to help development teams identify potential design issues and improve code organization. It can be used as a command-line application or imported as a library in other Go projects to enable programmatic access to code metrics functionality.

## 3. Purpose and Scope

### Purpose
Provide Go developers with quantitative metrics to evaluate package design quality according to Robert Martin's principles, helping teams identify packages that may need refactoring to improve maintainability.

### Scope
- Analyze any valid Go module source code
- Calculate core metrics (instability, abstractness, distance) for each package
- Generate reports in multiple formats
- Support both local and remote Git repositories
- Provide library functionality for integration with other Go projects

## 4. Product Features

### Core Features
- Calculate instability (I), abstractness (A), and distance from main sequence (D) metrics 
- Generate text, CSV and JSON reports
- Filter analysis by package patterns
- Support usage as a library in other Go projects with a well-defined API

## 5. User Stories

1. As a developer, I want to analyze a local Go module to understand its architectural health.
2. As a team lead, I want to generate reports of code metrics to identify problematic packages.
3. As a CI/CD engineer, I want to integrate metrics calculation into our pipeline to track architectural drift.
4. As an architect, I want to compare metrics between module versions to evaluate refactoring impact.
5. As a developer, I want to import Go Code Metrics as a library in my own project to programmatically analyze code and build custom tooling around the metrics.

## 6. Functional Requirements

### Metric Calculation
- **Instability (I)**: Calculate I = Ce/(Ca+Ce) where:
  - Ce: Number of packages this package depends on (efferent coupling)
  - Ca: Number of packages that depend on this package (afferent coupling)

- **Abstractness (A)**: Calculate A = Na/Nc where:
  - Na: Number of abstract types (interfaces) in the package
  - Nc: Total number of types in the package

- **Distance (D)**: Calculate D = |A + I - 1| to measure distance from the main sequence

### Input
- Accept path to local Go module directory only
- Allow package filtering with glob patterns

### Output
- Generate detailed metrics for each package
- Produce summary statistics for the entire module
- Support multiple output formats (text, CSV, JSON)
  
### Commands
```
gometrics [path/to/module] --format=<format>
```

### Library Usage
- Provide a clean, documented API for programmatic usage
- Support all core functionality available in the CLI version
- Allow customization of analysis parameters
- Enable integration with custom reporting and visualization systems
- Support streaming results for large-scale analysis
- Example library usage:
  ```go
  import "github.com/organization/gometrics/pkg/analyzer"
  
  metrics, err := analyzer.AnalyzeModule("path/to/module")
  if err != nil {
      // handle error
  }
  
  // Access metrics programmatically
  for pkg, data := range metrics.Packages {
      fmt.Printf("Package: %s, Instability: %.2f, Abstractness: %.2f, Distance: %.2f\n",
          pkg, data.Instability, data.Abstractness, data.Distance)
  }
  ```

## 7. Non-Functional Requirements

### Performance
- Complete analysis of medium-sized modules (50-100 packages) in under 15 seconds
- Handle large modules (500+ packages) without excessive memory usage

### Usability
- Provide clear, concise command-line interface with help documentation
- Generate readable, well-formatted reports
- Include color-coded output for quick status assessment
- Offer comprehensive API documentation for library users

### Compatibility
- Support Go versions 1.18 and higher
- Run on major operating systems (Linux, macOS, Windows)
- Ensure library API is stable and follows semantic versioning

## 8. Technical Requirements

### Implementation Details
- Parse Go source files using Go's AST package
- Build a directed graph of package dependencies
- Calculate metrics based on the graph and type information
- Use concurrent processing for performance optimization
- Implement clean separation of concerns to support library usage

### Installation
```
go install github.com/organization/gometrics@latest
```

### Library Integration
```
go get github.com/organization/gometrics
```

## 9. Constraints and Assumptions

### Constraints
- Analysis limited to Go language constructs
- May have reduced accuracy for packages using extensive reflection
- Cannot analyze binary-only packages

### Assumptions
- Standard Go module structure is used
- Analysis is limited to the module's own packages; external dependencies will not be analyzed
- Library API users will handle their own error management and logging

## 10. Acceptance Criteria

- Successfully analyzes official Go standard library packages
- Produces accurate metrics matching manual calculations for test cases
- Generates valid reports in all supported formats
- Completes analysis of the 100-package test module in under 15 seconds
- Library API provides all functionality of the CLI tool
- Documentation includes examples of library integration

## 11. Future Enhancements

- Integration with code editors and IDEs
- Trend analysis across multiple versions
- Integration with popular CI/CD platforms
- Expanded library API with additional analysis capabilities
- SDK for building plugins to extend functionality

## 12. Success Metrics

- Adoption by Go development teams
- Accuracy of metrics compared to manual calculation
- Performance on large codebases
- User satisfaction with generated reports
- Number of projects integrating the library functionality
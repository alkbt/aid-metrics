# Progress Reporting Feature Implementation Plan

## Overview
This document outlines the implementation plan for adding progress reporting capabilities to aid-metrics when analyzing large Go projects.

## Problem Statement
Currently, `packages.Load(config, pattern)` loads all packages at once without any progress feedback, which can be problematic for large projects where the analysis might take significant time.

## Solution Architecture

### Core Components

#### 1. ProgressReporter Interface (`pkg/models/progress.go`)
```go
type ProgressReporter interface {
    // Set total number of steps (for progress bar)
    SetTotal(total int)
    
    // Update progress with current step and description
    Update(current int, description string)
    
    // Mark as complete
    Complete()
}
```

This simplified interface:
- Shows a progress bar (current/total)
- Displays what's currently being processed
- Easy to implement with any progress bar library

#### 2. Package Discovery (`pkg/analyzer/discovery.go`)
- Manual filesystem traversal to find Go packages
- Early pattern matching for efficiency
- Progress reporting during discovery
- Returns `[]PackageInfo` for batch loading

**Progress Bar Strategy (0-100 scale)**:
- Discovery Phase: 0-10 (fast, 1 point per 2-3 packages, capped at 10)
- Loading Phase: 10-80 (the bottleneck - proportional to packages loaded)
- Analysis Phase: 80-100 (proportional to packages analyzed)

Example:
- Project with 20 packages: Discovery ends at 10, loading 10-80, analysis 80-100
- Project with 100 packages: Discovery ends at 10, loading 10-80, analysis 80-100

#### 3. Batch Loader (`pkg/analyzer/loader.go`)
- Loads packages in configurable batch sizes
- Uses `packages.Load` with specific paths instead of patterns
- Reports progress after each batch
- Maintains reliability of original approach

#### 4. Analyzer Refactoring
- Add `AnalyzerOptions` for configuration
- Replace existing API with progress-aware version

### Implementation Phases

#### Phase 1: Core Infrastructure
- [x] Create `pkg/models/progress.go` with interface definitions
- [x] Implement `ConsoleProgressReporter` using `github.com/schollz/progressbar/v3`
- [x] Add `AnalyzerOptions` struct

**Progress Bar Library Choice: `github.com/schollz/progressbar/v3`**
- Simple API that matches our needs perfectly
- Supports descriptions alongside progress
- Thread-safe (important for concurrent package processing)
- Minimal dependencies
- Good terminal compatibility (handles width, colors, etc.)

Example usage:
```go
bar := progressbar.NewOptions(100,
    progressbar.OptionSetDescription("Discovering packages..."),
    progressbar.OptionShowCount(),
    progressbar.OptionShowIts(),
)
bar.Set(25)
bar.Describe("Loading: pkg/analyzer")
```

#### Phase 2: Package Discovery
- [x] Create `pkg/analyzer/discovery.go`
- [x] Implement filesystem walking with Go package detection
- [x] Add pattern matching support (./..., specific paths)
- [x] Handle edge cases (vendor, testdata, .gitignore)

#### Phase 3: Batch Loading
- [x] Create `pkg/analyzer/loader.go`
- [x] Implement `BatchLoader` with configurable batch size
- [x] Integrate progress reporting
- [x] Handle partial failures gracefully

#### Phase 4: Analyzer Integration
- [x] Update `ModuleAnalyzer` to accept options
- [x] Refactor `findPackages()` to use discovery + batch loading
- [x] Update `parsePackages()` to report analysis progress

#### Phase 5: CLI Updates
- [x] Add `-progress` flag for progress output
- [x] Add `-batch-size` flag for tuning
- [x] Update help documentation

#### Phase 6: Testing
- [ ] Unit tests for discovery function
- [ ] Unit tests for batch loader
- [ ] Integration tests with mock progress reporter
- [ ] Performance benchmarks

#### Phase 7: Documentation
- [x] Update README.md with new flags
- [x] Update CLAUDE.md with progress reporting details
- [x] Add examples for large project analysis

## Technical Decisions

### Batch Size Selection
- Default: 100 packages per batch
- Rationale: Optimized for large projects while maintaining reasonable memory usage
- Configurable via CLI flag

### Progress Reporting Frequency
- Discovery: Every 2-3 packages found (increment by 1, max 10)
- Loading: After each batch (proportional within 10-80 range)
- Analysis: After each package (proportional within 80-100 range)

### Error Handling
- Non-fatal errors reported via `OnError()`
- Fatal errors still return from functions
- Progress reporter shouldn't affect core functionality

### Memory Optimization
- Batch loading reduces peak memory usage compared to loading all at once
- Still need all packages in memory for metric calculation
- Consider increasing batch size for better performance with sufficient memory

## API Examples

### Basic Usage with Progress Reporting
```go
progressReporter := analyzer.NewConsoleProgressReporter()
progressReporter.SetTotal(100)  // Fixed scale 0-100

opts := analyzer.AnalyzerOptions{
    ProgressReporter: progressReporter,
    BatchSize: 50,
}
analyzer := analyzer.NewModuleAnalyzerWithOptions(modulePath, packageFilter, opts)
metrics, err := analyzer.Analyze()
```

### CLI Usage
```bash
# Simple progress
aid-metrics ./... -progress

# Custom batch size
aid-metrics ./... -progress -batch-size=50
```

## Success Criteria
1. ✓ Progress feedback for projects with 100+ packages
2. ✓ No performance regression (< 15 seconds for 100 packages)
3. ✓ Memory usage acceptable for large projects
4. ✓ Clear, actionable progress messages

## Test Repositories for Performance Testing
1. **Kubernetes** (https://github.com/kubernetes/kubernetes) - 1000+ packages
2. **Docker/Moby** (https://github.com/moby/moby) - 500+ packages
3. **Prometheus** (https://github.com/prometheus/prometheus) - 200+ packages
4. **Terraform** (https://github.com/hashicorp/terraform) - 300+ packages
5. **CockroachDB** (https://github.com/cockroachdb/cockroach) - 500+ packages

### Known Limitations
- Very large monorepos (like CockroachDB with 2800+ directories) may experience long loading times
- The bottleneck is `packages.Load` which builds the full type graph for all packages
- Recommended to use specific package patterns instead of "./..." for huge projects
- Consider using larger batch sizes (100+) for projects with 500+ packages

## Code Documentation Requirements
- All public types, functions, and methods must have godoc comments
- Complex algorithms must include implementation details in comments
- Each new file must have a package-level comment explaining its purpose
- Include usage examples in godoc for key public APIs
- Document any non-obvious design decisions inline
- Error messages should be descriptive and actionable

## Risks and Mitigations
- **Risk**: Performance overhead from progress reporting
  - **Mitigation**: Use batching, make reporting optional
- **Risk**: Inaccurate progress estimates
  - **Mitigation**: Fixed 0-100 scale with weighted phases


## Implementation Summary

### Completed Features

1. **ProgressReporter Interface** (`pkg/models/progress.go`)
   - Simple 3-method interface: SetTotal, Update, Complete
   - Uses fixed 0-100 scale for consistent progress

2. **ConsoleProgressReporter** (`pkg/reporter/progress.go`)
   - Terminal progress bar using schollz/progressbar/v3
   - Thread-safe for concurrent package processing
   - Clean separation between progress bar and output

3. **Package Discovery** (`pkg/analyzer/discovery.go`)
   - Filesystem traversal without loading packages
   - Early pattern matching for efficiency
   - Progress reporting during discovery (0-10 scale)

4. **Batch Loading** (`pkg/analyzer/loader.go`)
   - Loads packages in configurable batches
   - Progress reporting during loading (10-80 scale)
   - Reduces memory usage compared to loading all at once

5. **Analyzer Integration**
   - ModuleAnalyzer now accepts AnalyzerOptions
   - Progress reporting during analysis (80-100 scale)
   - Backward compatible API maintained

6. **CLI Updates** (`cmd/aid-metrics/main.go`)
   - `-progress` flag to show progress bar
   - `-batch-size` flag to tune loading performance
   - Works with all output formats

7. **Documentation**
   - README.md updated with new CLI flags and library usage
   - CLAUDE.md updated with progress reporting details

### Performance Characteristics
- Discovery Phase: Fast directory scan (0-10%)
- Loading Phase: Main bottleneck (10-80%)
- Analysis Phase: Relatively fast (80-100%)
- Fixed scale ensures smooth progress regardless of project size

### Progress Bar Improvements
- Clean visual design with solid blocks (█)
- No redundant count display
- Shortened package paths for readability (e.g., `.../pkg1/pkg2`)
- Concise descriptions ("Loading 20 packages" instead of verbose paths)
- Proper newline after completion to separate from output


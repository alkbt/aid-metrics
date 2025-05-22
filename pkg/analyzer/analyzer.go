package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/alkbt/aid-metrics/pkg/models"
	"golang.org/x/tools/go/packages"
)

// ModuleAnalyzer performs analysis on a Go module
type ModuleAnalyzer struct {
	modulePath     string
	packageFilter  string
	dependencies   map[string][]string // Package -> dependencies
	reverseDepends map[string][]string // Package -> packages that depend on it
	abstractTypes  map[string]int      // Package -> number of interfaces
	totalTypes     map[string]int      // Package -> number of concrete types

	// Cache for the module path from go.mod
	moduleName string
}

// NewModuleAnalyzer creates a new ModuleAnalyzer
func NewModuleAnalyzer(modulePath string, packageFilter string) *ModuleAnalyzer {
	analyzer := &ModuleAnalyzer{
		modulePath:     modulePath,
		packageFilter:  packageFilter,
		dependencies:   make(map[string][]string),
		reverseDepends: make(map[string][]string),
		abstractTypes:  make(map[string]int),
		totalTypes:     make(map[string]int),
		moduleName:     readModuleName(modulePath),
	}

	return analyzer
}

// AnalyzeModule analyzes a Go module and returns metrics
func AnalyzeModule(modulePath string, packageFilter string) (*models.ModuleMetrics, error) {
	analyzer := NewModuleAnalyzer(modulePath, packageFilter)
	return analyzer.Analyze()
}

// Analyze performs the full analysis
func (a *ModuleAnalyzer) Analyze() (*models.ModuleMetrics, error) {
	// Step 1: Find all Go packages in the module
	pkgs, err := a.findPackages()
	if err != nil {
		return nil, fmt.Errorf("failed to find packages: %w", err)
	}

	// Step 2: Parse package dependencies and count types
	err = a.parsePackages(pkgs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse packages: %w", err)
	}

	// Step 3: Calculate metrics
	metrics := a.calculateMetrics()
	return metrics, nil
}

// findPackages finds all Go packages in the module
func (a *ModuleAnalyzer) findPackages() ([]*packages.Package, error) {
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes,
		Dir:  a.modulePath,
	}

	pattern := "./..."
	if a.packageFilter != "" {
		pattern = a.packageFilter
	}

	pkgs, err := packages.Load(config, pattern)
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}

// Define a struct to hold the package analysis results
type packageAnalysisResult struct {
	packageID       string
	dependencies    []string
	abstractCount   int
	totalTypesCount int
	err             error
}

// parsePackages parses all Go packages to extract dependencies and count types
func (a *ModuleAnalyzer) parsePackages(pkgs []*packages.Package) error {
	// Create a worker pool with a reasonable number of workers
	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8 // Cap at 8 workers to avoid excessive goroutines
	}

	// Create channels for input jobs and results
	jobs := make(chan *packages.Package, len(pkgs))
	results := make(chan packageAnalysisResult, len(pkgs))

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pkg := range jobs {
				// Process each package and send results through the channel
				result := a.analyzePackage(pkg)
				results <- result
			}
		}()
	}

	// Send all packages to be processed
	for _, pkg := range pkgs {
		jobs <- pkg
	}
	close(jobs) // No more jobs to send

	// Create a goroutine to close the results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results in the main goroutine
	for result := range results {
		if result.err != nil {
			return result.err
		}

		// Store the analysis results in the maps
		a.dependencies[result.packageID] = result.dependencies

		// Update reverse dependencies
		for _, dep := range result.dependencies {
			a.reverseDepends[dep] = append(a.reverseDepends[dep], result.packageID)
		}

		a.abstractTypes[result.packageID] = result.abstractCount
		a.totalTypes[result.packageID] = result.totalTypesCount
	}

	return nil
}

// analyzePackage analyzes a single package but doesn't modify shared maps
// Instead, it returns the analysis results to be processed by the main goroutine
func (a *ModuleAnalyzer) analyzePackage(pkg *packages.Package) packageAnalysisResult {
	result := packageAnalysisResult{
		packageID: pkg.ID,
	}

	// Skip standard library packages
	if isStandardLibraryPackage(pkg.ID, a.moduleName) || strings.HasPrefix(pkg.ID, "vendor/") {
		// Return empty result without error for skipped packages
		return result
	}

	// Get dependencies
	deps := make([]string, 0)
	for _, imp := range pkg.Imports {
		// Skip standard library packages
		if isStandardLibraryPackage(imp.ID, a.moduleName) || strings.HasPrefix(imp.ID, "vendor/") {
			continue
		}
		deps = append(deps, imp.ID)
	}
	result.dependencies = deps

	// Parse the package files to count abstract and concrete types
	var abstractCount, concreteCount int
	var funcCount int
	fset := token.NewFileSet()

	for _, filePath := range pkg.GoFiles {
		// Parse the file
		file, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
		if err != nil {
			result.err = fmt.Errorf("failed to parse file %s: %w", filePath, err)
			return result
		}

		// Count types and functions
		ast.Inspect(file, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.TypeSpec:
				if _, ok := t.Type.(*ast.InterfaceType); ok {
					abstractCount++
				} else if _, ok := t.Type.(*ast.StructType); ok {
					// Only count structs as concrete types
					concreteCount++
				}
				// Other types (like type aliases) are not counted
			case *ast.FuncDecl:
				// Count only standalone functions (not methods)
				if t.Recv == nil {
					funcCount++
				}
			}
			return true
		})
	}

	result.abstractCount = abstractCount
	// Include only structs and standalone functions as concrete types
	result.totalTypesCount = abstractCount + concreteCount + funcCount

	return result
}

// isStandardLibraryPackage checks if a package is part of the Go standard library
// It uses a more reliable method than just checking for dots in the package path
func isStandardLibraryPackage(pkgID, mainModulePath string) bool {
	// If we have the main module path, use it to determine if the package is a standard library package
	if mainModulePath != "" {
		// Standard library packages don't have a domain name as the first element
		// and they're not part of the main module
		parts := strings.Split(pkgID, "/")
		if len(parts) > 0 {
			// If the package has no dots in first path element and doesn't match the main module path
			if !strings.Contains(parts[0], ".") && !strings.HasPrefix(pkgID, mainModulePath) {
				// Standard library packages have no dots in the first path element
				return true
			}
		}

		// If the package is explicitly part of the main module, it's not a standard library package
		if strings.HasPrefix(pkgID, mainModulePath) {
			return false
		}
	}

	// If we couldn't determine based on the module path, fall back to the original behavior
	// This is less reliable but maintains backward compatibility
	if !strings.Contains(pkgID, ".") {
		return true
	}

	return false
}

// readModuleName reads the module name from the go.mod file
func readModuleName(modulePath string) string {
	goModPath := filepath.Join(modulePath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		// If we can't read go.mod, return empty string
		return ""
	}

	// Simple parsing of the module line from go.mod
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}

	// If no module declaration found, return empty string
	return ""
}

// calculateMetrics calculates metrics for all packages
func (a *ModuleAnalyzer) calculateMetrics() *models.ModuleMetrics {
	metrics := &models.ModuleMetrics{
		Path:     a.modulePath,
		Packages: make(map[string]models.PackageMetrics),
	}

	for pkg := range a.dependencies {
		ca := len(a.reverseDepends[pkg])
		ce := len(a.dependencies[pkg])
		na := a.abstractTypes[pkg]
		nc := a.totalTypes[pkg]

		// Calculate instability (I)
		instability := 0.0
		if ca+ce > 0 {
			instability = float64(ce) / float64(ca+ce)
		}

		// Calculate abstractness (A)
		abstractness := 0.0
		if nc > 0 {
			abstractness = float64(na) / float64(nc)
		}

		// Calculate distance from main sequence (D)
		distance := math.Abs(abstractness + instability - 1.0)

		metrics.Packages[pkg] = models.PackageMetrics{
			Name:         a.getRelativePackagePath(pkg),
			Ca:           ca,
			Ce:           ce,
			Na:           na,
			Nc:           nc,
			Instability:  instability,
			Abstractness: abstractness,
			Distance:     distance,
		}
	}

	return metrics
}

// getRelativePackagePath extracts the import path relative to the module name
func (a *ModuleAnalyzer) getRelativePackagePath(importPath string) string {
	// Use the cached module path if available
	if a.moduleName != "" {
		// If the import path starts with the module path, extract the relative part
		if strings.HasPrefix(importPath, a.moduleName) {
			// Special case for the root package
			if importPath == a.moduleName {
				// Get the last segment of the module path as the package name
				parts := strings.Split(a.moduleName, "/")
				return parts[len(parts)-1]
			}

			// For other packages, return the path relative to the module
			relPath := strings.TrimPrefix(importPath, a.moduleName+"/")
			if relPath == "" {
				// This is the root package of the module
				parts := strings.Split(a.moduleName, "/")
				return parts[len(parts)-1]
			}
			return relPath
		}
	}

	// Clean up package ID if it includes module metadata
	// E.g., "path/to/pkg [path/to/module]" -> "path/to/pkg"
	if pkgParts := strings.Split(importPath, " "); len(pkgParts) > 1 {
		importPath = pkgParts[0]
	}

	// Simple approach: use the full import path without assumptions
	// This preserves the package structure without making assumptions about
	// which parts are relevant
	parts := strings.Split(importPath, "/")

	// If it's a short path (1-2 segments), use it as is
	if len(parts) <= 2 {
		return importPath
	}

	// For longer paths, use the last two segments as a reasonable compromise
	// This is a simple heuristic that works well in practice without making
	// assumptions about versioning or domain structures
	return strings.Join(parts[len(parts)-2:], "/")
}

// getPackageName extracts the final package name from a full import path
// Kept for backwards compatibility
func getPackageName(importPath string) string {
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}

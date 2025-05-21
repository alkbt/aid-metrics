package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
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
	concreteTypes  map[string]int      // Package -> number of concrete types
}

// NewModuleAnalyzer creates a new ModuleAnalyzer
func NewModuleAnalyzer(modulePath string, packageFilter string) *ModuleAnalyzer {
	return &ModuleAnalyzer{
		modulePath:     modulePath,
		packageFilter:  packageFilter,
		dependencies:   make(map[string][]string),
		reverseDepends: make(map[string][]string),
		abstractTypes:  make(map[string]int),
		concreteTypes:  make(map[string]int),
	}
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

// parsePackages parses all Go packages to extract dependencies and count types
func (a *ModuleAnalyzer) parsePackages(pkgs []*packages.Package) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(pkgs))

	for _, pkg := range pkgs {
		wg.Add(1)
		go func(p *packages.Package) {
			defer wg.Done()
			if err := a.parsePackage(p); err != nil {
				errChan <- err
			}
		}(pkg)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// parsePackage parses a single package to extract dependencies and count types
func (a *ModuleAnalyzer) parsePackage(pkg *packages.Package) error {
	// Skip standard library packages
	if !strings.Contains(pkg.ID, ".") || strings.HasPrefix(pkg.ID, "vendor/") {
		return nil
	}

	// Get dependencies
	deps := make([]string, 0)
	for _, imp := range pkg.Imports {
		// Skip standard library packages
		if !strings.Contains(imp.ID, ".") || strings.HasPrefix(imp.ID, "vendor/") {
			continue
		}
		deps = append(deps, imp.ID)

		// Update reverse dependencies
		a.reverseDepends[imp.ID] = append(a.reverseDepends[imp.ID], pkg.ID)
	}
	a.dependencies[pkg.ID] = deps

	// Parse the package files to count abstract and concrete types
	var abstractCount, concreteCount int
	var funcCount int
	fset := token.NewFileSet()

	for _, filePath := range pkg.GoFiles {
		// Parse the file
		file, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", filePath, err)
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

	a.abstractTypes[pkg.ID] = abstractCount
	// Include only structs and standalone functions as concrete types
	a.concreteTypes[pkg.ID] = abstractCount + concreteCount + funcCount

	return nil
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
		nc := a.concreteTypes[pkg]

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
			Name:         getPackageName(pkg),
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

// getPackageName extracts the final package name from a full import path
func getPackageName(importPath string) string {
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}

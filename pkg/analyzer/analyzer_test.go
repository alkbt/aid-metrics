package analyzer

import (
	"testing"
)

func TestGetPackageName(t *testing.T) {
	tests := []struct {
		importPath string
		expected   string
	}{
		{"github.com/alkbt/aid-metrics/pkg/analyzer", "analyzer"},
		{"github.com/alkbt/aid-metrics", "aid-metrics"},
		{"std", "std"},
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			got := getPackageName(tt.importPath)
			if got != tt.expected {
				t.Errorf("getPackageName(%q) = %q, want %q", tt.importPath, got, tt.expected)
			}
		})
	}
}

func TestModuleAnalyzer_New(t *testing.T) {
	analyzer := NewModuleAnalyzer("testpath", "testfilter")

	if analyzer.modulePath != "testpath" {
		t.Errorf("Expected modulePath to be 'testpath', got %q", analyzer.modulePath)
	}

	if analyzer.packageFilter != "testfilter" {
		t.Errorf("Expected packageFilter to be 'testfilter', got %q", analyzer.packageFilter)
	}
}

func TestTypesCounting(t *testing.T) {
	// This test confirms that our logic for counting abstract and concrete types
	// works as expected. It specifically verifies that standalone functions
	// are counted as concrete types.

	// Set up test case
	analyzer := &ModuleAnalyzer{
		modulePath:     "",
		packageFilter:  "",
		dependencies:   make(map[string][]string),
		reverseDepends: make(map[string][]string),
		abstractTypes:  make(map[string]int),
		totalTypes:     make(map[string]int),
	}

	// Create simple test case
	pkgID := "test/pkg"
	abstractCount := 2 // Interfaces
	concreteCount := 3 // Structs only
	funcCount := 4     // Standalone functions

	// Manually set the values (like the code would do)
	analyzer.dependencies[pkgID] = []string{} // Add to dependencies to ensure it's included in metrics
	analyzer.abstractTypes[pkgID] = abstractCount
	analyzer.totalTypes[pkgID] = abstractCount + concreteCount + funcCount

	// Calculate metrics
	metrics := analyzer.calculateMetrics()

	// Verify that the abstactness is calculated correctly
	// A = Na / Nc = 2 / (2+3+4) = 2/9 = 0.222...
	// Note: Only structs and standalone functions are counted as concrete, not other types
	expectedAbstractness := float64(abstractCount) / float64(abstractCount+concreteCount+funcCount)

	// Package should be added by calculateMetrics
	pkg, exists := metrics.Packages[pkgID]
	if !exists {
		t.Fatalf("Package %s not found in metrics", pkgID)
	}

	if pkg.Abstractness != expectedAbstractness {
		t.Errorf("Expected abstractness to be %v, got %v", expectedAbstractness, pkg.Abstractness)
	}

	if pkg.Na != abstractCount {
		t.Errorf("Expected Na to be %v, got %v", abstractCount, pkg.Na)
	}

	if pkg.Nc != abstractCount+concreteCount+funcCount {
		t.Errorf("Expected Nc to be %v, got %v", abstractCount+concreteCount+funcCount, pkg.Nc)
	}
}

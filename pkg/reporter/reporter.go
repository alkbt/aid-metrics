package reporter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/alkbt/aid-metrics/pkg/models"
)

// FormatType represents the format of the report
type FormatType string

const (
	FormatText FormatType = "text"
	FormatCSV  FormatType = "csv"
	FormatJSON FormatType = "json"
)

// Reporter generates reports for module metrics
type Reporter struct {
	metrics *models.ModuleMetrics
	format  FormatType
}

// NewReporter creates a new Reporter
func NewReporter(metrics *models.ModuleMetrics, format FormatType) *Reporter {
	return &Reporter{
		metrics: metrics,
		format:  format,
	}
}

// Format returns the current format
func (r *Reporter) Format() FormatType {
	return r.format
}

// Generate generates a report in the specified format
func (r *Reporter) Generate(w io.Writer) error {
	switch r.format {
	case FormatText:
		return r.generateTextReport(w)
	case FormatCSV:
		return r.generateCSVReport(w)
	case FormatJSON:
		return r.generateJSONReport(w)
	default:
		return fmt.Errorf("unsupported format: %s", r.format)
	}
}

// generateTextReport generates a text report
func (r *Reporter) generateTextReport(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintf(tw, "MODULE: %s\n\n", r.metrics.Path)
	fmt.Fprintln(tw, "PACKAGE\tCa\tCe\tI\tNa\tNc\tA\tD")
	fmt.Fprintln(tw, "-------\t--\t--\t-\t--\t--\t-\t-")

	// Sort packages by name for consistent output
	packageNames := make([]string, 0, len(r.metrics.Packages))
	for pkgName := range r.metrics.Packages {
		packageNames = append(packageNames, pkgName)
	}
	sort.Strings(packageNames)

	for _, pkgName := range packageNames {
		pkg := r.metrics.Packages[pkgName]
		fmt.Fprintf(tw, "%s\t%d\t%d\t%.2f\t%d\t%d\t%.2f\t%.2f\n",
			pkg.Name, pkg.Ca, pkg.Ce, pkg.Instability, pkg.Na, pkg.Nc, pkg.Abstractness, pkg.Distance)
	}

	return nil
}

// generateCSVReport generates a CSV report
func (r *Reporter) generateCSVReport(w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	if err := csvWriter.Write([]string{"Package", "Ca", "Ce", "I", "Na", "Nc", "A", "D"}); err != nil {
		return err
	}

	// Sort packages by name for consistent output
	packageNames := make([]string, 0, len(r.metrics.Packages))
	for pkgName := range r.metrics.Packages {
		packageNames = append(packageNames, pkgName)
	}
	sort.Strings(packageNames)

	// Write data
	for _, pkgName := range packageNames {
		pkg := r.metrics.Packages[pkgName]
		record := []string{
			pkg.Name,
			strconv.Itoa(pkg.Ca),
			strconv.Itoa(pkg.Ce),
			fmt.Sprintf("%.2f", pkg.Instability),
			strconv.Itoa(pkg.Na),
			strconv.Itoa(pkg.Nc),
			fmt.Sprintf("%.2f", pkg.Abstractness),
			fmt.Sprintf("%.2f", pkg.Distance),
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// generateJSONReport generates a JSON report
func (r *Reporter) generateJSONReport(w io.Writer) error {
	// Create a simplified structure for JSON output
	type jsonPackage struct {
		Name         string  `json:"name"`
		Ca           int     `json:"ca"`
		Ce           int     `json:"ce"`
		Instability  float64 `json:"instability"`
		Na           int     `json:"na"`
		Nc           int     `json:"nc"`
		Abstractness float64 `json:"abstractness"`
		Distance     float64 `json:"distance"`
	}

	type jsonReport struct {
		Module   string        `json:"module"`
		Packages []jsonPackage `json:"packages"`
	}

	// Convert metrics to JSON format
	report := jsonReport{
		Module:   r.metrics.Path,
		Packages: make([]jsonPackage, 0, len(r.metrics.Packages)),
	}

	for _, pkg := range r.metrics.Packages {
		report.Packages = append(report.Packages, jsonPackage{
			Name:         pkg.Name,
			Ca:           pkg.Ca,
			Ce:           pkg.Ce,
			Instability:  pkg.Instability,
			Na:           pkg.Na,
			Nc:           pkg.Nc,
			Abstractness: pkg.Abstractness,
			Distance:     pkg.Distance,
		})
	}

	// Sort packages by name for consistent output
	sort.Slice(report.Packages, func(i, j int) bool {
		return report.Packages[i].Name < report.Packages[j].Name
	})

	// Encode JSON
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

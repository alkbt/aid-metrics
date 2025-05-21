package models

// PackageMetrics represents the metrics for a specific package
type PackageMetrics struct {
	Name         string  // Package name
	Ca           int     // Afferent coupling - packages that depend on this package
	Ce           int     // Efferent coupling - packages this package depends on
	Na           int     // Number of abstract types (interfaces)
	Nc           int     // Total number of types
	Instability  float64 // I = Ce/(Ca+Ce)
	Abstractness float64 // A = Na/Nc
	Distance     float64 // D = |A + I - 1|
}

// ModuleMetrics represents the metrics for an entire module
type ModuleMetrics struct {
	Path     string                    // Module path
	Packages map[string]PackageMetrics // Map of package metrics by package path
}

// Package analyzer provides functionality for analyzing Go modules and calculating design metrics.
// This file implements package discovery functionality that finds Go packages without loading them.
package analyzer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PackageInfo contains basic information about a discovered Go package.
// This is used during the discovery phase before packages are fully loaded.
type PackageInfo struct {
	// ImportPath is the full import path of the package (e.g., "github.com/user/project/pkg/analyzer")
	ImportPath string
	
	// Dir is the absolute filesystem path to the package directory
	Dir string
	
	// HasGoFiles indicates whether the directory contains any .go files
	HasGoFiles bool
}

// discoverPackages walks the filesystem to find all Go packages matching the given pattern.
// This is the first phase of the analysis process and provides quick package discovery
// without the overhead of loading package dependencies and type information.
//
// The pattern parameter supports standard Go package patterns:
//   - "./..." to find all packages recursively
//   - "." for just the current package
//   - specific package paths
//
// Progress is reported through the progressFunc callback, which is called for each
// package discovered. The discovery phase uses progress values 0-10 on the fixed
// 0-100 scale, incrementing by 1 for every 2-3 packages found (capped at 10).
func discoverPackages(modulePath, moduleName, pattern string, progressFunc func(found int)) ([]PackageInfo, error) {
	var packages []PackageInfo
	packagesFound := 0
	lastProgress := 0

	// Convert pattern to filesystem path
	searchPath := modulePath
	if pattern != "" && pattern != "./..." && pattern != "." {
		searchPath = filepath.Join(modulePath, pattern)
	}

	// Walk the filesystem
	err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		// Skip non-directories
		if !d.IsDir() {
			return nil
		}

		// Skip common non-package directories
		dirName := d.Name()
		if dirName == ".git" || dirName == ".idea" || dirName == "node_modules" ||
			dirName == "vendor" || dirName == "testdata" || strings.HasPrefix(dirName, ".") {
			return fs.SkipDir
		}

		// Check if directory contains Go files
		hasGoFiles := false
		entries, err := fs.ReadDir(fs.FS(dirFS{modulePath}), strings.TrimPrefix(path, modulePath+"/"))
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") &&
					!strings.HasSuffix(entry.Name(), "_test.go") {
					hasGoFiles = true
					break
				}
			}
		}

		if hasGoFiles {
			// Calculate import path
			relPath, err := filepath.Rel(modulePath, path)
			if err != nil {
				return nil
			}

			importPath := moduleName
			if relPath != "." {
				importPath = filepath.Join(moduleName, filepath.ToSlash(relPath))
			}

			// Check if this matches our pattern
			if matchesPattern(importPath, moduleName, pattern) {
				packages = append(packages, PackageInfo{
					ImportPath: importPath,
					Dir:        path,
					HasGoFiles: true,
				})

				packagesFound++
				
				// Update progress (0-10 range, 1 point per 2-3 packages)
				progress := packagesFound / 3
				if progress > 10 {
					progress = 10
				}
				if progress > lastProgress && progressFunc != nil {
					progressFunc(packagesFound)
					lastProgress = progress
				}
			}
		}

		return nil
	})

	return packages, err
}

// matchesPattern checks if an import path matches the given pattern
func matchesPattern(importPath, moduleName, pattern string) bool {
	// Empty pattern or "./..." matches everything in the module
	if pattern == "" || pattern == "./..." {
		return strings.HasPrefix(importPath, moduleName)
	}

	// "." matches only the root package
	if pattern == "." {
		return importPath == moduleName
	}

	// For other patterns, check if it's a prefix match
	fullPattern := filepath.Join(moduleName, pattern)
	return strings.HasPrefix(importPath, fullPattern)
}

// dirFS implements fs.FS for a directory
type dirFS struct {
	root string
}

func (d dirFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.Join(d.root, name))
}
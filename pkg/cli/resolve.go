package cli

import (
	"bytes"
	"fmt"
	"os"
	filepath "path/filepath"
)

// isGzipFile checks the first bytes of a file for the gzip magic number.
func isGzipFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	// gzip file signature (RFC 1952).
	gzipMagic := []byte{0x1F, 0x8B, 0x08}
	buf := make([]byte, 3)
	n, _ := f.Read(buf)
	return n == 3 && bytes.Equal(buf, gzipMagic)
}

// resolveChartRoot finds the chart root for a given path.
// Archives are returned as-is; other paths walk up to find Chart.yaml.
func resolveChartRoot(filePath string) (string, error) {
	dir, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path for %s: %w", filePath, err)
	}

	if isGzipFile(dir) {
		return dir, nil
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "Chart.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no Chart.yaml found in any parent of %s", filePath)
		}
		dir = parent
	}
}

// resolveUniqueCharts resolves each path to its chart root and deduplicates.
func resolveUniqueCharts(paths []string) ([]string, error) {
	seen := map[string]struct{}{}
	var roots []string
	for _, p := range paths {
		root, err := resolveChartRoot(p)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[root]; !ok {
			seen[root] = struct{}{}
			roots = append(roots, root)
		}
	}
	return roots, nil
}

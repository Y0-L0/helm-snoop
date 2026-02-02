package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/y0-l0/helm-snoop/pkg/path"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file1.json> <file2.json>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s old.json new.json\n", os.Args[0])
		os.Exit(1)
	}

	file1 := os.Args[1]
	file2 := os.Args[2]

	results1, err := loadGoldenFile(file1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file1, err)
		os.Exit(1)
	}

	results2, err := loadGoldenFile(file2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file2, err)
		os.Exit(1)
	}

	fmt.Printf("Comparing:\n  [1] %s\n  [2] %s\n\n", file1, file2)

	compareResults(results1, results2)
}

func loadGoldenFile(path string) (*snooper.ResultsJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var results snooper.ResultsJSON
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

func compareResults(r1, r2 *snooper.ResultsJSON) {
	fmt.Println("=== Referenced Paths ===")
	comparePaths(r1.Referenced, r2.Referenced)

	fmt.Println("\n=== Unused Paths ===")
	comparePaths(r1.Unused, r2.Unused)

	fmt.Println("\n=== Undefined Paths ===")
	comparePaths(r1.Undefined, r2.Undefined)
}

func comparePaths(paths1, paths2 path.PathsJSON) {
	// Convert to maps for easy lookup
	map1 := make(map[string]path.PathJSON)
	map2 := make(map[string]path.PathJSON)

	for _, p := range paths1 {
		map1[p.ID] = p
	}
	for _, p := range paths2 {
		map2[p.ID] = p
	}

	// Find paths only in file1
	onlyIn1 := []path.PathJSON{}
	for id, p := range map1 {
		if _, found := map2[id]; !found {
			onlyIn1 = append(onlyIn1, p)
		}
	}
	sort.Slice(onlyIn1, func(i, j int) bool {
		return onlyIn1[i].ID < onlyIn1[j].ID
	})

	// Find paths only in file2
	onlyIn2 := []path.PathJSON{}
	for id, p := range map2 {
		if _, found := map1[id]; !found {
			onlyIn2 = append(onlyIn2, p)
		}
	}
	sort.Slice(onlyIn2, func(i, j int) bool {
		return onlyIn2[i].ID < onlyIn2[j].ID
	})

	// Find common paths (and check for kind differences)
	common := []path.PathJSON{}
	kindDiffs := []struct {
		id     string
		kinds1 string
		kinds2 string
	}{}
	for id, p1 := range map1 {
		if p2, found := map2[id]; found {
			common = append(common, p1)
			if p1.Kinds != p2.Kinds {
				kindDiffs = append(kindDiffs, struct {
					id     string
					kinds1 string
					kinds2 string
				}{id, p1.Kinds, p2.Kinds})
			}
		}
	}
	sort.Slice(common, func(i, j int) bool {
		return common[i].ID < common[j].ID
	})

	if len(onlyIn1) > 0 {
		fmt.Printf("\nOnly in [1] (%d paths):\n", len(onlyIn1))
		for _, p := range onlyIn1 {
			fmt.Printf("  - %s (kinds: %s)\n", p.ID, p.Kinds)
		}
	}

	if len(onlyIn2) > 0 {
		fmt.Printf("\nOnly in [2] (%d paths):\n", len(onlyIn2))
		for _, p := range onlyIn2 {
			fmt.Printf("  + %s (kinds: %s)\n", p.ID, p.Kinds)
		}
	}

	if len(kindDiffs) > 0 {
		fmt.Printf("\nPaths with different kinds (%d):\n", len(kindDiffs))
		for _, diff := range kindDiffs {
			fmt.Printf("  ~ %s\n", diff.id)
			fmt.Printf("      [1]: %s\n", diff.kinds1)
			fmt.Printf("      [2]: %s\n", diff.kinds2)
		}
	}

	if len(common) > 0 {
		fmt.Printf("\nCommon paths: %d\n", len(common))
	}

	if len(onlyIn1) == 0 && len(onlyIn2) == 0 && len(kindDiffs) == 0 {
		fmt.Printf("✓ No differences (all %d paths match)\n", len(common))
	}
}

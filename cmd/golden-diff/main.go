// Package main implements a CLI tool for comparing golden test output files.
//
// It compares current .golden.json files against their committed versions in git,
// similar to how `git diff` shows changes against the last commit.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/y0-l0/helm-snoop/pkg/snooper"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	commit := fs.String("commit", "HEAD", "git commit to compare against")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [file.golden.json ...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Compares current .golden.json files against their versions in git.\n")
		fmt.Fprintf(os.Stderr, "With no files specified, diffs all modified .golden.json files.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}
	_ = fs.Parse(os.Args[1:])

	files := fs.Args()
	if len(files) == 0 {
		var err error
		files, err = findModifiedGoldenFiles(*commit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding modified golden files: %v\n", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Println("No modified .golden.json files found.")
			return
		}
	}

	for i, file := range files {
		if i > 0 {
			fmt.Println("\n" + strings.Repeat("─", 60))
		}

		oldResults, err := loadGoldenFileFromGit(*commit, file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot load %s from %s: %v (treating as empty)\n", file, *commit, err)
			oldResults = &snooper.ResultJSON{}
		}

		newResults, err := loadGoldenFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", file, err)
			os.Exit(1)
		}

		fmt.Printf("Comparing %s: %s vs working tree\n\n", file, *commit)
		compareResults(oldResults, newResults)
	}
}

// findModifiedGoldenFiles returns .golden.json files that differ from the given commit.
func findModifiedGoldenFiles(commit string) ([]string, error) {
	// Get both staged and unstaged changes.
	out, err := exec.CommandContext(context.Background(), "git", "diff", "--name-only", commit, "--", "*.golden.json").
		Output()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}

	var files []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// loadGoldenFileFromGit loads a golden file from a specific git commit.
func loadGoldenFileFromGit(commit, path string) (*snooper.ResultJSON, error) {
	// Convert to repo-relative path for git show.
	relPath, err := gitRelativePath(path)
	if err != nil {
		return nil, err
	}

	out, err := exec.CommandContext(context.Background(), "git", "show", commit+":"+relPath).Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s:%s: %w", commit, relPath, err)
	}

	var results snooper.ResultJSON
	if err := json.Unmarshal(out, &results); err != nil {
		return nil, err
	}
	return &results, nil
}

// gitRelativePath converts a file path to be relative to the git repo root.
func gitRelativePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	out, err := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	root := strings.TrimSpace(string(out))

	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return "", err
	}
	return rel, nil
}

func loadGoldenFile(path string) (*snooper.ResultJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var results snooper.ResultJSON
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

func compareResults(r1, r2 *snooper.ResultJSON) {
	fmt.Println("=== Referenced Paths ===")
	comparePaths(r1.Referenced, r2.Referenced)

	fmt.Println("\n=== Unused Paths ===")
	comparePaths(r1.Unused, r2.Unused)

	fmt.Println("\n=== Undefined Paths ===")
	comparePaths(r1.Undefined, r2.Undefined)
}

//nolint:gocognit // TODO: refactor to reduce cognitive complexity
func comparePaths(paths1, paths2 vpath.PathsJSON) {
	// Convert to maps for easy lookup
	map1 := make(map[string]vpath.PathJSON)
	map2 := make(map[string]vpath.PathJSON)

	for _, p := range paths1 {
		map1[p.ID] = p
	}
	for _, p := range paths2 {
		map2[p.ID] = p
	}

	// Find paths only in file1
	onlyIn1 := []vpath.PathJSON{}
	for id, p := range map1 {
		if _, found := map2[id]; !found {
			onlyIn1 = append(onlyIn1, p)
		}
	}
	sort.Slice(onlyIn1, func(i, j int) bool {
		return onlyIn1[i].ID < onlyIn1[j].ID
	})

	// Find paths only in file2
	onlyIn2 := []vpath.PathJSON{}
	for id, p := range map2 {
		if _, found := map1[id]; !found {
			onlyIn2 = append(onlyIn2, p)
		}
	}
	sort.Slice(onlyIn2, func(i, j int) bool {
		return onlyIn2[i].ID < onlyIn2[j].ID
	})

	// Find common paths (and check for kind differences)
	common := []vpath.PathJSON{}
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

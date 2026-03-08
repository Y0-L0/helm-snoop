package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/snooper"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// Options holds CLI-level settings that get merged with the config file.
type Options struct {
	ConfigPath  string      // --config flag (empty = auto-discover)
	NoConfig    bool        // --no-config flag
	Ignore      vpath.Paths // CLI -i flags
	ValuesFiles []string    // CLI -f flags
}

// Resolve loads the config file (if any), merges settings, and returns
// a ChartSettings for each chart path.
func Resolve(chartPaths []string, opts Options) ([]snooper.ChartSettings, error) {
	var cfg *fileConfig
	var configDir string

	if !opts.NoConfig {
		data, cfgPath, err := loadConfigFile(opts.ConfigPath)
		if err != nil {
			return nil, err
		}
		if data != nil {
			cfg, err = parse(data)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", cfgPath, err)
			}
			configDir = filepath.Dir(cfgPath)
		}
	}

	// Build merged global: config global + CLI globals.
	global := mergedGlobal(cfg, opts, configDir)

	charts := make([]snooper.ChartSettings, len(chartPaths))
	for i, chartPath := range chartPaths {
		charts[i] = resolveChart(chartPath, global, cfg, configDir)
	}
	return charts, nil
}

// mergedGlobal combines config file global settings with CLI flags.
type globalSettings struct {
	ignore      vpath.Paths
	valuesFiles []string
	extraValues map[string]any
}

func mergedGlobal(cfg *fileConfig, opts Options, configDir string) globalSettings {
	var g globalSettings

	if cfg != nil {
		g.ignore = parsePaths(cfg.Global.Ignore)
		g.valuesFiles = resolvePaths(cfg.Global.ValuesFiles, configDir)
		g.extraValues = cfg.Global.ExtraValues
	}

	// Append CLI flags.
	g.ignore = append(g.ignore, opts.Ignore...)
	g.valuesFiles = append(g.valuesFiles, opts.ValuesFiles...)

	return g
}

func resolveChart(
	chartPath string,
	global globalSettings,
	cfg *fileConfig,
	configDir string,
) snooper.ChartSettings {
	cs := snooper.ChartSettings{
		Path:        chartPath,
		Ignore:      global.ignore,
		ValuesFiles: global.valuesFiles,
		ExtraValues: global.extraValues,
	}

	if cfg == nil {
		return cs
	}

	// Find per-chart config by matching the chart path relative to the config dir.
	relPath, err := filepath.Rel(configDir, chartPath)
	if err != nil {
		return cs
	}
	chartCfg, ok := cfg.Charts[relPath]
	if !ok {
		return cs
	}

	// Append per-chart ignore and valuesFiles.
	cs.Ignore = append(cs.Ignore, parsePaths(chartCfg.Ignore)...)
	cs.ValuesFiles = append(cs.ValuesFiles, resolvePaths(chartCfg.ValuesFiles, configDir)...)

	// Merge extraValues: chart wins over global.
	if len(chartCfg.ExtraValues) > 0 {
		merged := make(map[string]any)
		maps.Copy(merged, cs.ExtraValues)
		maps.Copy(merged, chartCfg.ExtraValues)
		cs.ExtraValues = merged
	}

	return cs
}

// parsePaths converts string patterns to vpath.Paths, skipping invalid ones.
func parsePaths(patterns []string) vpath.Paths {
	var paths vpath.Paths
	for _, p := range patterns {
		parsed, err := vpath.ParsePath(p)
		if err != nil {
			continue
		}
		paths = append(paths, parsed)
	}
	return paths
}

// resolvePaths makes paths absolute relative to configDir.
func resolvePaths(paths []string, configDir string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		if filepath.IsAbs(p) {
			out[i] = p
		} else {
			out[i] = filepath.Join(configDir, p)
		}
	}
	return out
}

const configFileName = ".helm-snoop.yaml"

// loadConfigFile returns the config file contents and path.
// If explicit is set, reads that file. Otherwise walks up from CWD.
// Returns (nil, "", nil) if no config file is found.
func loadConfigFile(explicit string) ([]byte, string, error) {
	if explicit != "" {
		data, err := os.ReadFile(explicit)
		if err != nil {
			return nil, "", fmt.Errorf("reading config file: %w", err)
		}
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return nil, "", err
		}
		return data, abs, nil
	}

	path, err := discover(".")
	if err != nil || path == "" {
		return nil, "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("reading config file: %w", err)
	}
	return data, path, nil
}

// discover walks up from startDir looking for .helm-snoop.yaml.
// Returns the absolute path if found, or ("", nil) if not found.
func discover(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, configFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Dependency represents a Go module dependency
type Dependency struct {
	Path    string
	Version string
	Direct  bool
}

// OutdatedDependency represents an outdated dependency
type OutdatedDependency struct {
	Path         string
	CurrentVer   string
	LatestVer    string
	UpdateNeeded bool
}

// DependencyVisualization represents a simple visualization of dependencies
type DependencyVisualization struct {
	Graph     string
	NodeCount int
	Deps      []Dependency
}

// parseGoMod parses go.mod content and extracts dependencies
func parseGoMod(content string) ([]Dependency, error) {
	var deps []Dependency
	lines := strings.Split(content, "\n")

	// Regular expression to match require statements
	requireBlockRegex := regexp.MustCompile(`^require\s+\(\s*$`)
	singleRequireRegex := regexp.MustCompile(`^require\s+([^\s]+)\s+(.+)$`)
	depRegex := regexp.MustCompile(`^\s*([^\s]+)\s+(.+)$`)

	inRequireBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check for require block start
		if requireBlockRegex.MatchString(line) {
			inRequireBlock = true
			continue
		}

		// Check for require block end
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		// Parse single require line
		if !inRequireBlock {
			matches := singleRequireRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				deps = append(deps, Dependency{
					Path:    matches[1],
					Version: strings.TrimSpace(matches[2]),
					Direct:  true,
				})
			}
			continue
		}

		// Parse dependency in require block
		if inRequireBlock {
			matches := depRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				// Check if it has a comment like "// indirect"
				direct := !strings.Contains(line, "// indirect")

				deps = append(deps, Dependency{
					Path:    matches[1],
					Version: strings.TrimSpace(matches[2]),
					Direct:  direct,
				})
			}
		}
	}

	return deps, nil
}

// parseGoSum parses go.sum content to get more detailed version information
func parseGoSum(content string) map[string]string {
	versionMap := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			modulePath := parts[0]
			version := parts[1]
			versionMap[modulePath] = version
		}
	}

	return versionMap
}

// scanDependencies scans dependencies from go.mod content
func scanDependencies(goMod string) ([]Dependency, error) {
	return parseGoMod(goMod)
}

// checkOutdatedDependencies checks for outdated dependencies using go.mod content
func checkOutdatedDependencies(goMod string) ([]OutdatedDependency, error) {
	deps, err := parseGoMod(goMod)
	if err != nil {
		return nil, err
	}

	var outdated []OutdatedDependency

	// For each dependency, check if there's a newer version
	for _, dep := range deps {
		// Skip the main module
		if !strings.Contains(dep.Path, ".") {
			continue
		}

		// Check for newer versions using Go proxy API
		latestVer, err := getLatestVersion(dep.Path)
		if err != nil {
			// If we can't check, just assume current is latest
			outdated = append(outdated, OutdatedDependency{
				Path:         dep.Path,
				CurrentVer:   dep.Version,
				LatestVer:    dep.Version,
				UpdateNeeded: false,
			})
			continue
		}

		// Compare versions
		updateNeeded := latestVer != dep.Version

		outdated = append(outdated, OutdatedDependency{
			Path:         dep.Path,
			CurrentVer:   dep.Version,
			LatestVer:    latestVer,
			UpdateNeeded: updateNeeded,
		})
	}

	return outdated, nil
}

// getLatestVersion gets the latest version of a module using the Go proxy API
func getLatestVersion(modulePath string) (string, error) {
	// Use the Go proxy API to get the latest version
	url := fmt.Sprintf("https://proxy.golang.org/%s/@latest", modulePath)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get latest version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Version string `json:"Version"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Version, nil
}

// visualizeDependencies generates a simple visualization of dependencies from go.mod content
func visualizeDependencies(goModContent string) (*DependencyVisualization, error) {
	deps, err := parseGoMod(goModContent)
	if err != nil {
		return nil, err
	}

	// Extract module name from go.mod
	moduleName := "main module"
	moduleRegex := regexp.MustCompile(`^module\s+(.+)$`)
	lines := strings.Split(goModContent, "\n")
	for _, line := range lines {
		matches := moduleRegex.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) == 2 {
			moduleName = matches[1]
			break
		}
	}

	// Generate a simple ASCII graph
	var graph strings.Builder
	graph.WriteString("```\n")
	graph.WriteString("Dependency Graph:\n")
	graph.WriteString(fmt.Sprintf("%s\n", moduleName))

	for _, dep := range deps {
		graph.WriteString(fmt.Sprintf("  ├── %s@%s\n", dep.Path, dep.Version))
	}
	graph.WriteString("```\n")

	return &DependencyVisualization{
		Graph:     graph.String(),
		NodeCount: len(deps) + 1, // +1 for main module
		Deps:      deps,
	}, nil
}

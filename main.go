package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"github.com/bmatcuk/doublestar/v4"
)

var (
	pathFlag    = flag.String("path", ".", "Root directory to scan")
	excludeFlag = flag.String("exclude", "", "Comma-separated glob patterns to exclude (e.g. venv/*,*.log)")
)

func shouldExclude(path string, root string, patterns []string) bool {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	relPath = filepath.ToSlash(relPath)

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		match, err := doublestar.PathMatch(pattern, relPath)
		if err != nil {
			continue
		}
		if match {
			return true
		}
	}
	return false
}

func readFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("[Error reading file: %v]", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Sprintf("[Error reading file: %v]", err)
	}
	return string(content)
}

func renderPlainOutput(root string, files []string) string {
	var builder strings.Builder
	for _, file := range files {
		rel, _ := filepath.Rel(root, file)
		content := readFile(file)
		builder.WriteString(fmt.Sprintf(">>> START FILE %q\n%s\n<<< END FILE\n\n", rel, content))
	}
	return builder.String()
}

func collectFiles(root string, excludes []string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil {
			return nil
		}
		if shouldExclude(path, root, excludes) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func main() {
	flag.Parse()

	var excludePatterns []string
	if *excludeFlag != "" {
		excludePatterns = strings.Split(*excludeFlag, ",")
	}

	root := *pathFlag

	files, err := collectFiles(root, excludePatterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	output := renderPlainOutput(root, files)
	fmt.Print(output)
}

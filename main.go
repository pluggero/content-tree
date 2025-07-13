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
	includeFlag = flag.String("include", "", "Comma-separated glob patterns to include (e.g. **/*.go,cmd/**). When empty, all files are included.")
	maxLenFlag  = flag.Int("max-length", 0, "Maximum number of lines per prompt message (0 means no limit)")

)

// matchesPattern reports whether relPath matches at least one of the supplied
// doublestar glob patterns. The comparison is done using filesystem‑agnostic
// forward slashes.
func matchesPattern(relPath string, patterns []string) bool {
    for _, pattern := range patterns {
        pattern = strings.TrimSpace(pattern)
        if pattern == "" {
            continue
        }
        match, err := doublestar.PathMatch(pattern, relPath)
        if err != nil {
            // Ignore malformed patterns and continue.
            continue
        }
        if match {
            return true
        }
    }
    return false
}

// shouldProcess returns true if the given path should be processed, depending on
// the include and exclude patterns. Paths are checked relative to the root
// directory so that patterns are intuitive (e.g. "cmd/**" or "**/*.go").
func shouldProcess(path, root string, includePatterns, excludePatterns []string) bool {
    relPath, err := filepath.Rel(root, path)
    if err != nil {
        // If we cannot compute a relative path, play it safe and skip.
        return false
    }
    relPath = filepath.ToSlash(relPath)

    // Exclude takes precedence.
    if matchesPattern(relPath, excludePatterns) {
        return false
    }

    // If include patterns are provided, the path must match at least one of them.
    if len(includePatterns) > 0 && !matchesPattern(relPath, includePatterns) {
        return false
    }
    return true
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
        builder.WriteString(fmt.Sprintf(">>> START FILE %q\n%s\n<<< END FILE\n\n", filepath.ToSlash(rel), content))
    }
    return builder.String()
}

// collectFiles walks the directory tree rooted at root and returns the list of
// files that should be processed according to include/exclude rules.
func collectFiles(root string, includePatterns, excludePatterns []string) ([]string, error) {
    var files []string

    err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Always skip the root itself when matching patterns.
        if path == root {
            return nil
        }

        if d.IsDir() {
            // Skip entire directory if it matches an exclude pattern.
            rel, _ := filepath.Rel(root, path)
            if matchesPattern(filepath.ToSlash(rel), excludePatterns) {
                return filepath.SkipDir
            }
            return nil
        }

        if shouldProcess(path, root, includePatterns, excludePatterns) {
            files = append(files, path)
        }
        return nil
    })
    if err != nil {
        return nil, err
    }

    sort.Strings(files)
    return files, nil
}

// splitLines divides s into chunks containing at most maxLines lines. If
// maxLines <= 0, the original string is returned in a single-element slice.
func splitLines(s string, maxLines int) []string {
    if maxLines <= 0 {
        return []string{s}
    }

    lines := strings.Split(s, "\n")
    var parts []string
    for i := 0; i < len(lines); i += maxLines {
        end := i + maxLines
        if end > len(lines) {
            end = len(lines)
        }
        part := strings.Join(lines[i:end], "\n")
        parts = append(parts, part)
    }
    return parts
}

func main() {
    flag.Parse()

    // Split the include/exclude pattern lists once so we do not perform it for every path.
    var includePatterns, excludePatterns []string
    if *includeFlag != "" {
        includePatterns = strings.Split(*includeFlag, ",")
    }
    if *excludeFlag != "" {
        excludePatterns = strings.Split(*excludeFlag, ",")
    }

    root := *pathFlag

    files, err := collectFiles(root, includePatterns, excludePatterns)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
        os.Exit(1)
    }

    output := renderPlainOutput(root, files)

    parts := splitLines(output, *maxLenFlag)
    totalParts := len(parts)

    for i, part := range parts {
        if *maxLenFlag > 0 {
            fmt.Printf(">>>> START PROMPT PART %d OF %d\n%s\n<<<< END PROMPT PART %d OF %d\n\n", i+1, totalParts, part, i+1, totalParts)
        } else {
            fmt.Print(part)
        }
    }
}

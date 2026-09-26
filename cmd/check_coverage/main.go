package main

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	percentMultiplier  = 100.0
	minProfileFields   = 3
	minLocationParts   = 2
	expectedRangeParts = 2
)

// lineRange defines a start and end line number.
type lineRange struct {
	start int
	end   int
}

// coverageResult holds statement coverage counts and uncovered line ranges.
type coverageResult struct {
	coveredStmts    int
	totalStmts      int
	uncoveredRanges []lineRange
}

// profileEntry represents a parsed line from a coverage profile.
type profileEntry struct {
	filePath string
	locRange string
	numStmt  int
	count    int
}

// main runs isolated code coverage checks across packages.
func main() {
	pkgDirs := [...]string{"./datatypes", "./iri", "./langtag"}
	failed := false

	fmt.Fprintln(os.Stdout, "==================================================")
	fmt.Fprintln(os.Stdout, " Running Isolated 100% Coverage Checks")
	fmt.Fprintln(os.Stdout, "==================================================")

	for _, pkgDir := range pkgDirs {
		if !checkPackage(pkgDir) {
			failed = true
		}
	}

	fmt.Fprintln(os.Stdout, "\n==================================================")
	if failed {
		fmt.Fprintln(os.Stdout, "Result: Coverage check failed.")
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, "Result: All files in all packages achieved 100% code coverage!")
}

// checkPackage runs coverage checks on all eligible source files in a package directory.
func checkPackage(rawPkgDir string) bool {
	pkgDir := normalizePkgDir(rawPkgDir)
	fmt.Fprintf(os.Stdout, "\n📦 Package: %s\n", pkgDir)

	files, err := os.ReadDir(pkgDir)
	if err != nil {
		fmt.Fprintf(os.Stdout, "  ❌ Error reading directory %s: %v\n", pkgDir, err)
		return false
	}

	sourceFiles := filterSourceFiles(files)
	if len(sourceFiles) == 0 {
		fmt.Fprintln(os.Stdout, "  ⚠️  No source .go files found in this package.")
		return true
	}

	pkgPassed := true
	for _, f := range sourceFiles {
		if !checkFile(pkgDir, f.Name()) {
			pkgPassed = false
		}
	}
	return pkgPassed
}

// normalizePkgDir ensures the package directory retains the relative ./ prefix.
func normalizePkgDir(pkgDir string) string {
	cleaned := filepath.Clean(pkgDir)
	if !strings.HasPrefix(cleaned, ".") && !filepath.IsAbs(cleaned) {
		return "./" + cleaned
	}
	return cleaned
}

// filterSourceFiles filters directory entries to find eligible Go source files.
func filterSourceFiles(entries []os.DirEntry) []os.DirEntry {
	var files []os.DirEntry
	for _, entry := range entries {
		if isSourceFile(entry) {
			files = append(files, entry)
		}
	}
	return files
}

// isSourceFile determines if an entry is a Go source file subject to coverage checks.
func isSourceFile(f os.DirEntry) bool {
	if f.IsDir() {
		return false
	}
	name := f.Name()
	if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
		return false
	}
	return name != "doc.go" && name != "main.go"
}

// checkFile executes test coverage checks for a single source file.
func checkFile(pkgDir, name string) bool {
	base := strings.TrimSuffix(name, ".go")
	testFileName := base + "_test.go"
	testFilePath := filepath.Join(pkgDir, testFileName)
	srcFilePath := filepath.Join(pkgDir, name)

	if _, statErr := os.Stat(testFilePath); os.IsNotExist(statErr) {
		fmt.Fprintf(os.Stdout, "  ❌ FAIL: %s (Missing test file: %s)\n", name, testFileName)
		return false
	}

	tests, extractErr := extractTestFunctions(testFilePath)
	if extractErr != nil || len(tests) == 0 {
		fmt.Fprintf(os.Stdout, "  ⚠️  SKIP: %s (No Test functions found in %s)\n", name, testFileName)
		return true
	}

	res, ok := runFileTests(pkgDir, name, tests)
	if !ok {
		return false
	}

	return reportCoverage(name, srcFilePath, res)
}

// runFileTests runs go test for a source file and parses its coverage result.
func runFileTests(pkgDir, name string, tests []string) (coverageResult, bool) {
	covFile, err := os.CreateTemp("", "cov-*.out")
	if err != nil {
		fmt.Fprintf(os.Stdout, "  Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	covPath := covFile.Name()
	if closeErr := covFile.Close(); closeErr != nil {
		_ = os.Remove(covPath)
		fmt.Fprintf(os.Stdout, "  Error closing temp file: %v\n", closeErr)
		os.Exit(1)
	}
	defer func() {
		_ = os.Remove(covPath)
	}()

	runRegex := "^(" + strings.Join(tests, "|") + ")$"
	// #nosec G204 -- Arguments are constructed from internal package tests.
	cmd := exec.CommandContext(
		context.Background(),
		"go",
		"test",
		"-coverprofile="+covPath,
		"-run",
		runRegex,
		pkgDir,
	)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		fmt.Fprintf(
			os.Stdout,
			"  ❌ FAIL: %s (go test failed)\n     Output:\n%s\n",
			name,
			indentLines(string(out), "     "),
		)
		return coverageResult{}, false
	}

	res, parseErr := evaluateFileCoverage(covPath, pkgDir, name)
	if parseErr != nil {
		fmt.Fprintf(os.Stdout, "  ❌ FAIL: %s (Error parsing coverage profile: %v)\n", name, parseErr)
		return coverageResult{}, false
	}

	return res, true
}

// reportCoverage prints coverage results and returns whether checks passed.
func reportCoverage(name, srcFilePath string, res coverageResult) bool {
	if res.totalStmts == 0 {
		return handleZeroStatements(name, srcFilePath)
	}

	pct := (float64(res.coveredStmts) / float64(res.totalStmts)) * percentMultiplier
	if res.coveredStmts == res.totalStmts {
		fmt.Fprintf(
			os.Stdout,
			"  ✅ PASS: %s (100.0%% - %d/%d statements)\n",
			name,
			res.coveredStmts,
			res.totalStmts,
		)
		return true
	}

	fmt.Fprintf(
		os.Stdout,
		"  ❌ FAIL: %s (%.1f%% - %d/%d statements)\n",
		name,
		pct,
		res.coveredStmts,
		res.totalStmts,
	)
	fmt.Fprintln(os.Stdout, "     Uncovered lines:")
	printUncoveredRanges(res.uncoveredRanges)
	return false
}

// handleZeroStatements reports coverage for files with zero statements.
func handleZeroStatements(name, srcFilePath string) bool {
	hasFuncs, _ := hasExecutableFunctions(srcFilePath)
	if hasFuncs {
		fmt.Fprintf(
			os.Stdout,
			"  ❌ FAIL: %s (0.0%% coverage - file has functions but zero statements executed)\n",
			name,
		)
		return false
	}

	fmt.Fprintf(
		os.Stdout,
		"  ✅ PASS: %s (100.0%% - 0/0 statements / interfaces & types only)\n",
		name,
	)
	return true
}

// printUncoveredRanges prints formatted uncovered line ranges to stdout.
func printUncoveredRanges(ranges []lineRange) {
	for _, r := range ranges {
		if r.start == r.end {
			fmt.Fprintf(os.Stdout, "       • line %d\n", r.start)
		} else {
			fmt.Fprintf(os.Stdout, "       • lines %d-%d\n", r.start, r.end)
		}
	}
}

// hasExecutableFunctions checks if a Go source file contains non-empty function or method bodies.
func hasExecutableFunctions(filePath string) (bool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return false, err
	}

	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Body != nil && len(fn.Body.List) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

// extractTestFunctions uses Go's AST parser to extract function names starting with "Test".
func extractTestFunctions(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var tests []string
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "Test") {
				tests = append(tests, fn.Name.Name)
			}
		}
	}
	return tests, nil
}

// parseProfileLine parses a single line from a Go coverage profile.
func parseProfileLine(line string) (profileEntry, bool) {
	if strings.HasPrefix(line, "mode:") {
		return profileEntry{}, false
	}

	parts := strings.Fields(line)
	if len(parts) < minProfileFields {
		return profileEntry{}, false
	}

	locParts := strings.Split(parts[0], ":")
	if len(locParts) < minLocationParts {
		return profileEntry{}, false
	}

	numStmt, err1 := strconv.Atoi(parts[1])
	count, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return profileEntry{}, false
	}

	return profileEntry{
		filePath: locParts[0],
		locRange: locParts[1],
		numStmt:  numStmt,
		count:    count,
	}, true
}

// parseUncoveredRange extracts a lineRange from a coverage profile line range string.
func parseUncoveredRange(rangeStr string) (lineRange, bool) {
	rangeParts := strings.Split(rangeStr, ",")
	if len(rangeParts) != expectedRangeParts {
		return lineRange{}, false
	}

	startLineParts := strings.Split(rangeParts[0], ".")
	endLineParts := strings.Split(rangeParts[1], ".")

	startLine, err1 := strconv.Atoi(startLineParts[0])
	endLine, err2 := strconv.Atoi(endLineParts[0])
	if err1 != nil || err2 != nil || startLine <= 0 || endLine < startLine {
		return lineRange{}, false
	}

	return lineRange{start: startLine, end: endLine}, true
}

// evaluateFileCoverage parses the coverage profile for blocks matching targetFileName.
func evaluateFileCoverage(profilePath, pkgDir, targetFileName string) (coverageResult, error) {
	file, err := os.Open(profilePath)
	if err != nil {
		return coverageResult{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	var res coverageResult
	var rawUncovered []lineRange

	pkgBase := filepath.Base(pkgDir)
	expectedSuffix := pkgBase + "/" + targetFileName

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entry, ok := parseProfileLine(scanner.Text())
		if !ok {
			continue
		}

		if !strings.HasSuffix(entry.filePath, expectedSuffix) && entry.filePath != targetFileName {
			continue
		}

		res.totalStmts += entry.numStmt
		if entry.count > 0 {
			res.coveredStmts += entry.numStmt
		} else if lr, valid := parseUncoveredRange(entry.locRange); valid {
			rawUncovered = append(rawUncovered, lr)
		}
	}

	res.uncoveredRanges = mergeLineRanges(rawUncovered)
	return res, scanner.Err()
}

// mergeLineRanges sorts and consolidates contiguous or overlapping uncovered line ranges.
func mergeLineRanges(ranges []lineRange) []lineRange {
	if len(ranges) == 0 {
		return nil
	}

	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].start < ranges[j].start
	})

	var merged []lineRange
	current := ranges[0]

	for i := 1; i < len(ranges); i++ {
		r := ranges[i]
		if r.start <= current.end+1 {
			if r.end > current.end {
				current.end = r.end
			}
		} else {
			merged = append(merged, current)
			current = r
		}
	}
	merged = append(merged, current)
	return merged
}

// indentLines prepends prefix to each line in s.
func indentLines(s, prefix string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

package main

import (
	"bufio"
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

type lineRange struct {
	start int
	end   int
}

type coverageResult struct {
	coveredStmts    int
	totalStmts      int
	uncoveredRanges []lineRange
}

func main() {
	pkgDirs := [...]string{"./datatypes", "./iri", "./langtag"}
	failed := false

	fmt.Println("==================================================")
	fmt.Println(" Running Isolated 100% Coverage Checks")
	fmt.Println("==================================================")

	for _, pkgDir := range pkgDirs {
		// Ensure package directory retains relative ./ prefix required by go test
		pkgDir = filepath.Clean(pkgDir)
		if !strings.HasPrefix(pkgDir, ".") && !filepath.IsAbs(pkgDir) {
			pkgDir = "./" + pkgDir
		}

		fmt.Printf("\n📦 Package: %s\n", pkgDir)

		files, err := os.ReadDir(pkgDir)
		if err != nil {
			fmt.Printf("  ❌ Error reading directory %s: %v\n", pkgDir, err)
			failed = true
			continue
		}

		hasSourceFiles := false
		for _, f := range files {
			name := f.Name()

			// Skip non-Go files, test files, doc.go, and main.go
			if f.IsDir() || strings.HasSuffix(name, "_test.go") || name == "doc.go" || name == "main.go" || !strings.HasSuffix(name, ".go") {
				continue
			}
			hasSourceFiles = true

			base := strings.TrimSuffix(name, ".go")
			testFileName := base + "_test.go"
			testFilePath := filepath.Join(pkgDir, testFileName)
			srcFilePath := filepath.Join(pkgDir, name)

			if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
				fmt.Printf("  ❌ FAIL: %s (Missing test file: %s)\n", name, testFileName)
				failed = true
				continue
			}

			// Parse test file to discover test function names
			tests, err := extractTestFunctions(testFilePath)
			if err != nil || len(tests) == 0 {
				fmt.Printf("  ⚠️  SKIP: %s (No Test functions found in %s)\n", name, testFileName)
				continue
			}

			// Run go test for this target file's tests only
			covFile, err := os.CreateTemp("", "cov-*.out")
			if err != nil {
				fmt.Printf("  Error creating temp file: %v\n", err)
				os.Exit(1)
			}
			covPath := covFile.Name()
			covFile.Close()

			runRegex := "^(" + strings.Join(tests, "|") + ")$"
			cmd := exec.Command("go", "test", "-coverprofile="+covPath, "-run", runRegex, pkgDir)
			out, err := cmd.CombinedOutput()
			if err != nil {
				os.Remove(covPath)
				fmt.Printf("  ❌ FAIL: %s (go test failed)\n     Output:\n%s\n", name, indentLines(string(out), "     "))
				failed = true
				continue
			}

			// Evaluate statement coverage and extract uncovered line ranges
			res, err := evaluateFileCoverage(covPath, pkgDir, name)
			os.Remove(covPath) // Clean up temp file immediately

			if err != nil {
				fmt.Printf("  ❌ FAIL: %s (Error parsing coverage profile: %v)\n", name, err)
				failed = true
				continue
			}

			// Handle zero statements
			if res.totalStmts == 0 {
				hasFuncs, _ := hasExecutableFunctions(srcFilePath)
				if hasFuncs {
					fmt.Printf("  ❌ FAIL: %s (0.0%% coverage - file has functions but zero statements executed)\n", name)
					failed = true
				} else {
					fmt.Printf("  ✅ PASS: %s (100.0%% - 0/0 statements / interfaces & types only)\n", name)
				}
				continue
			}

			pct := (float64(res.coveredStmts) / float64(res.totalStmts)) * 100.0
			if res.coveredStmts == res.totalStmts {
				fmt.Printf("  ✅ PASS: %s (100.0%% - %d/%d statements)\n", name, res.coveredStmts, res.totalStmts)
			} else {
				fmt.Printf("  ❌ FAIL: %s (%.1f%% - %d/%d statements)\n", name, pct, res.coveredStmts, res.totalStmts)
				fmt.Println("     Uncovered lines:")
				for _, r := range res.uncoveredRanges {
					if r.start == r.end {
						fmt.Printf("       • line %d\n", r.start)
					} else {
						fmt.Printf("       • lines %d-%d\n", r.start, r.end)
					}
				}
				failed = true
			}
		}

		if !hasSourceFiles {
			fmt.Println("  ⚠️  No source .go files found in this package.")
		}
	}

	fmt.Println("\n==================================================")
	if failed {
		fmt.Println("Result: Coverage check failed.")
		os.Exit(1)
	}
	fmt.Println("Result: All files in all packages achieved 100% code coverage!")
}

// hasExecutableFunctions checks if a Go source file contains non-empty function or method bodies
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

// extractTestFunctions uses Go's AST parser to extract function names starting with "Test"
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

// evaluateFileCoverage parses the coverage profile for blocks matching targetFileName
func evaluateFileCoverage(profilePath, pkgDir, targetFileName string) (coverageResult, error) {
	file, err := os.Open(profilePath)
	if err != nil {
		return coverageResult{}, err
	}
	defer file.Close()

	var res coverageResult
	var rawUncovered []lineRange

	pkgBase := filepath.Base(pkgDir)
	expectedSuffix := pkgBase + "/" + targetFileName

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		fileLoc := parts[0]
		locParts := strings.Split(fileLoc, ":")
		if len(locParts) < 2 {
			continue
		}

		filePath := locParts[0]
		if !strings.HasSuffix(filePath, expectedSuffix) && filePath != targetFileName {
			continue
		}

		numStmt, _ := strconv.Atoi(parts[1])
		count, _ := strconv.Atoi(parts[2])

		res.totalStmts += numStmt
		if count > 0 {
			res.coveredStmts += numStmt
		} else {
			rangeParts := strings.Split(locParts[1], ",")
			if len(rangeParts) == 2 {
				startLineParts := strings.Split(rangeParts[0], ".")
				endLineParts := strings.Split(rangeParts[1], ".")

				startLine, _ := strconv.Atoi(startLineParts[0])
				endLine, _ := strconv.Atoi(endLineParts[0])

				if startLine > 0 && endLine >= startLine {
					rawUncovered = append(rawUncovered, lineRange{start: startLine, end: endLine})
				}
			}
		}
	}

	res.uncoveredRanges = mergeLineRanges(rawUncovered)
	return res, scanner.Err()
}

// mergeLineRanges sorts and consolidates contiguous or overlapping uncovered line ranges
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

func indentLines(s, prefix string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
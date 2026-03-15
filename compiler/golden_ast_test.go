package compiler

// Golden AST comparison test: parses each .ivy file in ivy-lang-examples/
// with both Python and Go, serializes the AST to text, and compares.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/parser"
)

// testdataDir returns the absolute path to the testdata/ directory.
func testdataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "testdata")
}

func pythonTestHelperDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "pytesthelper")
}

// examplesDir returns the absolute path to the ivy-lang-examples/ directory.
func examplesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "ivy-lang-examples")
}

// pythonDumper returns the path to the Python AST dump script.
func pythonDumper() string {
	return filepath.Join(pythonTestHelperDir(), "ivy_ast_dump.py")
}

// pythonAvailable checks if python3 and the dumper script are available.
func pythonAvailable() bool {
	dumper := pythonDumper()
	if _, err := os.Stat(dumper); err != nil {
		return false
	}
	_, err := exec.LookPath("python3")
	return err == nil
}

// parsePythonAST runs the Python AST dumper on a file and returns the output lines.
func parsePythonAST(t *testing.T, ivyFile string) ([]string, error) {
	t.Helper()
	dumper := pythonDumper()
	// Detect version from the file
	data, err := os.ReadFile(ivyFile)
	if err != nil {
		return nil, err
	}
	ver := "1.7"
	src := string(data)
	if strings.HasPrefix(src, "#lang ivy") {
		line := src[:strings.Index(src, "\n")]
		v := strings.TrimPrefix(line, "#lang ivy")
		if v != "" {
			ver = v
		}
	}

	// Run the Python dumper from the ivy root directory (where the ivy package is)
	_, thisFile, _, _ := runtime.Caller(0)
	ivyRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	cmd := exec.Command("python3", dumper, "--version", ver, ivyFile)
	cmd.Dir = ivyRoot
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		// Parse error from Python — return the error output as a single line
		if strings.HasPrefix(outStr, "PARSE_ERROR:") || strings.HasPrefix(outStr, "ERROR:") {
			return []string{outStr}, nil
		}
		return nil, fmt.Errorf("python dumper failed: %s\n%s", err, outStr)
	}
	if outStr == "" {
		return nil, nil
	}
	return strings.Split(outStr, "\n"), nil
}

// parseGoAST parses a file with the Go parser and returns the output in the
// same format as the Python dumper: "[i] DeclType name"
func parseGoAST(t *testing.T, ivyFile string) ([]string, error) {
	t.Helper()
	data, err := os.ReadFile(ivyFile)
	if err != nil {
		return nil, err
	}
	src := string(data)

	// Detect version
	ver := lexer.Version{1, 7}
	if strings.HasPrefix(src, "#lang ivy") {
		line := src[:strings.Index(src, "\n")]
		v := strings.TrimPrefix(line, "#lang ivy")
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			major, minor := 1, 7
			fmt.Sscanf(parts[0], "%d", &major)
			fmt.Sscanf(parts[1], "%d", &minor)
			ver = lexer.Version{major, minor}
		}
		src = src[strings.Index(src, "\n")+1:]
	}

	p := parser.New(src, ver)
	decls, parseErr := p.Parse()
	if parseErr != nil && len(decls) == 0 {
		return []string{fmt.Sprintf("PARSE_ERROR: %s", parseErr)}, nil
	}

	var lines []string
	for i, d := range decls {
		typeName := fmt.Sprintf("%T", d)
		typeName = typeName[strings.LastIndex(typeName, ".")+1:]
		name := goDeclName(d)
		lines = append(lines, fmt.Sprintf("[%d] %s %s", i, typeName, name))
	}
	return lines, nil
}

// goDeclName extracts a stable name from a Go AST declaration.
func goDeclName(d ast.Node) string {
	switch n := d.(type) {
	case *ast.TypeDecl:
		if len(n.Args()) > 0 {
			if td, ok := n.Args()[0].(*ast.TypeDef); ok {
				if sym, ok := td.Name.(*ast.Symbol); ok {
					return sym.Rep
				}
				if atom, ok := td.Name.(*ast.Atom); ok {
					return atom.Rep
				}
				return fmt.Sprint(td.Name)
			}
		}
	case *ast.ConstantDecl:
		if len(n.Args()) > 0 {
			if a, ok := n.Args()[0].(*ast.Atom); ok {
				if len(a.Terms) > 0 {
					return fmt.Sprintf("%s(%d)", a.Rep, len(a.Terms))
				}
				return a.Rep
			}
		}
	case *ast.ActionDecl:
		if len(n.Args()) > 0 {
			if ad, ok := n.Args()[0].(*ast.ActionDef); ok {
				return ad.Defines()
			}
		}
	case *ast.MixinDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			// Truncate for readability
			if len(s) > 40 {
				s = s[:40]
			}
			return s
		}
	case *ast.ConjectureDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *ast.PropertyDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *ast.AxiomDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *ast.ExportDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ast.ImportDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ast.ObjectDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ast.ModuleDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ast.DerivedDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *ast.IsolateDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ast.InterpretDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ast.DefinitionDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *ast.InstantiateDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ast.InitDecl:
		return "init"
	}
	return ""
}

// TestGoldenAST walks ivy-lang-examples/ and compares Python vs Go AST output
// for every .ivy file. Only the declaration count and types are compared
// (not the internal details like auto-generated label names).
func TestGoldenAST(t *testing.T) {
	if !pythonAvailable() {
		t.Skip("python3 or ivy_ast_dump.py not available")
	}

	beg := 514
	end := 761

	dir := examplesDir()
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("ivy-lang-examples/ not found at %s", dir)
	}

	var total, matched, diffCount, skipCount int

	var ex []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() || !strings.HasSuffix(path, ".ivy") {
			return nil
		}
		total++

		//x, _ := filepath.Rel(dir, path)
		ex = append(ex, path)
		return nil
	})

	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	//vv("will compare a total of %v paths", len(ex))
	for i, path := range ex {
		if i < beg {
			continue
		}
		if i > end {
			break
		}

		if i%10 == 0 {
			//vv("out of %v: i=%v; skipCount=%v ; matched = %v", len(ex), i, skipCount, matched)
		}

		// Get Python AST
		pyLines, pyErr := parsePythonAST(t, path)
		if pyErr != nil {
			skipCount++
			t.Skipf("Python error: %v", pyErr)
			continue
		}

		// Check if Python had a parse error
		if len(pyLines) == 1 && (strings.HasPrefix(pyLines[0], "PARSE_ERROR:") || strings.HasPrefix(pyLines[0], "ERROR:")) {
			// Python couldn't parse it either — skip comparison
			skipCount++
			//t.Skipf("Python parse error: %s", pyLines[0])
			continue
		}

		// Get Go AST
		goLines, goErr := parseGoAST(t, path)
		if goErr != nil {
			t.Fatalf("path='%v': Go parse error: %v", path, goErr)
			continue
		}

		// Check if Go had a parse error
		if len(goLines) == 1 && strings.HasPrefix(goLines[0], "PARSE_ERROR:") {
			// If Python succeeded but Go failed, that's a real difference
			if len(pyLines) > 0 {
				diffCount++
				t.Fatalf("path='%v': Go parse error but Python succeeded (%d decls):\n  Go: %s", path, len(pyLines), goLines[0])
			}
			continue
		}

		// Compare declaration count
		if len(pyLines) != len(goLines) {
			diffCount++
			t.Fatalf("i=%v path='%v': declaration count mismatch: Python=%d Go=%d\n  Python:\n    %s\n  Go:\n    %s",
				i, path,
				len(pyLines), len(goLines),
				strings.Join(pyLines, "\n    "),
				strings.Join(goLines, "\n    "))
			continue
		}

		// Compare each declaration's type (the word after [N])
		for i := 0; i < len(pyLines) && i < len(goLines); i++ {
			pyType := extractDeclType(pyLines[i])
			goType := extractDeclType(goLines[i])
			if pyType != goType {
				diffCount++
				t.Fatalf("path='%v': declaration [%d] type mismatch:\n  Python: %s\n  Go:     %s", path, i, pyLines[i], goLines[i])
				continue
			}
		}
		matched++
	}

	t.Logf("Results: %d total, %d matched, %d diffs, %d skipped", total, matched, diffCount, skipCount)
	if diffCount > 0 {
		t.Fatalf("%d files had mismatches out of %d tested", diffCount, matched+diffCount)
	}
}

// extractDeclType extracts the declaration type from a line like "[0] TypeDecl client"
func extractDeclType(line string) string {
	// Skip "[N] "
	idx := strings.Index(line, "] ")
	if idx < 0 {
		return line
	}
	rest := line[idx+2:]
	// Take the first word
	sp := strings.Index(rest, " ")
	if sp < 0 {
		return rest
	}
	return rest[:sp]
}

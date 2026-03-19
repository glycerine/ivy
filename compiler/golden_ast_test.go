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
	ivyRoot := filepath.Join(filepath.Dir(thisFile), "..") // , "..")

	ivyHomeDir := os.Getenv("IVY_HOME")
	if ivyHomeDir != "" {
		ivyRoot = ivyHomeDir
	}
	//vv("ivyRoot = '%v'", ivyRoot) // /Users/jaten/go/src/github.com/glycerine

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
	//t.Skip("skip TestGoldenAST because takes 2+ minutes to run ")

	if !pythonAvailable() {
		t.Skip("python3 or ivy_ast_dump.py not available")
	}

	beg := 0
	end := 761

	// to speed up the test, now we only run those which
	// python ivy can parse.
	// 133s -> 43s test time and we are not testing anything less.
	//
	// I don't know why Python Ivy stumbles on
	// half the .ivy example inputs. Some are probably tutorial
	// files that are not expected to build. But most? (462 of 760)?
	// That's strange. Test on the 298 that do parse:
	pyIvyNoError := []int{0, 1, 2, 3, 41, 43, 44, 49, 53, 54, 55, 57, 59, 86,
		94, 95, 96, 97, 99, 104, 107, 108, 114, 126, 131, 132, 133, 154, 155,
		156, 157, 178, 183, 184, 189, 198, 210, 211, 235, 239, 240, 242, 243,
		244, 252, 280, 281, 291, 300, 302, 303, 304, 306, 307, 308, 309, 310,
		311, 314, 315, 317, 318, 320, 322, 324, 325, 326, 327, 329, 330, 332,
		333, 335, 337, 338, 339, 340, 341, 342, 345, 346, 350, 354, 356, 357,
		358, 359, 361, 362, 364, 365, 372, 374, 375, 377, 379, 394, 396, 397,
		399, 401, 415, 417, 418, 420, 422, 436, 438, 439, 441, 444, 463, 464,
		469, 470, 473, 474, 475, 480, 482, 483, 484, 485, 486, 487, 489, 490,
		491, 492, 500, 501, 503, 504, 505, 506, 508, 510, 511, 512, 513, 514,
		515, 519, 520, 521, 522, 523, 525, 532, 533, 534, 535, 536, 539, 541,
		542, 544, 545, 546, 547, 548, 549, 551, 552, 553, 554, 555, 556, 557,
		558, 559, 560, 561, 562, 563, 564, 565, 566, 567, 568, 569, 570, 571,
		572, 573, 575, 576, 577, 578, 581, 583, 584, 585, 586, 587, 590, 591,
		592, 593, 594, 596, 597, 600, 602, 604, 605, 606, 609, 610, 611, 612,
		613, 614, 617, 618, 619, 620, 621, 622, 623, 624, 625, 626, 627, 629,
		630, 632, 633, 634, 635, 636, 637, 638, 639, 640, 641, 643, 644, 646,
		647, 648, 649, 652, 660, 662, 663, 664, 665, 667, 669, 670, 671, 672,
		679, 682, 687, 689, 690, 691, 692, 693, 694, 695, 696, 697, 699, 700,
		701, 702, 703, 704, 705, 706, 710, 712, 719, 720, 721, 722, 723, 724,
		725, 729, 732, 733, 734, 735, 736, 739, 743, 744, 745, 746, 747, 748,
		755, 756, 757}

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
	//for i, path := range ex {
	for _, i := range pyIvyNoError {
		path := ex[i]

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
			//t.Skipf("Python error: %v", pyErr)
			//t.Fatalf("i=%v; path='%v'; Python error: %v", i, path, pyErr)
			continue
		}

		// Check if Python had a parse error
		if len(pyLines) == 1 && (strings.HasPrefix(pyLines[0], "PARSE_ERROR:") || strings.HasPrefix(pyLines[0], "ERROR:")) {
			// Python couldn't parse it either — skip comparison
			skipCount++
			//t.Skipf("Python parse error: %s", pyLines[0])
			//t.Fatalf("i=%v; path='%v'; Python error: %v", i, path, pyLines[0])
			continue
		}

		//fmt.Printf("%v,", i)

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
			t.Logf("i=%v path='%v': count mismatch: Py=%d Go=%d",
				i, path, len(pyLines), len(goLines))
			continue
		}

		// Compare each declaration's type (the word after [N])
		typeMismatch := false
		for i := 0; i < len(pyLines) && i < len(goLines); i++ {
			pyType := extractDeclType(pyLines[i])
			goType := extractDeclType(goLines[i])
			if pyType != goType {
				if !typeMismatch {
					diffCount++
					t.Logf("path='%v': decl [%d] type mismatch: Py=%s Go=%s", path, i, pyType, goType)
				}
				typeMismatch = true
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

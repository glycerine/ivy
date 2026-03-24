package lalr_full

// Golden AST comparison test: parses each .ivy file in ivy-lang-examples/
// with both Python and Go LALR parser, serializes the AST to text, and compares.
// Adapted from compiler/golden_ast_test.go — uses lalr_full.Parse instead of
// the hand-rolled parser.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/lexer"
)

// examplesDir returns the absolute path to the ivy-lang-examples/ directory.
func examplesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "ivy-lang-examples")
}

// pythonTestHelperDir returns the path to the Python helper scripts.
func pythonTestHelperDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "pytesthelper")
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

	_, thisFile, _, _ := runtime.Caller(0)
	ivyRoot := filepath.Join(filepath.Dir(thisFile), "..")

	ivyHomeDir := os.Getenv("IVY_HOME")
	if ivyHomeDir != "" {
		ivyRoot = ivyHomeDir
	}

	args := []string{dumper, "--version", ver, ivyFile}
	cmd := exec.Command("python3", args...)
	cmd.Dir = ivyRoot
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
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

// parseLALRFullAST parses a file with the LALR full parser and returns the output
// in the same format as the Python dumper: "[i] DeclType name"
func parseLALRFullAST(t *testing.T, ivyFile string) ([]string, error) {
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

	result, parseErr := Parse(src, ver)
	if parseErr != nil {
		return []string{fmt.Sprintf("PARSE_ERROR: %s", parseErr)}, nil
	}
	if result == nil || len(result.Decls) == 0 {
		return nil, nil
	}

	var lines []string
	for i, d := range result.Decls {
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

// extractDeclType extracts the declaration type from a line like "[0] TypeDecl client"
func extractDeclType(line string) string {
	idx := strings.Index(line, "] ")
	if idx < 0 {
		return line
	}
	rest := line[idx+2:]
	sp := strings.Index(rest, " ")
	if sp < 0 {
		return rest
	}
	return rest[:sp]
}

// pyIvyNoError is the list of file indices that Python Ivy can parse without error.
// Copied from compiler/golden_ast_test.go.
var pyIvyNoError = []int{0, 1, 2, 3, 41, 43, 44, 49, 53, 54, 55, 57, 59, 86,
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

// TestGoldenLALR walks ivy-lang-examples/ and compares Python vs Go LALR parser
// AST output for every .ivy file that Python can parse.
func TestGoldenLALR(t *testing.T) {
	t.Skip("skip TestGoldenLALR, takes forev.")
	return // for now
	if testing.Short() {
		t.Skip("skip TestGoldenLALR in short mode (takes ~1 minute)")
	}

	if !pythonAvailable() {
		t.Skip("python3 or ivy_ast_dump.py not available")
	}

	dir := examplesDir()
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("ivy-lang-examples/ not found at %s", dir)
	}

	// Collect all .ivy files
	var allFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".ivy") {
			return nil
		}
		allFiles = append(allFiles, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	var total, matched, diffCount, parseErrCount, skipCount int

	for _, idx := range pyIvyNoError {
		if idx >= len(allFiles) {
			break
		}
		path := allFiles[idx]
		total++

		// Get Python AST
		pyLines, pyErr := parsePythonAST(t, path)
		if pyErr != nil {
			skipCount++
			continue
		}
		if len(pyLines) == 1 && (strings.HasPrefix(pyLines[0], "PARSE_ERROR:") || strings.HasPrefix(pyLines[0], "ERROR:")) {
			skipCount++
			continue
		}

		// Get Go LALR AST
		goLines, goErr := parseLALRFullAST(t, path)
		if goErr != nil {
			t.Logf("i=%d path=%s: Go LALR error: %v", idx, filepath.Base(path), goErr)
			parseErrCount++
			continue
		}
		if len(goLines) == 1 && strings.HasPrefix(goLines[0], "PARSE_ERROR:") {
			if len(pyLines) > 0 {
				parseErrCount++
				t.Logf("i=%d path=%s: LALR parse error (Python got %d decls): %s",
					idx, filepath.Base(path), len(pyLines), goLines[0])
			}
			continue
		}

		// Compare declaration count
		if len(pyLines) != len(goLines) {
			diffCount++
			t.Logf("i=%d path=%s: count mismatch: Py=%d Go=%d",
				idx, filepath.Base(path), len(pyLines), len(goLines))
			continue
		}

		// Compare each declaration type
		typeMismatch := false
		for j := 0; j < len(pyLines) && j < len(goLines); j++ {
			pyType := extractDeclType(pyLines[j])
			goType := extractDeclType(goLines[j])
			if pyType != goType {
				if !typeMismatch {
					diffCount++
					t.Logf("i=%d path=%s: decl [%d] type mismatch: Py=%s Go=%s",
						idx, filepath.Base(path), j, pyType, goType)
				}
				typeMismatch = true
			}
		}
		if !typeMismatch {
			matched++
		}
	}

	t.Logf("Results: %d tested, %d matched, %d type-diffs, %d parse-errors, %d skipped",
		total, matched, diffCount, parseErrCount, skipCount)
}

// TestOrdLive: do we parse this demanding
// file the same as python Ivy?
// The python helper cannot load ord_live.ivy
// without an "isolate=cf_live" to check
func TestOrdLive(t *testing.T) {

	// ivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

	path := "/Users/jaten/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("target path not found at %s", path)
	}

	args := []string{"isolate=cf_live"}

	// Get Python AST
	ivyPipe, pyErr := ivy_check(t, args, path)
	if pyErr != nil {
		t.Fatalf("%v had Python error: %v", path, pyErr)
		panic(pyErr)
		return
	}
	if ivyPipe == nil {
		panic("nil pipe but no error?")
	}
	defer ivyPipe.Close()
	ivyR := bufio.NewReader(ivyPipe)

	// Get Go AST, + parse xtrace

	goivyPipe, goErr := goivy_check_xtrace(t, args, path)
	if goErr != nil {
		t.Fatalf("path='%v': Go parse error: %v", path, goErr)
		return
	}
	defer goivyPipe.Close()
	goivyR := bufio.NewReader(goivyPipe)

	// read a line from each, and compare
	// Normalize file paths so that different install locations
	// (e.g. ~/goivy/... vs ~/pyivy/ivy/...) don't cause false diffs.
	normalizeLine := func(line string) string {
		// Strip known path prefixes for include files
		for _, prefix := range []string{
			"/Users/jaten/goivy/ivy-lang-examples/ivy/include/",
			"/Users/jaten/pyivy/ivy/ivy/include/",
			"/Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/ivy/include/",
		} {
			if strings.Contains(line, prefix) {
				line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>/")
			}
		}
		for _, prefix := range []string{
			"/Users/jaten/goivy/ivy-lang-examples/",
			"/Users/jaten/go/src/github.com/glycerine/goivy/ivy-lang-examples/",
		} {
			if strings.Contains(line, prefix) {
				line = strings.ReplaceAll(line, prefix, "<IVY_EXAMPLES>/")
			}
		}
		for _, prefix := range []string{
			"/Users/jaten/pyivy/ivy/ivy/include/",
		} {
			if strings.Contains(line, prefix) {
				line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>/")
			}
		}
		return line
	}

	var goLast10 []string
	var pyLast10 []string

	for i := 0; ; i++ {
		goCheck, err := goivyR.ReadString('\n')
		if err != nil {
			fmt.Printf("stopping on goivy_check_xtrace error %v\n", err)
			return
		}
		ivCheck, err := ivyR.ReadString('\n')
		if err != nil {
			fmt.Printf("stopping on ivy_check error %v\n", err)
			return
		}
		goNorm := normalizeLine(goCheck)
		ivNorm := normalizeLine(ivCheck)

		// on mismatch, report last 10 for context.
		goLast10 = append(goLast10, goNorm)
		if len(goLast10) > 10 {
			goLast10 = goLast10[1:]
		}
		pyLast10 = append(pyLast10, ivNorm)
		if len(pyLast10) > 10 {
			pyLast10 = pyLast10[1:]
		}

		if goNorm != ivNorm {
			n := len(pyLast10)
			if i > 10 {
				fmt.Printf("(omit prior matching xtrace from 0 - %v, for speed...)\n", i-n)
			}
			for j, pys := range pyLast10 {
				fmt.Printf("%05d  go : %v", i-n+j+1, goLast10[j])
				fmt.Printf("       py : %v\n", pys)
			}
			// If both lines are HASH lines with canon= data, show a structured diff.
			if strings.Contains(goNorm, "HASH") && strings.Contains(ivNorm, "HASH") &&
				strings.Contains(goNorm, "canon=") && strings.Contains(ivNorm, "canon=") {
				goCanon := goNorm[strings.Index(goNorm, "canon=")+6:]
				pyCanon := ivNorm[strings.Index(ivNorm, "canon=")+6:]
				goCanon = strings.TrimSpace(goCanon)
				pyCanon = strings.TrimSpace(pyCanon)
				diff := iu.DiffSexp(goCanon, pyCanon)
				if diff != "" {
					fmt.Printf("\n=== S-expression diff (go vs py) ===\n%s\n", diff)
				}
			}
			t.Fatalf("ivy_check and goivy_check differ at line %v, counting from 0.", i)
		}
	}
}

// ivy_check calls ivy_check.
// It streams output back on r, a pipe, asynchronously.
func ivy_check(t *testing.T, args []string, ivyFile string) (r io.ReadCloser, err error) {
	t.Helper()

	_, thisFile, _, _ := runtime.Caller(0)
	ivyRoot := filepath.Join(filepath.Dir(thisFile), "..")

	ivyHomeDir := os.Getenv("IVY_HOME")
	if ivyHomeDir != "" {
		ivyRoot = ivyHomeDir
	}
	//vv("ivyRoot = '%v'", ivyRoot) // /Users/jaten/go/src/github.com/glycerine
	pr, pw := io.Pipe()
	if err != nil {
		panic(err)
	}

	args = append(args, ivyFile)
	cmd := exec.Command("ivy_check", args...)
	cmd.Dir = ivyRoot
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	go func() {
		cmd.Wait()
		pw.Close() // must close write end so reader sees EOF
	}()

	return pr, nil
}

// goivy_check_xtrace re-makes and then runs goivy_check_xtrace.
// It streams output back on r, a pipe, asynchronously.
func goivy_check_xtrace(t *testing.T, args []string, ivyFile string) (r io.ReadCloser, err error) {
	t.Helper()

	_, thisFile, _, _ := runtime.Caller(0)
	goivyRoot := filepath.Join(filepath.Dir(thisFile), "..")

	fmt.Printf("build goivy_check_xtrace so we know it is up to date.\n")
	cmd := exec.Command("make", "tr")
	cmd.Dir = goivyRoot // parent dir.
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Printf("done refreshing goivy_check_xtrace\n\n")

	pr, pw := io.Pipe()
	if err != nil {
		panic(err)
	}

	args = append(args, ivyFile)
	exe := "goivy_check_xtrace"
	cmd = exec.Command(exe, args...)
	cmd.Dir = goivyRoot
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start '%v': %v", exe, err)
	}

	go func() {
		cmd.Wait()
		pw.Close() // must close write end so reader sees EOF
	}()

	return pr, nil
}

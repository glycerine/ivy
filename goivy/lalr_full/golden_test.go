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
	"syscall"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	"github.com/glycerine/ivy/goivy/lexer"
)

// examplesDir returns the absolute path to the ivy-lang-examples/ directory.
func examplesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "ivy-lang-examples")
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

// read a line from each, and compare
// Normalize file paths so that different install locations
// (e.g. ~/goivy/... vs ~/pyivy/ivy/...) don't cause false diffs.
func normalizeLine(repo, line string) string {
	// Strip known path prefixes for include files
	gopath := os.Getenv("GOPATH")
	home := os.Getenv("HOME")
	if gopath == "" {
		gopath = filepath.Join(home, "go")
	}
	for _, prefix := range []string{
		home + "/ivy/ivy-lang-examples/ivy/include/",
		filepath.Join(repo, "/ivy-lang-examples/ivy/include/"),
		home + "/ivy/pyivy/ivy/ivy/include/",
		gopath + "/src/github.com/glycerine/ivy/ivy-lang-examples/ivy/include/",
	} {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>/")
		}
	}
	for _, prefix := range []string{
		home + "/ivy/ivy-lang-examples/",
		filepath.Join(repo, "/ivy-lang-examples/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/ivy-lang-examples/"),
	} {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_EXAMPLES>/")
		}
	}
	for _, prefix := range []string{
		home + "/ivy/pyivy/ivy/ivy/include/",
		filepath.Join(repo, "/ivy/pyivy/ivy/ivy/include/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/pyivy/ivy/ivy/include/"),
	} {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>/")
		}
	}
	return line
}

func mustGetRepoDir(t *testing.T) (dir string) {
	// _, filename, _, ok := runtime.Caller(0)
	// filename is the absolute path to THIS test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("unable to get current .go test file path")
		return
	}

	dir = filepath.Dir(filename)
	dir = filepath.Dir(dir)
	dir = filepath.Dir(dir)
	return
}

// TestOrdLive: do we parse this demanding
// file the same as python Ivy?
// The python helper cannot load ord_live.ivy
// without an "isolate=cf_live" to check
func TestOrdLive(t *testing.T) {
	path := "ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	GoldenPathCompareIvyCheck(t, false, true, path)
}

// TestVerboseOrdLive is the same as TestOrdLive but prints every
// matching trace line, not just the last 10 before the divergence.
func TestVerboseOrdLive(t *testing.T) {
	path := "ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	GoldenPathCompareIvyCheck(t, true, true, path)
}

// TestVerboseNonstopOrdLive does not stop
// at the first divergence. It prints all parsed
// and xtraced lines.
func TestVerboseNonstopOrdLive(t *testing.T) {
	path := "ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	GoldenPathCompareIvyCheck(t, true, false, path)
}

func GoldenPathCompareIvyCheck(t *testing.T, verbose, diffStop bool, repoRelPath string) {
	//return // off to check everything else under make test.
	t.Helper()

	repo := mustGetRepoDir(t)
	path := filepath.Join(repo, repoRelPath)

	// ivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

	if _, err := os.Stat(path); err != nil {
		t.Skipf("target path not found at %s", path)
	}

	args := []string{"isolate=cf_live"}

	// Get Python AST
	ivyPipe, pyProc, pyErr := ivy_check(t, args, path, repo)
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

	goivyPipe, goProc, goErr := goivy_check_xtrace(t, args, path, repo)
	if goErr != nil {
		t.Fatalf("path='%v': Go parse error: %v", path, goErr)
		return
	}
	defer goivyPipe.Close()

	// Kill child process groups when the test exits (pass or fail).
	// Each child runs in its own process group (Setpgid: true), so
	// killing with -pgid terminates the child and all its descendants.
	// Without this, ivy_check and goivy_check_xtrace linger after
	// the test stops early at the first divergence.
	t.Cleanup(func() {
		if pyProc != nil {
			// Kill entire process group: negative PID = process group.
			syscall.Kill(-pyProc.Pid, syscall.SIGKILL)
		}
		if goProc != nil {
			syscall.Kill(-goProc.Pid, syscall.SIGKILL)
		}
	})
	goivyR := bufio.NewReader(goivyPipe)

	var goLast30 []string
	var pyLast30 []string
	var goCheck, ivCheck string
	var err error

	for i := 0; ; i++ {

		for {
			goCheck, err = goivyR.ReadString('\n')
			if err != nil {
				handleEOF(t, "go", goCheck, ivyR, "py", i)
				return
			}
			if strings.HasPrefix(goCheck, "XTRACE:") {
				break
			}
			if showNonXtraceLines {
				// allow stack traces/other debug prints through
				fmt.Printf("~go[after i=%v]: %v", i-1, goCheck)
			}
		}
		for {
			ivCheck, err = ivyR.ReadString('\n')
			if err != nil {
				handleEOF(t, "py", ivCheck, goivyR, "go", i)
				return
			}
			if strings.HasPrefix(ivCheck, "XTRACE:") {
				break
			}
			if showNonXtraceLines {
				// allow stack traces/other debug prints through
				fmt.Printf("~py[after i=%v]: %v", i-1, ivCheck)
			}
		}
		goNorm := normalizeLine(repo, goCheck)
		ivNorm := normalizeLine(repo, ivCheck)

		if verbose {
			fmt.Printf("%06d  go : %v", i, goNorm)
			fmt.Printf("        py : %v\n", ivNorm)
		}

		if !diffStop {
			continue
		}

		// on mismatch, report last 30 for context.
		goLast30 = append(goLast30, goNorm)
		if len(goLast30) > 30 {
			goLast30 = goLast30[1:]
		}
		pyLast30 = append(pyLast30, ivNorm)
		if len(pyLast30) > 30 {
			pyLast30 = pyLast30[1:]
		}

		if goNorm != ivNorm {
			if !verbose {
				n := len(pyLast30)
				if i > 30 {
					fmt.Printf("(omit prior matching xtrace from 0 - %v, for speed...)\n", i-n)
				}
				// note: truncate to first 300 bytes to
				// avoid regurgitating very long canonical
				// strings for modules of matching stuff.
				for j, pys := range pyLast30 {
					if false {
						fmt.Printf("%06d  go : %.300s", i-n+j+1, goLast30[j])
						if len(goLast30[j]) > 300 {
							fmt.Printf(" ...(truncated long line to 300 bytes)\n")
						}
						fmt.Printf("        py : %.300s\n", pys)
						if len(pys) > 300 {
							fmt.Printf(" ...(truncated long line to 300 bytes)\n")
						}
					} else {
						fmt.Printf("%06d  go : %s", i-n+j+1, goLast30[j])
						fmt.Printf("        py : %s\n", pys)
					}
				}
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
					fmt.Printf("\n=== S-expression diff (go '-' vs py '+') ===\n%s\n", diff)
				}
			}

			// show any trailing after \n prints (like stacks) that come
			// before the next XTRACE.

			var sourceShownGo bool
			var sourceShownPy bool
			for {
				goCheck, err = goivyR.ReadString('\n')
				if err != nil {
					break
				}
				if strings.HasPrefix(goCheck, "XTRACE:") {
					break
				}
				if showNonXtraceLines {
					// allow stack traces/other debug prints through
					if sourceShownGo {
						fmt.Printf("%v", goCheck)
					} else {
						sourceShownGo = true
						fmt.Printf("========== trailing ~go[after i=%7d]:\n%v", i, goCheck)
					}
				}
			}
			for {
				ivCheck, err = ivyR.ReadString('\n')
				if err != nil {
					break
				}
				if strings.HasPrefix(ivCheck, "XTRACE:") {
					break
				}
				if showNonXtraceLines {
					// allow stack traces/other debug prints through
					if sourceShownPy {
						fmt.Printf("%v", ivCheck)
					} else {
						sourceShownPy = true
						fmt.Printf("========== trailing ~py[after i=%7d]:\n%v", i, ivCheck)
					}
				}
			}

			t.Fatalf("ivy_check and goivy_check differ at line %v, counting from 0.", i)
		}
	}
}

// handleEOF is called when one process hits EOF. It prints the final
// data from the dead process, shows what the surviving process does
// next, and fails the test.
func handleEOF(t *testing.T, deadName string, deadData string,
	aliveReader *bufio.Reader, aliveName string, i int) {
	t.Helper()
	// 1. Print any data returned alongside the EOF (ReadString returns
	//    partial data before the error — the current code was discarding this).
	if trimmed := strings.TrimSpace(deadData); trimmed != "" {
		fmt.Printf("~%s[final at i=%d]: %s\n", deadName, i, trimmed)
	}
	// 2. Read up to 10 more XTRACE lines from the surviving process
	//    to show where it continues that the dead process didn't.
	fmt.Printf("\n=== %s died at XTRACE line %d, but %s continues: ===\n", deadName, i, aliveName)
	shown := 0
	for shown < 10 {
		line, err := aliveReader.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		if line != "" {
			fmt.Printf("  %s[i=%d+%d]: %s\n", aliveName, i, shown, line)
			if strings.HasPrefix(line, "XTRACE:") {
				shown++
			}
		}
		if err != nil {
			fmt.Printf("  (%s also hit %v)\n", aliveName, err)
			break
		}
	}
	t.Fatalf("%s exited at XTRACE line %d while %s continues", deadName, i, aliveName)
}

const fullXtraceToDir string = ".."

const writeFullLogFile = true

const showNonXtraceLines = true

// ivy_check calls ivy_check.
// It streams output back on r, a pipe, asynchronously.
func ivy_check(t *testing.T, args []string, ivyFile, repo string) (r io.ReadCloser, proc *os.Process, err error) {
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

	var w io.Writer = pw
	var f *os.File
	if writeFullLogFile {
		outPath := filepath.Join(fullXtraceToDir, "out.py.xtrace")
		var ferr error
		f, ferr = os.Create(outPath)
		if ferr != nil {
			t.Fatalf("failed to create %s: %v", outPath, ferr)
		}
		w = io.MultiWriter(pw, f)
	}

	// We need to normalize lines before writing to w, so pipe
	// the command's raw output through a filter goroutine.
	cmdPr, cmdPw := io.Pipe()

	args = append(args, ivyFile)
	cmd := exec.Command("ivy_check", args...)
	cmd.Dir = ivyRoot
	cmd.Stdout = cmdPw
	cmd.Stderr = cmdPw
	// Put the child in its own process group so we can kill all
	// its descendants (including any grandchildren) on cleanup.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	go func() {
		cmd.Wait()
		cmdPw.Close()
	}()

	// Filter goroutine: read raw lines, normalize, write to w.
	go func() {
		scanner := bufio.NewScanner(cmdPr)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := normalizeLine(repo, scanner.Text())
			fmt.Fprintf(w, "%s\n", line)
		}
		pw.Close() // must close write end so reader sees EOF
		if f != nil {
			f.Close()
		}
	}()

	return pr, cmd.Process, nil
}

// goivy_check_xtrace re-makes and then runs goivy_check_xtrace.
// It streams output back on r, a pipe, asynchronously.
func goivy_check_xtrace(t *testing.T, args []string, ivyFile, repo string) (r io.ReadCloser, proc *os.Process, err error) {
	t.Helper()

	_, thisFile, _, _ := runtime.Caller(0)
	// parent dir.
	goivyRoot := filepath.Join(filepath.Dir(thisFile), "..")
	// cmd/goivy_check dir
	goivyCheckCmdDir := filepath.Join(goivyRoot, "cmd", "goivy_check")

	// we will compile goivy_check_xtrace now to make
	// sure it is up-to-date, and place it into the gobin directory.
	gobin := os.Getenv("GOBIN")
	// fallback places; if GOBIN is not set.
	home := os.Getenv("HOME")
	gopath := os.Getenv("GOPATH")
	if gobin == "" {
		switch {
		case gopath != "":
			gobin = filepath.Join(gopath, "bin")
		case home != "":
			gobin = filepath.Join(home, "go", "bin")
			if dirExists(gobin) {
				break
			}
			fallthrough
		default:
			// write to root of repo as last resort.
			gobin = repo
		}
	}
	target := filepath.Join(gobin, "goivy_check_xtrace")
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")
	fmt.Printf("build goivy_check_xtrace so we know it is up to date: 'cd %v && %v build -o %v'\n", goivyCheckCmdDir, goBinary, target)
	cmd := exec.Command(goBinary, "build", "-o", target)
	cmd.Dir = goivyCheckCmdDir
	err = cmd.Run()
	if err != nil {
		panicf("could not run 'make tr' to build goivy_check_xtrace; error: '%v'", err)
	}
	fmt.Printf("done refreshing goivy_check_xtrace\n\n")

	pr, pw := io.Pipe()
	if err != nil {
		panic(err)
	}

	var w io.Writer = pw
	var f *os.File
	if writeFullLogFile {
		outPath := filepath.Join(fullXtraceToDir, "out.go.xtrace")
		var ferr error
		f, ferr = os.Create(outPath)
		if ferr != nil {
			t.Fatalf("failed to create %s: %v", outPath, ferr)
		}
		w = io.MultiWriter(pw, f)
	}

	// We need to normalize lines before writing to w, so pipe
	// the command's raw output through a filter goroutine.
	cmdPr, cmdPw := io.Pipe()

	args = append(args, ivyFile)
	exe := "goivy_check_xtrace"
	cmd = exec.Command(exe, args...)
	cmd.Dir = goivyRoot
	cmd.Stdout = cmdPw
	cmd.Stderr = cmdPw
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start '%v': %v", exe, err)
	}

	go func() {
		cmd.Wait()
		cmdPw.Close()
	}()

	// Filter goroutine: read raw lines, normalize, write to w.
	go func() {
		scanner := bufio.NewScanner(cmdPr)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := normalizeLine(repo, scanner.Text())
			fmt.Fprintf(w, "%s\n", line)
		}
		pw.Close() // must close write end so reader sees EOF
		if f != nil {
			f.Close()
		}
	}()

	return pr, cmd.Process, nil
}

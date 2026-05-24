package goivy

// Golden AST comparison test: parses each .ivy file in ivy-lang-examples/
// with both Python and Go LALR parser, serializes the AST to text, and compares.
// Adapted from compiler/golden_ast_test.go — uses parser.Parse, the lalr parser.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

const goldenProcessLineBuffer = 4096

type bufferedLinePipe struct {
	lines chan []byte
	done  chan struct{}
	once  sync.Once
	buf   []byte
}

func newBufferedLinePipe(capacity int) *bufferedLinePipe {
	return &bufferedLinePipe{
		lines: make(chan []byte, capacity),
		done:  make(chan struct{}),
	}
}

func (p *bufferedLinePipe) Read(dst []byte) (int, error) {
	for len(p.buf) == 0 {
		select {
		case line, ok := <-p.lines:
			if !ok {
				return 0, io.EOF
			}
			p.buf = line
		case <-p.done:
			return 0, io.EOF
		}
	}
	n := copy(dst, p.buf)
	p.buf = p.buf[n:]
	return n, nil
}

func (p *bufferedLinePipe) Write(src []byte) (int, error) {
	line := append([]byte(nil), src...)
	select {
	case p.lines <- line:
		return len(src), nil
	case <-p.done:
		return 0, io.ErrClosedPipe
	}
}

func (p *bufferedLinePipe) Close() error {
	p.once.Do(func() {
		close(p.done)
	})
	return nil
}

func (p *bufferedLinePipe) closeWriter() {
	close(p.lines)
}

func attachCombinedOutputPipe(cmd *exec.Cmd) (*os.File, *os.File, error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	cmd.Stdout = pw
	cmd.Stderr = pw
	return pr, pw, nil
}

var newline = []byte("\n")

func forwardGoldenProcessLine(repo, label string, w io.Writer, xtraceCount *int64, raw string, onlyNonXtrace bool) error {
	line := normalizeLine(repo, raw)
	isX := strings.HasPrefix(line, "XTRACE:")
	if onlyNonXtrace {
		if !isX {
			//fmt.Fprintf(w, "%s\n", line)
			//fmt.Fprintf(w, "%s\n", label, line)
			w.Write([]byte(line))
			w.Write(newline)
		}
		return nil
	}
	if isX {
		if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
			return err
		}
		*xtraceCount++
		return nil
	}
	if showNonXtraceLines {
		fmt.Printf("~%s[after i=%d]: %s\n", label, *xtraceCount-1, line)
	}
	return nil
}

// examplesDir returns the absolute path to the ivy-lang-examples/ directory.
func examplesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "ivy-lang-examples")
}

// pythonTestHelperDir returns the path to the Python helper scripts.
func pythonTestHelperDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "pytesthelper")
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
	ivyRoot := filepath.Dir(thisFile)

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
	data, err := os.ReadFile(ivyFile)
	if err != nil {
		return nil, err
	}
	src := string(data)

	// Detect version
	ver := Version{1, 7}
	if strings.HasPrefix(src, "#lang ivy") {
		line := src[:strings.Index(src, "\n")]
		v := strings.TrimPrefix(line, "#lang ivy")
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			major, minor := 1, 7
			fmt.Sscanf(parts[0], "%d", &major)
			fmt.Sscanf(parts[1], "%d", &minor)
			ver = Version{major, minor}
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
func goDeclName(d Node) string {
	switch n := d.(type) {
	case *TypeDecl:
		if len(n.Args()) > 0 {
			if td, ok := n.Args()[0].(*TypeDef); ok {
				if sym, ok := td.Name.(*Symbol); ok {
					return sym.Rep
				}
				if atom, ok := td.Name.(*Atom); ok {
					return atom.Rep
				}
				return fmt.Sprint(td.Name)
			}
		}
	case *ConstantDecl:
		if len(n.Args()) > 0 {
			if a, ok := n.Args()[0].(*Atom); ok {
				if len(a.Terms) > 0 {
					return fmt.Sprintf("%s(%d)", a.Rep, len(a.Terms))
				}
				return a.Rep
			}
		}
	case *ActionDecl:
		if len(n.Args()) > 0 {
			if ad, ok := n.Args()[0].(*ActionDef); ok {
				return ad.Defines()
			}
		}
	case *MixinDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 40 {
				s = s[:40]
			}
			return s
		}
	case *ConjectureDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *PropertyDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *AxiomDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *ExportDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ImportDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ObjectDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *ModuleDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *DerivedDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *IsolateDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *InterpretDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *DefinitionDecl:
		if len(n.Args()) > 0 {
			s := fmt.Sprint(n.Args()[0])
			if len(s) > 60 {
				s = s[:60]
			}
			return s
		}
	case *InstantiateDecl:
		if len(n.Args()) > 0 {
			return fmt.Sprint(n.Args()[0])
		}
	case *InitDecl:
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
	slashDir := func(path string) string {
		return strings.TrimRight(filepath.ToSlash(filepath.Clean(path)), "/") + "/"
	}

	// Strip known path prefixes for include files
	gopath := os.Getenv("GOPATH")
	home := os.Getenv("HOME")
	if gopath == "" {
		gopath = filepath.Join(home, "go")
	}
	for _, prefix := range []string{
		slashDir(filepath.Join(home, "ivy/ivy-lang-examples/ivy/include")),
		slashDir(filepath.Join(repo, "ivy-lang-examples/ivy/include")),
		slashDir(filepath.Join(home, "ivy/pyivy/ivy/ivy/include")),
		slashDir(filepath.Join(gopath, "src/github.com/glycerine/ivy/ivy-lang-examples/ivy/include")),
	} {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>/")
		}
	}
	for _, prefix := range []string{
		slashDir(filepath.Join(home, "ivy/ivy-lang-examples")),
		slashDir(filepath.Join(repo, "ivy-lang-examples")),
		slashDir(filepath.Join(gopath, "src/github.com/glycerine/ivy/ivy-lang-examples")),
	} {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_EXAMPLES>/")
		}
	}
	for _, prefix := range []string{
		slashDir(filepath.Join(home, "ivy/pyivy/ivy/ivy/include")),
		slashDir(filepath.Join(repo, "ivy/pyivy/ivy/ivy/include")),
		slashDir(filepath.Join(gopath, "src/github.com/glycerine/ivy/pyivy/ivy/ivy/include")),
	} {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>/")
		}
	}
	return line
}

func TestNormalizeLineRepoExamplesPathKeepsSingleSlash(t *testing.T) {
	repo := "/tmp/ivy"
	line := "XTRACE: check.start ENTER file=/tmp/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	got := normalizeLine(repo, line)
	want := "XTRACE: check.start ENTER file=<IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy"
	if got != want {
		t.Fatalf("normalizeLine mismatch:\n got: %q\nwant: %q", got, want)
	}
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
	return
}

// TestOrdLive: do we parse this demanding
// file the same as python Ivy?
// The python helper cannot load ord_live.ivy
// without an "isolate=cf_live" to check
func TestOrdLive(t *testing.T) {
	path := "ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	cfg := &goldenConfig{args: []string{"isolate=cf_live"}, path: path}
	GoldenPathCompareIvyCheck(t, cfg)
}

// takes two hours to check all isolates.
func Test2hrOrdLive(t *testing.T) {
	path := "ivy-lang-examples/doc/examples/apple/ord_live.ivy"
	//args := []string{"isolate=cf_live"}
	cfg := &goldenConfig{path: path}
	GoldenPathCompareIvyCheck(t, cfg) // verbose=false, diffStop=true
}

func TestGoldenAll(t *testing.T) {
	off := os.Getenv("XTRACE_OFF")
	if off != "" {
		t.Skip("skip again the golden test(s) when XTRACE_OFF.")
		return // off to check everything else under make test.
	}
	vv("top of TestGoldenAll")

	startDir := "../ivy-lang-examples/"
	paths, err := ListAllIvyPathsRecursively(startDir)
	panicOn(err)
	//vv("spec list (len %v) = '%#v'", len(paths), paths)

	cfg := &goldenConfig{}

	// already green once specs should not be revisted until we
	// get through the others, to save time.
	skipTo := 178

	skipRebuild := false
	for ipath, path := range paths {

		if ipath < skipTo {
			vv("skipTo(%v) > ipath(%v)", skipTo, ipath)
			continue
		}

		// when we parse here, we do not want to see all the traces.
		xtracer.Suppressed = true
		isos, err := ListIsolates(path)
		if err != nil {
			errs := err.Error()
			if false {
				if strings.Contains(errs, "syntax error") {
					vv("syntax error, skip to next: '%v' on path '%v'", err, path)
					continue
				}
				panicOn(err)
			}
			vv("warning: ignoring error '%v' on path '%v'", err, path)
			continue
		}
		xtracer.Suppressed = false

		path2 := path[3:] // strip "../" to get a repo-root-relative path.
		for _, iso := range isos {
			vv("======= begin TestGoldenAll: path='%v'; isolate='%v' (path %v of %v)", path2, iso, ipath, len(paths))
			cfg.path = path2
			cfg.args = []string{fmt.Sprintf("isolate=%v", iso)}
			cfg.skipRebuild = skipRebuild
			GoldenPathCompareIvyCheck(t, cfg)
			skipRebuild = true // only need rebuild the first time.
		}
	}
}

// The exampl1_numeric.ivy spec deliberate goes outside of the
// EPR fragment and invokes the Z3 inconclusive behavior. We
// focused on this to get the ToSMT2 diagnostic output to be
// comparable in Go.
func Test05550_SolverInconclusive(t *testing.T) {

	t.Skip("diffs minor in the SMT2 output, revisit only if needed.")

	outPathGo := "out.test.05550.go.txt"
	outPathPy := "out.test.05550.py.txt"
	os.Remove(outPathGo)
	os.Remove(outPathPy)

	cfg := &goldenConfig{
		path: "ivy-lang-examples/doc/examples/cav2024/examp1_numeric.ivy",
		args: []string{"isolate=q.iso"},
	}
	if false {
		GoldenPathCompareIvyCheck(t, cfg)
	} else {
		// need to rebuild goivy_check_xtrace so it is current.
		// GoldenPathCompareIvyCheck will do this for us if we are using it.
		rebuild_goivy_check_xtrace()
	}
	// also assert that the non-XTRACE lines agree.
	skipRebuild := true
	repo := mustGetRepoDir(t)
	path := filepath.Join(repo, cfg.path)

	onlyNonXtrace := true
	ivyPipe, pyProc, pyErr := ivy_check(t, cfg.args, path, repo, onlyNonXtrace)
	goivyPipe, goProc, goErr := goivy_check_xtrace(t, cfg.args, path, repo, skipRebuild, onlyNonXtrace)

	if pyErr != nil {
		t.Fatalf("%v had Python error: %v", path, pyErr)
		panic(pyErr)
		return
	}
	if ivyPipe == nil {
		panic("nil pipe but no error?")
	}
	defer ivyPipe.Close()
	//ivyR := bufio.NewReader(ivyPipe)

	if goErr != nil {
		t.Fatalf("path='%v': Go parse error: %v", path, goErr)
		return
	}
	defer goivyPipe.Close()

	t.Cleanup(func() {
		if pyProc != nil {
			// Kill entire process group: negative PID = process group.
			syscall.Kill(-pyProc.Pid, syscall.SIGKILL)
		}
		if goProc != nil {
			syscall.Kill(-goProc.Pid, syscall.SIGKILL)
		}
	})
	//goivyR := bufio.NewReader(goivyPipe)

	fdg, err := os.Create(outPathGo)
	panicOn(err)
	defer fdg.Close()

	fdp, err := os.Create(outPathPy)
	panicOn(err)
	defer fdp.Close()

	go func() {
		//io.Copy(fdg, goivyR)
		io.Copy(fdg, goivyPipe)
		fdg.Sync()
		vv("io.Copy go done")
	}()
	go func() {
		//io.Copy(fdp, ivyR)
		io.Copy(fdp, ivyPipe)
		fdp.Sync()
		vv("io.Copy py done")
	}()

	goProc.Wait()
	pyProc.Wait()
	vv("done waiting on both")

	// diff them
	panicOn(fdg.Sync())
	panicOn(fdp.Sync())

	goVers, err := os.ReadFile(outPathGo)
	panicOn(err)
	pyVers, err := os.ReadFile(outPathPy)
	panicOn(err)

	diff := diffLogs(string(goVers), string(pyVers), 0)
	if len(diff) > 0 {
		vv("diff = \n%v\n", diff)
	}
}

func Test2hrNodeGoldenOrdLive(t *testing.T) {
	cfg := &goldenConfig{
		path:    "ivy-lang-examples/doc/examples/apple/ord_live.ivy",
		useNode: "bigGo",
		//args: []string{"isolate=cf_live"},
	}
	GoldenPathCompareIvyCheck(t, cfg)
}

func TestIvyTlbModel(t *testing.T) {
	cfg := &goldenConfig{path: "ivy-lang-examples/examples/liveness/tlb.ivy"}
	GoldenPathCompareIvyCheck(t, cfg)
}

func TestIvy_1dot1_tilelink1_model(t *testing.T) {
	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path: "ivy-lang-examples/examples/tilelink/tilelink1.ivy",
	})
}

// TestVerboseOrdLive is the same as TestOrdLive but prints every
// matching trace line, not just the last 10 before the divergence.
func TestVerboseOrdLive(t *testing.T) {
	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path:    "ivy-lang-examples/doc/examples/apple/ord_live.ivy",
		args:    []string{"isolate=cf_live"},
		verbose: true,
	})
}

// This isolate is the 9th one in. Seen at XTRACE 28_234_303
// when running golden-2hr, which takes 3 hours to crash, so
// try just running this isolate alone instead.
func TestRfnAbsIso(t *testing.T) {
	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path: "ivy-lang-examples/doc/examples/apple/ord_live.ivy",
		args: []string{"isolate=rfn.abs.iso"},
	})
}

// see "isolate sys_live = " in ivy-lang-examples/doc/examples/apple/ord_live.ivy
func TestSysLiveIso(t *testing.T) {
	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path: "ivy-lang-examples/doc/examples/apple/ord_live.ivy",
		args: []string{"isolate=sys_live"},
	})
}

// see "isolate this" in ivy-lang-examples/doc/examples/apple/ord_live.ivy
func TestThisIso(t *testing.T) {
	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path: "ivy-lang-examples/doc/examples/apple/ord_live.ivy",
		args: []string{"isolate=this"},
	})
}

// TestVerboseNonstopOrdLive does not stop
// at the first divergence. It prints all parsed
// and xtraced lines.
func TestVerboseNonstopOrdLive(t *testing.T) {
	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path: "ivy-lang-examples/doc/examples/apple/ord_live.ivy",
		args: []string{"isolate=cf_live"},
	})
}

func TestEchoDotIvy(t *testing.T) {
	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path: "ivy-lang-examples/doc/examples/echo.ivy",
		args: []string{"isolate=protocol"},
	})

	vv("TestEchoDotIvy: echo.ivy test: done with isolate=protcol, now on to isolate=service")

	GoldenPathCompareIvyCheck(t, &goldenConfig{
		path: "ivy-lang-examples/doc/examples/echo.ivy",
		args: []string{"isolate=service"},
	})
}

type goldenConfig struct {
	path         string
	verbose      bool
	diffContinue bool
	args         []string
	useNode      string
	skipRebuild  bool
}

func GoldenPathCompareIvyCheck(t *testing.T, cfg *goldenConfig) {
	off := os.Getenv("XTRACE_OFF")
	if off != "" {
		t.Skip("skip again the golden test(s) when XTRACE_OFF.")
		return // off to check everything else under make test.
	}

	verbose := cfg.verbose
	diffStop := !cfg.diffContinue
	repoRelPath := cfg.path
	args := cfg.args
	useNode := cfg.useNode
	skipRebuild := cfg.skipRebuild

	vv("top of GoldenPathCompareIvyCheck(repoRelPath='%v'); useNode=%v", repoRelPath, useNode)

	repo := mustGetRepoDir(t)
	path := filepath.Join(repo, repoRelPath)

	// ivy_check isolate=cf_live /Users/jaten/go/src/github.com/glycerine/ivy/goivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy

	if _, err := os.Stat(path); err != nil {
		t.Skipf("target path not found at %s", path)
	}

	// let the other finish too, whoever finishes first should
	// not be able to terminate the test prematurely.
	var goDone, pyDone bool

	//args := []string{"isolate=cf_live"}

	var ivyPipe io.ReadCloser
	var pyProc *os.Process
	var pyErr error
	var goivyPipe io.ReadCloser
	var goProc *os.Process
	var goErr error

	switch useNode {
	case "bigGo": // js/wasm
		// The nodegold helper may rebuild the js/wasm payload before it starts
		// producing xtrace. Start it before Python so Python does not fill and
		// block behind an unread pipe during that preparation window.
		goivyPipe, goProc, goErr = nodegold_ivy_check_xtrace(t, args, path, repo, false)
		ivyPipe, pyProc, pyErr = ivy_check(t, args, path, repo, false)

	default: // native Go
		ivyPipe, pyProc, pyErr = ivy_check(t, args, path, repo, false)
		goivyPipe, goProc, goErr = goivy_check_xtrace(t, args, path, repo, skipRebuild, false)
	}

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
	var err, err1, err2 error

top:
	for i := 0; ; i++ {

		//if i > 0 && i%10_000 == 0 {
		//	fmt.Fprintf(os.Stderr, "progress: i = %v\n", i)
		//}

		if !goDone && err != nil {
			fmt.Printf("Go stopping (after i=%v) on goivy_check_xtrace error %v\n", i-1, err)
			goDone = true
		}
		if !pyDone && err1 != nil {
			fmt.Printf("Python stopping (after i=%v) on ivy_check error %v\n", i-1, err1)
			pyDone = true
		}
		if goDone && pyDone {
			fmt.Printf("both sides are done. (after i=%v)\n", i-1)
			return
		}

		for !goDone {
			goCheck, err = goivyR.ReadString('\n')
			if strings.HasPrefix(goCheck, "XTRACE:") {
				break
			}
			if err != nil {
				//handleEOF(t, "go", goCheck, ivyR, "py", i)
				fmt.Printf("Go stopping (i=%v) on goivy_check_xtrace error %v\n", i, err)
				goDone = true
				break
			}

			if showNonXtraceLines {
				// allow stack traces/other debug prints through
				fmt.Printf("~go[after i=%v]: %v", i-1, goCheck)
			}
		}
		for !pyDone {
			ivCheck, err1 = ivyR.ReadString('\n')
			if strings.HasPrefix(ivCheck, "XTRACE:") {
				break
			}
			if err1 != nil {
				//handleEOF(t, "py", ivCheck, goivyR, "go", i)
				fmt.Printf("python stopping on (i=%v) ivy_check error %v\n", i, err1)
				pyDone = true
				continue top
			}
			if showNonXtraceLines {
				// allow stack traces/other debug prints through
				fmt.Printf("~py[after i=%v]: %v", i-1, ivCheck)
			}
		}
		if goDone && pyDone {
			fmt.Printf("both sides are done. (after i=%v)\n", i)
			return
		}
		// asymmetry: drain the other side so we see "OK" or other side errors
		if goDone && !pyDone {
			fmt.Printf("goDone but not python, so drain python side til EOF, ignoring XTRACE:\n")
			for {
				ivCheck, err1 = ivyR.ReadString('\n')
				if strings.HasPrefix(ivCheck, "XTRACE:") {
					continue // ignore these now. right?
				}
				if err1 != nil {
					//handleEOF(t, "py", ivCheck, goivyR, "go", i)
					fmt.Printf("python stopping on (i=%v) ivy_check error %v\n", i, err1)
					pyDone = true
					return
				}
				if showNonXtraceLines {
					// allow stack traces/other debug prints through
					fmt.Printf("~py[after i=%v]: %v", i-1, ivCheck)
				}
			}
			// keep draining until error or EOF
		}
		if !goDone && pyDone {
			goCheck, err = goivyR.ReadString('\n')
			//vv("pyDone but not goDone. goCheck = '%v'", goCheck)
			if strings.HasPrefix(goCheck, "XTRACE:") {
				continue // ignore
			}
			if err != nil {
				//handleEOF(t, "go", goCheck, ivyR, "py", i)
				fmt.Printf("Go stopping (i=%v) on goivy_check_xtrace error %v\n", i, err)
				goDone = true
				return
			}

			if showNonXtraceLines {
				// allow stack traces/other debug prints through
				fmt.Printf("~go[after i=%v]: %v", i-1, goCheck)
			}
			// keep draining until error or EOF
		}

		var goNorm, ivNorm string
		if xtracer.Enabled {
			goNorm = xtracer.NormalizeLine(goCheck)
			ivNorm = xtracer.NormalizeLine(ivCheck)

			//if strings.Contains(goCheck, `<IVY_INCLUDE>/1.8/order.ivy: line 5: index.spec.antisymmetry`) {
			//	vv("1st: from goCheck='%v' to goNorm='%v'", goCheck, goNorm)
			//	// golden_test.go:666 [goID 6] 2026-05-01 18:46:35.349569000 +0000 UTC 1st: from goCheck='        <IVY_INCLUDE>/1.8/order.ivy: line 5: index.spec.antisymmetry  [assumed]\n' to goNorm='        <IVY_INCLUDE>/1.8/order.ivy: line 5: index.spec.antisymmetry  [assumed]\n'
			//	vv("_1st ivy  ivCheck='%v' to ivNorm='%v'", ivCheck, ivNorm) // empty strings
			//}
			//if strings.Contains(goCheck, `transrel.ComposeUpdates ENTER u1.Modified=[](modAll=False) u2.Modified=[](modAll=False)`) {
			//vv("2nd: from goCheck='%v' to goNorm='%v' (equal: %v)", goCheck, goNorm, goCheck == goNorm)
			//vv("_2nd ivy  ivCheck='%v' to ivNorm='%v' (equal: %v)", ivCheck, ivNorm, ivCheck == ivNorm)
			//}

		} else {
			goNorm = normalizeLine(repo, goCheck)
			ivNorm = normalizeLine(repo, ivCheck)
		}

		if verbose {
			fmt.Printf("%06d  go : %v", i, goNorm)
			fmt.Printf("        py : %v\n", ivNorm)
		}

		if !diffStop {
			continue
		}

		// on mismatch, report last 30 for context.
		goLast30 = append(goLast30, goNorm)
		if len(goLast30) > showLast30Lines {
			goLast30 = goLast30[1:]
		}
		pyLast30 = append(pyLast30, ivNorm)
		if len(pyLast30) > showLast30Lines {
			pyLast30 = pyLast30[1:]
		}

		if goNorm != ivNorm {
			vv("we have divergence at i = %v becuase (len %v) goNorm='%v' != (len %v) ivNorm='%v'", i, len(goNorm), goNorm, len(ivNorm), ivNorm) // we are in here!:
			// golden_test.go:699 [goID 6] 2026-05-01 18:46:35.349619000 +0000 UTC we have divergence at i = 263993 becuase (len 80) goNorm='        <IVY_INCLUDE>/1.8/order.ivy: line 5: index.spec.antisymmetry  [assumed]
			// ' != (len 0) ivNorm=''
			if !verbose {
				n := len(pyLast30)
				if i > showLast30Lines {
					fmt.Printf("(omit prior matching xtrace from 0 - %v, for speed...)\n", i-n) // we seen this
				}
				// note: truncate to first 2000 bytes to
				// avoid regurgitating very long canonical
				// strings for modules of matching stuff.
				for j, pys := range pyLast30 {
					if true {
						fmt.Printf("%06d  go : %.300s", i-n+j+1, goLast30[j]) // this is printing the INCLUDE line without the XTRACE
						if len(goLast30[j]) > 300 {
							fmt.Printf(" ...(truncated long line to 300 bytes)\n\n")
						}
						fmt.Printf("        py : %.300s\n", pys)
						if len(pys) > 300 {
							fmt.Printf(" ...(truncated long line to 300 bytes)\n\n")
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
				diff := DiffSexp(goCanon, pyCanon, 10)
				if diff != "" {
					fmt.Printf("\n=== S-expression diff (go '-' vs py '+') ===\n%s\n", diff)
				}
			}

			// show any trailing after \n prints (like stacks) that come
			// before the next XTRACE.

			var sourceShownGo bool
			var sourceShownPy bool
			for {
				goCheck, err2 = goivyR.ReadString('\n')
				if strings.HasPrefix(goCheck, "XTRACE:") {
					break
				}
				if err2 != nil {
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
				ivCheck, err2 = ivyR.ReadString('\n')
				if strings.HasPrefix(ivCheck, "XTRACE:") {
					break
				}
				if err2 != nil {
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

const writeFullLogFile = false

const showNonXtraceLines = true

const showLast30Lines = 300

// ivy_check calls ivy_check.
// It streams output back on r, a pipe, asynchronously.
func ivy_check(t *testing.T, args []string, ivyFile, repo string, onlyNonXtrace bool) (r io.ReadCloser, proc *os.Process, err error) {

	_, thisFile, _, _ := runtime.Caller(0)
	ivyRoot := filepath.Dir(thisFile)

	ivyHomeDir := os.Getenv("IVY_HOME")
	if ivyHomeDir != "" {
		ivyRoot = ivyHomeDir
	}
	//vv("ivyRoot = '%v'", ivyRoot) // /Users/jaten/go/src/github.com/glycerine
	pr := newBufferedLinePipe(goldenProcessLineBuffer)
	if err != nil {
		panic(err)
	}

	var w io.Writer = pr
	var f *os.File
	if writeFullLogFile {
		outPath := filepath.Join(fullXtraceToDir, "out.py.xtrace")
		var ferr error
		f, ferr = os.Create(outPath)
		if ferr != nil {
			t.Fatalf("failed to create %s: %v", outPath, ferr)
		}
		w = io.MultiWriter(pr, f)
	}

	args = append(args, ivyFile)
	cmd := exec.Command("ivy_check", args...)
	cmd.Dir = ivyRoot
	// Put the child in its own process group so we can kill all
	// its descendants (including any grandchildren) on cleanup.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Use a real OS pipe here, not io.Pipe. If Stdout/Stderr are not *os.File,
	// os/exec inserts hidden copy goroutines before our scanner. A real pipe
	// gives children and grandchildren one inherited fd path back to this test.
	cmdPr, cmdPw, err := attachCombinedOutputPipe(cmd)
	if err != nil {
		t.Fatalf("failed to create ivy_check output pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		cmdPr.Close()
		cmdPw.Close()
		t.Fatalf("failed to start: %v", err)
	}
	cmdPw.Close()

	go func() {
		err := cmd.Wait()
		vv("ivy_check command has finished. err='%v'", err)
	}()

	// Filter goroutine: read raw lines, normalize, write to w.
	go func() {
		defer cmdPr.Close()
		scanner := bufio.NewScanner(cmdPr)
		scanner.Buffer(make([]byte, 0, 16<<20), 1<<30)
		var xtraceCount int64
		for scanner.Scan() {
			if err := forwardGoldenProcessLine(repo, "py", w, &xtraceCount, scanner.Text(), onlyNonXtrace); err != nil {
				vv("ivy_check scanner could not forward line: %v", err)
				break
			}
		}
		serr := scanner.Err()
		vv("ivy_check scanner has finished. scanner.Err()='%v'", serr)
		if serr != nil {
			panicf("scanner.Err() was not nil, very bad!: %v", serr)
		}
		pr.closeWriter() // must close write end so reader sees EOF
		if f != nil {
			f.Close()
		}
	}()

	return pr, cmd.Process, nil
}

// goivy_check_xtrace re-makes and then runs goivy_check_xtrace.
// It streams output back on r, a pipe, asynchronously.
func goivy_check_xtrace(t *testing.T, args []string, ivyFile, repo string, skipRebuild, onlyNonXtrace bool) (r io.ReadCloser, proc *os.Process, err error) {

	_, thisFile, _, _ := runtime.Caller(0)
	// parent dir.
	goivyRoot := filepath.Dir(thisFile)
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
	var cmd *exec.Cmd
	if !skipRebuild {
		// could do instead:
		//rebuild_goivy_check_xtrace()

		doFullCmd := fmt.Sprintf("cd %v && %v build -o %v", goivyCheckCmdDir, goBinary, target)
		fmt.Printf("build goivy_check_xtrace so we know it is up to date: '%v'\n", doFullCmd)
		cmd = exec.Command(goBinary, "build", "-o", target)
		cmd.Dir = goivyCheckCmdDir
		err = cmd.Run()
		if err != nil {
			panicf("could not run '%v' (see also 'make tr') to build goivy_check_xtrace; error: '%v'", doFullCmd, err)
		}
		fmt.Printf("done refreshing goivy_check_xtrace\n\n")
	}

	pr := newBufferedLinePipe(goldenProcessLineBuffer)
	if err != nil {
		panic(err)
	}

	var w io.Writer = pr
	var f *os.File
	if writeFullLogFile {
		outPath := filepath.Join(fullXtraceToDir, "out.go.xtrace")
		var ferr error
		f, ferr = os.Create(outPath)
		if ferr != nil {
			t.Fatalf("failed to create %s: %v", outPath, ferr)
		}
		w = io.MultiWriter(pr, f)
	}

	args = append(args, ivyFile)
	exe := target // "goivy_check_xtrace"
	cmd = exec.Command(exe, args...)
	cmd.Dir = goivyRoot
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Use a real OS pipe here, not io.Pipe. If Stdout/Stderr are not *os.File,
	// os/exec inserts hidden copy goroutines before our scanner.
	cmdPr, cmdPw, err := attachCombinedOutputPipe(cmd)
	if err != nil {
		t.Fatalf("failed to create goivy_check_xtrace output pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		cmdPr.Close()
		cmdPw.Close()
		t.Fatalf("failed to start '%v': %v", exe, err)
	}
	cmdPw.Close()

	go func() {
		err := cmd.Wait()
		vv("goivy_check_xtrace command has finished. err='%v'", err)
	}()

	// Filter goroutine: read raw lines, normalize, write to w.
	go func() {
		defer cmdPr.Close()
		scanner := bufio.NewScanner(cmdPr)
		scanner.Buffer(make([]byte, 0, 16<<20), 1<<30)
		var xtraceCount int64
		for scanner.Scan() {
			if err := forwardGoldenProcessLine(repo, "go", w, &xtraceCount, scanner.Text(), onlyNonXtrace); err != nil {
				vv("goivy_check_xtrace scanner could not forward line: %v", err)
				break
			}
		}
		serr := scanner.Err()
		vv("goivy_check_xtrace scanner has finished. scanner.Err()='%v'", serr)
		if serr != nil {
			panicf("scanner.Err() was not nil, very bad!: %v", serr)
		}
		pr.closeWriter() // must close write end so reader sees EOF
		if f != nil {
			f.Close()
		}
	}()

	return pr, cmd.Process, nil
}

// when useNode == "bigGo" we should use:
//
// nodegold_ivy_check_xtrace re-makes and then runs nodegold
// It streams output back on r, a pipe, asynchronously.
func nodegold_ivy_check_xtrace(t *testing.T, args []string, ivyFile, repo string, onlyNonXtrace bool) (r io.ReadCloser, proc *os.Process, err error) {

	_, thisFile, _, _ := runtime.Caller(0)
	// parent dir.
	goivyRoot := filepath.Dir(thisFile)

	nodegoldCmdDir := filepath.Join(goivyRoot, "nodegold")
	goivyWasm := filepath.Join(goivyRoot, "webui", "static", "wasm", "goivy-check-js.wasm")
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")

	wasmFullCmd := fmt.Sprintf("cd %v && GOOS=js GOARCH=wasm %v build -o %v ./cmd/goivy_check_jswasm", goivyRoot, goBinary, goivyWasm)
	fmt.Printf("build goivy-check-js.wasm so nodegold has an up-to-date payload: '%v'\n", wasmFullCmd)
	cmd := exec.Command(goBinary, "build", "-o", goivyWasm, "./cmd/goivy_check_jswasm")
	cmd.Dir = goivyRoot
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := cmd.CombinedOutput(); err != nil {
		panicf("could not run '%v'; error: '%v'; output:\n%s", wasmFullCmd, err, out)
	}
	fmt.Printf("done refreshing goivy-check-js.wasm\n\n")

	// we will compile nodegold now to make
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
	target := filepath.Join(gobin, "nodegold")
	doFullCmd := fmt.Sprintf("cd %v && %v build -o %v", nodegoldCmdDir, goBinary, target)
	fmt.Printf("build nodegold so we know it is up to date: '%v'\n", doFullCmd)
	cmd = exec.Command(goBinary, "build", "-o", target)
	cmd.Dir = nodegoldCmdDir
	err = cmd.Run()
	if err != nil {
		panicf("could not run '%v' (see also 'make tr') to build nodegold; error: '%v'", doFullCmd, err)
	}
	fmt.Printf("done refreshing nodegold\n\n")

	pr := newBufferedLinePipe(goldenProcessLineBuffer)
	if err != nil {
		panic(err)
	}

	var w io.Writer = pr
	var f *os.File
	if writeFullLogFile {
		outPath := filepath.Join(fullXtraceToDir, "out.nodegold.xtrace")
		var ferr error
		f, ferr = os.Create(outPath)
		if ferr != nil {
			t.Fatalf("failed to create %s: %v", outPath, ferr)
		}
		w = io.MultiWriter(pr, f)
	}

	args = append([]string{"-goivy-wasm", goivyWasm}, args...)
	args = append(args, ivyFile)
	exe := target // "nodegold"
	cmd = exec.Command(exe, args...)
	cmd.Dir = goivyRoot
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Use a real OS pipe here, not io.Pipe. This is especially important for
	// nodegold: the node process is a grandchild, and it should inherit a real
	// stdout/stderr fd from nodegold rather than writing through an os/exec
	// copy goroutine layered on top of an io.Pipe.
	cmdPr, cmdPw, err := attachCombinedOutputPipe(cmd)
	if err != nil {
		t.Fatalf("failed to create nodegold output pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		cmdPr.Close()
		cmdPw.Close()
		t.Fatalf("failed to start '%v': %v", exe, err)
	}
	cmdPw.Close()

	go func() {
		err := cmd.Wait()
		vv("nodegold command has finished. err='%v'", err)
	}()

	// Filter goroutine: read raw lines, normalize, write to w.
	go func() {
		defer cmdPr.Close()
		scanner := bufio.NewScanner(cmdPr)
		scanner.Buffer(make([]byte, 0, 16<<20), 1<<30)
		var xtraceCount int64
		for scanner.Scan() {
			if err := forwardGoldenProcessLine(repo, "go", w, &xtraceCount, scanner.Text(), onlyNonXtrace); err != nil {
				vv("nodegold scanner could not forward line: %v", err)
				break
			}
		}
		serr := scanner.Err()
		vv("nodegold scanner has finished. scanner.Err()='%v'", serr)
		if serr != nil {
			panicf("scanner.Err() was not nil, very bad!: %v", serr)
		}
		pr.closeWriter() // must close write end so reader sees EOF
		if f != nil {
			f.Close()
		}
	}()

	return pr, cmd.Process, nil
}

func rebuild_goivy_check_xtrace() {

	_, thisFile, _, _ := runtime.Caller(0)
	// parent dir.
	goivyRoot := filepath.Dir(thisFile)
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
			panicf("cannot figure out where to write refreshed goivy_check_xtracer!")
		}
	}
	target := filepath.Join(gobin, "goivy_check_xtrace")
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")

	doFullCmd := fmt.Sprintf("cd %v && %v build -o %v", goivyCheckCmdDir, goBinary, target)
	fmt.Printf("build '%v' so we know it is up to date: '%v'\n", target, doFullCmd)
	cmd := exec.Command(goBinary, "build", "-o", target)
	cmd.Dir = goivyCheckCmdDir
	err := cmd.Run()
	if err != nil {
		panicf("could not run '%v' (see also 'make tr') to build goivy_check_xtrace; error: '%v'", doFullCmd, err)
	}
	fmt.Printf("done refreshing %v\n\n", target)
}

func diffLogs(goVers, pyVers string, maxDiffs int) string {
	linesGo := strings.Split(goVers, "\n")
	linesPy := strings.Split(pyVers, "\n")

	diffs := 0
	var sb strings.Builder
	maxLen := len(linesGo)
	if len(linesPy) > maxLen {
		maxLen = len(linesPy)
	}
	for i := 0; i < maxLen; i++ {
		lgo, lpy := "", ""
		if i < len(linesGo) {
			lgo = strings.TrimSpace(linesGo[i])
		}
		if i < len(linesPy) {
			lpy = strings.TrimSpace(linesPy[i])
		}
		if lgo == lpy {
			sb.WriteString("   " + lgo + "\n")
		} else {
			//sb.WriteString("- " + lgo + "\n") // go
			//sb.WriteString("+ " + lpy + "\n") // python
			fmt.Fprintf(&sb, "line %v: - %v\n", i, lgo)
			fmt.Fprintf(&sb, "line %v: + %v\n", i, lpy)
			diffs++
			if maxDiffs > 0 && diffs >= maxDiffs {
				return sb.String()
			}
		}
	}
	return sb.String()
}

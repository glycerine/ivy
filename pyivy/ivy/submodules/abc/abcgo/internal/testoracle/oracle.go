// Package testoracle lets tests compare the port against the original ABC (the
// C sources, programs compiled from them, and the native abc binary) without
// needing any of it at test time.
//
// Tests start external programs with testoracle.Command instead of exec.Command,
// look them up with testoracle.LookPath, and read files that live outside the
// module (the C sources, shared benchmarks) with testoracle.ReadFile.
//
// Replay (the default) needs nothing but the module: every command returns
// the stdout, stderr and exit status recorded for it, the files it wrote are
// written again (into the test's current temporary directories or the
// package directory), and testoracle.ReadFile returns the recorded contents.
// Recordings live in each package's test_vectors/oracle/<TestName>.json.
//
// Recording (ABCGO_ORACLE_RECORD=1) runs the real commands and reads the real
// files, and saves what they produced. It needs the C tree next to the module
// (../src, as in the ivy repository) and, for native comparisons, ABC_BINARY
// set to the native abc. Re-record after changing a test's commands, inputs or
// the C reference:
//
//	ABCGO_ORACLE_RECORD=1 ABC_BINARY=$PWD/.scratch/abc-release/abc go test ./pkg -run TestX
package testoracle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unicode/utf8"
)

// Recording reports whether this run records (ABCGO_ORACLE_RECORD=1) rather
// than replays.
func Recording() bool { return os.Getenv("ABCGO_ORACLE_RECORD") == "1" }

// ExitError is the error of a command that exited with a nonzero status,
// recorded or real.
type ExitError struct {
	Code    int
	Message string
	Status  syscall.WaitStatus // the raw wait status (signals included)
}

// Sys returns the wait status, as exec.ExitError's ProcessState.Sys does.
func (e *ExitError) Sys() any { return e.Status }

func (e *ExitError) Error() string { return e.Message }

// ExitCode returns the command's exit status.
func (e *ExitError) ExitCode() int { return e.Code }

// Cmd is the subset of exec.Cmd the tests use.
type Cmd struct {
	Path   string
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Dir    string
	Env    []string

	t        testing.TB
	ctx      context.Context
	expired  bool
	exitCode int
	ran      bool
}

// Command is exec.Command for tests: t names the recording.
func Command(t testing.TB, name string, arg ...string) *Cmd {
	t.Helper()
	return &Cmd{Path: name, Args: append([]string{name}, arg...), t: t}
}

// CommandContext is exec.CommandContext for tests. While recording, the
// context bounds the run (a killed run is recorded as such); replay ignores it.
func CommandContext(t testing.TB, ctx context.Context, name string, arg ...string) *Cmd {
	t.Helper()
	c := Command(t, name, arg...)
	c.ctx = ctx
	return c
}

// LookPath is exec.LookPath while recording; in replay every program
// "exists" (its runs are replayed).
func LookPath(file string) (string, error) {
	if Recording() {
		return exec.LookPath(file)
	}
	return file, nil
}

// NativeABC returns the native abc binary: ABC_BINARY while recording (the
// test is skipped without it), a placeholder in replay.
func NativeABC(t testing.TB) string {
	t.Helper()
	bin := NativeABCPath()
	if bin == "" {
		t.Skip("recording native comparisons needs ABC_BINARY (a native abc built from ../src)")
	}
	return bin
}

// NativeABCPath stands for os.Getenv("ABC_BINARY"): the absolute ABC_BINARY
// while recording ("" when unset), a placeholder path in replay (Stat reports
// it as an executable; its runs are replayed).
func NativeABCPath() string {
	if !Recording() {
		return nativePlaceholder()
	}
	bin := os.Getenv("ABC_BINARY")
	if bin == "" {
		return ""
	}
	abs, err := filepath.Abs(bin)
	if err != nil {
		return bin
	}
	return abs
}

func nativePlaceholder() string {
	return filepath.Join(os.TempDir(), "abcgo-oracle-replay", "abc")
}

// ContextExpired reports whether CommandContext's context had ended when the
// run finished (recorded, so it holds in replay too).
func (c *Cmd) ContextExpired() bool { return c.expired }

// ExitCode is the exit status of the last run (0 before any run).
func (c *Cmd) ExitCode() int { return c.exitCode }

// String renders the command line.
func (c *Cmd) String() string { return strings.Join(c.Args, " ") }

// Run runs the command.
func (c *Cmd) Run() error { return c.run("run", c.Stdout, c.Stderr) }

// Output runs the command and returns its stdout.
func (c *Cmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, fmt.Errorf("oracle: Stdout already set")
	}
	var out bytes.Buffer
	err := c.run("output", &out, c.Stderr)
	return out.Bytes(), err
}

// CombinedOutput runs the command and returns stdout and stderr together.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil || c.Stderr != nil {
		return nil, fmt.Errorf("oracle: Stdout or Stderr already set")
	}
	var out bytes.Buffer
	err := c.run("combined", &out, &out)
	return out.Bytes(), err
}

// entry is one recorded command run or file read.
type entry struct {
	Key     string          `json:"key"`
	Stdout  blob            `json:"stdout,omitempty"`
	Stderr  blob            `json:"stderr,omitempty"`
	Exit    int             `json:"exit,omitempty"`
	Wait    uint32          `json:"wait,omitempty"`    // raw wait status
	Expired bool            `json:"expired,omitempty"` // the context ended the run
	Temps   []string        `json:"temps,omitempty"`   // $TMP/NNN named by the command, in order
	ErrText string          `json:"err,omitempty"`
	Files   map[string]blob `json:"files,omitempty"` // written by the command, normalized path -> contents
	Empty   []string        `json:"empty,omitempty"` // compiler outputs: recreated empty and executable
}

type recording struct {
	mu      sync.Mutex
	file    string
	entries []*entry
	used    map[string]int // replay: next occurrence per key
	dirty   bool
}

// recordDir and recordName place recordings (tests of this package change them).
var (
	recordDir  = filepath.Join("test_vectors", "oracle")
	recordName = func(t testing.TB) string { return t.Name() }
)

var (
	recordingsMu sync.Mutex
	recordings   = map[testing.TB]*recording{}
)

func sanitize(name string) string {
	return regexp.MustCompile(`[^A-Za-z0-9_.-]+`).ReplaceAllString(name, "_")
}

// recordingFor loads (replay) or starts (record) t's recording.
func recordingFor(t testing.TB) *recording {
	recordingsMu.Lock()
	defer recordingsMu.Unlock()
	if r := recordings[t]; r != nil {
		return r
	}
	dir := recordDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(startDir, dir) // tests may t.Chdir
	}
	r := &recording{file: filepath.Join(dir, sanitize(recordName(t))+".json"), used: map[string]int{}}
	if !Recording() {
		if data, err := readRecording(r.file); err == nil {
			if err := json.Unmarshal(data, &r.entries); err != nil {
				t.Fatalf("oracle: %s: %v", r.file, err)
			}
		}
	}
	recordings[t] = r
	t.Cleanup(func() {
		recordingsMu.Lock()
		delete(recordings, t)
		recordingsMu.Unlock()
		if !Recording() || !r.dirty {
			return
		}
		data, err := json.MarshalIndent(r.entries, "", " ")
		if err == nil {
			err = writeRecording(r.file, append(data, '\n'))
		}
		if err != nil {
			t.Errorf("oracle: writing %s: %v", r.file, err)
		}
	})
	return r
}

// next returns the next recorded entry for key (occurrences in order).
func (r *recording) next(t testing.TB, key string) *entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	want := r.used[key]
	r.used[key] = want + 1
	seen := 0
	for _, e := range r.entries {
		if e.Key == key {
			if seen == want {
				return e
			}
			seen++
		}
	}
	t.Fatalf("oracle: nothing recorded for %q (occurrence %d) in %s; record with ABCGO_ORACLE_RECORD=1 (needs ../src and ABC_BINARY)", key, want, r.file)
	return nil
}

func (r *recording) add(e *entry) {
	r.mu.Lock()
	r.entries = append(r.entries, e)
	r.dirty = true
	r.mu.Unlock()
}

// Path normalization: the module root, the C tree next to it, the native
// binary, the temporary directories and the working directory are replaced by
// placeholders, so recordings are independent of the machine and the run.

var (
	modRootOnce sync.Once
	modRoot     string
)

// startDir is the package directory: the working directory when the test
// binary starts, before any test changes it.
var startDir, _ = os.Getwd()

func moduleRoot() string {
	modRootOnce.Do(func() {
		dir := startDir
		if dir == "" {
			return
		}
		for d := dir; ; d = filepath.Dir(d) {
			if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
				modRoot = d
				return
			}
			if filepath.Dir(d) == d {
				return
			}
		}
	})
	return modRoot
}

// placeholders returns (real prefix, placeholder) pairs, longest first.
func placeholders() [][2]string {
	var pairs [][2]string
	add := func(real, ph string) {
		if real != "" && real != "/" {
			pairs = append(pairs, [2]string{filepath.Clean(real), ph})
		}
	}
	if root := moduleRoot(); root != "" {
		add(filepath.Join(filepath.Dir(root), "src"), "$SRC")
		add(filepath.Dir(root), "$ABCROOT")
		add(root, "$MOD")
	}
	if Recording() {
		if bin := os.Getenv("ABC_BINARY"); bin != "" {
			if abs, err := filepath.Abs(bin); err == nil {
				add(abs, "$ABC")
			}
		}
		add(NativeLibABCPath(), "$LIBABC")
	} else {
		add(nativePlaceholder(), "$ABC")
		add(NativeLibABCPath(), "$LIBABC")
	}
	sort.Slice(pairs, func(i, j int) bool { return len(pairs[i][0]) > len(pairs[j][0]) })
	return pairs
}

var (
	tempOnce    sync.Once
	tempPattern *regexp.Regexp // t.TempDir(): <tmp>/<TestName><random>/<NNN>
)

func tempRegexp() *regexp.Regexp {
	tempOnce.Do(func() {
		tempPattern = regexp.MustCompile(regexp.QuoteMeta(filepath.Clean(os.TempDir())) + `/[A-Za-z0-9_-]*?[0-9]+/([0-9]{3})`)
	})
	return tempPattern
}

// tempDirsOf lists the t.TempDir() directories a command names (arguments,
// directory, input), each once, in order.
func (c *Cmd) tempDirsOf(stdin []byte) []string {
	var dirs []string
	seen := map[string]bool{}
	for _, s := range append(append(append([]string{}, c.Args...), c.Dir), string(stdin)) {
		for _, m := range tempRegexp().FindAllString(s, -1) {
			if !seen[m] {
				seen[m] = true
				dirs = append(dirs, m)
			}
		}
	}
	return dirs
}

// normalize replaces machine- and run-specific paths by placeholders.
func normalize(s string) string {
	tmp := filepath.Clean(os.TempDir())
	s = tempRegexp().ReplaceAllString(s, "$$TMP/$1")
	// os.MkdirTemp names: <tmp>/<pattern><digits>
	s = regexp.MustCompile(regexp.QuoteMeta(tmp)+`/([A-Za-z_.-]*?)[0-9]{6,}`).ReplaceAllString(s, "$$TMPDIR_$1")
	if wd, err := os.Getwd(); err == nil {
		s = strings.ReplaceAll(s, wd+"/", "$CWD/")
		if strings.HasSuffix(s, wd) {
			s = strings.TrimSuffix(s, wd) + "$CWD"
		}
	}
	for _, p := range placeholders() {
		s = strings.ReplaceAll(s, p[0], p[1])
	}
	return s
}

func (c *Cmd) key(mode string, stdin []byte) string {
	var b strings.Builder
	b.WriteString(mode)
	b.WriteString(": ")
	for i, a := range c.Args {
		if i == 0 {
			// programs of the test (fixtures it built) keep their
			// normalized path; system programs and the native abc are
			// named by base name (LookPath gives /usr/bin/cc only when
			// recording)
			if n := normalize(a); strings.HasPrefix(n, "$") && n != "$ABC" {
				a = n
			} else {
				a = filepath.Base(a)
			}
		} else {
			a = normalize(a)
		}
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(a)
	}
	if c.Dir != "" {
		b.WriteString(" [dir " + normalize(c.Dir) + "]")
	}
	if len(stdin) > 0 {
		sum := sha256.Sum256([]byte(tempNumbers.ReplaceAllString(normalize(string(stdin)), "$$TMP")))
		b.WriteString(" [stdin " + hex.EncodeToString(sum[:8]) + "]")
	}
	// t.TempDir() numbers depend on the order a test creates its
	// directories (map-ordered cases vary): keys omit them
	return tempNumbers.ReplaceAllString(b.String(), "$$TMP")
}

var tempNumbers = regexp.MustCompile(`\$TMP/[0-9]{3}`)

func isCompiler(name string) bool {
	// build tools: their outputs (objects, archives, programs) are not
	// recorded, only recreated empty, since replay never runs them
	switch filepath.Base(name) {
	case "cc", "gcc", "clang", "c++", "g++", "clang++", "ar", "ld", "make", "ranlib":
		return true
	}
	return false
}

// watchedDirs are the directories whose new or changed files a command's
// recording keeps: temporary directories named in its arguments, its Dir, and
// the working directory.
func (c *Cmd) watchedDirs() []string {
	tmp := filepath.Clean(os.TempDir())
	set := map[string]bool{}
	for _, a := range append(append([]string{}, c.Args[1:]...), c.Dir) {
		for _, m := range regexp.MustCompile(regexp.QuoteMeta(tmp)+`/[^\s"'=:,;()]+`).FindAllString(a, -1) {
			// the top temporary directory of the path
			rel := strings.TrimPrefix(m, tmp+"/")
			parts := strings.Split(rel, "/")
			d := filepath.Join(tmp, parts[0])
			if len(parts) > 1 && regexp.MustCompile(`^[0-9]{3}$`).MatchString(parts[1]) {
				d = filepath.Join(d, parts[1])
			}
			if st, err := os.Stat(d); err == nil && st.IsDir() {
				set[d] = true
			}
		}
	}
	if wd, err := os.Getwd(); err == nil {
		set[wd] = true
	}
	if c.Dir != "" {
		if abs, err := filepath.Abs(c.Dir); err == nil {
			set[abs] = true
		}
	}
	return slices.Sorted(maps.Keys(set))
}

type fileState struct {
	mod  time.Time
	size int64
}

func snapshot(dirs []string, cwd string) map[string]fileState {
	snap := map[string]fileState{}
	for _, d := range dirs {
		walk := func(p string, de fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if de.IsDir() {
				// the working directory: its own files only
				if d == cwd && p != d {
					return filepath.SkipDir
				}
				return nil
			}
			if info, err := de.Info(); err == nil {
				snap[p] = fileState{info.ModTime(), info.Size()}
			}
			return nil
		}
		_ = filepath.WalkDir(d, walk)
	}
	return snap
}

func (c *Cmd) run(mode string, stdout, stderr io.Writer) error {
	t := c.t
	t.Helper()
	if c.ran {
		return fmt.Errorf("oracle: command already run")
	}
	c.ran = true
	var stdin []byte
	if c.Stdin != nil {
		data, err := io.ReadAll(c.Stdin)
		if err != nil {
			return err
		}
		stdin = data
	}
	if isSelf(c.Path) {
		// the test binary re-run as a subprocess tests Go: always live
		_, err := c.execute(stdin, mode, stdout, stderr)
		return err
	}
	key := c.key(mode, stdin)
	r := recordingFor(t)
	if Recording() {
		return c.record(r, key, mode, stdin, stdout, stderr)
	}
	e := r.next(t, key)
	// the recording's temporary directories are this run's, in the order
	// the command names them
	temps := map[string]string{}
	current := c.tempDirsOf(stdin)
	for i, rec := range e.Temps {
		if i < len(current) {
			temps[rec] = current[i]
		}
	}
	if stdout != nil {
		_, _ = stdout.Write([]byte(denormText(t, string(e.Stdout), temps)))
	}
	if stderr != nil && mode != "combined" {
		_, _ = stderr.Write([]byte(denormText(t, string(e.Stderr), temps)))
	}
	for _, p := range slices.Sorted(maps.Keys(e.Files)) {
		data := e.Files[p]
		real := denormalize(t, p, temps)
		if err := os.MkdirAll(filepath.Dir(real), 0o755); err != nil {
			t.Fatalf("oracle: %v", err)
		}
		if isText(data) {
			data = []byte(denormText(t, string(data), temps))
		}
		if err := os.WriteFile(real, data, 0o644); err != nil {
			t.Fatalf("oracle: %v", err)
		}
	}
	for _, p := range e.Empty {
		real := denormalize(t, p, temps)
		_ = os.MkdirAll(filepath.Dir(real), 0o755)
		_ = os.WriteFile(real, nil, 0o755)
	}
	c.exitCode, c.expired = e.Exit, e.Expired
	if e.ErrText != "" {
		return &ExitError{Code: e.Exit, Message: e.ErrText, Status: syscall.WaitStatus(e.Wait)}
	}
	return nil
}

// isText reports whether a written file is text (paths in it are
// normalized): UTF-8 with no NUL byte before the last (super ends its
// library with one). Binary formats such as AIGER (length-prefixed,
// NUL-terminated names followed by a comment) are kept byte for byte.
func isText(data []byte) bool {
	return utf8.Valid(data) && (len(data) == 0 || bytes.IndexByte(data[:len(data)-1], 0) < 0)
}

// isSelf reports whether path is the running test binary.
func isSelf(path string) bool {
	if path == os.Args[0] {
		return true
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(path)
	return err == nil && abs == exe
}

// execute runs the command for real, writing to stdout and stderr.
func (c *Cmd) execute(stdin []byte, mode string, stdout, stderr io.Writer) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	if c.ctx != nil {
		cmd = exec.CommandContext(c.ctx, c.Path, c.Args[1:]...)
	} else {
		cmd = exec.Command(c.Path, c.Args[1:]...)
	}
	cmd.Dir, cmd.Env = c.Dir, c.Env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	c.expired = c.ctx != nil && c.ctx.Err() != nil
	if x, ok := err.(*exec.ExitError); ok {
		c.exitCode = x.ExitCode()
		ws, _ := x.Sys().(syscall.WaitStatus)
		return cmd, &ExitError{Code: x.ExitCode(), Message: err.Error(), Status: ws}
	}
	return cmd, err
}

func (c *Cmd) record(r *recording, key, mode string, stdin []byte, stdout, stderr io.Writer) error {
	var cmd *exec.Cmd
	if c.ctx != nil {
		cmd = exec.CommandContext(c.ctx, c.Path, c.Args[1:]...)
	} else {
		cmd = exec.Command(c.Path, c.Args[1:]...)
	}
	cmd.Dir, cmd.Env = c.Dir, c.Env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var out, errb bytes.Buffer
	if mode == "combined" {
		cmd.Stdout, cmd.Stderr = &out, &out
	} else {
		cmd.Stdout, cmd.Stderr = &out, &errb
	}
	cwd, _ := os.Getwd()
	dirs := c.watchedDirs()
	before := snapshot(dirs, cwd)
	runErr := cmd.Run()
	after := snapshot(dirs, cwd)
	temps := c.tempDirsOf(stdin)
	e := &entry{Key: key, Stdout: blob(normalizeOut(out.String(), temps)), Expired: c.ctx != nil && c.ctx.Err() != nil}
	for _, d := range temps {
		e.Temps = append(e.Temps, normalize(d))
	}
	if mode != "combined" {
		e.Stderr = blob(normalizeOut(errb.String(), temps))
	}
	if runErr != nil {
		e.ErrText = runErr.Error()
		if x, ok := runErr.(*exec.ExitError); ok {
			e.Exit = x.ExitCode()
			if ws, ok := x.Sys().(syscall.WaitStatus); ok {
				e.Wait = uint32(ws)
			}
		} else {
			e.Exit = -1
		}
	}
	var changed []string
	for _, p := range slices.Sorted(maps.Keys(after)) {
		st := after[p]
		if old, ok := before[p]; !ok || !old.mod.Equal(st.mod) || old.size != st.size {
			changed = append(changed, p)
		}
	}
	sort.Strings(changed)
	for _, p := range changed {
		np := normalizeFile(p)
		if isCompiler(c.Path) {
			e.Empty = append(e.Empty, np)
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if e.Files == nil {
			e.Files = map[string]blob{}
		}
		// binary files (AIGER) may hold length-prefixed paths: kept as
		// written; text files have their paths normalized
		if isText(data) {
			data = []byte(normalizeOut(string(data), temps))
		}
		e.Files[np] = blob(data)
	}
	r.add(e)
	if stdout != nil {
		_, _ = stdout.Write(out.Bytes())
	}
	if stderr != nil && mode != "combined" {
		_, _ = stderr.Write(errb.Bytes())
	}
	c.exitCode, c.expired = e.Exit, e.Expired
	if e.ErrText != "" {
		return &ExitError{Code: e.Exit, Message: e.ErrText, Status: syscall.WaitStatus(e.Wait)}
	}
	return nil
}

// normalizeFile normalizes a written file's path; files in the working
// directory become $CWD/<name>.
func normalizeFile(p string) string { return normalize(p) }

// normalizeOut normalizes recorded output: paths as normalize, and also the
// command's temporary directories with '/' turned into '_' (ABC derives file
// names from paths that way), as $TMPU/NNN.
func normalizeOut(s string, dirs []string) string {
	for _, d := range dirs {
		s = strings.ReplaceAll(s, strings.ReplaceAll(d, "/", "_"), "$TMPU/"+strings.TrimPrefix(normalize(d), "$TMP/"))
	}
	return normalize(s)
}

// denormText puts this run's paths back into recorded output, which was
// stored with the recording run's paths replaced by placeholders.
func denormText(t testing.TB, s string, temps map[string]string) string {
	if !strings.Contains(s, "$") {
		return s
	}
	s = regexp.MustCompile(`\$TMPU/([0-9]{3})`).ReplaceAllStringFunc(s, func(m string) string {
		if dir, ok := temps["$TMP/"+m[len("$TMPU/"):]]; ok {
			return strings.ReplaceAll(dir, "/", "_")
		}
		return m
	})
	s = regexp.MustCompile(`\$TMP/([0-9]{3})`).ReplaceAllStringFunc(s, func(m string) string {
		if dir, ok := temps[m]; ok {
			return dir
		}
		if dir := findTempDir(t, m[len("$TMP/"):]); dir != "" {
			return dir
		}
		return m
	})
	wd, _ := os.Getwd()
	s = strings.ReplaceAll(s, "$CWD", wd)
	phs := placeholders()
	sort.Slice(phs, func(i, j int) bool { return len(phs[i][1]) > len(phs[j][1]) })
	for _, ph := range phs {
		s = strings.ReplaceAll(s, ph[1], ph[0])
	}
	return s
}

// denormalize maps a recorded path back to this run: $CWD to the working
// directory, $TMP/NNN to t's temporary directory NNN, $MOD/$SRC to the module
// and C roots.
func denormalize(t testing.TB, p string, temps map[string]string) string {
	if rest, ok := strings.CutPrefix(p, "$CWD/"); ok {
		wd, _ := os.Getwd()
		return filepath.Join(wd, rest)
	}
	if rest, ok := strings.CutPrefix(p, "$TMP/"); ok {
		parts := strings.SplitN(rest, "/", 2)
		dir, ok := temps["$TMP/"+parts[0]]
		if !ok {
			dir = tempDir(t, parts[0])
		}
		if len(parts) == 1 {
			return dir
		}
		return filepath.Join(dir, parts[1])
	}
	for _, ph := range placeholders() {
		if rest, ok := strings.CutPrefix(p, ph[1]); ok {
			return ph[0] + rest
		}
	}
	return p
}

// tempDir finds t's temporary directory number nnn (t.TempDir() names its
// directories 001, 002, ... under one per-test parent).
func tempDir(t testing.TB, nnn string) string {
	dir := findTempDir(t, nnn)
	if dir == "" {
		t.Fatalf("oracle: no temporary directory %s of %s to replay into (call t.TempDir() before the command)", nnn, t.Name())
	}
	return dir
}

// findTempDir is tempDir, or "" when t has no such directory.
func findTempDir(t testing.TB, nnn string) string {
	tmp := filepath.Clean(os.TempDir())
	prefix := sanitizeTempName(t.Name())
	entries, _ := os.ReadDir(tmp)
	var best string
	var bestTime time.Time
	for _, de := range entries {
		if !de.IsDir() || !strings.HasPrefix(de.Name(), prefix) {
			continue
		}
		cand := filepath.Join(tmp, de.Name(), nnn)
		if st, err := os.Stat(cand); err == nil && st.ModTime().After(bestTime) {
			best, bestTime = cand, st.ModTime()
		}
	}
	return best
}

// sanitizeTempName mirrors testing's naming of t.TempDir()'s parent.
func sanitizeTempName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r < 0x80 && (r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

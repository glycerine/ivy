package testoracle

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Gate stands for an opt-in switch of a C comparison (FRAIG_C_DIFFERENTIAL
// and the like): such comparisons always run, from their recordings.
func Gate(name string) string { return "1" }

// SourceRoot is the original ABC source tree, ../src next to the module (it
// need not exist in replay: reads of it are recorded).
func SourceRoot() string {
	return filepath.Join(filepath.Dir(moduleRoot()), "src")
}

// NativeLibABCPath stands for os.Getenv("ABC_LIBABC"), a native libabc.a:
// while recording the variable, or ../libabc.a when that exists; a
// placeholder in replay (only build commands, which are replayed, use it).
func NativeLibABCPath() string {
	if !Recording() {
		return filepath.Join(os.TempDir(), "abcgo-oracle-replay", "libabc.a")
	}
	if lib := os.Getenv("ABC_LIBABC"); lib != "" {
		abs, _ := filepath.Abs(lib)
		return abs
	}
	lib := filepath.Join(filepath.Dir(moduleRoot()), "libabc.a")
	if _, err := os.Stat(lib); err == nil {
		return lib
	}
	return ""
}

// walkEntry is one entry of a recorded walk, relative to its root.
type walkEntry struct {
	rel  string
	dir  bool
	mode fs.FileMode
}

func walkKey(p string) string { return "$WALK/" + strings.TrimPrefix(p, "$") }

// recordedWalk returns the entries under root (root itself first), from the
// file system while recording (and records them) or from the recording.
func recordedWalk(root string) ([]walkEntry, error) {
	p, out := outside(root)
	if !out {
		return nil, nil
	}
	if Recording() {
		var entries []walkEntry
		var names []string
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			e := walkEntry{rel: rel, dir: d.IsDir(), mode: d.Type()}
			entries = append(entries, e)
			name := rel
			if e.dir {
				name += "/"
			} else if !e.mode.IsRegular() {
				name += "@"
			}
			names = append(names, name)
			return nil
		})
		if err != nil {
			recordMeta(walkKey(p), meta{Missing: true})
			return nil, err
		}
		recordMeta(walkKey(p), meta{Entries: names})
		return entries, nil
	}
	m, ok := readMeta(walkKey(p))
	if !ok {
		return nil, &fs.PathError{Op: "lstat", Path: root, Err: unrecorded(root)}
	}
	if m.Missing {
		return nil, &fs.PathError{Op: "lstat", Path: root, Err: fs.ErrNotExist}
	}
	var entries []walkEntry
	for _, n := range m.Entries {
		switch {
		case strings.HasSuffix(n, "/"):
			entries = append(entries, walkEntry{rel: strings.TrimSuffix(n, "/"), dir: true, mode: fs.ModeDir})
		case strings.HasSuffix(n, "@"):
			entries = append(entries, walkEntry{rel: strings.TrimSuffix(n, "@"), mode: fs.ModeSymlink})
		default:
			entries = append(entries, walkEntry{rel: n})
		}
	}
	return entries, nil
}

type walkDirEntry struct{ e walkEntry }

func (d walkDirEntry) Name() string      { return filepath.Base(d.e.rel) }
func (d walkDirEntry) IsDir() bool       { return d.e.dir }
func (d walkDirEntry) Type() fs.FileMode { return d.e.mode.Type() }
func (d walkDirEntry) Info() (fs.FileInfo, error) {
	return fileInfo{filepath.Base(d.e.rel), meta{Dir: d.e.dir, Mode: d.e.mode}}, nil
}

// WalkDir is filepath.WalkDir; under a root outside the module the walk is
// recorded (a SkipDir from fn is honored).
func WalkDir(root string, fn fs.WalkDirFunc) error {
	if _, out := outside(root); !out {
		return filepath.WalkDir(root, fn)
	}
	entries, err := recordedWalk(root)
	if err != nil {
		return fn(root, nil, err)
	}
	var skip []string
	for _, e := range entries {
		path := filepath.Join(root, e.rel)
		if skipped(path, skip) {
			continue
		}
		if err := fn(path, walkDirEntry{e}, nil); err != nil {
			if err == filepath.SkipDir {
				if e.dir {
					skip = append(skip, path)
					continue
				}
				skip = append(skip, filepath.Dir(path))
				continue
			}
			if err == filepath.SkipAll {
				return nil
			}
			return err
		}
	}
	return nil
}

// Walk is filepath.Walk, recorded like WalkDir.
func Walk(root string, fn filepath.WalkFunc) error {
	if _, out := outside(root); !out {
		return filepath.Walk(root, fn)
	}
	return WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fn(path, nil, err)
		}
		info, _ := d.Info()
		return fn(path, info, nil)
	})
}

func skipped(path string, skip []string) bool {
	for _, s := range skip {
		if path == s || strings.HasPrefix(path, s+"/") {
			return true
		}
	}
	return false
}

// NativeABCAbs stands for filepath.Abs of the native abc next to the C tree
// (../abc): NativeABCPath, with a nil error.
func NativeABCAbs() (string, error) { return NativeABCPath(), nil }

// BuildFixture compiles a C fixture of the test's package (a path such as
// test_vectors/x.c) against the original ABC: headers from SourceRoot and
// ../libabc.a (built from ../src with assertions on: make ABC_USE_STDINT_H=1
// ABC_USE_NO_READLINE=1 libabc.a). It returns the program, whose runs are
// recorded like any command; in replay nothing is compiled.
func BuildFixture(t testing.TB, fixture string) string {
	t.Helper()
	lib := NativeLibABCPath()
	if lib == "" {
		t.Skipf("recording %s needs ../libabc.a (make ABC_USE_STDINT_H=1 ABC_USE_NO_READLINE=1 libabc.a in the ABC tree)", fixture)
	}
	exe := filepath.Join(t.TempDir(), strings.TrimSuffix(filepath.Base(fixture), ".c"))
	out, err := Command(t, "gcc", "-DABC_USE_STDINT_H", "-I", SourceRoot(), fixture, lib, "-lm", "-ldl", "-lpthread", "-lstdc++", "-o", exe).CombinedOutput()
	if err != nil {
		t.Fatalf("compile %s: %v\n%s", fixture, err, out)
	}
	return exe
}

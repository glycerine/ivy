package testoracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Files outside the module (the C tree ../src, ../i10.aig, ../test_vectors:
// everything under the module's parent directory but not in the module) are
// read through Stat, ReadFile, Open, ReadDir and Glob. Elsewhere these are the
// os functions. While recording, what they return is saved under the
// module's test_vectors/oracle/_files/<placeholder>/<path>; in replay it is
// read back from there, so the C tree need not exist.

// outside reports whether path is in the module's parent but not the module,
// and returns its placeholder form ($SRC/..., $ABCROOT/...).
func outside(path string) (string, bool) {
	root := moduleRoot()
	if root == "" {
		return "", false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	parent := filepath.Dir(root)
	if abs == root || strings.HasPrefix(abs, root+"/") || !(abs == parent || strings.HasPrefix(abs, parent+"/")) {
		return "", false
	}
	return normalize(abs), true
}

// storePath is where the recording of placeholder path p lives.
func storePath(p string) string {
	rel := strings.TrimPrefix(p, "$")
	return filepath.Join(moduleRoot(), "test_vectors", "oracle", "_files", filepath.FromSlash(rel))
}

// meta is the recorded result of a Stat (and the absence of a file).
type meta struct {
	Missing bool        `json:"missing,omitempty"`
	Dir     bool        `json:"dir,omitempty"`
	Size    int64       `json:"size"`
	Mode    fs.FileMode `json:"mode"`
	Entries []string    `json:"entries,omitempty"` // ReadDir: names, dirs with a trailing /
}

func metaFile(p string) string { return storePath(p) + ".oracle-meta.json" }

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	tmp := path + ".tmp-" + hex.EncodeToString(sum[:4]) + fmt.Sprint(os.Getpid())
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func recordMeta(p string, m meta) {
	old, ok := readMeta(p)
	if ok && old.Entries != nil && m.Entries == nil {
		m.Entries = old.Entries
	}
	data, _ := json.MarshalIndent(m, "", " ")
	if err := writeAtomic(metaFile(p), append(data, '\n')); err != nil {
		panic(fmt.Sprintf("oracle: recording %s: %v", p, err))
	}
}

func readMeta(p string) (meta, bool) {
	var m meta
	data, err := os.ReadFile(metaFile(p))
	if err != nil {
		return m, false
	}
	if err := json.Unmarshal(data, &m); err != nil {
		panic(fmt.Sprintf("oracle: %s: %v", metaFile(p), err))
	}
	return m, true
}

func unrecorded(p string) error {
	return fmt.Errorf("oracle: %s was not recorded (record with ABCGO_ORACLE_RECORD=1 next to the C tree): %w", p, fs.ErrNotExist)
}

type fileInfo struct {
	name string
	m    meta
}

func (f fileInfo) Name() string       { return f.name }
func (f fileInfo) Size() int64        { return f.m.Size }
func (f fileInfo) Mode() fs.FileMode  { return f.m.Mode }
func (f fileInfo) ModTime() time.Time { return time.Time{} }
func (f fileInfo) IsDir() bool        { return f.m.Dir }
func (f fileInfo) Sys() any           { return nil }

func statOf(path string, lstat bool) (fs.FileInfo, error) {
	if lstat {
		return os.Lstat(path)
	}
	return os.Stat(path)
}

func stat(path string, lstat bool) (fs.FileInfo, error) {
	if !Recording() && path == nativePlaceholder() {
		return fileInfo{"abc", meta{Size: 1, Mode: 0o755}}, nil
	}
	p, out := outside(path)
	if !out {
		return statOf(path, lstat)
	}
	if Recording() {
		info, err := statOf(path, lstat)
		if err != nil {
			recordMeta(p, meta{Missing: true})
			return nil, err
		}
		recordMeta(p, meta{Dir: info.IsDir(), Size: info.Size(), Mode: info.Mode()})
		return info, nil
	}
	m, ok := readMeta(p)
	if !ok {
		return nil, &fs.PathError{Op: "stat", Path: path, Err: unrecorded(p)}
	}
	if m.Missing {
		return nil, &fs.PathError{Op: "stat", Path: path, Err: fs.ErrNotExist}
	}
	return fileInfo{filepath.Base(path), m}, nil
}

// Stat is os.Stat; outside the module it is recorded.
func Stat(path string) (fs.FileInfo, error) { return stat(path, false) }

// Lstat is os.Lstat; outside the module it is recorded.
func Lstat(path string) (fs.FileInfo, error) { return stat(path, true) }

// ReadFile is os.ReadFile; outside the module it is recorded.
func ReadFile(path string) ([]byte, error) {
	p, out := outside(path)
	if !out {
		return os.ReadFile(path)
	}
	if Recording() {
		data, err := os.ReadFile(path)
		if err != nil {
			if _, serr := os.Stat(path); serr != nil {
				recordMeta(p, meta{Missing: true})
			}
			return nil, err
		}
		if err := writeAtomic(storePath(p), data); err != nil {
			panic(fmt.Sprintf("oracle: recording %s: %v", p, err))
		}
		info, _ := os.Stat(path)
		recordMeta(p, meta{Size: int64(len(data)), Mode: info.Mode()})
		return data, nil
	}
	if m, ok := readMeta(p); ok && m.Missing {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	data, err := os.ReadFile(storePath(p))
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: unrecorded(p)}
	}
	return data, nil
}

// Path returns a path at which the contents of path can be read: path itself
// inside the module or while recording, a copy of the recorded contents in
// replay. Use it for files outside the module that the code under test opens.
func Path(path string) string {
	if _, out := outside(path); !out || Recording() {
		_, _ = ReadFile(path) // record
		return path
	}
	p, _ := outside(path)
	if _, err := os.Stat(storePath(p)); err != nil {
		return path // missing: let the caller see the error on the real path
	}
	return storePath(p)
}

// Open is os.Open; outside the module the recorded contents are opened.
func Open(path string) (*os.File, error) {
	p, out := outside(path)
	if !out {
		return os.Open(path)
	}
	if Recording() {
		if _, err := ReadFile(path); err != nil {
			return nil, err
		}
		return os.Open(path)
	}
	if m, ok := readMeta(p); ok && m.Missing {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	f, err := os.Open(storePath(p))
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: unrecorded(p)}
	}
	return f, nil
}

type dirEntry struct {
	name string
	dir  bool
}

func (d dirEntry) Name() string               { return d.name }
func (d dirEntry) IsDir() bool                { return d.dir }
func (d dirEntry) Type() fs.FileMode          { return map[bool]fs.FileMode{true: fs.ModeDir}[d.dir] }
func (d dirEntry) Info() (fs.FileInfo, error) { return fileInfo{d.name, meta{Dir: d.dir}}, nil }

// ReadDir is os.ReadDir; outside the module the listing is recorded.
func ReadDir(path string) ([]os.DirEntry, error) {
	p, out := outside(path)
	if !out {
		return os.ReadDir(path)
	}
	if Recording() {
		entries, err := os.ReadDir(path)
		if err != nil {
			recordMeta(p, meta{Missing: true})
			return nil, err
		}
		names := []string{}
		for _, e := range entries {
			n := e.Name()
			if e.IsDir() {
				n += "/"
			}
			names = append(names, n)
		}
		recordMeta(p, meta{Dir: true, Mode: fs.ModeDir | 0o755, Entries: names})
		return entries, nil
	}
	m, ok := readMeta(p)
	if !ok || (!m.Missing && m.Entries == nil) {
		return nil, &fs.PathError{Op: "open", Path: path, Err: unrecorded(p)}
	}
	if m.Missing {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	var list []os.DirEntry
	for _, n := range m.Entries {
		list = append(list, dirEntry{strings.TrimSuffix(n, "/"), strings.HasSuffix(n, "/")})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}

// Glob is filepath.Glob; a pattern outside the module has its matches
// recorded.
func Glob(pattern string) ([]string, error) {
	p, out := outside(pattern)
	if !out {
		return filepath.Glob(pattern)
	}
	key := "$GLOB/" + strings.NewReplacer("*", "STAR", "?", "QUESTION", "[", "LB", "]", "RB").Replace(strings.TrimPrefix(p, "$"))
	if Recording() {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		var names []string
		for _, m := range matches {
			np, _ := outside(m)
			names = append(names, np)
		}
		recordMeta(key, meta{Entries: append([]string{}, names...)})
		return matches, nil
	}
	m, ok := readMeta(key)
	if !ok {
		return nil, unrecorded(pattern)
	}
	var matches []string
	for _, n := range m.Entries {
		matches = append(matches, denormalizePath(n))
	}
	return matches, nil
}

// denormalizePath maps $SRC/... and $ABCROOT/... back to this machine.
func denormalizePath(p string) string {
	for _, ph := range placeholders() {
		if rest, ok := strings.CutPrefix(p, ph[1]); ok {
			return ph[0] + rest
		}
	}
	return p
}

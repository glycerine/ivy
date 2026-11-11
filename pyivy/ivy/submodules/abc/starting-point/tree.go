package startingpoint

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TreeFile struct {
	Source   string
	Output   string
	Package  string
	SHA256   string
	Issues   int
	Fallback bool
}
type TreeReport struct {
	Source     string
	Output     string
	Files      []TreeFile
	Fallbacks  int
	DurationNS int64
}

// TranslateTree mirrors C/C++ files into a new destination. The destination
// must not exist. Unsupported syntax produces an explicit source-preserving
// fallback draft and diagnostic; filesystem failures stop the run with an error.
func TranslateTree(source, destination string, opts Options) (TreeReport, error) {
	start := time.Now()
	r := TreeReport{}
	var err error
	r.Source, err = filepath.Abs(source)
	if err != nil {
		return r, err
	}
	r.Output, err = filepath.Abs(destination)
	if err != nil {
		return r, err
	}
	info, err := os.Stat(r.Source)
	if err != nil {
		return r, err
	}
	if !info.IsDir() {
		return r, fmt.Errorf("not a source directory: %s", source)
	}
	if _, err = os.Lstat(r.Output); !os.IsNotExist(err) {
		if err != nil {
			return r, err
		}
		return r, fmt.Errorf("output already exists: %s", r.Output)
	}
	if opts.Package != "" {
		if _, err = ResolvePackage("", opts.Package); err != nil {
			return r, err
		}
	}
	// Inventory before creating output, and never follow symlinks or descend into
	// prior generated trees. The latter matters when processing ABC from its root.
	var files []string
	err = filepath.WalkDir(r.Source, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != r.Source && (path == r.Output || strings.HasPrefix(d.Name(), "starting-point-cpp2go-") || d.Name() == ".git" || d.Name() == ".hg" || d.Name() == ".svn") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".c", ".cpp", ".h":
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return r, err
	}
	if err = os.Mkdir(r.Output, 0755); err != nil {
		return r, err
	}
	for _, path := range files {
		rel, err := filepath.Rel(r.Source, path)
		if err != nil {
			return r, err
		}
		target := filepath.Join(r.Output, rel+".go")
		pkg, err := ResolvePackage(target, opts.Package)
		if err != nil {
			return r, err
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return r, err
		}
		fileOpts := opts
		fileOpts.Package = pkg
		result, analyzeErr := AnalyzeWithOptions(path, string(bytes), fileOpts)
		fallback := analyzeErr != nil
		if fallback {
			r.Fallbacks++
			result = Result{Schema: 2, Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256(bytes)),
				Draft:  "package " + pkg + "\n\n//TODO: translation could not analyze this file; original source follows.\n// " + strings.ReplaceAll(string(bytes), "\n", "\n// ") + "\n",
				Issues: []Issue{{ID: path + "::analysis", Category: "analysis-failure", Status: "open", Message: analyzeErr.Error(), Start: 0, End: len(bytes)}}}
		}
		if fallback {
			result.Draft, err = SanitizeCommentUTF8(result.Draft)
			if err != nil {
				return r, err
			}
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return r, err
		}
		if err = os.WriteFile(target, []byte(result.Draft), 0644); err != nil {
			return r, err
		}
		meta := filepath.Join(r.Output, "_starting-point", rel)
		if err = os.MkdirAll(filepath.Dir(meta), 0755); err != nil {
			return r, err
		}
		if err = os.WriteFile(meta+".original", bytes, 0644); err != nil {
			return r, err
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return r, err
		}
		if err = os.WriteFile(meta+".json", append(data, '\n'), 0644); err != nil {
			return r, err
		}
		r.Files = append(r.Files, TreeFile{path, rel + ".go", pkg, result.SHA256, len(result.Issues), fallback})
	}
	r.DurationNS = time.Since(start).Nanoseconds()
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return r, err
	}
	err = os.WriteFile(filepath.Join(r.Output, "run.json"), append(data, '\n'), 0644)
	return r, err
}

// DefaultTreeOutput names the destination in the current working directory.
func DefaultTreeOutput(source string) (string, error) {
	absolute, err := filepath.Abs(source)
	if err != nil {
		return "", err
	}
	return "starting-point-cpp2go-" + filepath.Base(absolute), nil
}

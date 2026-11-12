package startingpoint

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TreeFile struct {
	LibraryConverted int                 `json:",omitempty"`
	LibraryReview    int                 `json:",omitempty"`
	Conditionals     []ConditionalReview `json:",omitempty"`
	Source           string
	Output           string
	Package          string
	SHA256           string
	Issues           int
	Fallback         bool
	FallbackReason   string          `json:",omitempty"`
	FallbackLocation *SourceLocation `json:",omitempty"`
}
type TreeReport struct {
	Configuration map[string]int64 `json:",omitempty"`
	HelperFiles   []string         `json:",omitempty"`
	Source        string
	Output        string
	Files         []TreeFile
	Fallbacks     int
	DurationNS    int64
}

// TranslateTree mirrors C/C++ files into a new destination. The destination
// must not exist. Unsupported syntax produces an explicit source-preserving
// fallback draft and diagnostic; filesystem failures stop the run with an error.
func TranslateTree(source, destination string, opts Options) (TreeReport, error) {
	start := time.Now()
	r := TreeReport{Configuration: opts.Defines}
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
	indexed, err := CollectTypeNames(files)
	if err != nil {
		return r, err
	}
	for name := range opts.TypeNames {
		indexed[name] = true
	}
	opts.TypeNames = indexed
	opts.IncludePaths = append(append([]string{}, opts.IncludePaths...), r.Source, filepath.Join(r.Source, "src"))
	rels := make([]string, len(files))
	for i, path := range files {
		rels[i], err = filepath.Rel(r.Source, path)
		if err != nil {
			return r, err
		}
	}
	outputNames, err := PlanGoFilenames(rels)
	if err != nil {
		return r, err
	}
	if err = os.Mkdir(r.Output, 0755); err != nil {
		return r, err
	}
	helperPackages := map[string]string{}
	for i, path := range files {
		rel := rels[i]
		target := filepath.Join(r.Output, outputNames[i])
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
		reason := ""
		var location *SourceLocation
		var diagnostic *SourceError
		if errors.As(analyzeErr, &diagnostic) {
			loc := diagnostic.Location
			location = &loc
		}
		if fallback {
			reason = analyzeErr.Error()
		}
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
		if result.NeedsAssertHelper() {
			helperPackages[filepath.Dir(target)] = pkg
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
		converted, review := 0, 0
		for _, change := range result.Library {
			if change.Status == "converted" {
				converted++
			} else {
				review++
			}
		}
		r.Files = append(r.Files, TreeFile{LibraryConverted: converted, LibraryReview: review, Conditionals: result.Conditionals, Source: path, Output: outputNames[i], Package: pkg, SHA256: result.SHA256, Issues: len(result.Issues), Fallback: fallback, FallbackReason: reason, FallbackLocation: location})
	}
	var helperDirs []string
	for dir := range helperPackages {
		helperDirs = append(helperDirs, dir)
	}
	sort.Strings(helperDirs)
	for _, dir := range helperDirs {
		path, err := WriteAssertHelper(dir, helperPackages[dir])
		if err != nil {
			return r, err
		}
		rel, err := filepath.Rel(r.Output, path)
		if err != nil {
			return r, err
		}
		r.HelperFiles = append(r.HelperFiles, rel)
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

package goivy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestParseV16CorpusScan(t *testing.T) {
	if os.Getenv("V16_CORPUS_SCAN") == "" {
		t.Skip("set V16_CORPUS_SCAN=1 to parse every #lang ivy1.6 file under the example corpora")
	}

	roots := v16CorpusRoots(t)
	files := v16CorpusFiles(t, roots)
	if len(files) == 0 {
		t.Fatalf("no #lang ivy1.6 files found under %v", roots)
	}

	var failures []string
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("READ %s: %v", path, err))
			continue
		}
		src := stripIvyLangLine(string(data))
		if _, err := Parse(src, Version{1, 6}, WithFilename(path)); err != nil {
			failures = append(failures, fmt.Sprintf("PARSE %s: %v", path, err))
		}
	}

	for _, failure := range failures {
		t.Log(failure)
	}
	if len(failures) > 0 {
		t.Fatalf("v1.6 corpus parse scan failed: scanned=%d failed=%d", len(files), len(failures))
	}
	t.Logf("v1.6 corpus parse scan passed: scanned=%d", len(files))
}

func v16CorpusRoots(t *testing.T) []string {
	t.Helper()
	if raw := os.Getenv("V16_CORPUS_ROOTS"); raw != "" {
		var roots []string
		for _, part := range strings.Split(raw, ",") {
			root := strings.TrimSpace(part)
			if root != "" {
				roots = append(roots, root)
			}
		}
		if len(roots) == 0 {
			t.Fatal("V16_CORPUS_ROOTS was set but contained no roots")
		}
		return roots
	}

	candidates := []string{"../ivy-lang-examples", "../pyivy/ivy"}
	var roots []string
	for _, root := range candidates {
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			roots = append(roots, root)
		}
	}
	if len(roots) == 0 {
		t.Skipf("no default v1.6 corpus roots found: %v", candidates)
	}
	return roots
}

func v16CorpusFiles(t *testing.T, roots []string) []string {
	t.Helper()
	var files []string
	for _, root := range roots {
		if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".ivy") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.HasPrefix(string(data), "#lang ivy1.6") {
				files = append(files, filepath.Clean(path))
			}
			return nil
		}); err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	sort.Strings(files)
	return files
}

func stripIvyLangLine(src string) string {
	if i := strings.IndexByte(src, '\n'); i >= 0 {
		return src[i+1:]
	}
	return ""
}

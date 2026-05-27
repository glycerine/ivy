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
	var expectedErrors int
	for _, path := range files {
		//vv("path = '%v'", path)
		data, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("READ %s: %v", path, err))
			continue
		}
		src := stripIvyLangLine(string(data))
		_, err = Parse(src, Version{1, 6}, WithFilename(path))
		if want, ok := expectedV16CorpusParseError(path); ok {
			expectedErrors++
			if err == nil {
				failures = append(failures, fmt.Sprintf("PARSE %s: got success, want Python-compatible parse error containing %q (%s)", path, want.Contains, want.Reason))
				continue
			}
			if !strings.Contains(err.Error(), want.Contains) {
				failures = append(failures, fmt.Sprintf("PARSE %s: got %q, want Python-compatible parse error containing %q (%s)", path, err.Error(), want.Contains, want.Reason))
				continue
			}
			//vv("good: got EXPECTED %s: %v (%s)", path, err, want.Reason)
			continue
		}
		if err != nil {
			failures = append(failures, fmt.Sprintf("PARSE %s: %v", path, err))
		}
	}

	for _, failure := range failures {
		t.Log(failure)
	}
	if len(failures) > 0 {
		t.Fatalf("v1.6 corpus parse scan failed: scanned=%d expectedErrors=%d failures=%d", len(files), expectedErrors, len(failures))
	}
	//vv("v1.6 corpus parse scan passed: scanned=%d expectedErrors=%d", len(files), expectedErrors)
}

func expectedV16CorpusParseError(path string) (v16ExpectedParseError, bool) {
	for suffix, expected := range expectedV16CorpusParseErrors {
		if strings.HasSuffix(path, suffix) {
			return expected, true
		}
	}
	return v16ExpectedParseError{}, false
}

type v16ExpectedParseError struct {
	Contains string
	Reason   string
}

// expectedV16CorpusParseErrors documents #lang ivy1.6 corpus files that are
// expected to fail according to the Python grammar we are porting. These are
// not skipped: each file must fail, and the Go diagnostic must match the same
// token-level syntax error shape Python reports via ivy_utils.p_error.
var expectedV16CorpusParseErrors = map[string]v16ExpectedParseError{
	"examples/ivy/arrrel.ivy": {
		Contains: "token 'r': syntax error",
		Reason:   "comma-less struct fields; Python tterms require commas",
	},
	"examples/raft/raft.ivy": {
		Contains: "token 'RV_option_wf': syntax error",
		Reason:   "uppercase label [RV_option_wf]; Python LABEL uses SYMBOL/PRESYMBOL",
	},
	"test/impltype1.ivy": {
		Contains: "token 'as': syntax error",
		Reason:   "uses an `as` cast form with no Python grammar token",
	},
	"test/marcelocrash3.ivy": {
		Contains: "token 'x': syntax error",
		Reason:   "omits the semicolon between consecutive simple actions",
	},
	"test/proving4.ivy": {
		Contains: "token ';': syntax error",
		Reason:   "uses an unbraced proof sequence; Python optproof takes one proofstep",
	},
	"test/recursion1.ivy": {
		Contains: "token '->': syntax error",
		Reason:   "uses function-sort schema parameters outside Python atype grammar",
	},
	"test/yacc1.ivy": {
		Contains: "token ')': syntax error",
		Reason:   "uses empty argument parentheses; Python optargs require lparams",
	},
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

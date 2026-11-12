package main

import (
	"bytes"
	"encoding/json"
	startingpoint "github.com/glycerine/abc/starting-point"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedOutputPackage(t *testing.T) {
	for _, override := range []string{"", "satoko"} {
		t.Run("override="+override, func(t *testing.T) {
			root := t.TempDir()
			var sources []string
			for _, dir := range []string{"vec", "utils"} {
				path := filepath.Join(root, dir, "input.h")
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("static inline int f() { return 1; }"), 0644); err != nil {
					t.Fatal(err)
				}
				sources = append(sources, path)
			}
			out := filepath.Join(root, "unrelated_destination")
			args := []string{"-out", out}
			if override != "" {
				args = append(args, "-package", override)
			}
			args = append(args, sources...)
			if err := run(args); err != nil {
				t.Fatal(err)
			}
			want := override
			if want == "" {
				want = "vec"
			}
			files, err := filepath.Glob(filepath.Join(out, "*.go"))
			if err != nil || len(files) != 2 {
				t.Fatal(files, err)
			}
			for _, path := range files {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(string(data), "package "+want+"\n") {
					t.Fatalf("%s: %s", path, data)
				}
			}
			data, err := os.ReadFile(filepath.Join(out, "run.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report struct{ Package string }
			if err := json.Unmarshal(data, &report); err != nil {
				t.Fatal(err)
			}
			if report.Package != want {
				t.Fatal(report)
			}
		})
	}
}

func TestInvalidPackageBeforeOutputCreation(t *testing.T) {
	out := filepath.Join(t.TempDir(), "output")
	if err := run([]string{"-out", out, "-package", "bad-name", "file.h"}); err == nil {
		t.Fatal("expected invalid package error")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("created output before validating package", err)
	}
}

func TestDirectoryCommand(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "abc")
	out := filepath.Join(root, "out")
	for _, dir := range []string{"vec", "utils"} {
		path := filepath.Join(source, dir, "input.h")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("static inline int f() { return 1; }"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := run([]string{"-out", out, source}); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"vec", "utils"} {
		data, err := os.ReadFile(filepath.Join(out, dir, "input.go"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), "package "+dir+"\n") {
			t.Fatal(string(data))
		}
	}
}

func TestSourceArgumentCheckoutName(t *testing.T) {
	parent := t.TempDir()
	checkout := filepath.Join(parent, "abc")
	if err := os.Mkdir(checkout, 0755); err != nil {
		t.Fatal(err)
	}
	// Reproduce ABC's same-named executable inside its source checkout.
	if err := os.WriteFile(filepath.Join(checkout, "abc"), []byte("executable"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, cwd := range []string{parent, checkout} {
		source, isDir, err := resolveSourceArgument("abc", cwd)
		if err != nil || !isDir || source != checkout {
			t.Fatal(source, isDir, err)
		}
	}
	if _, _, err := resolveSourceArgument("missing", parent); err == nil || !strings.Contains(err.Error(), "cannot access source") {
		t.Fatal(err)
	}
	file, isDir, err := resolveSourceArgument(filepath.Join(checkout, "abc"), parent)
	if err != nil || isDir || file != filepath.Join(checkout, "abc") {
		t.Fatal(file, isDir, err)
	}
}

func TestFallbackDetailsPrinted(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "abc")
	out := filepath.Join(root, "out")
	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "broken.c"), []byte("/* unfinished"), 0644); err != nil {
		t.Fatal(err)
	}
	report, err := startingpoint.TranslateTree(source, out, startingpoint.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	printTreeReport(&output, report, true)
	for _, want := range []string{"1 source-preserving fallbacks", "commented original source; analysis failed", "broken.c -> broken.go: unterminated comment at source line 1, column 1 (byte offset 0)"} {
		if !strings.Contains(output.String(), want) {
			t.Fatal(output.String())
		}
	}
	output.Reset()
	printTreeReport(&output, report, false)
	if strings.Contains(output.String(), "choose branch") {
		t.Fatal("preprocessor listing enabled by default")
	}
	data, err := os.ReadFile(filepath.Join(out, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved startingpoint.TreeReport
	if err = json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Files[0].FallbackLocation == nil || saved.Files[0].FallbackLocation.Line != 1 || saved.Files[0].FallbackLocation.Column != 1 || saved.Files[0].FallbackLocation.ByteOffset != 0 {
		t.Fatal(saved)
	}
	if saved.Files[0].FallbackReason != "unterminated comment at source line 1, column 1 (byte offset 0)" {
		t.Fatal(saved)
	}
	output.Reset()
	printTreeReport(&output, startingpoint.TreeReport{}, false)
	if strings.Contains(output.String(), "Fallbacks (") {
		t.Fatal("unexpected fallback listing")
	}
}

func TestConditionalLocationsPrintedAndSaved(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "abc")
	out := filepath.Join(root, "out")
	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}
	input := "// heading\n  #if OUTER\n#if INNER\nint x;\n#else\nint y;\n#endif\n#endif\n"
	if err := os.WriteFile(filepath.Join(source, "choice.c"), []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	report, err := startingpoint.TranslateTree(source, out, startingpoint.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	printTreeReport(&output, report, true)
	for _, want := range []string{"0 source-preserving fallbacks", "branch decisions required", "choice.c:2:3 -> choice.go: preserved #if OUTER", "choice.c:3:1 -> choice.go: preserved #if INNER"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q:\n%s", want, output.String())
		}
	}
	output.Reset()
	printTreeReport(&output, report, false)
	if strings.Contains(output.String(), "choose branch") {
		t.Fatal("preprocessor listing enabled by default")
	}
	data, err := os.ReadFile(filepath.Join(out, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved startingpoint.TreeReport
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.Files[0].Conditionals) != 2 || saved.Files[0].Conditionals[0].Location.Column != 3 {
		t.Fatal(saved)
	}
	output.Reset()
	printConditionals(&output, "choice.c", "choice.go", saved.Files[0].Conditionals)
	if !strings.Contains(output.String(), "choice.c:2:3") {
		t.Fatal(output.String())
	}
}

func TestLibraryReviewCountsPersistAndPrint(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "src")
	out := filepath.Join(root, "out")
	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "input.c"), []byte("int f(char *s) { int n = atoi(s); malloc(n); return n; }"), 0644); err != nil {
		t.Fatal(err)
	}
	report, err := startingpoint.TranslateTree(source, out, startingpoint.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 1 || report.Files[0].LibraryConverted != 1 || report.Files[0].LibraryReview != 1 {
		t.Fatal(report)
	}
	var output bytes.Buffer
	printTreeReport(&output, report, false)
	if !strings.Contains(output.String(), "Standard library: 1 converted sites, 1 retained review sites") {
		t.Fatal(output.String())
	}
	data, err := os.ReadFile(filepath.Join(out, "_starting-point", "input.c.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result startingpoint.Result
	if err = json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Library) != 2 || result.Library[0].Location.Line < 1 || result.Library[0].SourceLine == "" {
		t.Fatal(result.Library)
	}
}

func TestBatchAssertHelper(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "assert.c")
	if err := os.WriteFile(source, []byte("void check(int x) { assert(x); }"), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "out")
	if err := run([]string{"-out", out, "-package", "checks", source}); err != nil {
		t.Fatal(err)
	}
	helper, err := os.ReadFile(filepath.Join(out, "assert2.go"))
	if err != nil || !strings.HasPrefix(string(helper), "package checks\n") || !strings.Contains(string(helper), `panic(where + "assertion failed, we do not see: " + message)`) {
		t.Fatal(string(helper), err)
	}
	draft, err := os.ReadFile(filepath.Join(out, "assert.go"))
	if err != nil || !strings.Contains(string(draft), `assert((x != 0), "x")`) {
		t.Fatal(string(draft), err)
	}
}

func TestConfigFlagAndDefault(t *testing.T) {
	root := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(previous)
	source := filepath.Join(root, "input.c")
	if err = os.WriteFile(source, []byte("#ifdef WIN32\nint windows;\n#else\nint portable;\n#endif\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile("starting.cfg", []byte("WIN32 = 0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile("alternate.cfg", []byte("WIN32 = 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{"default", nil, "var portable int32;"},
		{"override", []string{"-cfg", "alternate.cfg"}, "var windows int32;"},
	} {
		out := filepath.Join(root, test.name)
		args := append([]string{"-out", out}, test.args...)
		args = append(args, source)
		if err := run(args); err != nil {
			t.Fatal(err)
		}
		draft, err := os.ReadFile(filepath.Join(out, "input.go"))
		if err != nil || !strings.Contains(string(draft), test.want) || strings.Contains(string(draft), "#ifdef") {
			t.Fatal(string(draft), err)
		}
	}
	badOut := filepath.Join(root, "bad")
	if err := run([]string{"-cfg", "missing.cfg", "-out", badOut, source}); err == nil {
		t.Fatal("missing explicit config accepted")
	}
	if _, err := os.Stat(badOut); !os.IsNotExist(err) {
		t.Fatal("created output before config validation")
	}
}

func TestMissingConfigTemplate(t *testing.T) {
	root := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(previous)
	source := filepath.Join(root, "input.c")
	text := "#if defined(WIN32) || LEVEL >= 0xFFu // ignoredWord\nint first;\n#elif OTHER\nint second;\n#endif\n"
	if err = os.WriteFile(source, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	if err = run([]string{"-out", "out", source}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("_starting.cfg")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := startingpoint.ParseConfig(string(data))
	if err != nil || len(cfg) != 3 {
		t.Fatal(string(data), cfg, err)
	}
	for _, name := range []string{"WIN32", "LEVEL", "OTHER"} {
		if value, ok := cfg[name]; !ok || value != 0 {
			t.Fatal(cfg)
		}
	}
	if !strings.Contains(string(data), "input.c:1: #if") {
		t.Fatal(string(data))
	}
	// An edited template is never overwritten by another run.
	if err = os.WriteFile("_starting.cfg", []byte("WIN32 = 1 // edited\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = run([]string{"-out", "out2", source}); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile("_starting.cfg")
	if string(data) != "WIN32 = 1 // edited\n" {
		t.Fatal(string(data))
	}
}

// CLI tests may generate a config template in their working directory.
func TestMain(m *testing.M) {
	previous, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir, err := os.MkdirTemp("", "starting-point-cli-tests-")
	if err != nil {
		panic(err)
	}
	if err = os.Chdir(dir); err != nil {
		panic(err)
	}
	code := m.Run()
	os.Chdir(previous)
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestConfigTemplateTreeAndExplicitPath(t *testing.T) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(previous)
	for _, explicit := range []bool{false, true} {
		root := t.TempDir()
		if err = os.Chdir(root); err != nil {
			t.Fatal(err)
		}
		if err = os.Mkdir("src", 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile("src/a.c", []byte("#if WIN32\nint x;\n#else\nint y;\n#endif\n"), 0644); err != nil {
			t.Fatal(err)
		}
		args := []string{"-out", "out"}
		if explicit {
			if err = os.WriteFile("custom.cfg", []byte("WIN32=0// disabled\n"), 0644); err != nil {
				t.Fatal(err)
			}
			args = append(args, "-cfg", "custom.cfg")
		}
		args = append(args, "src")
		if err = run(args); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile("_starting.cfg")
		if explicit {
			if !os.IsNotExist(err) {
				t.Fatal("explicit config generated template", err)
			}
		} else if err != nil || !strings.Contains(string(data), "WIN32 = 0") {
			t.Fatal(string(data), err)
		}
	}
}

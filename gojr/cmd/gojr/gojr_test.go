package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShouldPrintValueSuppressesOnlyActualNil(t *testing.T) {
	tests := []struct {
		name   string
		result evalResult
		want   bool
	}{
		{
			name:   "empty",
			result: evalResult{},
			want:   false,
		},
		{
			name:   "actual nil",
			result: evalResult{Value: "<nil>", ValueIsNil: true},
			want:   false,
		},
		{
			name:   "string containing nil spelling",
			result: evalResult{Value: "<nil>"},
			want:   true,
		},
		{
			name:   "normal value",
			result: evalResult{Value: "42"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPrintValue(tt.result); got != tt.want {
				t.Fatalf("shouldPrintValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripPackageClausePreservesLineNumbers(t *testing.T) {
	source := "// doc\npackage demo // comment\n\nfunc F() int { return 1 }\n"
	stripped, found, name, err := stripPackageClausePreservingLinesWithName(source)
	if err != nil {
		t.Fatalf("stripPackageClausePreservingLinesWithName() error = %v", err)
	}
	if !found {
		t.Fatalf("package clause not found")
	}
	if name != "demo" {
		t.Fatalf("package name = %q, want demo", name)
	}
	if strings.Contains(stripped, "package demo") {
		t.Fatalf("package clause was not stripped:\n%s", stripped)
	}
	if got, want := strings.Count(stripped, "\n"), strings.Count(source, "\n"); got != want {
		t.Fatalf("newline count = %d, want %d", got, want)
	}
	if !strings.Contains(stripped, "func F() int") {
		t.Fatalf("function body missing after strip:\n%s", stripped)
	}
}

func TestReadPackageDirCombinesNonTestGoFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "b.go"), "package demo\n\nimport \"fmt\"\n\nfunc B() string { return fmt.Sprintf(\"%v\", 2) }\n")
	writeTestFile(t, filepath.Join(dir, "a.go"), "package demo\n\nimport \"fmt\"\n\nfunc A() int { return 1 }\n")
	writeTestFile(t, filepath.Join(dir, "a_test.go"), "package demo\n\nfunc TestIgnored() {}\n")

	files, err := readPackageDir(dir)
	if err != nil {
		t.Fatalf("readPackageDir() error = %v", err)
	}
	combined := combinedTestSource(files)
	if strings.Contains(combined, "package demo") {
		t.Fatalf("combined source still contains package clause:\n%s", combined)
	}
	if !strings.Contains(combined, "func A() int") || !strings.Contains(combined, "func B() string") {
		t.Fatalf("combined source missing package files:\n%s", combined)
	}
	if strings.Contains(combined, "TestIgnored") {
		t.Fatalf("combined source included _test.go:\n%s", combined)
	}
	if len(files) != 2 || filepath.Base(files[0].Filename) != "a.go" || filepath.Base(files[1].Filename) != "b.go" {
		t.Fatalf("files are not sorted by filename: %#v", files)
	}
}

func TestReadPackageDirRejectsPackageMismatch(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.go"), "package one\n\nfunc A() {}\n")
	writeTestFile(t, filepath.Join(dir, "b.go"), "package two\n\nfunc B() {}\n")

	_, err := readPackageDir(dir)
	if err == nil {
		t.Fatalf("readPackageDir() succeeded; want package mismatch error")
	}
	if !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("readPackageDir() error = %v, want package mismatch", err)
	}
}

func TestReadBuildTargetKeepsPackageClausesAndSkipsTestFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "b.go"), "package demo\n\nfunc B() {}\n")
	writeTestFile(t, filepath.Join(dir, "a.go"), "package demo\n\nfunc A() {}\n")
	writeTestFile(t, filepath.Join(dir, "a_test.go"), "package demo\n\nfunc TestA() {}\n")

	files, err := readBuildTarget(dir)
	if err != nil {
		t.Fatalf("readBuildTarget() error = %v", err)
	}
	if len(files) != 2 || filepath.Base(files[0].Filename) != "a.go" || filepath.Base(files[1].Filename) != "b.go" {
		t.Fatalf("files are not sorted package files: %#v", files)
	}
	combined := combinedTestSource(files)
	if !strings.Contains(combined, "package demo") {
		t.Fatalf("build source should retain package clause:\n%s", combined)
	}
	if strings.Contains(combined, "TestA") {
		t.Fatalf("build source included _test.go:\n%s", combined)
	}
}

func TestDeriveBuildImportPathUsesGOPATHSrc(t *testing.T) {
	old := os.Getenv("GOPATH")
	t.Cleanup(func() {
		if old == "" {
			_ = os.Unsetenv("GOPATH")
		} else {
			_ = os.Setenv("GOPATH", old)
		}
	})

	gopath := t.TempDir()
	if err := os.Setenv("GOPATH", gopath); err != nil {
		t.Fatalf("Setenv GOPATH error = %v", err)
	}
	dir := filepath.Join(gopath, "src", "example.com", "demo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	got := deriveBuildImportPath(dir, "demo")
	if got != "example.com/demo" {
		t.Fatalf("deriveBuildImportPath() = %q, want example.com/demo", got)
	}
}

func TestReadTestTargetDirectoryIncludesTestFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.go"), "package demo\n\nfunc A() int { return 1 }\n")
	writeTestFile(t, filepath.Join(dir, "a_test.go"), "package demo\n\nfunc TestA() {}\n")

	files, err := readTestTarget(dir)
	if err != nil {
		t.Fatalf("readTestTarget(dir) error = %v", err)
	}
	combined := combinedTestSource(files)
	if strings.Contains(combined, "package demo") {
		t.Fatalf("combined source still contains package clause:\n%s", combined)
	}
	if !strings.Contains(combined, "func A() int") || !strings.Contains(combined, "func TestA()") {
		t.Fatalf("combined test source missing package or test files:\n%s", combined)
	}
}

func TestReadTestTargetTestFileIncludesSiblingPackageFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.go"), "package demo\n\nfunc A() int { return 1 }\n")
	writeTestFile(t, filepath.Join(dir, "b_test.go"), "package demo\n\nfunc TestB() { _ = A() }\n")
	writeTestFile(t, filepath.Join(dir, "c_test.go"), "package demo\n\nfunc TestC() {}\n")

	files, err := readTestTarget(filepath.Join(dir, "b_test.go"))
	if err != nil {
		t.Fatalf("readTestTarget(file) error = %v", err)
	}
	combined := combinedTestSource(files)
	if !strings.Contains(combined, "func A() int") || !strings.Contains(combined, "func TestB()") {
		t.Fatalf("combined test source missing target test or package files:\n%s", combined)
	}
	if strings.Contains(combined, "func TestC()") {
		t.Fatalf("combined test source included another _test.go file:\n%s", combined)
	}
}

func TestNodeRuntimeUsesEnvironmentRandomSeed(t *testing.T) {
	t.Setenv("GOJR_RANDOM_SEED", "gojr-select-seed")
	moduleBundle, err := runtimeModuleBundle()
	if err != nil {
		t.Fatalf("runtimeModuleBundle() error = %v", err)
	}
	rt, err := newNodeRuntime(moduleBundle)
	if err != nil {
		t.Fatalf("newNodeRuntime() error = %v", err)
	}
	t.Cleanup(rt.Close)

	source := `ch1 := make(chan int, 2)
ch2 := make(chan int, 2)
ch1 <- 1
ch1 <- 3
ch2 <- 2
ch2 <- 4
a := 0
b := 0
select {
case a = <-ch1:
case a = <-ch2:
}
select {
case b = <-ch1:
case b = <-ch2:
}
return a, b
`

	result := mustEval(t, rt, source)
	if result.Value != "2, 1" {
		t.Fatalf("GOJR_RANDOM_SEED gojr-select-seed value = %q, want 2, 1", result.Value)
	}

	for _, source := range []string{
		"c := make(chan int)",
		"go func() { c <- 1 }()",
		"gotFromC := <-c",
	} {
		result, err := rt.Eval(source)
		if err != nil {
			t.Fatalf("Eval(%q) error = %v", source, err)
		}
		if !result.OK {
			t.Fatalf("Eval(%q) diagnostics = %v", source, result.Diagnostics)
		}
	}

	result = mustEval(t, rt, "gotFromC")
	if result.Value != "1" {
		t.Fatalf("Eval(gotFromC) value = %q, want 1", result.Value)
	}

	_ = mustEval(t, rt, "dead := make(chan int)")
	_ = mustEval(t, rt, "sink := 0")
	result, err = rt.Eval("sink = <-dead")
	if err != nil {
		t.Fatalf("Eval(deadlock) error = %v", err)
	}
	if result.OK {
		t.Fatalf("Eval(deadlock) succeeded; want deadlock diagnostic")
	}
	if len(result.Diagnostics) != 1 || !strings.Contains(result.Diagnostics[0], "GOJR_DEADLOCK001") {
		t.Fatalf("Eval(deadlock) diagnostics = %v, want GOJR_DEADLOCK001", result.Diagnostics)
	}

	_ = mustEval(t, rt, "go func() { dead <- 3 }()")
	result = mustEval(t, rt, "sink")
	if result.Value != "0" {
		t.Fatalf("Eval(sink) after deadlock value = %q, want 0", result.Value)
	}
	_ = mustEval(t, rt, "sink = <-dead")
	result = mustEval(t, rt, "sink")
	if result.Value != "3" {
		t.Fatalf("Eval(sink) after fresh receive value = %q, want 3", result.Value)
	}
}

func mustEval(t *testing.T, rt *nodeRuntime, source string) evalResult {
	t.Helper()

	result, err := rt.Eval(source)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	if !result.OK {
		t.Fatalf("Eval() diagnostics = %v", result.Diagnostics)
	}
	return result
}

func combinedTestSource(files []sourceFile) string {
	var out strings.Builder
	for _, file := range files {
		out.WriteString(file.Source)
		out.WriteByte('\n')
	}
	return out.String()
}

func writeTestFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

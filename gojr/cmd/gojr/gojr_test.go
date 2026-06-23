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

	source, err := readPackageDir(dir)
	if err != nil {
		t.Fatalf("readPackageDir() error = %v", err)
	}
	if strings.Contains(source, "package demo") {
		t.Fatalf("combined source still contains package clause:\n%s", source)
	}
	if !strings.Contains(source, "func A() int") || !strings.Contains(source, "func B() string") {
		t.Fatalf("combined source missing package files:\n%s", source)
	}
	if strings.Contains(source, "TestIgnored") {
		t.Fatalf("combined source included _test.go:\n%s", source)
	}
	if strings.Index(source, "a.go") > strings.Index(source, "b.go") {
		t.Fatalf("combined source is not sorted by filename:\n%s", source)
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

func TestReadTestTargetDirectoryIncludesTestFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.go"), "package demo\n\nfunc A() int { return 1 }\n")
	writeTestFile(t, filepath.Join(dir, "a_test.go"), "package demo\n\nfunc TestA() {}\n")

	source, err := readTestTarget(dir)
	if err != nil {
		t.Fatalf("readTestTarget(dir) error = %v", err)
	}
	if strings.Contains(source, "package demo") {
		t.Fatalf("combined source still contains package clause:\n%s", source)
	}
	if !strings.Contains(source, "func A() int") || !strings.Contains(source, "func TestA()") {
		t.Fatalf("combined test source missing package or test files:\n%s", source)
	}
}

func TestReadTestTargetTestFileIncludesSiblingPackageFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.go"), "package demo\n\nfunc A() int { return 1 }\n")
	writeTestFile(t, filepath.Join(dir, "b_test.go"), "package demo\n\nfunc TestB() { _ = A() }\n")
	writeTestFile(t, filepath.Join(dir, "c_test.go"), "package demo\n\nfunc TestC() {}\n")

	source, err := readTestTarget(filepath.Join(dir, "b_test.go"))
	if err != nil {
		t.Fatalf("readTestTarget(file) error = %v", err)
	}
	if !strings.Contains(source, "func A() int") || !strings.Contains(source, "func TestB()") {
		t.Fatalf("combined test source missing target test or package files:\n%s", source)
	}
	if strings.Contains(source, "func TestC()") {
		t.Fatalf("combined test source included another _test.go file:\n%s", source)
	}
}

func writeTestFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

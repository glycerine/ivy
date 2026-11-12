package startingpoint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAssertConditionText(t *testing.T) {
	for _, condition := range []string{"x", "p != nil", "count > 0 && ready", `name == "quoted\\value"`, "next() /* keep */", "first &&\n second"} {
		source := "package p\nfunc f() { assert(" + condition + "); }"
		draft, changes, err := RewriteStandardLibrary(source)
		want := "assert(" + condition + ", " + strconv.Quote(condition) + ")"
		if err != nil || !strings.Contains(draft, want) {
			t.Fatalf("%v: missing %q in %s", err, want, draft)
		}
		if len(changes) != 1 || changes[0].Status != "converted" || anonymousFunctions(draft) != 0 {
			t.Fatal(changes, draft)
		}
	}
}

func TestAssertNestedRewriteAndProtection(t *testing.T) {
	draft, _, err := RewriteStandardLibrary(`package p
func f() { assert(strlen(s) > 0); obj.assert(x); use(assert(x)); }
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`assert(len(s) > 0, "strlen(s) > 0")`, "obj.assert(x)", `use(assert(x, "x"))`} {
		if !strings.Contains(draft, want) {
			t.Fatal(draft)
		}
	}
	protected, _, err := RewriteStandardLibrary("package p\nfunc assert(x int) {}\nfunc f() { assert(x); }")
	if err != nil || strings.Contains(protected, "panic(") {
		t.Fatal(protected, err)
	}
}

func TestAssertPackageHelper(t *testing.T) {
	root := t.TempDir()
	src, out := filepath.Join(root, "src"), filepath.Join(root, "out")
	putSource(t, src, "one/assert.c", "void check(int x) { assert(x > 0); }")
	putSource(t, src, "one/assert2.c", "void second(int x) { assert(x != 0); }")
	putSource(t, src, "two/a.c", "void third(int x) { assert(x > 1); }")
	putSource(t, src, "plain/a.c", "void plain(void) {}")
	report, err := TranslateTree(src, out, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(report.HelperFiles, ",") != "one/assert3.go,two/assert.go" {
		t.Fatal(report.HelperFiles)
	}
	helper, err := os.ReadFile(filepath.Join(out, "one/assert3.go"))
	if err != nil || !strings.Contains(string(helper), "package main") || !strings.Contains(string(helper), "func assert(x bool, message string)") {
		t.Fatal(string(helper), err)
	}
	// Compile and execute translated assertions alongside the generated helper.
	main := `package main
import "fmt"
func main() {
 check(1)
 defer func() { fmt.Println(recover()) }()
 check(0)
}`
	mainFile := filepath.Join(out, "one/main.go")
	if err := os.WriteFile(mainFile, []byte(main), 0644); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(out, "one/*.go"))
	cmd := exec.Command("go", append([]string{"run"}, files...)...)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil || string(output) != "assert.go:3: assertion failed, we do not see: x > 0\n" {
		t.Fatal(string(output), err)
	}
	if _, err := os.Stat(filepath.Join(out, "plain/assert.go")); !os.IsNotExist(err) {
		t.Fatal("unneeded helper", err)
	}
}

func TestAssertDefaultPackage(t *testing.T) {
	root := t.TempDir()
	src, out := filepath.Join(root, "src"), filepath.Join(root, "out")
	putSource(t, src, "logic/a.c", "void check(int x) { assert(x); }")
	_, err := TranslateTree(src, out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	helper, err := os.ReadFile(filepath.Join(out, "logic/assert.go"))
	if err != nil || !strings.HasPrefix(string(helper), "package logic\n") {
		t.Fatal(string(helper), err)
	}
}

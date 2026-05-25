package ivy2go

// Test policy for ivy2go (mirrors ivy2cpp_test.go's policy with Go
// substitutions):
// - Keep tests fast for the inner development loop.
// - Do not shell out from these tests. Stay in-process and call the
//   ivy2go library API directly.
// - Do not invoke `go build` on emitted output from these tests.
//   Validate the generated Go shape with focused string/format-source
//   assertions. Real build smoke lives behind SLOW_GO_TEST and runs in
//   a separate file (added by M5).

import (
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// SlowGoTest true means we actually run `go build` on emitted output.
// Off by default; flipped on by the SLOW_GO_TEST environment variable.
var SlowGoTest bool

func init() {
	_, SlowGoTest = os.LookupEnv("SLOW_GO_TEST")
}

// compileIvySource parses an Ivy source string into a goivy Module.
// Mirrors ivy2cpp_test.go compileIvySource. An empty src is normalized
// to a minimal valid Ivy 1.7 program so callers can pass "" when they
// only need module framing.
func compileIvySource(t *testing.T, src string) *goivy.Module {
	t.Helper()
	if strings.TrimSpace(src) == "" {
		src = "#lang ivy1.7\n"
	} else if !strings.HasPrefix(strings.TrimSpace(src), "#lang") {
		src = "#lang ivy1.7\n" + src
	}
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	sig := goivy.NewSigOn(mod.Cfg.IuCfg)
	if err := goivy.SourceString("test.ivy", src, mod, sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatalf("compile ivy source: %v", err)
	}
	return mod
}

// hasLineWithAllTerms is the same line-scan helper ivy2cpp_test.go
// uses. Returns true iff any single line of s contains all the given
// terms.
func hasLineWithAllTerms(s string, terms ...string) bool {
	for _, line := range strings.Split(s, "\n") {
		ok := true
		for _, term := range terms {
			if !strings.Contains(line, term) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// requireHasLineWithAllTerms fails the test if hasLineWithAllTerms
// returns false. The output is included in the failure message.
func requireHasLineWithAllTerms(t *testing.T, text string, terms ...string) {
	t.Helper()
	if !hasLineWithAllTerms(text, terms...) {
		t.Fatalf("missing line containing all of %v:\n%s", terms, text)
	}
}

// assertGoSourceValid asserts that text parses as a syntactically
// valid Go source file. Useful for Tier 3 static checks beyond the
// format.Source roundtrip finalize already performs.
func assertGoSourceValid(t *testing.T, name, text string) {
	t.Helper()
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, name, text, parser.AllErrors); err != nil {
		t.Fatalf("emitted %s is not valid Go: %v\n---\n%s", name, err, text)
	}
}

// assertGoSourceGofmt asserts that text matches its gofmt output. This
// is stronger than parse-only: it catches stylistic deviations that
// finalize is supposed to normalize away. Trailing whitespace is
// trimmed before comparison so the assertion is robust to the single
// trailing-blank-line gofmt strips.
func assertGoSourceGofmt(t *testing.T, name, text string) {
	t.Helper()
	formatted, err := format.Source([]byte(text))
	if err != nil {
		t.Fatalf("emitted %s is not valid Go: %v\n---\n%s", name, err, text)
	}
	if strings.TrimRight(string(formatted), "\n") != strings.TrimRight(text, "\n") {
		t.Fatalf("emitted %s is not gofmt-clean. diff:\nwant:\n%s\ngot:\n%s", name, string(formatted), text)
	}
}

// --- M1: skeleton tests ----------------------------------------------

func TestGenerateEmptyModuleProducesAlwaysOnFiles(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, name := range []string{"state.go", "init.go", "runtime.go"} {
		if _, ok := out.Files[name]; !ok {
			t.Errorf("Output.Files missing required key %q. keys=%v", name, keysOf(out.Files))
		}
	}
}

func TestGenerateEmptyModuleEveryFileIsGofmtClean(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, text := range out.Files {
		assertGoSourceValid(t, name, text)
		assertGoSourceGofmt(t, name, text)
	}
}

func TestGenerateEveryFileStartsWithPackageClause(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "myivy"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, text := range out.Files {
		if !strings.HasPrefix(text, "package myivy") {
			t.Errorf("%s does not start with 'package myivy':\n%s", name, text)
		}
	}
}

func TestGenerateCustomPackageNameFlowsToOutput(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "class", PackageName: "demo"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.PackageName != "demo" {
		t.Errorf("PackageName = %q, want %q", out.PackageName, "demo")
	}
}

func TestGenerateDefaultPackageDerivedFromBase(t *testing.T) {
	mod := compileIvySource(t, "")
	mod.Name = "MyProtocol"
	out, err := Generate(mod, Config{Target: "impl"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// "MyProtocol" lowercases to "myprotocol"; that's the default.
	if out.PackageName != "myprotocol" {
		t.Errorf("default PackageName = %q, want %q", out.PackageName, "myprotocol")
	}
}

func TestGenerateClassTargetDisablesEmitMain(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "class"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.EmitMain {
		t.Error("target=class should leave EmitMain false")
	}
	if _, has := out.Files["main.go"]; has {
		t.Error("target=class should not emit main.go")
	}
}

func TestGenerateNonMainPackageElidesMain(t *testing.T) {
	// EmitMain requested but package is not "main"; finalize should
	// silently disable so emitted files compile.
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "mypkg"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.EmitMain {
		t.Error("EmitMain should auto-disable when package != main")
	}
	if _, has := out.Files["main.go"]; has {
		t.Error("non-main package should not emit main.go")
	}
}

func TestGenerateMainPackageEmitsMain(t *testing.T) {
	mod := compileIvySource(t, "")
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !out.EmitMain {
		t.Error("package=main should have EmitMain true")
	}
	if _, has := out.Files["main.go"]; !has {
		t.Errorf("package=main should emit main.go. keys=%v", keysOf(out.Files))
	}
}

func TestNormalizeConfigDefaultsTargetToGen(t *testing.T) {
	cfg, _, err := normalizeConfig(Config{})
	if err != nil {
		t.Fatalf("normalizeConfig: %v", err)
	}
	if cfg.Target != "gen" {
		t.Errorf("default Target = %q, want %q", cfg.Target, "gen")
	}
	if cfg.RequestedTarget != "gen" {
		t.Errorf("default RequestedTarget = %q, want %q", cfg.RequestedTarget, "gen")
	}
	if cfg.GoivyImportPath != DefaultGoivyImportPath {
		t.Errorf("default GoivyImportPath = %q, want %q", cfg.GoivyImportPath, DefaultGoivyImportPath)
	}
}

func TestNormalizeConfigRejectsUnknownTarget(t *testing.T) {
	_, _, err := normalizeConfig(Config{Target: "rocket"})
	if err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
}

func TestNormalizeConfigClassRewritesTargetToRepl(t *testing.T) {
	cfg, _, err := normalizeConfig(Config{Target: "class"})
	if err != nil {
		t.Fatalf("normalizeConfig: %v", err)
	}
	if cfg.Target != "repl" {
		t.Errorf("class Target rewrite = %q, want %q", cfg.Target, "repl")
	}
	if cfg.RequestedTarget != "class" {
		t.Errorf("class RequestedTarget = %q, want %q", cfg.RequestedTarget, "class")
	}
	if cfg.EmitMain {
		t.Error("class target should set EmitMain false")
	}
}

func TestMergeParamsRoutesKnownKeys(t *testing.T) {
	cfg, ivy, err := mergeParams(map[string]string{
		"target":   "impl",
		"package":  "demo",
		"trace":    "true",
		"isolate":  "iso1",
		"gomodule": "example.com/demo",
	}, Config{})
	if err != nil {
		t.Fatalf("mergeParams: %v", err)
	}
	if cfg.Target != "impl" {
		t.Errorf("Target = %q, want %q", cfg.Target, "impl")
	}
	if cfg.PackageName != "demo" {
		t.Errorf("PackageName = %q, want %q", cfg.PackageName, "demo")
	}
	if !cfg.Trace {
		t.Error("Trace should be true")
	}
	if cfg.GoModule != "example.com/demo" {
		t.Errorf("GoModule = %q, want %q", cfg.GoModule, "example.com/demo")
	}
	if ivy["isolate"] != "iso1" {
		t.Errorf("ivyParams[isolate] = %q, want %q", ivy["isolate"], "iso1")
	}
}

func TestMergeParamsRejectsUnknown(t *testing.T) {
	_, _, err := mergeParams(map[string]string{"foo": "bar"}, Config{})
	if err == nil {
		t.Fatal("expected error for unknown param, got nil")
	}
}

func TestMergeParamsAcceptsClassnameAlias(t *testing.T) {
	// Accept "classname" as an alias for "package" so ivy2cpp argv
	// invocations work against ivy2go too.
	cfg, _, err := mergeParams(map[string]string{"classname": "demo"}, Config{})
	if err != nil {
		t.Fatalf("mergeParams: %v", err)
	}
	if cfg.PackageName != "demo" {
		t.Errorf("classname → PackageName = %q, want %q", cfg.PackageName, "demo")
	}
}

func TestParseArgsHappyPath(t *testing.T) {
	params, file, err := ParseArgs([]string{"target=impl", "package=demo", "foo.ivy"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if file != "foo.ivy" {
		t.Errorf("file = %q, want foo.ivy", file)
	}
	if params["target"] != "impl" || params["package"] != "demo" {
		t.Errorf("params = %v", params)
	}
}

func TestParseArgsRequiresIvyExtension(t *testing.T) {
	_, _, err := ParseArgs([]string{"target=impl"})
	if err == nil {
		t.Fatal("expected error when no .ivy file given, got nil")
	}
}

func TestVarNameMangling(t *testing.T) {
	cases := map[string]string{
		"<":          "__lt",
		"<=":         "__le",
		">=":         "__ge",
		"loc:x":      "loc__x",
		"a.b.c":      "a__b__c",
		"arr[0]":     "arr__0__",
		"ret:result": "result",
	}
	for in, want := range cases {
		if got := varName(in); got != want {
			t.Errorf("varName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGoIdentReservedWordsSuffixed(t *testing.T) {
	for _, kw := range []string{"type", "select", "len", "func"} {
		got := goIdent(kw)
		if got != kw+"_" {
			t.Errorf("goIdent(%q) = %q, want %q", kw, got, kw+"_")
		}
	}
	if got := goIdent("ordinary"); got != "ordinary" {
		t.Errorf("goIdent(ordinary) = %q, want ordinary", got)
	}
}

func TestGoPackageNameDigitPrefix(t *testing.T) {
	if got := goPackageName("3foo"); got != "pkg_3foo" {
		t.Errorf("goPackageName(3foo) = %q, want pkg_3foo", got)
	}
	if got := goPackageName("MyProto"); got != "myproto" {
		t.Errorf("goPackageName(MyProto) = %q, want myproto", got)
	}
	if got := goPackageName(""); got != "ivygen" {
		t.Errorf("goPackageName('') = %q, want ivygen", got)
	}
}

func TestGoExportedNameDigitPrefix(t *testing.T) {
	if got := goExportedName("3foo"); got != "X3foo" {
		t.Errorf("goExportedName(3foo) = %q, want X3foo", got)
	}
	if got := goExportedName("foo"); got != "Foo" {
		t.Errorf("goExportedName(foo) = %q, want Foo", got)
	}
}

func TestCheckMemberNamesDetectsCollision(t *testing.T) {
	mod := compileIvySource(t, "type color = {red, green, blue}")
	// Force a collision: configure PackageName to the same name as
	// a sort in the module.
	g := &Generator{
		Mod:         mod,
		Config:      Config{PackageName: "color"},
		PackageName: "color",
	}
	if err := g.checkMemberNames(); err == nil {
		t.Error("expected collision error for PackageName=color, got nil")
	}
}

// keysOf returns sorted keys of m, for stable test failure messages.
func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sortStrings(out)
	return out
}

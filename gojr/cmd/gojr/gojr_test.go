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

func TestBuildSourceRootsUsesExplicitInferredAndGOPATHRoots(t *testing.T) {
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
	workspace := t.TempDir()
	targetDir := filepath.Join(workspace, "example.com", "app")
	if err := os.MkdirAll(targetDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	explicit := filepath.Join(workspace, "manual-root")

	roots := buildSourceRoots(targetDir, "example.com/app", []string{explicit, explicit})
	want := []string{
		filepath.Clean(explicit),
		filepath.Clean(workspace),
		filepath.Join(gopath, "src"),
	}
	if len(roots) < len(want) || strings.Join(roots[:len(want)], "\n") != strings.Join(want, "\n") {
		t.Fatalf("buildSourceRoots() = %#v, want prefix %#v", roots, want)
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

func assertNodeRuntimeCacheRootListAndClear(t *testing.T, rt *nodeRuntime) {
	t.Helper()
	parent := t.TempDir()
	pathResult, err := rt.Cache(cacheRequest{Action: "path", PackageCacheParent: parent})
	if err != nil {
		t.Fatalf("Cache(path) error = %v", err)
	}
	root := pathResult.Root
	if root != filepath.Join(parent, "gojr_js") {
		t.Fatalf("Cache(path).Root = %q, want gojr_js child", root)
	}
	if ok, err := runCache(rt, []string{"path", "-pkgdir", parent, "-artifact-root", filepath.Join(parent, "exact")}); err == nil || ok {
		t.Fatalf("runCache(path with both roots) ok=%v err=%v, want error", ok, err)
	}

	if err := os.MkdirAll(filepath.Join(root, "example.com"), 0o700); err != nil {
		t.Fatalf("MkdirAll(cache) error = %v", err)
	}
	writeTestFile(t, filepath.Join(root, "example.com", "demo.a"), "artifact")
	writeTestFile(t, filepath.Join(root, "example.com", "ignored.txt"), "nope")
	writeTestFile(t, filepath.Join(root, "top.a"), "top")

	listResult, err := rt.Cache(cacheRequest{Action: "list", ArtifactRoot: root})
	if err != nil {
		t.Fatalf("Cache(list) error = %v", err)
	}
	entries := listResult.Entries
	if len(entries) != 2 {
		t.Fatalf("Cache(list).Entries = %#v, want 2 archive entries", entries)
	}
	if entries[0].ImportPath != "example.com/demo" || entries[1].ImportPath != "top" {
		t.Fatalf("cache import paths = %#v, want example.com/demo and top", entries)
	}
	if ok, err := runCache(rt, []string{"list", "-artifact-root", root}); err != nil || !ok {
		t.Fatalf("runCache(list) ok=%v err=%v, want success", ok, err)
	}
	if ok, err := runCache(rt, []string{"clear", "-artifact-root", root}); err == nil || ok {
		t.Fatalf("runCache(clear without --yes) ok=%v err=%v, want confirmation error", ok, err)
	}
	if ok, err := runCache(rt, []string{"clear", "-artifact-root", root, "--yes"}); err != nil || !ok {
		t.Fatalf("runCache(clear) ok=%v err=%v, want success", ok, err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("cache root still exists after clear, stat err=%v", err)
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

	assertNodeRuntimeBuildSkipsFreshDiskArtifact(t, rt)
	assertNodeRuntimeBuildsSourcePackageGraph(t, rt)
	assertNodeRuntimeCacheRootListAndClear(t, rt)
	assertNodeRuntimeEvaluatesSourcePackageGraphFromSourceRoot(t, rt)
	assertNodeRuntimeRunsFixtureWithSourcePackageGraph(t, rt)
	assertTopLevelCommandsUseNodeRuntime(t, rt)

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

	_ = mustEval(t, rt, "loopC := make(chan int)")
	_ = mustEval(t, rt, "go func() { for i := range 5 { loopC <- i } }()")
	for _, want := range []string{"0", "1", "2", "3", "4"} {
		result = mustEval(t, rt, "<-loopC")
		if result.Value != want {
			t.Fatalf("Eval(<-loopC) value = %q, want %s", result.Value, want)
		}
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

func assertTopLevelCommandsUseNodeRuntime(t *testing.T, rt *nodeRuntime) {
	t.Helper()

	if ok, err := runEval(rt, []string{"--sheet-json", `{"A1":40}`, "sheet.A1.(int64) + 2"}); err != nil || !ok {
		t.Fatalf("runEval() ok=%v err=%v, want success", ok, err)
	}
	evalResult, err := rt.Eval("sheet.A1.(int64) + 2")
	if err != nil {
		t.Fatalf("Eval(sheet.A1.(int64) + 2) error = %v", err)
	}
	if !evalResult.OK || evalResult.Value != "42" {
		t.Fatalf("Eval(sheet.A1.(int64) + 2) = ok %v value %q diagnostics %v, want 42", evalResult.OK, evalResult.Value, evalResult.Diagnostics)
	}
	if len(evalResult.ObservedDeps) != 1 || evalResult.ObservedDeps[0].Kind != "cell" || evalResult.ObservedDeps[0].Sheet != "sheet" || evalResult.ObservedDeps[0].Cell != "A1" {
		t.Fatalf("Eval(sheet.A1.(int64) + 2) observed deps = %#v, want sheet A1", evalResult.ObservedDeps)
	}
	iotaResult := mustEval(t, rt, "const (\n  abit, amask = 1 << iota, 1<<iota - 1\n  bbit, bmask\n)\nreturn abit, amask, bbit, bmask")
	if iotaResult.Value != "1, 0, 2, 1" {
		t.Fatalf("Eval(iota const group) value = %q diagnostics %v, want 1, 0, 2, 1", iotaResult.Value, iotaResult.Diagnostics)
	}

	runFile := filepath.Join(t.TempDir(), "run.go")
	writeTestFile(t, runFile, "package demo\n\nreturn 6 * 7\n")
	if ok, err := runSource(rt, []string{runFile}); err != nil || !ok {
		t.Fatalf("runSource() ok=%v err=%v, want success", ok, err)
	}

	pkgDir := t.TempDir()
	writeTestFile(t, filepath.Join(pkgDir, "counter.go"), `package counter

var Count int

func init() {
	Count = 40
}

func Next() int {
	Count++
	return Count
}
`)
	if ok, err := runEval(rt, []string{"--pkg", "example.com/counter=" + pkgDir, `import counter "example.com/counter"; return counter.Next()`}); err != nil || !ok {
		t.Fatalf("runEval(--pkg) ok=%v err=%v, want success", ok, err)
	}
	pkgSpec, err := readRuntimePackageSpecs([]string{"example.com/counter=" + pkgDir})
	if err != nil {
		t.Fatalf("readRuntimePackageSpecs() error = %v", err)
	}
	pkgResult, err := rt.EvalWithPackages(evalWithPackagesRequest{
		Source:   `import counter "example.com/counter"; return counter.Next()`,
		Packages: pkgSpec,
	})
	if err != nil {
		t.Fatalf("EvalWithPackages() error = %v", err)
	}
	if !pkgResult.OK || pkgResult.Value != "41" {
		t.Fatalf("EvalWithPackages() ok=%v value=%q diagnostics=%v, want 41", pkgResult.OK, pkgResult.Value, pkgResult.Diagnostics)
	}

	if ok, err := runCompile(rt, []string{"--sheet-json", `{"A1":40}`, "--expr", "sheet.A1.(int64) + 2"}); err != nil || !ok {
		t.Fatalf("runCompile(expr) ok=%v err=%v, want success", ok, err)
	}
	compileResult, err := rt.Compile(compileRequest{
		Files: []sourceFile{{
			Filename: "bad.go",
			Source:   "var a = 10\nvar s = \"hi\"\nreturn a + s\n",
		}},
	})
	if err != nil {
		t.Fatalf("Compile(bad) error = %v", err)
	}
	if compileResult.OK || len(compileResult.Diagnostics) != 1 || !strings.Contains(compileResult.Diagnostics[0], "GOJR_TYPE001") {
		t.Fatalf("Compile(bad) = ok %v diagnostics %v, want GOJR_TYPE001", compileResult.OK, compileResult.Diagnostics)
	}

	testDir := t.TempDir()
	writeTestFile(t, filepath.Join(testDir, "calc.go"), "package calc\n\nfunc Add(a, b int) int { return a + b }\n")
	writeTestFile(t, filepath.Join(testDir, "calc_test.go"), "package calc\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) { if Add(2, 5) != 7 { t.Fatalf(\"bad add\") } }\n")
	if ok, err := runCompile(rt, []string{testDir}); err != nil || !ok {
		t.Fatalf("runCompile(dir) ok=%v err=%v, want success", ok, err)
	}
	if ok, err := runTest(rt, []string{testDir}); err != nil || !ok {
		t.Fatalf("runTest() ok=%v err=%v, want success", ok, err)
	}
}

func assertNodeRuntimeBuildSkipsFreshDiskArtifact(t *testing.T, rt *nodeRuntime) {
	t.Helper()

	artifactRoot := filepath.ToSlash(t.TempDir())
	request := buildRequest{
		ImportPath:   "example.com/cache",
		ArtifactRoot: artifactRoot,
		Files: []sourceFile{{
			Filename: "cache.go",
			Source:   "package cache\nfunc F() int { return 1 }\n",
		}},
	}
	artifactPath := artifactRoot + "/example.com/cache.a"

	first, err := rt.Build(request)
	if err != nil {
		t.Fatalf("Build(first) error = %v", err)
	}
	if !first.OK {
		t.Fatalf("Build(first) diagnostics = %v", first.Diagnostics)
	}
	if len(first.Built) != 1 || first.Built[0] != artifactPath {
		t.Fatalf("Build(first).Built = %#v, want %s", first.Built, artifactPath)
	}
	if len(first.Artifacts) != 1 || first.Artifacts[0].Action != "built" {
		t.Fatalf("Build(first).Artifacts = %#v, want built artifact", first.Artifacts)
	}
	if _, err := os.Stat(filepath.FromSlash(artifactPath)); err != nil {
		t.Fatalf("built artifact missing: %v", err)
	}

	second, err := rt.Build(request)
	if err != nil {
		t.Fatalf("Build(second) error = %v", err)
	}
	if !second.OK {
		t.Fatalf("Build(second) diagnostics = %v", second.Diagnostics)
	}
	if len(second.Built) != 0 || len(second.Skipped) != 1 || second.Skipped[0] != artifactPath {
		t.Fatalf("Build(second) built=%#v skipped=%#v, want skip %s", second.Built, second.Skipped, artifactPath)
	}
	if len(second.Artifacts) != 1 || second.Artifacts[0].Action != "skipped" {
		t.Fatalf("Build(second).Artifacts = %#v, want skipped artifact", second.Artifacts)
	}

	inspectRoot := filepath.ToSlash(t.TempDir())
	inspectRequest := buildRequest{
		ImportPath:   "example.com/inspect",
		ArtifactRoot: inspectRoot,
		Files: []sourceFile{{
			Filename: "inspect.go",
			Source:   "package inspect\nfunc Answer() int { return 42 }\n",
		}},
	}
	inspect, err := rt.InspectJS(inspectRequest)
	if err != nil {
		t.Fatalf("InspectJS() error = %v", err)
	}
	if !inspect.OK {
		t.Fatalf("InspectJS() diagnostics = %v", inspect.Diagnostics)
	}
	if !strings.Contains(inspect.Source, "gojrPackageArtifact") || !strings.Contains(inspect.Source, "example.com/inspect") {
		t.Fatalf("InspectJS().Source missing artifact envelope:\n%s", inspect.Source)
	}
	inspectArtifact := filepath.FromSlash(inspectRoot + "/example.com/inspect.a")
	if _, err := os.Stat(inspectArtifact); !os.IsNotExist(err) {
		t.Fatalf("InspectJS wrote artifact %s, stat err=%v", inspectArtifact, err)
	}
}

func assertNodeRuntimeBuildsSourcePackageGraph(t *testing.T, rt *nodeRuntime) {
	t.Helper()

	artifactRoot := filepath.ToSlash(t.TempDir())
	request := buildRequest{
		ImportPath:   "example.com/app",
		ArtifactRoot: artifactRoot,
		PackageSources: packageSources{
			"example.com/lib": []sourceFile{{
				Filename: "lib.go",
				Source:   "package lib\n\nfunc One() int { return 1 }\n",
			}},
		},
		Files: []sourceFile{{
			Filename: "app.go",
			Source:   "package app\n\nimport lib \"example.com/lib\"\n\nfunc Two() int { return lib.One() + 1 }\n",
		}},
	}
	result, err := rt.Build(request)
	if err != nil {
		t.Fatalf("Build(graph) error = %v", err)
	}
	if !result.OK {
		t.Fatalf("Build(graph) diagnostics = %v", result.Diagnostics)
	}
	wantBuilt := []string{
		artifactRoot + "/example.com/lib.a",
		artifactRoot + "/example.com/app.a",
	}
	if strings.Join(result.Built, "\n") != strings.Join(wantBuilt, "\n") {
		t.Fatalf("Build(graph).Built = %#v, want %#v", result.Built, wantBuilt)
	}
	if len(result.Artifacts) != 2 || result.Artifacts[0].ImportPath != "example.com/lib" || result.Artifacts[1].ImportPath != "example.com/app" {
		t.Fatalf("Build(graph).Artifacts = %#v, want lib then app", result.Artifacts)
	}
	if len(result.Artifacts[1].DependencyCacheKeys) != 1 || !strings.HasPrefix(result.Artifacts[1].DependencyCacheKeys[0], "example.com/lib:") {
		t.Fatalf("Build(graph) app dependency cache keys = %#v", result.Artifacts[1].DependencyCacheKeys)
	}
	cacheResult, err := rt.Cache(cacheRequest{Action: "list", ArtifactRoot: artifactRoot})
	if err != nil {
		t.Fatalf("Cache(list graph) error = %v", err)
	}
	cacheEntries := cacheResult.Entries
	if len(cacheEntries) != 2 {
		t.Fatalf("Cache(list graph).Entries = %#v, want lib and app", cacheEntries)
	}
	var appEntry cacheEntry
	for _, entry := range cacheEntries {
		if entry.ImportPath == "example.com/app" {
			appEntry = entry
		}
	}
	if appEntry.ImportPath != "example.com/app" || appEntry.PackageName != "app" || appEntry.CacheKey == "" {
		t.Fatalf("app cache metadata = %#v, want package app with cache key", appEntry)
	}
	if strings.Join(appEntry.Dependencies, ",") != "example.com/lib" || len(appEntry.DependencyCacheKeys) != 1 {
		t.Fatalf("app cache dependencies = deps %#v keys %#v, want lib dependency", appEntry.Dependencies, appEntry.DependencyCacheKeys)
	}
	if len(appEntry.Exports) != 1 || appEntry.Exports[0].Name != "Two" || appEntry.Exports[0].Kind != "func" {
		t.Fatalf("app cache exports = %#v, want exported func Two", appEntry.Exports)
	}

	libDir := t.TempDir()
	appDir := t.TempDir()
	cliRoot := filepath.ToSlash(t.TempDir())
	writeTestFile(t, filepath.Join(libDir, "lib.go"), "package lib\n\nfunc One() int { return 1 }\n")
	writeTestFile(t, filepath.Join(appDir, "app.go"), "package app\n\nimport lib \"example.com/lib\"\n\nfunc Two() int { return lib.One() + 1 }\n")
	if ok, err := runBuild(rt, []string{
		"--pkg", "example.com/lib=" + libDir,
		"-importpath", "example.com/app",
		"-artifact-root", cliRoot,
		appDir,
	}); err != nil || !ok {
		t.Fatalf("runBuild(--pkg) ok=%v err=%v, want success", ok, err)
	}
	if _, err := os.Stat(filepath.FromSlash(cliRoot + "/example.com/lib.a")); err != nil {
		t.Fatalf("runBuild(--pkg) missing lib artifact: %v", err)
	}
	if _, err := os.Stat(filepath.FromSlash(cliRoot + "/example.com/app.a")); err != nil {
		t.Fatalf("runBuild(--pkg) missing app artifact: %v", err)
	}

	srcRoot := t.TempDir()
	providerRoot := filepath.ToSlash(t.TempDir())
	libProviderDir := filepath.Join(srcRoot, "example.com", "lib")
	appProviderDir := filepath.Join(srcRoot, "example.com", "app")
	if err := os.MkdirAll(libProviderDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(libProviderDir) error = %v", err)
	}
	if err := os.MkdirAll(appProviderDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(appProviderDir) error = %v", err)
	}
	writeTestFile(t, filepath.Join(libProviderDir, "lib.go"), "package lib\n\nfunc One() int { return 1 }\n")
	writeTestFile(t, filepath.Join(appProviderDir, "app.go"), "package app\n\nimport lib \"example.com/lib\"\n\nfunc Two() int { return lib.One() + 1 }\n")
	if ok, err := runBuild(rt, []string{
		"--srcroot", srcRoot,
		"-importpath", "example.com/app",
		"-artifact-root", providerRoot,
		appProviderDir,
	}); err != nil || !ok {
		t.Fatalf("runBuild(--srcroot) ok=%v err=%v, want success", ok, err)
	}
	if _, err := os.Stat(filepath.FromSlash(providerRoot + "/example.com/lib.a")); err != nil {
		t.Fatalf("runBuild(--srcroot) missing lib artifact: %v", err)
	}
	if _, err := os.Stat(filepath.FromSlash(providerRoot + "/example.com/app.a")); err != nil {
		t.Fatalf("runBuild(--srcroot) missing app artifact: %v", err)
	}
}

func assertNodeRuntimeEvaluatesSourcePackageGraphFromSourceRoot(t *testing.T, rt *nodeRuntime) {
	t.Helper()

	srcRoot := t.TempDir()
	libDir := filepath.Join(srcRoot, "example.com", "lib")
	appDir := filepath.Join(srcRoot, "example.com", "app")
	if err := os.MkdirAll(libDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(libDir) error = %v", err)
	}
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(appDir) error = %v", err)
	}
	writeTestFile(t, filepath.Join(libDir, "lib.go"), "package lib\n\nfunc One() int { return 1 }\n")
	writeTestFile(t, filepath.Join(appDir, "app.go"), "package app\n\nimport lib \"example.com/lib\"\n\nfunc Two() int { return lib.One() + 1 }\n")

	result, err := rt.EvalWithPackages(evalWithPackagesRequest{
		Source:      `import app "example.com/app"; return app.Two()`,
		SourceRoots: []string{srcRoot},
	})
	if err != nil {
		t.Fatalf("EvalWithPackages(srcroot) error = %v", err)
	}
	if !result.OK || result.Value != "2" {
		t.Fatalf("EvalWithPackages(srcroot) ok=%v value=%q diagnostics=%v, want 2", result.OK, result.Value, result.Diagnostics)
	}

	if ok, err := runEval(rt, []string{
		"--srcroot", srcRoot,
		`import app "example.com/app"; return app.Two()`,
	}); err != nil || !ok {
		t.Fatalf("runEval(--srcroot) ok=%v err=%v, want success", ok, err)
	}

	script := filepath.Join(t.TempDir(), "script.go")
	writeTestFile(t, script, `import app "example.com/app"

return app.Two()
`)
	if ok, err := runSource(rt, []string{"--srcroot", srcRoot, script}); err != nil || !ok {
		t.Fatalf("runSource(--srcroot) ok=%v err=%v, want success", ok, err)
	}

	testDir := t.TempDir()
	writeTestFile(t, filepath.Join(testDir, "use_app_test.go"), `package useapp

import (
	app "example.com/app"
	"testing"
)

func TestTwo(t *testing.T) {
	if app.Two() != 2 {
		t.Fatalf("bad Two")
	}
}
`)
	testResult, err := rt.TestFilesWithPackages(evalWithPackagesRequest{
		Files:       mustReadTestTarget(t, testDir),
		SourceRoots: []string{srcRoot},
	})
	if err != nil {
		t.Fatalf("TestFilesWithPackages(srcroot) error = %v", err)
	}
	if !testResult.OK || !strings.Contains(testResult.Output, "PASS") {
		t.Fatalf("TestFilesWithPackages(srcroot) ok=%v output=%q diagnostics=%v, want PASS", testResult.OK, testResult.Output, testResult.Diagnostics)
	}
	if ok, err := runTest(rt, []string{"--srcroot", srcRoot, testDir}); err != nil || !ok {
		t.Fatalf("runTest(--srcroot) ok=%v err=%v, want success", ok, err)
	}

	compileResult, err := rt.Compile(compileRequest{
		Files: []sourceFile{{
			Filename: "formula.go",
			Source:   `import app "example.com/app"; return app.Two()`,
		}},
		SourceRoots: []string{srcRoot},
	})
	if err != nil {
		t.Fatalf("Compile(srcroot) error = %v", err)
	}
	if !compileResult.OK {
		t.Fatalf("Compile(srcroot) diagnostics = %v, want success", compileResult.Diagnostics)
	}
	badCompile, err := rt.Compile(compileRequest{
		Files: []sourceFile{{
			Filename: "bad_formula.go",
			Source:   `import app "example.com/app"; return app.Nope()`,
		}},
		SourceRoots: []string{srcRoot},
	})
	if err != nil {
		t.Fatalf("Compile(bad srcroot) error = %v", err)
	}
	if badCompile.OK || len(badCompile.Diagnostics) == 0 || !strings.Contains(badCompile.Diagnostics[0], "Nope") {
		t.Fatalf("Compile(bad srcroot) ok=%v diagnostics=%v, want missing selector diagnostic", badCompile.OK, badCompile.Diagnostics)
	}
	if ok, err := runCompile(rt, []string{
		"--srcroot", srcRoot,
		"--expr", `import app "example.com/app"; return app.Two()`,
	}); err != nil || !ok {
		t.Fatalf("runCompile(--srcroot) ok=%v err=%v, want success", ok, err)
	}
}

func assertNodeRuntimeRunsFixtureWithSourcePackageGraph(t *testing.T, rt *nodeRuntime) {
	t.Helper()

	srcRoot := t.TempDir()
	pkgDir := filepath.Join(srcRoot, "example.com", "mathx")
	if err := os.MkdirAll(pkgDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(pkgDir) error = %v", err)
	}
	writeTestFile(t, filepath.Join(pkgDir, "mathx.go"), `package mathx

func Add(a, b int) int {
	return a + b
}
`)
	fixtureJSON := `{
  "cells": {
    "A1": 40,
    "B1": {
      "formula": "import mathx \"example.com/mathx\"\nreturn mathx.Add(int(sheet.A1.(int64)), 2)"
    }
  }
}`
	result, err := rt.RunFixtureWithPackages(fixtureWithPackagesRequest{
		FixtureJSON: fixtureJSON,
		SourceRoots: []string{srcRoot},
	})
	if err != nil {
		t.Fatalf("RunFixtureWithPackages() error = %v", err)
	}
	if !result.OK || result.Sheets["sheet"]["B1"] != "42" {
		t.Fatalf("RunFixtureWithPackages() ok=%v sheets=%#v diagnostics=%v, want B1=42", result.OK, result.Sheets, result.Diagnostics)
	}

	fixtureFile := filepath.Join(t.TempDir(), "fixture.json")
	writeTestFile(t, fixtureFile, fixtureJSON)
	if ok, err := runFixture(rt, []string{"--srcroot", srcRoot, fixtureFile}); err != nil || !ok {
		t.Fatalf("runFixture(--srcroot) ok=%v err=%v, want success", ok, err)
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

func mustReadTestTarget(t *testing.T, target string) []sourceFile {
	t.Helper()

	files, err := readTestTarget(target)
	if err != nil {
		t.Fatalf("readTestTarget(%s) error = %v", target, err)
	}
	return files
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

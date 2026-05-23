package ivy2cpp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func generateThunkEmissionFixture(t *testing.T, className string, rhsNames ...string) *Output {
	t.Helper()
	var src strings.Builder
	src.WriteString(`#lang ivy1.7
type node
relation rel(X:node)
`)
	for i := range rhsNames {
		src.WriteString(fmt.Sprintf("action step%d = {}\n", i))
	}
	mod := compileIvySource(t, src.String())
	node := mod.Sig.Sorts.Get("node")
	fnSort, err := goivy.NewFunctionSort(node, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	rel := goivy.NewConst("rel", fnSort)
	x, err := goivy.NewVariable("X", node)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	lhs := goivy.MustApply(rel, x)
	for i, rhsName := range rhsNames {
		mod.Actions.Set(fmt.Sprintf("step%d", i), goivy.NewAssignAction(lhs, goivy.NewConst(rhsName, goivy.Boolean)))
	}
	out, err := Generate(mod, Config{ClassName: className})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNoUnsupportedCPP(t, out)
	return out
}

func TestThunkStructEmittedAtFileScope(t *testing.T) {
	out := generateThunkEmissionFixture(t, "thunkscope", "true")
	structIdx := strings.Index(out.Impl, "struct __thunk__0 : thunk<int, bool>")
	ctorIdx := strings.Index(out.Impl, "thunkscope::thunkscope()")
	methodIdx := strings.Index(out.Impl, "void thunkscope::step0()")
	if structIdx < 0 || ctorIdx < 0 || methodIdx < 0 {
		t.Fatalf("missing thunk/constructor/method markers:\n%s", out.Impl)
	}
	if !(structIdx < ctorIdx && structIdx < methodIdx) {
		t.Fatalf("thunk should be file-scope before constructor and method: struct=%d ctor=%d method=%d\n%s", structIdx, ctorIdx, methodIdx, out.Impl)
	}
	methodBody := out.Impl[methodIdx:]
	if strings.Contains(methodBody, "struct __thunk__0") {
		t.Fatalf("thunk struct should not be emitted inside the method body:\n%s", methodBody)
	}
	if !strings.Contains(methodBody, "rel = hash_thunk<int, bool>(new __thunk__0())") {
		t.Fatalf("method should instantiate the file-scope thunk:\n%s", methodBody)
	}
	compileGeneratedCPP(t, out)
}

func TestIdenticalThunksEmittedOnce(t *testing.T) {
	out := generateThunkEmissionFixture(t, "thunkonce", "true", "true")
	ctorIdx := strings.Index(out.Impl, "thunkonce::thunkonce()")
	if ctorIdx < 0 {
		t.Fatalf("missing constructor:\n%s", out.Impl)
	}
	if got := strings.Count(out.Impl[:ctorIdx], "struct __thunk__"); got != 1 {
		t.Fatalf("identical thunk bodies should emit once, got %d:\n%s", got, out.Impl)
	}
	if !strings.Contains(out.Impl, "void thunkonce::step0()") || !strings.Contains(out.Impl, "void thunkonce::step1()") {
		t.Fatalf("missing generated methods:\n%s", out.Impl)
	}
	if got := strings.Count(out.Impl, "new __thunk__0()"); got != 2 {
		t.Fatalf("both actions should instantiate the memoized thunk, got %d:\n%s", got, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestDifferentBodiesGetDifferentThunkStructs(t *testing.T) {
	out := generateThunkEmissionFixture(t, "thunktwo", "true", "false")
	for _, want := range []string{
		"struct __thunk__0 : thunk<int, bool>",
		"struct __thunk__1 : thunk<int, bool>",
		"return true;",
		"return false;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in different-thunk output:\n%s", want, out.Impl)
		}
	}
	ctorIdx := strings.Index(out.Impl, "thunktwo::thunktwo()")
	if ctorIdx < 0 {
		t.Fatalf("missing constructor:\n%s", out.Impl)
	}
	if got := strings.Count(out.Impl[:ctorIdx], "struct __thunk__"); got != 2 {
		t.Fatalf("different bodies should emit two thunks, got %d:\n%s", got, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestThunkNumberingStableAcrossRuns(t *testing.T) {
	first := generateThunkEmissionFixture(t, "thunkstable", "true", "false")
	second := generateThunkEmissionFixture(t, "thunkstable", "true", "false")
	if first.Impl != second.Impl {
		t.Fatalf("Generate should be byte-stable across runs\nfirst:\n%s\nsecond:\n%s", first.Impl, second.Impl)
	}
}

func TestThunkScopeNoShadowWarning(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to run this test")
	}
	out := generateThunkEmissionFixture(t, "thunkshadow", "true", "false")
	prefix := "#if defined(__GNUC__) || defined(__clang__)\n#pragma GCC diagnostic warning \"-Wshadow\"\n#pragma GCC diagnostic error \"-Wshadow\"\n#endif\n"
	include := "#include \"" + out.BaseName + ".h\"\n"
	if !strings.Contains(out.Impl, include) {
		t.Fatalf("generated impl missing header include %q:\n%s", include, out.Impl)
	}
	scoped := *out
	scoped.Impl = strings.Replace(out.Impl, include, include+prefix, 1)
	compileGeneratedCPP(t, &scoped)
}

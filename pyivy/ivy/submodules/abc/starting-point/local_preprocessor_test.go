package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestLocalFeatureDefineBeforeIfdef(t *testing.T) {
	r, err := Analyze("x.cpp", "#define USE_SIMP_SOLVER 1\n#ifdef USE_SIMP_SOLVER\nint chosen;\n#else\nint other;\n#endif\n")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Draft, "#ifdef") || strings.Contains(r.Draft, "other") || !strings.Contains(r.Draft, "var chosen int32") {
		t.Fatal(r.Draft)
	}
}

func TestUsingNamespaceComment(t *testing.T) {
	r, err := Analyze("x.cpp", "using namespace Gluco2;\nint value;\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "//using namespace Gluco2;") {
		t.Fatal(r.Draft)
	}
}

func TestAbcGlucose2FunctionsTranslated(t *testing.T) {
	path := "../src/sat/glucose2/AbcGlucose2.cpp"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := AnalyzeWithOptions(path, string(source), Options{IncludePaths: []string{"../src"}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Draft, "#ifdef USE_SIMP_SOLVER") || !strings.Contains(r.Draft, "//using namespace Gluco2;") {
		t.Fatal("feature/namespace not converted")
	}
	if len(r.Functions) != 61 {
		t.Fatal("missing functions", len(r.Functions))
	}
	for _, f := range r.Functions {
		if !strings.HasPrefix(strings.TrimSpace(f.Draft), "func ") {
			t.Errorf("retained signature: %s", f.Name)
		}
	}
	if !strings.Contains(r.Draft, "func glucose2_solver_stop(S *Gluco2.SimpSolver)") {
		t.Error("missing qualified signature")
	}
}

func TestLocalPreprocessorChoices(t *testing.T) {
	for _, tc := range []struct {
		name, prefix, test, want string
		cfg                      map[string]int64
	}{
		{"zero present", "#define X 0\n", "#ifdef X", "yes", nil},
		{"zero numeric", "#define X 0\n", "#if X", "no", nil},
		{"defined zero", "#define X 0\n", "#if defined(X)", "yes", nil},
		{"empty present", "#define X\n", "#ifdef X", "yes", nil},
		{"function present", "#define X(a) a\n", "#ifdef X", "yes", nil},
		{"undef", "#define X 1\n#undef X\n", "#ifndef X", "yes", nil},
		{"redefine", "#define X 0\n#define X 2u\n", "#if X > 1", "yes", nil},
		{"configured override", "#define X 1\n", "#ifdef X", "no", map[string]int64{"X": 0}},
		{"unknown redefine", "#define X 1\n#if UNKNOWN\n#undef X\n#endif\n", "#ifdef X", "opaque", nil},
		{"include barrier", "#define X 1\n#include \"external.h\"\n", "#ifdef X", "opaque", nil},
		{"unknown expression", "#define X dynamic()\n", "#if X", "opaque", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := tc.prefix + tc.test + "\nint yes;\n#else\nint no;\n#endif\n"
			got, err := selectConfiguredBranches(source, tc.cfg)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(source) || strings.Count(got, "\n") != strings.Count(source, "\n") {
				t.Fatal("coordinates changed")
			}
			if tc.want == "opaque" {
				if !strings.Contains(got, tc.test) {
					t.Fatal(got)
				}
				return
			}
			if !strings.Contains(got, "int "+tc.want+";") || strings.Contains(got, tc.test) {
				t.Fatal(got)
			}
			other := "yes"
			if tc.want == "yes" {
				other = "no"
			}
			if strings.Contains(got, "int "+other+";") {
				t.Fatal(got)
			}
		})
	}
	source := "#define FIRST 1\n#if FIRST\n#define SECOND 1\n#else\n#define SECOND 0\n#endif\n#if SECOND\nint yes;\n#elif UNKNOWN\nint no;\n#endif\n"
	got, err := selectConfiguredBranches(source, nil)
	if err != nil || !strings.Contains(got, "int yes;") || strings.Contains(got, "int no;") {
		t.Fatal(err, got)
	}
}

func TestUsingNamespaceShapes(t *testing.T) {
	source := `using namespace ::outer::inner; int keep;
void f() { using /* note */ namespace local; call(); }
using
namespace multiline;
using Alias = int;
using std::vector;
const char *s = "using namespace untouched;";
// using namespace already;
#if UNKNOWN
using namespace hidden;
#endif
`
	got, err := commentUsingNamespaces(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"//using namespace ::outer::inner;\n int keep;", "//using /* note */ namespace local;\n call();", "//using\n//namespace multiline;", "using Alias = int;", "using std::vector;", `"using namespace untouched;"`, "// using namespace already;", "//using namespace hidden;"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
	again, err := commentUsingNamespaces(got)
	if err != nil || again != got {
		t.Fatal("not idempotent", err, again)
	}
}

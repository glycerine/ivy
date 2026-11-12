package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestPackageVariables(t *testing.T) {
	source := `static int nThreadsRunning = 0;
pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
int first, second = 0;
static int nonzero = 4;
int prototype(int x);
struct record { int field; };
void f(void) { int local = 0; }
#if MODE
static int conditional = 0;
#endif
`
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var nThreadsRunning int32;", "var mutex sync.Mutex", "var first, second int32;", "var nonzero int32 = 4;", "// int prototype(int x);", "var local int32;", "#if MODE\nstatic int conditional = 0;\n#endif"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestStarterPackageVariables(t *testing.T) {
	source, err := os.ReadFile("../src/starter.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("starter.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var nThreadsRunning int32;", "var mutex sync.Mutex"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

func TestForwardDeclarationsAndExternVariables(t *testing.T) {
	source := `static void Abc_NtkPrintCuts( void * p, Abc_Ntk_t * pNtk, int fSeq );
static void Abc_NtkPrintCuts_( void * p, Abc_Ntk_t * pNtk, int fSeq );
extern int nTotal, nGood, nEqual;
extern int initialized = 0;
extern char *name;
static int multiline(
  int arg /* keep comment */
); int following;
int sameLine(void); int body(void) { return 1; }
int (*callback)(int);
#if CHOICE
extern int opaque;
#endif
`
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"// static void Abc_NtkPrintCuts(", "// static void Abc_NtkPrintCuts_(",
		"var nTotal, nGood, nEqual int32;", "var initialized int32;", "var name string;",
		"// static int multiline(\n//   int arg /* keep comment */\n// );",
		"var following int32;", "// int sameLine(void);\n", "func body() int32",
		"int (*callback)(int);", "#if CHOICE\nextern int opaque;\n#endif",
	} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestAbcCutForwardDeclarations(t *testing.T) {
	source, err := os.ReadFile("../src/opt/cut/abcCut.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("abcCut.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"// static void Abc_NtkPrintCuts(", "// static void Abc_NtkPrintCuts_(", "var nTotal, nGood, nEqual int32;", "func Abc_NtkPrintCuts(", "var pProgress *ProgressBar;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

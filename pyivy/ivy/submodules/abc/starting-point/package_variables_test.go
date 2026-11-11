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
	for _, want := range []string{"var nThreadsRunning int32;", "var mutex pthread_mutex_t = PTHREAD_MUTEX_INITIALIZER;", "var first, second int32;", "var nonzero int32 = 4;", "int prototype(int x);", "var local int32;", "#if MODE\nstatic int conditional = 0;\n#endif"} {
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
	for _, want := range []string{"var nThreadsRunning int32;", "var mutex pthread_mutex_t = PTHREAD_MUTEX_INITIALIZER;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
}

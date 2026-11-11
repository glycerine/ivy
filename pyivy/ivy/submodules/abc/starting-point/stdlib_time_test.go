package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestTimeLibraryDirectCalls(t *testing.T) {
	draft, changes, err := RewriteStandardLibrary(`package p
func f() { use(time(nil), time(&stored), clock(), difftime(12, 7)); }
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"time.Now().Unix()", "time(&stored)", "time.Now().UnixNano()", "difftime(12, 7)"} {
		if !strings.Contains(draft, want) {
			t.Fatal(draft)
		}
	}
	if anonymousFunctions(draft) != 0 {
		t.Fatal(draft)
	}
	reviews := 0
	for _, c := range changes {
		if c.Status == "review" {
			reviews++
		}
	}
	if reviews != 2 {
		t.Fatal(changes)
	}
}

func TestStarterClockConversion(t *testing.T) {
	source, err := os.ReadFile("../src/starter.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("starter.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`import "time"`, "var clk int64", "time.Now().UnixNano() - clk", "int64(time.Second)"} {
		// Multiple requested imports may be grouped.
		if want == `import "time"` {
			want = `"time"`
		}
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
	found := false
	for _, change := range r.Library {
		if change.Name == "clock" && strings.Contains(change.Message, "CPU time to wall-clock") {
			found = true
		}
	}
	if !found {
		t.Fatal("clock policy not recorded")
	}
}

func TestTimeImportCollisionAndOpaqueRegions(t *testing.T) {
	draft, _, err := RewriteStandardLibrary("package p\nfunc f() { var time int; use(time, clock()); }")
	if err != nil || !strings.Contains(draft, "spTime.Now()") {
		t.Fatal(draft, err)
	}
	r, err := Analyze("x.c", "#if X\ntime_t stamp = time(NULL);\n#endif\n")
	if err != nil || !strings.Contains(r.Draft, "time(NULL)") || strings.Contains(r.Draft, `import "time"`) {
		t.Fatal(r.Draft, err)
	}
}

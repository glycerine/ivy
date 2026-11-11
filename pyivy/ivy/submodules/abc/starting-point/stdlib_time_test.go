package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestTimeLibraryRuntime(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
func main() {
 var stored int64
 stamp := time(&stored)
 fmt.Println(stamp == stored, stamp > 0, time(nil) >= stamp)
 first := clock()
 second := clock()
 fmt.Println(second >= first, CLOCKS_PER_SEC == 1000000000)
 fmt.Println(difftime(12, 7), difftime(-3, 4))
 fmt.Println(difftime(int64(9223372036854775807), int64(-9223372036854775807-1)) > 1e19)
}`)
	if output != "true true true\ntrue true\n5 -7\ntrue\n" {
		t.Fatal(output)
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

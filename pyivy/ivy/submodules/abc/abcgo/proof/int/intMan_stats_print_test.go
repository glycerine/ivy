package int

import (
	"io"
	"os"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// Inter_ManStop -v prints intMan.c's ABC_PRTP table; the port printed one
// line of Go durations ("total 124ms (rewriting 0s, ...)").
func TestInterManStopRuntimeLikeABC(t *testing.T) {
	sec := util.CLOCKS_PER_SEC
	p := &Inter_Man_t{fVerbose: 1, timeRwr: sec / 4, timeCnf: 0, timeSat: sec / 2, timeInt: sec / 4, timeEqu: 0, timeTotal: 2 * sec}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = w
	Inter_ManStop(p, 0)
	os.Stdout = saved
	w.Close()
	out, _ := io.ReadAll(r)
	want := "Runtime statistics:\n" +
		"Rewriting   =     0.25 sec ( 12.50 %)\n" +
		"CNF mapping =     0.00 sec (  0.00 %)\n" +
		"SAT solving =     0.50 sec ( 25.00 %)\n" +
		"Interpol    =     0.25 sec ( 12.50 %)\n" +
		"Containment =     0.00 sec (  0.00 %)\n" +
		"Other       =     1.00 sec ( 50.00 %)\n" +
		"TOTAL       =     2.00 sec (100.00 %)\n"
	if string(out) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

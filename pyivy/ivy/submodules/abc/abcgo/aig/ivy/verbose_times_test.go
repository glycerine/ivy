package ivy

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// captureIvyStdout returns what action prints to the process stdout.
func captureIvyStdout(t *testing.T, action func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan []byte)
	go func() {
		var b bytes.Buffer
		io.Copy(&b, r)
		done <- b.Bytes()
	}()
	func() {
		defer func() { os.Stdout = stdout }()
		action()
	}()
	w.Close()
	return string(<-done)
}

var ivySecTimes = regexp.MustCompile(`= +[0-9]+\.[0-9]{2} sec`)

// Times are ABC_PRT times ("Time =%9.2f sec") after the same text as ABC;
// the ports printed Go durations.
func TestIvyVerboseTimesLikeABC(t *testing.T) {
	build := func() *Ivy_Man_t {
		p := Ivy_ManStart()
		a, b, c, d := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
		Ivy_ObjCreatePo(p, Ivy_And(p, Ivy_And(p, a, b), Ivy_And(p, c, d)))
		return p
	}
	p := build()
	out := ivySecTimes.ReplaceAllString(captureIvyStdout(t, func() { Ivy_FastMapPerform(p, 3, 1, 1) }), "= T sec")
	want := "Delay oriented mapping:  : Delay =   2. Area =      3. Time = T sec\n" +
		"Area recovery 2       :  : Delay =   2. Area =      3. Time = T sec\n" +
		"Area recovery 1       :  : Delay =   2. Area =      3. Time = T sec\n" +
		"Area recovery 2       :  : Delay =   2. Area =      3. Time = T sec\n"
	if out != want {
		t.Errorf("Ivy_FastMapPerform -v:\n%q\nwant\n%q", out, want)
	}
	Ivy_FastMapStop(p)

	for i, test := range []func(*Ivy_Man_t){Ivy_ManTestCutsAll, Ivy_ManTestCutsTravAll} {
		out = ivySecTimes.ReplaceAllString(captureIvyStdout(t, func() { test(build()) }), "= T sec")
		if !regexp.MustCompile(`^Total cuts = +[0-9]+\. Trivial = +7\.   Nodes = +3\. Satur = +0\.  Time = T sec\n$`).MatchString(out) {
			t.Errorf("cut test %d: %q", i, out)
		}
	}

	out = ivySecTimes.ReplaceAllString(captureIvyStdout(t, func() { Ivy_FraigMiterPrint(build(), "SAT solving", util.Abc_Clock(), 1) }), "= T sec")
	if want := "Nodes =       3.  Levels =    2.  SAT solving = T sec\n"; out != want || strings.Contains(out, "ms") {
		t.Errorf("Ivy_FraigMiterPrint: %q", out)
	}
}

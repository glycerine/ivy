package gia

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// captureGiaStdout returns what action prints to the process stdout.
func captureGiaStdout(t *testing.T, action func()) string {
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

// Times are ABC times: "Time =%9.2f sec" after the counters, as
// Abc_PrintTime prints them; the ports printed Go durations.
func TestGiaVerboseTimesLikeABC(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashAlloc(g)
	Gia_ManAppendCo(g, Gia_ManHashOr(g, a, b))
	man := Gia_QbfAlloc(g, 1, 0, 0)
	defer Gia_QbfFree(man)
	values := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(values, 1)
	out := captureGiaStdout(t, func() { Gia_QbfPrint(man, values, 3) })
	want := regexp.MustCompile(`^    3 : 1  Var =      1  Cla =      0  Conf =        0  Time = +[0-9]+\.[0-9]{2} sec\n$`)
	if !want.MatchString(out) {
		t.Errorf("Gia_QbfPrint: %q", out)
	}
}

func TestHcdComputeChoicesVerboseLikeABC(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.AND(a, b), c))
	out := captureGiaStdout(t, func() { Hcd_ComputeChoices(g, 1000, 1, 0, 1) })
	out = giaSecTimes.ReplaceAllString(out, "= T sec")
	for _, want := range []string{
		"Synthesis time = T sec\nSim = T sec\n",
		"Ref = T sec\nLEVEL   1  (out of   2)   nSaved = 0   nRecords = 0   nUndec = 0   nClassRefs = 0   nMiniSat = 0   nTas = 0\nTas     = T sec\nMiniSat = T sec\nSim     = T sec\nTotal   = T sec\n",
		"Iters =   1   Time = T sec\n",
		"Choicing time = T sec\nChoices   : Reprs =     0. Equivs =     1. Choices =     1.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

var giaSecTimes = regexp.MustCompile(`= +[0-9]+\.[0-9]{2} sec`)

func TestGiaSwiSimulateVerboseLikeABC(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashAlloc(g)
	Gia_ManAppendCo(g, Gia_ManHashAnd(g, a, b))
	var pars Gia_ParSwi_t
	Gia_ManSetDefaultParamsSwi(&pars)
	pars.fVerbose = 1
	out := captureGiaStdout(t, func() { Gia_ManSwiSimulate(g, &pars) })
	want := regexp.MustCompile(`^Obj = +5 \( +3\)\. F = +[0-9]+\. AIG = +[0-9.]+ MB\. F-mem = +[0-9.]+ MB\. Other = +[0-9.]+ MB\.  Time = +[0-9]+\.[0-9]{2} sec\n` +
		`Simulated 48 frames with 10 words\. Simulation time = +[0-9]+\.[0-9]{2} sec\n$`)
	if !want.MatchString(out) {
		t.Errorf("Gia_ManSwiSimulate -v:\n%s", out)
	}
}

func TestGiaSimVerboseTimesLikeABC(t *testing.T) {
	sims := vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(sims, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(sims, 0xcccccccccccccccc)
	out := captureGiaStdout(t, func() {
		Gia_SimAbsPerformOne(nil, []uint64{0x5555555555555555}, []uint64{0xaaaaaaaaaaaaaaaa}, sims, 1, 1)
	})
	// Gia_SimAbsSolve pads the solution to 12 columns of "%6d "
	want := "Solution  1 for covering problem [   64 x     2]:      0 " + strings.Repeat("       ", 11) + "   Time = T sec\n"
	if got := giaSecTimes.ReplaceAllString(out, "= T sec"); !strings.Contains(got, want) || !strings.HasSuffix(got, "Resubstitution time = T sec\n") {
		t.Errorf("Gia_SimAbsPerformOne -v:\n%q\nwant\n%q", got, want)
	}

	p := Gia_ManStart(24)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(p)
	}
	x := Gia_ManAppendAnd(p, a, b^1)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCo(p, x)
	}
	Gia_ManSetRegNum(p, 5)
	out = captureGiaStdout(t, func() { Gia_ManSimDeriveResets(p) })
	if !regexp.MustCompile(`^Logic0 = [0-9]+ \([0-9]+\)\. Logic1 = [0-9]+ \([0-9]+\)\. Time = +[0-9]+\.[0-9]{2} sec\n$`).MatchString(out) {
		t.Errorf("Gia_ManSimDeriveResets: %q", out)
	}

	nodes := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(nodes, x>>1)
	out = captureGiaStdout(t, func() { Gia_ManCreateNodeSupps(p, nodes, 1) })
	if !regexp.MustCompile(`^Support computation = +[0-9]+\.[0-9]{2} sec\n$`).MatchString(out) {
		t.Errorf("Gia_ManCreateNodeSupps -v: %q", out)
	}
}

func TestGiaResubTimesLikeABC(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	out := giaSecTimes.ReplaceAllString(captureGiaStdout(t, func() { Gia_ManTryResub(g) }), "= T sec")
	if want := "Total nodes =     1.  Non-realizable =     1.  Gain =      0.  Time = T sec\nPure resub time = T sec\n"; out != want {
		t.Errorf("Gia_ManTryResub:\n%q\nwant\n%q", out, want)
	}
	out = captureGiaStdout(t, func() { Gia_ManResubTest(g) })
	if !regexp.MustCompile(`for pivots\.   Time = +[0-9]+\.[0-9]{2} sec\n$`).MatchString(out) {
		t.Errorf("Gia_ManResubTest: %q", out)
	}
}

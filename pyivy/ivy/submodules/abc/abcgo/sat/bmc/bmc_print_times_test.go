package bmc

import (
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// abcTime matches Abc_PrintTime's "%s =%9.2f sec\n".
func abcTime(label string) string {
	return regexp.QuoteMeta(label) + ` = *[0-9]+\.[0-9]{2} sec\n`
}

// Gia_ManTestSatEnum prints its times with Abc_PrintTime, as bmcGen.c;
// the port printed "Time = 1.2ms. SAT solver time = 3µs."
func TestSatEnumPrintsAbcTimes(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a)
	out := captureMeshOutput(t, func() { Gia_ManTestSatEnum(p) })
	want := regexp.MustCompile(`\nFinished generating [0-9]+ assignments\.  ` + abcTime("Time") + abcTime("SAT solver time") + `$`)
	if !want.MatchString(out) {
		t.Fatalf("output:\n%q", out)
	}
}

// Gia_ManFaultDumpNewFaults reports the dump as bmcFault.c: two spaces,
// then Abc_PrintTime; the port printed ". Time = 12µs".
func TestDumpNewFaultsPrintsAbcTime(t *testing.T) {
	t.Chdir(t.TempDir())
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	prior := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	next := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	out := captureMeshOutput(t, func() { Gia_ManFaultDumpNewFaults(p, 1, prior, next, &Bmc_ParFf_t{}) })
	want := regexp.MustCompile(`Dumped [0-9]+ new multiple faults into file "newfaults\.txt"\.  ` + abcTime("Time") + `$`)
	if !want.MatchString(out) {
		t.Fatalf("output:\n%q", out)
	}
}

// Maj3_ManExactSynthesis prints bmcMaj3.c's banner, per-iteration line and
// Abc_PrintTime total; the port printed "Iteration 1: mismatch 3" and
// "Total runtime: 1.2ms".
func TestMaj3ExactSynthesisPrintsLikeABC(t *testing.T) {
	levels := vec.Vec_IntAllocArrayCopy([]int32{1, 3, 1}, 3)
	out := captureMeshOutput(t, func() { Maj3_ManExactSynthesis(5, 5, 1, levels) })
	iter := regexp.MustCompile(`(?m)^Iter +[0-9]+ : [01]{5}  Var = *[0-9]+  Cla = *[0-9]+  Conf = *[0-9]+  Lazy = *[0-9]+  ` + abcTime("Time"))
	if !strings.HasPrefix(out, "Running exact synthesis for 5-input majority with 5 MAJ3 gates...\n") || !iter.MatchString(out) ||
		!regexp.MustCompile(`\n`+abcTime("Total runtime")+`$`).MatchString(out) {
		t.Fatalf("output:\n%s", out)
	}
}

// Zyx_ManExactSynthesis ends with Abc_PrintTime's "Total runtime", as
// bmcMaj3.c; the port printed "Total runtime: 1.2ms".
func TestZyxExactSynthesisPrintsAbcTime(t *testing.T) {
	t.Chdir(t.TempDir())
	out := captureMeshOutput(t, func() { Zyx_ManExactSynthesis(&Bmc_EsPar_t{NVars: 3, NNodes: 1, NLutSize: 3, FMajority: 1}) })
	if !regexp.MustCompile(`\n` + abcTime("Total runtime") + `$`).MatchString(out) {
		t.Fatalf("output:\n%s", out)
	}
}

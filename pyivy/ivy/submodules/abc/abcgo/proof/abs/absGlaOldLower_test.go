package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldGlaLowerNativeGates(t *testing.T) {
	for _, mux := range []bool{false, true} {
		for polarity := int32(0); polarity < 16; polarity++ {
			g := gia.Gia_ManStart(16)
			a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			var out int32
			if mux {
				g.EnableMuxes()
				out = gia.Gia_ManAppendMuxReal(g, a^(polarity&1), b^((polarity>>1)&1), c^((polarity>>2)&1))
			} else {
				out = gia.Gia_ManAppendXorReal(g, a^(polarity&1), b^((polarity>>1)&1))
			}
			gia.Gia_ManAppendCo(g, out^((polarity>>3)&1))
			lowered, owner, copies := glaOldLower(g)
			if len(owner) != int(gia.Gia_ManObjNum(lowered)) || len(copies) != int(gia.Gia_ManObjNum(g)) {
				t.Fatal("incorrect lowering provenance")
			}
			for bits := 0; bits < 8; bits++ {
				in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
				want, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				got, err := lowered.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				if got[0] != want[0] {
					t.Fatalf("mux=%t polarity=%d input=%d got=%v want=%v", mux, polarity, bits, got, want)
				}
			}
			gia.Gia_ManStop(lowered)
		}
	}
}

func TestOldGlaNativeGateCounterexamplesAndClasses(t *testing.T) {
	for _, mux := range []bool{false, true} {
		for _, seeded := range []bool{false, true} {
			g := gia.Gia_ManStart(12)
			a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			var out int32
			if mux {
				g.EnableMuxes()
				out = gia.Gia_ManAppendMuxReal(g, a, b, c)
			} else {
				out = gia.Gia_ManAppendXorReal(g, a, b)
			}
			gia.Gia_ManAppendCo(g, out)
			if seeded {
				classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
				vec.Vec_IntWriteEntry(classes, 0, 1)
				vec.Vec_IntWriteEntry(classes, int(out>>1), 1)
				gia.Gia_ManSetGateClasses(g, classes)
			}
			var pars Abs_Par_t
			Abs_ParSetDefaults(&pars)
			pars.nFramesMax = 2
			if got := Gia_ManPerformGlaOld(g, &pars, 0); got != 0 {
				t.Fatalf("mux=%t seeded=%t result=%d", mux, seeded, got)
			}
			if cex := g.SequentialCounterexample(); cex == nil || gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
				t.Fatalf("mux=%t seeded=%t invalid witness", mux, seeded)
			}
		}
	}
}

func TestOldGlaNativeSequentialWitness(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		x := gia.Gia_ManAppendCi(g)
		q := gia.Gia_ManAppendCi(g)
		var bad int32
		if mux {
			g.EnableMuxes()
			// q ? x : 0
			bad = gia.Gia_ManAppendMuxReal(g, q, x, 0)
		} else {
			bad = gia.Gia_ManAppendXorReal(g, q, x)
		}
		// At frame zero the property is false; the latch becomes one.
		gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd2(g, bad, q))
		gia.Gia_ManAppendCo(g, 1)
		gia.Gia_ManSetRegNum(g, 1)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.nFramesMax = 4
		if got := Gia_ManPerformGlaOld(g, &pars, 0); got != 0 {
			t.Fatalf("mux=%t result=%d", mux, got)
		}
		cex := g.SequentialCounterexample()
		if cex == nil || cex.IFrame != 1 || gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
			t.Fatalf("mux=%t invalid sequential witness: %+v", mux, cex)
		}
	}
}

func TestOldGlaNativeClassesReconstructRealGates(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := gia.Gia_ManAppendXorReal(g, a^1, b)
	m := gia.Gia_ManAppendMuxReal(g, c, x, a^1)
	gia.Gia_ManAppendCo(g, m)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(x>>1), 1)
	vec.Vec_IntWriteEntry(classes, int(m>>1), 1)
	abs := Gia_ManDupAbsGates(g, classes)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		got, err := abs.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if got[0] != want[0] {
			t.Fatalf("input=%d got=%v want=%v", bits, got, want)
		}
	}
}

func TestOldGlaMuxAbstractionIncludesControlBoundary(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	control := gia.Gia_ManAppendAnd(g, c, d)
	g.EnableMuxes()
	m := gia.Gia_ManAppendMuxReal(g, control, a, b)
	gia.Gia_ManAppendCo(g, m)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(m>>1), 1)
	abs := Gia_ManDupAbsGates(g, classes)
	if n := len(abs.InputLiterals()); n != 3 {
		t.Fatalf("got %d abstract inputs, want data inputs and control boundary", n)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := abs.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		want := in[1]
		if in[2] {
			want = in[0]
		}
		if got[0] != want {
			t.Fatalf("inputs=%v got=%v want=%v", in, got, want)
		}
	}
}

func TestOldGlaNativeGateClassProjection(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := gia.Gia_ManStart(16)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		var x int32
		if mux {
			g.EnableMuxes()
			x = gia.Gia_ManAppendMuxReal(g, c, a, b)
		} else {
			x = gia.Gia_ManAppendXorReal(g, a, b)
		}
		buffer := gia.Gia_ManAppendBuf(g, x)
		root := gia.Gia_ManAppendAnd(g, x, buffer^1)
		gia.Gia_ManAppendCo(g, root)
		var pars Abs_Par_t
		Abs_ParSetDefaults(&pars)
		pars.nFramesMax = 1
		pars.nRatioMin = 0
		if got := Gia_ManPerformGlaOld(g, &pars, 0); got != -1 {
			t.Fatalf("mux=%t result=%d", mux, got)
		}
		classes := gia.Gia_ManGateClasses(g)
		if classes == nil || vec.Vec_IntSize(classes) != int(gia.Gia_ManObjNum(g)) {
			t.Fatalf("mux=%t missing original gate map", mux)
		}
		abs := Gia_ManDupAbsGates(g, classes)
		if n := len(abs.InputLiterals()); n > 8 {
			t.Fatalf("too many abstract inputs: %d", n)
		}
		for bits := 0; bits < 1<<len(abs.InputLiterals()); bits++ {
			in := make([]bool, len(abs.InputLiterals()))
			for i := range in {
				in[i] = bits&(1<<i) != 0
			}
			got, err := abs.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			if got[0] {
				t.Fatalf("mux=%t projected abstraction has spurious frame-zero failure on %v", mux, in)
			}
		}
	}
}

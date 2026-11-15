package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestCounterexampleProbeOwnsSourceScratch(t *testing.T) {
	p := Gia_ManStart(16)
	x, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, x, r))
	Gia_ManAppendCo(p, x)
	Gia_ManSetRegNum(p, 1)
	c := util.Abc_CexAlloc(1, 1, 37)
	c.IFrame = 36
	for i := int32(0); i < 37; i++ {
		bit := 1 + i
		c.PData[bit>>5] |= 1 << uint(bit&31)
	}
	Gia_ManCounterExampleValueStart(p, c)
	if p.pData2 == nil {
		t.Fatal("probe did not occupy source scratch storage")
	}
	for frame := int32(0); frame < 37; frame++ {
		want := int32(0)
		if frame != 0 {
			want = 1
		}
		if got := Gia_ManCounterExampleValueLookup(p, r>>1, frame); got != want {
			t.Fatalf("frame%d state %d want%d", frame, got, want)
		}
	}
	Gia_ManCounterExampleValueStop(p)
	if p.pData2 != nil {
		t.Fatal("probe scratch not released")
	}
}

func TestCounterexampleMinCexSourceDiagnostics(t *testing.T) {
	p := cexReviewThreeInputGraph(7)
	c := util.Abc_CexAlloc(1, 3, 1)
	text := hashReviewStdout(t, func() { Gia_ManMinCex(p, c) })
	if strings.Count(text, "Status -1.  Selected") != 4 || strings.Count(text, "LEXUNSAT") != 2 {
		t.Fatalf("source MinCex diagnostic passes missing: %q", text)
	}
}

func TestCounterexampleTwoRegisterLayoutMatrix(t *testing.T) {
	traces, rows := 0, 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				for init := 0; init < 4; init++ {
					for seed := 0; seed < 4; seed++ {
						p := windowReviewGraph(signs, typed, layout)
						Gia_ManSetRegNum(p, 2)
						c := util.Abc_CexAlloc(2, 2, 37)
						c.IFrame = 36
						c.PData[0] = uint32(init)
						random := uint32(seed + 1)
						for f := int32(0); f < 37; f++ {
							random = random*1664525 + 1013904223
							for i := int32(0); i < 2; i++ {
								if random>>uint(31-i)&1 != 0 {
									bit := 2 + f*2 + i
									c.PData[bit>>5] |= 1 << uint(bit&31)
								}
							}
						}
						current, _ := cexReviewSimulate(p, c)
						final := current[len(current)-1]
						expectedFirst := int32(-1)
						for po := int32(0); po < 2; po++ {
							c.IPo = po
							want := giaBool(final[Gia_ObjId(p, Gia_ManPo(p, po))])
							if Gia_ManVerifyCex(p, c, 0) != want {
								t.Fatal("final failure disagrees with scalar simulator")
							}
							if want != 0 && expectedFirst < 0 {
								expectedFirst = po
							}
						}
						c.IPo = 0
						if got := Gia_ManVerifyCex(p, c, 1); got != giaBool(final[Gia_ObjId(p, Gia_ManPo(p, 0))] != final[Gia_ObjId(p, Gia_ManPo(p, 1))]) {
							t.Fatal("dual failure")
						}
						if Gia_ManFindFailedPoCex(p, c, 0) != expectedFirst {
							t.Fatal("final failed output")
						}
						zero := util.Abc_CexDup(c, c.NRegs)
						zeroFrames, _ := cexReviewSimulate(p, zero)
						firstPo, firstFrame := int32(-1), c.IFrame
					search:
						for f, values := range zeroFrames {
							for po := int32(0); po < 2; po++ {
								if values[Gia_ObjId(p, Gia_ManPo(p, po))] {
									firstPo, firstFrame = po, int32(f)
									break search
								}
							}
						}
						expectedBits := c.NBits
						if firstPo >= 0 {
							expectedBits = c.NRegs + (firstFrame+1)*c.NPis
						}
						first := util.Abc_CexDup(c, c.NRegs)
						if Gia_ManSetFailedPoCex(p, first) != firstPo || first.IFrame != firstFrame || first.NBits != expectedBits {
							t.Fatal("first failure or trace truncation")
						}
						if expectedFirst >= 0 {
							c.IPo = expectedFirst
							all := Gia_ManCexExtendToIncludeAllObjects(p, c)
							state := Gia_ManCexExtendToIncludeCurrentStates(p, c)
							for frame, values := range current {
								for id, want := range values {
									bit := int64(frame)*int64(p.nObjs) + int64(id)
									if got := all.PData[bit>>5]>>uint(bit&31)&1 != 0; got != want {
										t.Fatal("all-object current frame")
									}
								}
								for i, id := range Vec_IntArray(p.vCis) {
									bit := frame*4 + i
									if got := state.PData[bit>>5]>>uint(bit&31)&1 != 0; got != values[id] {
										t.Fatal("logical CI/state ordering")
									}
								}
							}
							if init == 0 {
								Gia_ManCounterExampleValueStart(p, c)
								if len(p.pData2) != (37*int(p.nObjs)+31)/32 {
									t.Fatal("probe is not packed")
								}
								for frame, values := range current {
									for id, want := range values {
										if got := Gia_ManCounterExampleValueLookup(p, int32(id), int32(frame)); got != giaBool(want) {
											t.Fatal("packed probe")
										}
									}
								}
								Gia_ManCounterExampleValueStop(p)
							}
						}
						rows += len(current)
						traces++
					}
				}
			}
		}
	}
	t.Logf("%d signed/typed/layout/initial-state traces and %d independent state rows", traces, rows)
}

func cexReviewMustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected a rejected invalid operation")
		}
	}()
	fn()
}
func TestCounterexampleProbeGuardsAndReuse(t *testing.T) {
	p := cexReviewGraph(0, 0)
	c := cexReviewWitness(2, 3, 0)
	// Find a valid final property while retaining a zero initial state.
	for c.PData[0] = 0; c.PData[0] < 8; c.PData[0] += 2 {
		if Gia_ManFindFailedPoCex(p, c, 0) >= 0 {
			break
		}
	}
	c.IPo = Gia_ManFindFailedPoCex(p, c, 0)
	if c.IPo < 0 {
		t.Fatal("invalid guard fixture")
	}
	cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueStop(p) })
	cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueLookup(p, 0, 0) })
	for repeat := 0; repeat < 2; repeat++ {
		Gia_ManCounterExampleValueStart(p, c)
		cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueStart(p, c) })
		for _, pair := range [][2]int32{{-1, 0}, {p.nObjs, 0}, {0, -1}, {0, c.IFrame + 1}} {
			cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueLookup(p, pair[0], pair[1]) })
		}
		Gia_ManCounterExampleValueStop(p)
	}
	bad := *c
	bad.PData = append([]uint32(nil), c.PData...)
	bad.PData[0] |= 1
	cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueStart(p, &bad) })
	bad = *c
	bad.NBits--
	cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueStart(p, &bad) })
	bad = *c
	bad.IPo = 2
	cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueStart(p, &bad) })
	if p.pData2 != nil {
		t.Fatal("failed start occupied scratch")
	}
}

func TestCounterexampleCareConstantAndWordBoundaries(t *testing.T) {
	for _, constant := range []bool{false, true} {
		p := Gia_ManStart(16)
		x, _ := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		root := x
		if constant {
			root = 1
		}
		Gia_ManAppendCo(p, root)
		Gia_ManAppendCo(p, 0)
		Gia_ManSetRegNum(p, 1)
		c := util.Abc_CexAlloc(1, 1, 37)
		c.IFrame = 36
		c.PData[37>>5] |= 1 << uint(37&31)
		for effort := int32(0); effort < 2; effort++ {
			care := Bmc_CexCareSatBasedMinimizeAig(p, c, effort, 0)
			want := util.Abc_CexAlloc(1, 1, 37)
			if !constant {
				want.PData[37>>5] |= 1 << uint(37&31)
			}
			if !slices.Equal(care.PData, want.PData) {
				t.Fatalf("constant%v care %x want%x", constant, care.PData, want.PData)
			}
			if care.IFrame != c.IFrame || care.NBits != c.NBits || care.NRegs != 1 {
				t.Fatal("care dimensions")
			}
		}
	}
}

func TestCounterexampleProbeValidatesBeforeAllocating(t *testing.T) {
	p := cexReviewGraph(0, 0)
	c := cexReviewWitness(2, 0, 0)
	c.IFrame = 4_000_000 // Deliberately inconsistent with its small bit buffer.
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	cexReviewMustPanic(t, func() { Gia_ManCounterExampleValueStart(p, c) })
	runtime.ReadMemStats(&after)
	if after.TotalAlloc-before.TotalAlloc > 1<<20 {
		t.Fatalf("invalid witness allocated %d bytes before validation", after.TotalAlloc-before.TotalAlloc)
	}
	if p.pData2 != nil {
		t.Fatal("invalid probe occupied scratch")
	}
}

func TestCounterexampleCareForcedTrueAfterUnrolling(t *testing.T) {
	p := Gia_ManStart(32)
	Gia_ManHashStart(p)
	x, s := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, cexReviewFunction(p, 14, x, s))
	Gia_ManAppendCo(p, cexReviewFunction(p, 3, x, s))
	Gia_ManSetRegNum(p, 1)
	Gia_ManHashStop(p)
	c := cexReviewWitness(2, 0, 0)
	for effort := int32(0); effort < 2; effort++ {
		care := Bmc_CexCareSatBasedMinimizeAig(p, c, effort, 0)
		if util.Abc_CexCountOnes(care) != 0 {
			t.Fatal("forced-true output retained irrelevant input bits")
		}
	}
}

func TestCounterexamplePriorityReportsEachSourceOrder(t *testing.T) {
	p := cexReviewThreeInputGraph(7)
	c := util.Abc_CexAlloc(1, 3, 1)
	text := hashReviewStdout(t, func() { Gia_ManCexCareMinimize(p, 3, c, 4, 1, 1) })
	for _, label := range []string{"Original :    ", "PI-  PPI-:    ", "PI+  PPI-:    ", "PI-  PPI+:    ", "PI+  PPI+:    ", "Final    :    "} {
		if !strings.Contains(text, label) {
			t.Fatalf("source priority label %q missing: %q", label, text)
		}
	}
}

func BenchmarkCounterexamplePriority(b *testing.B) {
	p := cexReviewThreeInputGraph(0xe8)
	c := util.Abc_CexAlloc(1, 3, 1)
	c.PData[0] = 14
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Gia_ManCexCareMinimize(p, 3, c, 4, 0, 0)
	}
}

func cexReviewSequentialGraph(signs, typed, layout int) *Gia_Man_t {
	p := Gia_ManStart(32)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	var r0, r1 int32
	if layout&1 == 0 {
		r0, r1 = Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	}
	property := Gia_ManAppendAnd(p, a^int32(signs&1), b^int32(signs>>1&1))
	if layout&1 != 0 {
		Gia_ManAppendCo(p, property)
		r0, r1 = Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	}
	var next0, next1 int32
	if typed != 0 {
		p.EnableMuxes()
		next0 = Gia_ManAppendXorReal(p, a, r0^int32(signs>>2&1))
		next1 = Gia_ManAppendMuxReal(p, b, r1^int32(signs&1), r0)
	} else {
		next0 = Gia_ManAppendAnd(p, a, r0^int32(signs>>2&1))
		next1 = Gia_ManAppendAnd(p, r1^int32(signs&1), r0)
	}
	if layout&1 == 0 {
		Gia_ManAppendCo(p, property)
	}
	Gia_ManAppendCo(p, next0)
	Gia_ManAppendCo(p, next1)
	if layout&2 != 0 {
		for _, ports := range []*Vec_Int_t{p.vCis, p.vCos} {
			slices.Reverse(Vec_IntArray(ports))
			for i, id := range Vec_IntArray(ports) {
				Gia_ObjSetCioId(p.pObjs[id], int32(i))
			}
		}
	}
	Gia_ManSetRegNum(p, 2)
	p.SetIdentity("cex-frames", "cex-frames.aig")
	return p
}
func TestCounterexampleMinimizationFramesMatrix(t *testing.T) {
	graphs, rows := 0, 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				for _, frames := range []int32{1, 2, 3, 5} {
					p := cexReviewSequentialGraph(signs, typed, layout)
					out := Gia_ManFramesForCexMin(p, frames)
					if Gia_ManRegNum(out) != 0 || Gia_ManCiNum(out) != 2*frames || Gia_ManCoNum(out) != 3 {
						t.Fatal("unrolled interface")
					}
					if name, spec := out.Identity(); name != "cex-frames" || spec != "cex-frames.aig" {
						t.Fatal("unrolled identity")
					}
					for mask := 0; mask < 1<<uint(2*frames); mask++ {
						c := util.Abc_CexAlloc(2, 2, frames)
						c.IFrame = frames - 1
						inputs := make([]bool, 2*frames)
						for i := range inputs {
							inputs[i] = mask>>i&1 != 0
							if inputs[i] {
								bit := 2 + i
								c.PData[bit>>5] |= 1 << uint(bit&31)
							}
						}
						values, _ := cexReviewSimulate(p, c)
						final := values[len(values)-1]
						want := make([]bool, 3)
						for i := range want {
							want[i] = final[Gia_ObjId(p, Gia_ManCo(p, int32(i)))]
						}
						checkQuantificationOutputs(t, out, inputs, want)
						rows++
					}
					graphs++
				}
			}
		}
	}
	t.Logf("%d typed/signed/layout unrollings and %d exhaustive input histories match scalar transitions", graphs, rows)
}

func TestCounterexampleCareLiteralMaskBoundaries(t *testing.T) {
	c := util.Abc_CexAlloc(3, 5, 17)
	c.IFrame = 16
	c.IPo = 2
	for i := range c.PData {
		c.PData[i] = ^uint32(0)
	}
	first := int32(11)
	lits := []int32{first << 1, (first+31)<<1 | 1, (first + 32) << 1, (first+63)<<1 | 1, (first + 31) << 1}
	care := Bmc_CexCareDeriveCex(c, first, lits, int32(len(lits)))
	if care.IPo != 2 || care.IFrame != 16 || care.NRegs != 3 || care.NPis != 5 || care.NBits != c.NBits {
		t.Fatal("derived dimensions")
	}
	for i := int32(0); i < c.NBits; i++ {
		want := i == 3 || i == 34 || i == 35 || i == 66
		if got := util.Abc_InfoHasBit(care.PData, i); got != want {
			t.Fatalf("mask bit%d %v want%v", i, got, want)
		}
	}
	if !slices.Equal(c.PData, []uint32{^uint32(0), ^uint32(0), ^uint32(0)}) {
		t.Fatal("literal derivation changed witness")
	}
	if util.Abc_CexCountOnes(Bmc_CexCareDeriveCex(c, first, nil, 0)) != 0 {
		t.Fatal("empty mask")
	}
	if util.Abc_CexCountOnes(Bmc_CexCareDeriveCex(c, first, []int32{first << 1, -2}, 1)) != 1 {
		t.Fatal("selected prefix")
	}
	for _, bad := range []struct {
		lits  []int32
		count int32
	}{{lits, -1}, {lits, 6}, {[]int32{(first - 1) << 1}, 1}, {[]int32{(first + c.NBits - c.NRegs) << 1}, 1}} {
		cexReviewMustPanic(t, func() { Bmc_CexCareDeriveCex(c, first, bad.lits, bad.count) })
	}
}

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

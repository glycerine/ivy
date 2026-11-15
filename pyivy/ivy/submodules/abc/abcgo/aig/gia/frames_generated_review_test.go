package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func framesReviewGenerated(seed uint32) *Gia_Man_t {
	g := Gia_ManStart(128)
	rng := seed
	next := func() uint32 { rng = rng*1664525 + 1013904223; return rng >> 8 }
	pool := []int32{0}
	for i := 0; i < 4; i++ {
		pool = append(pool, Gia_ManAppendCi(g))
	}
	for i := 0; i < 12; i++ {
		a, b := int(next())%len(pool), int(next())%len(pool)
		if a == b {
			b = (b + 1) % len(pool)
		}
		x, y := pool[a]^int32(next()&1), pool[b]^int32(next()&1)
		pool = append(pool, Gia_ManAppendAnd(g, x, y))
	}
	for i := 0; i < 4; i++ {
		index := len(pool) - 1 - i
		if i >= 2 {
			index = int(next()) % len(pool)
		}
		Gia_ManAppendCo(g, pool[index]^int32(next()&1))
	}
	Gia_ManSetRegNum(g, 2)
	return g
}

func framesReviewDump(out *strings.Builder, g *Gia_Man_t, tag string) {
	fmt.Fprintf(out, "%s ci%d co%d regs%d objs%d\n", tag, Gia_ManCiNum(g), Gia_ManCoNum(g), g.nRegs, g.nObjs)
	for id, o := range g.pObjs {
		fmt.Fprintf(out, "%d %d %d %d %d %d %d\n", id, o.iDiff0, o.iDiff1, o.fCompl0, o.fCompl1, o.fTerm, o.Value)
	}
}

// Simulate logical PIs/state independently for each complete frame history.
func framesReviewCheckSequence(t *testing.T, src, out *Gia_Man_t, p Gia_ParFra_t) int {
	t.Helper()
	nPi, nPo, nReg := int(Gia_ManPiNum(src)), int(Gia_ManPoNum(src)), int(src.nRegs)
	nInput, nOutput, regs := nPi*int(p.NFrames), nPo*int(p.NFrames), 0
	if p.FInit == 0 {
		nInput += nReg
		nOutput += nReg
		regs = nReg
		if p.FOrPos != 0 {
			nOutput = nPo + nReg
		}
	}
	if out == nil || int(Gia_ManCiNum(out)) != nInput || int(Gia_ManCoNum(out)) != nOutput || int(out.nRegs) != regs {
		t.Fatalf("frames%d init%d disable%d or%d: interface must be ci%d/co%d/regs%d", p.NFrames, p.FInit, p.FDisableSt, p.FOrPos, nInput, nOutput, regs)
	}
	for assignment := 0; assignment < 1<<nInput; assignment++ {
		state := make([]bool, nReg)
		if p.FInit == 0 {
			for i := range state {
				state[i] = assignment>>(nPi*int(p.NFrames)+i)&1 != 0
			}
		}
		var expected []bool
		merged := make([]bool, nPo)
		for f := 0; f < int(p.NFrames); f++ {
			input := assignment >> (f * nPi) & ((1 << nPi) - 1)
			for i, bit := range state {
				if bit {
					input |= 1 << (nPi + i)
				}
			}
			values := cofReviewValues(src, input)
			for po := 0; po < nPo; po++ {
				v := values[Gia_ManPo(src, int32(po)).id]
				merged[po] = merged[po] || v
				if p.FInit != 0 || p.FOrPos == 0 {
					expected = append(expected, v)
				}
			}
			for i := range state {
				state[i] = values[Gia_ManRi(src, int32(i)).id]
			}
		}
		if p.FInit == 0 {
			if p.FOrPos != 0 {
				expected = append(expected, merged...)
			}
			expected = append(expected, state...)
		}
		actual := cofReviewValues(out, assignment)
		for co, want := range expected {
			if got := actual[Gia_ManCo(out, int32(co)).id]; got != want {
				t.Fatalf("frames%d init%d disable%d or%d assignment%d output%d got%t want%t", p.NFrames, p.FInit, p.FDisableSt, p.FOrPos, assignment, co, got, want)
			}
		}
	}
	return 1 << nInput
}

func TestFramesGeneratedNativeDrivers(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := exec.Command(exe, "--generated-drivers").CombinedOutput()
	if err != nil {
		t.Fatalf("C generated drivers: %v\n%s", err, native)
	}
	var got strings.Builder
	cases, rows := 0, 0
	for seed := uint32(0); seed < 32; seed++ {
		for frames := int32(1); frames <= 3; frames++ {
			for init := int32(0); init < 2; init++ {
				for disable := int32(0); disable < 2; disable++ {
					for or := int32(0); or < 2; or++ {
						g := framesReviewGenerated(seed)
						p := Gia_ParFra_t{NFrames: frames, FInit: init, FDisableSt: disable, FOrPos: or}
						out := Gia_ManFrames(g, &p)
						rows += framesReviewCheckSequence(t, g, out, p)
						fmt.Fprintf(&got, "generated %d f%d init%d disable%d or%d\n", seed, frames, init, disable, or)
						framesReviewDump(&got, out, "result")
						Gia_ManStop(out)
						Gia_ManStop(g)
						cases++
					}
				}
			}
		}
	}
	esopReviewCompare(t, native, got.String())
	t.Logf("%d corrected-source configurations and %d independent sequential assignments", cases, rows)
}

func TestFramesGeneratedNativeHistory(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := exec.Command(exe, "--histories").CombinedOutput()
	if err != nil {
		t.Fatalf("C generated history: %v\n%s", err, native)
	}
	var got strings.Builder
	vector := func(tag string, values []int32) {
		got.WriteString(tag)
		for _, v := range values {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
	}
	snapshot := func(p *Gia_ManUnr_t) {
		vector("limits", p.limits)
		vector("delays", p.delays)
		got.WriteString("first")
		for _, v := range p.first {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
		vector("history", p.store)
	}
	for seed := uint32(0); seed < 32; seed++ {
		for save := int32(0); save < 2; save++ {
			g := framesReviewGenerated(seed)
			p := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 3, fSaveLastLit: save}).(*Gia_ManUnr_t)
			fmt.Fprintf(&got, "history %d save%d step0\n", seed, save)
			snapshot(p)
			framesReviewDump(&got, p.pOrder, "ordered")
			for step := int32(1); step <= 3; step++ {
				q := Gia_ManUnrollAdd(p, step).(*Gia_Man_t)
				fmt.Fprintf(&got, "history %d save%d step%d lit%d\n", seed, save, step, p.LastLit)
				snapshot(p)
				framesReviewDump(&got, q, "step")
			}
			q := p.pNew
			Gia_ManUnrollStop(p)
			Gia_ManStop(q)
			Gia_ManStop(g)
		}
	}
	esopReviewCompare(t, native, got.String())
}

func TestFramesMixedDriversSequentialTruth(t *testing.T) {
	cases, rows := 0, 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				t.Run(fmt.Sprintf("phase%d/typed%d/layout%d", signs, typed, layout), func(t *testing.T) {
					for frames := int32(1); frames <= 3; frames++ {
						for init := int32(0); init < 2; init++ {
							for disable := int32(0); disable < 2; disable++ {
								for or := int32(0); or < 2; or++ {
									g := windowReviewGraph(signs, typed, layout)
									Gia_ManSetRegNum(g, 2)
									g.pName, g.PSpec = "frame-owner", "frame-owner.aig"
									p := Gia_ParFra_t{NFrames: frames, FInit: init, FDisableSt: disable, FOrPos: or}
									beforeParams := p
									q := Gia_ManFrames(g, &p)
									rows += framesReviewCheckSequence(t, g, q, p)
									if p != beforeParams || q.pName != g.pName || q.PSpec != g.PSpec {
										t.Fatal("changed parameters or lost graph identity")
									}
									before := cofReviewValues(q, 0)
									Gia_ManStop(g)
									Gia_ManLevelNum(q)
									if !slices.Equal(before, cofReviewValues(q, 0)) {
										t.Fatal("returned frame graph depended on disposed source")
									}
									Gia_ManStop(q)
									cases++
								}
							}
						}
					}
				})
			}
		}
	}
	t.Logf("%d mixed-gate/layout/flag configurations and %d independent sequential assignments", cases, rows)
}

package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestFxTypedInterleavedSequentialInterfaces(t *testing.T) {
	for kind := 0; kind < 4; kind++ {
		for signs := int32(0); signs < 16; signs++ {
			for reorder := 0; reorder < 2; reorder++ {
				for reverse := int32(0); reverse < 2; reverse++ {
					for _, limit := range []int32{0, 4} {
						p := Gia_ManStart(64)
						p.EnableMuxes()
						a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
						var x int32
						switch kind {
						case 0:
							x = Gia_ManAppendAnd(p, a^(signs&1), b)
						case 1, 2:
							x = Gia_ManAppendXorReal(p, a^(signs&1), b)
						case 3:
							x = Gia_ManAppendBuf(p, a^(signs&1))
						}
						c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
						var root int32
						switch kind {
						case 0:
							root = Gia_ManAppendAnd(p, x, c^(signs>>1&1))
						case 1:
							root = Gia_ManAppendXorReal(p, x, c^(signs>>1&1))
						case 2:
							root = Gia_ManAppendMuxReal(p, d, x, c^(signs>>1&1))
						case 3:
							root = Gia_ManAppendBuf(p, x^(signs>>1&1))
						}
						Gia_ManAppendCo(p, root^(signs>>2&1))
						side := Gia_ManAppendAnd(p, a, c)
						Gia_ManAppendCo(p, side^1)
						Gia_ManAppendCo(p, x^(signs>>3&1))
						Gia_ManSetRegNum(p, 1)
						if reorder != 0 {
							ids := Vec_IntArray(p.vCis)
							ids[0], ids[3] = ids[3], ids[0]
							Gia_ObjSetCioId(Gia_ManCi(p, 0), 0)
							Gia_ObjSetCioId(Gia_ManCi(p, 3), 3)
						}
						p.VMapping = Vec_IntStart(int(p.nObjs))
						xCut := []int32{a >> 1, b >> 1}
						if kind == 3 {
							xCut = xCut[:1]
						}
						for _, lut := range []struct {
							root int32
							cut  []int32
						}{{x >> 1, xCut}, {root >> 1, []int32{a >> 1, b >> 1, c >> 1, d >> 1}}, {side >> 1, []int32{a >> 1, c >> 1}}} {
							Vec_IntWriteEntry(p.VMapping, int(lut.root), int32(Vec_IntSize(p.VMapping)))
							Vec_IntPush(p.VMapping, int32(len(lut.cut)))
							for _, id := range lut.cut {
								Vec_IntPush(p.VMapping, id)
							}
							Vec_IntPush(p.VMapping, lut.root)
						}
						if err := p.SetAigerProperties(1, 1); err != nil {
							t.Fatal(err)
						}
						p.aigerLatchInits = []int32{-1}
						p.aigerComment = []byte("typed extraction")
						p.VNamesIn = Vec_PtrAlloc(4)
						p.VNamesOut = Vec_PtrAlloc(3)
						for i := 0; i < 4; i++ {
							Vec_PtrPush(p.VNamesIn, fmt.Sprintf("ci%d", i))
						}
						for i := 0; i < 3; i++ {
							Vec_PtrPush(p.VNamesOut, fmt.Sprintf("co%d", i))
						}
						inputNames, outputNames := p.InputNames(), p.OutputNames()
						arrivals := Vec_IntStart(4)
						required := Vec_IntStart(3)
						p.vCiArrs, p.vCoReqs = arrivals, required
						q := Gia_ManPerformFx(p, limit, 0, reverse, 0, 0)
						cofReviewCheckRebuild(t, p, q)
						if q.nConstrs != 1 || q.aigerBadOutputs != 1 || !slices.Equal(q.aigerLatchInits, []int32{-1}) || string(q.aigerComment) != "typed extraction" || !slices.Equal(q.InputNames(), inputNames) || !slices.Equal(q.OutputNames(), outputNames) {
							t.Fatalf("kind%d signs%d reorder%d reverse%d limit%d: lost interface", kind, signs, reorder, reverse, limit)
						}
						if q.vCiArrs != arrivals || q.vCoReqs != required || p.vCiArrs != nil || p.vCoReqs != nil || p.VNamesIn != nil || p.VNamesOut != nil {
							t.Fatal("timing/name ownership did not move")
						}
						Gia_ManStop(p)
						if q.aigerLatchInits[0] != -1 || q.InputNames()[0] != "ci0" {
							t.Fatal("stopping source invalidated result")
						}
						Gia_ManStop(q)
					}
				}
			}
		}
	}
}

func TestFxNoAndDriverNeedsNoMapping(t *testing.T) {
	for _, inputs := range []int{0, 4} {
		for _, registers := range []int32{0, 1} {
			if inputs == 0 && registers != 0 {
				continue
			}
			p := Gia_ManStart(32)
			vars := make([]int32, inputs)
			for i := range vars {
				vars[i] = Gia_ManAppendCi(p)
			}
			Gia_ManAppendCo(p, 0)
			Gia_ManAppendCo(p, 1)
			last := int32(0)
			if inputs != 0 {
				last = vars[inputs-1] ^ 1
			}
			Gia_ManAppendCo(p, last)
			Gia_ManSetRegNum(p, registers)
			if err := p.SetAigerProperties(1, 1); err != nil {
				t.Fatal(err)
			}
			if registers != 0 {
				p.aigerLatchInits = []int32{-1}
			}
			p.aigerComment = []byte("no AND mapping")
			p.VNamesIn = Vec_PtrAlloc(inputs)
			p.VNamesOut = Vec_PtrAlloc(3)
			for i := 0; i < inputs; i++ {
				Vec_PtrPush(p.VNamesIn, fmt.Sprintf("in%d", i))
			}
			for i := 0; i < 3; i++ {
				Vec_PtrPush(p.VNamesOut, fmt.Sprintf("out%d", i))
			}
			inNames, outNames := p.InputNames(), p.OutputNames()
			q := Gia_ManPerformFx(p, 4, 0, 0, 0, 0)
			cofReviewCheckRebuild(t, p, q)
			if q.nConstrs != 1 || q.aigerBadOutputs != 1 || !slices.Equal(q.aigerLatchInits, p.aigerLatchInits) || !slices.Equal(q.InputNames(), inNames) || !slices.Equal(q.OutputNames(), outNames) || string(q.aigerComment) != "no AND mapping" {
				t.Fatal("no-AND interface lost")
			}
			if p.VNamesIn != nil || p.VNamesOut != nil {
				t.Fatal("names were not transferred")
			}
			Gia_ManStop(p)
			Gia_ManStop(q)
		}
	}
}

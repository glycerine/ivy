package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func fxReviewMappedGraph(mode, signs int32, reorder bool) *Gia_Man_t {
	g := Gia_ManStart(64)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a^(signs&1), b)
	y := Gia_ManAppendAnd(g, x, c^(signs>>1&1))
	z := Gia_ManAppendAnd(g, x, d^(signs>>2&1))
	w := Gia_ManAppendAnd(g, y^(signs>>3&1), z)
	roots := []int32{x >> 1, y >> 1, z >> 1, w >> 1}
	cuts := [][]int32{{a >> 1, b >> 1}, {x >> 1, c >> 1}, {x >> 1, d >> 1}, {y >> 1, z >> 1}}
	if mode == 1 {
		cuts[1] = []int32{a >> 1, b >> 1, c >> 1}
		cuts[2] = []int32{a >> 1, b >> 1, d >> 1}
	}
	if mode == 2 {
		cuts[3] = []int32{a >> 1, b >> 1, c >> 1, d >> 1}
	}
	Gia_ManAppendCo(g, w^(signs&1))
	if mode < 2 {
		Gia_ManAppendCo(g, y)
		Gia_ManAppendCo(g, z^1)
	}
	if reorder {
		ids := Vec_IntArray(g.vCis)
		ids[0], ids[3] = ids[3], ids[0]
		Gia_ObjSetCioId(Gia_ManCi(g, 0), 0)
		Gia_ObjSetCioId(Gia_ManCi(g, 3), 3)
	}
	g.VMapping = Vec_IntStart(int(g.nObjs))
	first := int(mode)
	if mode == 2 {
		first = 3
	}
	for i := first; i < 4; i++ {
		Vec_IntWriteEntry(g.VMapping, int(roots[i]), int32(Vec_IntSize(g.VMapping)))
		Vec_IntPush(g.VMapping, int32(len(cuts[i])))
		for j := len(cuts[i]) - 1; j >= 0; j-- {
			Vec_IntPush(g.VMapping, cuts[i][j])
		}
		Vec_IntPush(g.VMapping, roots[i])
	}
	return g
}

func TestFxNativeMappedRetrievalAndDriver(t *testing.T) {
	exe := compileGiaSourceOracle(t, "fx_mapped_review.c")
	raw, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("original C mapped extraction: %v\n%s", err, raw)
	}
	// The unchanged branch calls original Gia_ManDup, which appends CIs in
	// physical order and does not restore the logical port vectors.
	if !strings.Contains(string(raw), "driver m0 s0 r0 o1 limit0 ci4 co3 and4 8000 8080 77ff") {
		t.Fatal("original C port-order fault changed")
	}
	native, err := exec.Command(exe, "--correct-ports").CombinedOutput()
	if err != nil {
		t.Fatalf("corrected C mapped extraction: %v\n%s", err, native)
	}
	var got strings.Builder
	for mode := int32(0); mode < 3; mode++ {
		for signs := int32(0); signs < 16; signs++ {
			for reverse := int32(0); reverse < 2; reverse++ {
				for reorder := 0; reorder < 2; reorder++ {
					p := fxReviewMappedGraph(mode, signs, reorder != 0)
					var phases *Vec_Str_t
					cubes := Gia_ManFxRetrieve(p, &phases, reverse)
					fmt.Fprintf(&got, "retrieve m%d s%d r%d o%d phases", mode, signs, reverse, reorder)
					for i := 0; i < Vec_StrSize(phases); i++ {
						fmt.Fprint(&got, Vec_StrEntry(phases, i))
					}
					for _, obj := range p.pObjs {
						fmt.Fprintf(&got, " v%v", obj.Value)
					}
					for i := 0; i < Vec_WecSize(cubes); i++ {
						got.WriteString(" [")
						for j, lit := range Vec_IntArray(Vec_WecEntry(cubes, i)) {
							if j > 0 {
								got.WriteByte(',')
							}
							fmt.Fprint(&got, lit)
						}
						got.WriteByte(']')
					}
					got.WriteByte('\n')
					// Fanin order is an intentional mutation of the original mapped network.
					for id := int32(1); id < p.nObjs; id++ {
						if Gia_ObjIsLut(p, id) == 0 {
							continue
						}
						fanins := Gia_ObjLutFanins(p, id)
						for i := 1; i < len(fanins); i++ {
							if reverse != 0 && fanins[i-1] >= fanins[i] || reverse == 0 && fanins[i-1] <= fanins[i] {
								t.Fatal("LUT fanins not sorted in requested direction")
							}
						}
					}
					Vec_WecFree(cubes)
					Vec_StrFree(phases)
					Gia_ManStop(p)
					for limit := int32(0); limit <= 4; limit += 4 {
						p = fxReviewMappedGraph(mode, signs, reorder != 0)
						q := Gia_ManPerformFx(p, limit, 0, reverse, 0, 0)
						cofReviewCheckRebuild(t, p, q)
						fmt.Fprintf(&got, "driver m%d s%d r%d o%d limit%d ci%d co%d and%d", mode, signs, reverse, reorder, limit, Gia_ManCiNum(q), Gia_ManCoNum(q), Gia_ManAndNum(q))
						masks := make([]uint32, Gia_ManCoNum(q))
						for input := 0; input < 16; input++ {
							values := cofReviewValues(q, input)
							for i := int32(0); i < Gia_ManCoNum(q); i++ {
								obj := Gia_ManCo(q, i)
								if values[Gia_ObjFaninId0p(q, obj)] != (obj.fCompl0 != 0) {
									masks[i] |= 1 << input
								}
							}
						}
						for _, mask := range masks {
							fmt.Fprintf(&got, " %04x", mask)
						}
						got.WriteByte('\n')
						Gia_ManStop(q)
						Gia_ManStop(p)
					}
				}
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}

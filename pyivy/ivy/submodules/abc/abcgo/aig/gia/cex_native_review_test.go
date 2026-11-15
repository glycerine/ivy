package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"os/exec"
	"strings"
	"testing"
)

func cexReviewFunction(p *Gia_Man_t, table int, x, s int32) int32 {
	root := int32(0)
	for mask := 0; mask < 4; mask++ {
		if table>>mask&1 != 0 {
			cube := Gia_ManHashAnd(p, x^int32(1^(mask&1)), s^int32(1^(mask>>1&1)))
			root = Gia_ManHashOr(p, root, cube)
		}
	}
	return root
}
func cexReviewGraph(transition, property int) *Gia_Man_t {
	left, right := []int{8, 2, 14, 0}, []int{6, 8, 4, 0}
	p := Gia_ManStart(32)
	Gia_ManHashStart(p)
	x, s := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for _, table := range []int{left[property], right[property], transition} {
		Gia_ManAppendCo(p, cexReviewFunction(p, table, x, s))
	}
	Gia_ManSetRegNum(p, 1)
	Gia_ManHashStop(p)
	return Gia_ManCleanup(p)
}
func cexReviewWitness(frames, seed, init int) *util.Abc_Cex_t {
	c := util.Abc_CexAlloc(1, 1, int32(frames))
	c.IFrame = int32(frames - 1)
	c.PData[0] = uint32(init)
	random := uint32(seed + 1)
	for f := 0; f < frames; f++ {
		random = random*1664525 + 1013904223
		if random>>31 != 0 {
			bit := 1 + f
			c.PData[bit>>5] |= 1 << uint(bit&31)
		}
	}
	return c
}
func cexReviewWords(out *strings.Builder, words []uint32) {
	for _, v := range words {
		fmt.Fprintf(out, "%08x", v)
	}
}

// The original all-object extension snapshots before evaluation. Keep that
// stale snapshot as explicit oracle evidence, and check Go's current-frame
// extension separately against this scalar simulator.
func cexReviewSimulate(p *Gia_Man_t, c *util.Abc_Cex_t) (current, legacy [][]bool) {
	old := make([]bool, p.nObjs)
	state := make([]bool, p.nRegs)
	bit := func(index int32) bool { return c.PData[index>>5]>>uint(index&31)&1 != 0 }
	for i := range state {
		state[i] = bit(int32(i))
		old[Gia_ObjId(p, Gia_ManCo(p, Gia_ManPoNum(p)+int32(i)))] = state[i]
	}
	for frame := int32(0); frame <= c.IFrame; frame++ {
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			old[Gia_ObjId(p, Gia_ManCi(p, i))] = bit(c.NRegs + frame*c.NPis + i)
		}
		for i, v := range state {
			old[Gia_ObjId(p, Gia_ManCi(p, c.NPis+int32(i)))] = v
		}
		legacy = append(legacy, append([]bool(nil), old...))
		now := append([]bool(nil), old...)
		read := func(lit int32) bool { return now[lit>>1] != (lit&1 != 0) }
		for id := int32(1); id < p.nObjs; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 1:
			case 2:
				now[id] = read(a) && read(b)
			case 3:
				now[id] = read(a) != read(b)
			case 4:
				if read(a) {
					now[id] = read(b)
				} else {
					now[id] = read(c)
				}
			case 5, 6:
				now[id] = read(a)
			default:
				panic("invalid scalar CEX gate")
			}
		}
		current = append(current, now)
		for i := range state {
			state[i] = now[Gia_ObjId(p, Gia_ManCo(p, Gia_ManPoNum(p)+int32(i)))]
		}
		old = append([]bool(nil), now...)
	}
	return
}
func cexReviewPack(rows [][]bool) []uint32 {
	if len(rows) == 0 {
		return nil
	}
	width := len(rows[0])
	words := make([]uint32, (len(rows)*width+31)/32)
	for f, row := range rows {
		for i, v := range row {
			if v {
				bit := f*width + i
				words[bit>>5] |= 1 << uint(bit&31)
			}
		}
	}
	return words
}
func TestCounterexampleReplayAndValuesMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cex_replay.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native CEX: %v\n%s", err, native)
	}
	var got strings.Builder
	cases := 0
	for transition := 0; transition < 16; transition++ {
		for property := 0; property < 4; property++ {
			for _, frames := range []int{1, 2, 5, 37} {
				for seed := 0; seed < 4; seed++ {
					for init := 0; init < 2; init++ {
						p := cexReviewGraph(transition, property)
						c := cexReviewWitness(frames, seed, init)
						current, legacy := cexReviewSimulate(p, c)
						c.IPo = 0
						v0, dual := Gia_ManVerifyCex(p, c, 0), Gia_ManVerifyCex(p, c, 1)
						c.IPo = 1
						v1 := Gia_ManVerifyCex(p, c, 0)
						found := Gia_ManFindFailedPoCex(p, c, 0)
						fmt.Fprintf(&got, "case %d %d %d %d %d objects%d verify%d:%d:%d find%d:%d:%d", transition, property, frames, seed, init, p.nObjs, v0, v1, dual, found, Gia_ManFindFailedPoCex(p, c, 1), Gia_ManFindFailedPoCex(p, c, 2))
						first := util.Abc_CexDup(c, c.NRegs)
						Gia_ManSetFailedPoCex(p, first)
						fmt.Fprintf(&got, " first%d:%d:%d", first.IPo, first.IFrame, first.NBits)
						if found >= 0 {
							c.IPo = found
							ext := Gia_ManCexExtendToIncludeCurrentStates(p, c)
							got.WriteString(" states")
							cexReviewWords(&got, ext.PData)
							ext = Gia_ManCexExtendToIncludeAllObjects(p, c)
							got.WriteString(" legacy-all")
							cexReviewWords(&got, cexReviewPack(legacy))
							expected := cexReviewPack(current)
							if len(ext.PData) != len(expected) {
								t.Fatal("current-frame extension width")
							}
							for i, w := range expected {
								if ext.PData[i] != w {
									t.Fatalf("current-frame extension word%d: %x want%x", i, ext.PData[i], w)
								}
							}
							if init == 0 {
								Gia_ManCounterExampleValueStart(p, c)
								got.WriteString(" probe")
								cexReviewWords(&got, p.pData2)
								for frame := int32(0); frame < c.IFrame+1; frame++ {
									for id := int32(0); id < p.nObjs; id++ {
										bit := int64(frame)*int64(p.nObjs) + int64(id)
										want := int32(ext.PData[bit>>5] >> uint(bit&31) & 1)
										if actual := Gia_ManCounterExampleValueLookup(p, id, frame); actual != want {
											t.Fatalf("probe disagrees with all-object extension: %d want%d", actual, want)
										}
									}
								}
								Gia_ManCounterExampleValueStop(p)
							}
						}
						got.WriteByte('\n')
						cases++
					}
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("CEX source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("native output length Go%d C%d", got.Len(), len(native))
	}
	t.Logf("%d native replay, first/final failure, dual-output, state/object-extension and packed-probe cases", cases)
}

package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
)

func mappingReviewBoxGraph(seed uint32, fault bool) (p *Gia_Man_t, lut, choice []int32) {
	state := seed
	next := func() uint32 { state = state*1664525 + 1013904223; return state }
	pis, boxes := 2+int(seed%3), 2+int(seed%3)
	ins, outs := make([]int, boxes), make([]int, boxes)
	totalIns, totalOuts := 0, 0
	for i := 0; i < boxes; i++ {
		ins[i] = 1 + (int(seed)+i)%3
		outs[i] = 1 + (int(seed)+2*i)%2
		totalIns += ins[i]
		totalOuts += outs[i]
	}
	p = Gia_ManStart(256)
	ci := make([]int32, pis+totalOuts)
	for i := range ci {
		ci[i] = Gia_ManAppendCi(p)
	}
	available := append([]int32(nil), ci[:pis]...)
	ciOffset := pis
	expectedLut, expectedChoice := map[int32]int32{}, map[int32]int32{}
	var roots []int32
	appendAnd := func(a, b int32) int32 {
		root := Gia_ManAppendAnd(p, a, b)
		roots = append(roots, root>>1)
		expectedLut[root>>1] = max(expectedLut[a>>1], expectedLut[b>>1]) + 1
		expectedChoice[root>>1] = max(expectedChoice[a>>1], expectedChoice[b>>1]) + 1
		return root
	}
	appendCo := func(root int32) int32 {
		co := Gia_ManAppendCo(p, root)
		expectedLut[co>>1] = expectedLut[root>>1]
		expectedChoice[co>>1] = expectedChoice[root>>1]
		return co >> 1
	}
	for box := 0; box < boxes; box++ {
		boxInputs := make([]int32, ins[box])
		for j := 0; j < ins[box]; j++ {
			root := available[int(next())%len(available)]
			if fault && box == 0 && j == 0 {
				root = ci[pis+outs[0]]
			}
			depth := 1 + (int(seed)+box+j)%3
			for k := 0; k < depth; k++ {
				b := available[int(next())%len(available)]
				if b>>1 == root>>1 {
					if b>>1 == available[0]>>1 {
						b = available[1]
					} else {
						b = available[0]
					}
				}
				pa, pb := int32(next()&1), int32(next()&1)
				root = appendAnd(root^pa, b^pb)
			}
			boxInputs[j] = appendCo(root)
		}
		for output := 0; output < outs[box]; output++ {
			id := ci[ciOffset] >> 1
			var l, c int32
			for j, input := range boxInputs {
				if (int(seed)+box+output*ins[box]+j)%4 != 0 {
					l = max(l, expectedLut[input]+1)
				}
				c = max(c, expectedChoice[input]+1)
			}
			expectedLut[id], expectedChoice[id] = l, c
			available = append(available, ci[ciOffset])
			ciOffset++
		}
	}
	for j := 0; j < 2; j++ {
		root := ci[pis+totalOuts-1-j%totalOuts]
		for k := 0; k < 2+j; k++ {
			root = appendAnd(root, ci[k%pis])
		}
		appendCo(root ^ int32(j&1))
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, root := range roots {
		o := p.pObjs[root]
		Vec_IntWriteEntry(p.VMapping, int(root), int32(Vec_IntSize(p.VMapping)))
		for _, v := range []int32{2, Gia_ObjFaninId0(o, root), Gia_ObjFaninId1(o, root), root} {
			Vec_IntPush(p.VMapping, v)
		}
	}
	tm := tim.Tim_ManStart(pis+totalOuts, totalIns+2)
	tables := make([][]float32, boxes)
	ciOffset = pis
	coOffset := 0
	for box := 0; box < boxes; box++ {
		table := make([]float32, 3+ins[box]*outs[box])
		table[0], table[1], table[2] = float32(box), float32(ins[box]), float32(outs[box])
		for j := 0; j < ins[box]*outs[box]; j++ {
			table[3+j] = float32((int(seed)+j)%7) / 2
			if (int(seed)+box+j)%4 == 0 {
				table[3+j] = -float32(giaInfinity)
			}
		}
		tables[box] = table
		tim.Tim_ManCreateBox(tm, coOffset, ins[box], ciOffset, outs[box], box, 0)
		ciOffset += outs[box]
		coOffset += ins[box]
	}
	tim.Tim_ManSetDelayTables(tm, tables)
	p.PManTime = tm
	lut, choice = make([]int32, p.nObjs), make([]int32, p.nObjs)
	for id, value := range expectedLut {
		lut[id] = value
	}
	for id, value := range expectedChoice {
		choice[id] = value
	}
	return p, lut, choice
}
func mappingReviewStoredLevels(p *Gia_Man_t) []int32 {
	levels := make([]int32, p.nObjs)
	for i := range levels {
		levels[i] = Gia_ObjLevelId(p, int32(i))
	}
	return levels
}
func TestMappingNativeBoxedMetricsAndChoiceLevels(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mapping_box_levels_review.c")
	cases := 0
	for seed := 1; seed <= 64; seed++ {
		for fault := 0; fault < 2; fault++ {
			source, err := testoracle.Command(t, exe, strconv.Itoa(seed), strconv.Itoa(fault)).CombinedOutput()
			if err != nil {
				t.Fatalf("native seed%d fault%d: %v %s", seed, fault, err, source)
			}
			p, lutExpected, choiceExpected := mappingReviewBoxGraph(uint32(seed), fault != 0)
			got := hashReviewStdout(t, func() {
				for repeat := 0; repeat < 2; repeat++ {
					var luts, edges, depth int32
					Gia_ManLutParams(p, &luts, &edges, &depth)
					fmt.Printf("params %d %d %d\n", luts, edges, depth)
					mappingReviewArray(os.Stdout, "lut-levels", mappingReviewStoredLevels(p))
					if fault == 0 {
						want := int32(0)
						for i := Gia_ManCoNum(p) - 2; i < Gia_ManCoNum(p); i++ {
							want = max(want, lutExpected[Gia_ManCo(p, i).id])
						}
						if depth != want {
							t.Fatalf("seed%d LUT depth%d want%d", seed, depth, want)
						}
						for id, want := range lutExpected {
							if Gia_ObjLevelId(p, int32(id)) != want {
								t.Fatalf("seed%d object%d LUT level%d want%d", seed, id, Gia_ObjLevelId(p, int32(id)), want)
							}
						}
						depth = Gia_ManChoiceLevel(p)
						fmt.Printf("choice %d\n", depth)
						mappingReviewArray(os.Stdout, "choice-levels", mappingReviewStoredLevels(p))
						want = 0
						for _, o := range p.pObjs {
							if Gia_ObjIsTerm(o) != 0 {
								want = max(want, choiceExpected[o.id])
							}
						}
						if depth != want {
							t.Fatalf("seed%d choice depth%d want%d", seed, depth, want)
						}
						for id, want := range choiceExpected {
							if Gia_ObjLevelId(p, int32(id)) != want {
								t.Fatalf("seed%d object%d choice level%d want%d", seed, id, Gia_ObjLevelId(p, int32(id)), want)
							}
						}
					}
				}
			})
			if string(source) != got {
				t.Logf("seed%d fault%d", seed, fault)
				esopReviewCompare(t, source, got)
			}
			if fault != 0 && !strings.Contains(got, "Boxes are not in a topological order") {
				t.Fatal("fault case did not exercise fallback")
			}
			Gia_ManStop(p)
			cases++
		}
	}
	t.Logf("compared %d repeated boxed/fallback states with independent valid-network expectations", cases)
}

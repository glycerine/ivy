package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func cutEnumReviewGraph(seed int) *Gia_Man_t {
	g := Gia_ManStart(128)
	Gia_ManHashStart(g)
	var lits [80]int32
	count := 40
	if seed >= 16 {
		count = 80
	}
	for i := 0; i < 8; i++ {
		lits[i] = Gia_ManAppendCi(g)
	}
	for i := 8; i < 12; i++ {
		lits[i] = Gia_ManHashAnd(g, lits[2*(i-8)], lits[2*(i-8)+1])
	}
	lits[12] = Gia_ManHashAnd(g, lits[8], lits[9])
	lits[13] = Gia_ManHashAnd(g, lits[10], lits[11])
	lits[14] = Gia_ManHashAnd(g, lits[12], lits[13])
	random := uint32(seed + 307)
	for i := 15; i < count; i++ {
		random = random*1664525 + 1013904223
		a := lits[int(random)%i]
		random = random*1664525 + 1013904223
		a ^= int32(random >> 31)
		random = random*1664525 + 1013904223
		b := lits[int(random)%i]
		random = random*1664525 + 1013904223
		b ^= int32(random >> 31)
		if i%3 == 0 {
			if seed >= 8 && seed < 16 && a>>1 != b>>1 {
				lits[i] = Gia_ManAppendXorReal(g, a, b)
			} else {
				lits[i] = Gia_ManHashXor(g, a, b)
			}
		} else {
			lits[i] = Gia_ManHashAnd(g, a, b)
		}
	}
	Gia_ManAppendCo(g, lits[14])
	Gia_ManAppendCo(g, lits[count-1])
	Gia_ManAppendCo(g, lits[count-3]^1)
	Gia_ManHashStop(g)
	return g
}

func cutEnumReviewEnumerate(g *Gia_Man_t, size, num, truth, minimum int32) *Gia_Sto_t {
	p := Gia_StoAlloc(g, size, num, truth, minimum, 0)
	for id := int32(0); id < g.nObjs; id++ {
		Gia_StoRefObj(p, id)
	}
	Gia_StoComputeCutsConst0(p, 0)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		Gia_StoComputeCutsCi(p, Gia_ManCi(g, i).id)
	}
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsAnd(g.pObjs[id]) != 0 {
			Gia_StoComputeCutsNode(p, id)
		}
	}
	return p
}

// cutEnumReviewValidate checks cut semantics without using the cut merging,
// expansion, minimization or dominance routines under review.
func cutEnumReviewValidate(t *testing.T, p *Gia_Sto_t) int {
	t.Helper()
	g := p.pGia
	var assignments [256][]bool
	for mask := range assignments {
		inputs := make([]bool, 8)
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		values, _, err := g.evaluateObjects(inputs)
		if err != nil {
			t.Fatal(err)
		}
		assignments[mask] = values
	}
	checked := 0
	for id := int32(0); id < g.nObjs; id++ {
		data := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(id)))
		if len(data) == 0 {
			continue
		}
		if data[0] < 1 || data[0] > p.nCutNum {
			t.Fatal("cut count exceeds storage contract")
		}
		position := 1
		var records [][]int32
		for c := int32(0); c < data[0]; c++ {
			n := int(data[position])
			position++
			if n < 0 || int32(n) > p.nCutSize || position+n >= len(data) {
				t.Fatal("invalid serialized cut")
			}
			leaves := data[position : position+n]
			function := data[position+n]
			position += n + 1
			if !slices.IsSorted(leaves) {
				t.Fatal("unsorted cut leaves")
			}
			for i, leaf := range leaves {
				if leaf < 0 || leaf >= g.nObjs || i > 0 && leaf == leaves[i-1] {
					t.Fatal("invalid/repeated cut leaf")
				}
			}
			records = append(records, leaves)
			if p.fCutMin == 0 {
				continue
			}
			truth := vec.Vec_MemReadEntry(p.vTtMem, int(function>>1))
			bit := func(row int) bool { return (truth[row>>6]>>uint(row&63)&1 != 0) != (function&1 != 0) }
			for mask, values := range assignments {
				row := 0
				for i, leaf := range leaves {
					if values[leaf] {
						row |= 1 << uint(i)
					}
				}
				if bit(row) != values[id] {
					t.Fatalf("node%d cut%v function%d CI assignment%d disagrees", id, leaves, function, mask)
				}
			}
			if p.fTruthMin == 0 {
				cache, seen := make([]bool, g.nObjs), make([]bool, g.nObjs)
				for row := 0; row < 1<<uint(n); row++ {
					clear(seen)
					var evaluate func(int32) bool
					evaluate = func(node int32) bool {
						if index := slices.Index(leaves, node); index >= 0 {
							return row&(1<<uint(index)) != 0
						}
						if node == 0 {
							return false
						}
						if seen[node] {
							return cache[node]
						}
						a, b, _, kind := g.Gate(node)
						if kind != 2 && kind != 3 {
							t.Fatalf("node%d cut%v misses boundary node%d kind%d", id, leaves, node, kind)
						}
						va, vb := evaluate(a>>1) != (a&1 != 0), evaluate(b>>1) != (b&1 != 0)
						value := va && vb
						if kind == 3 {
							value = va != vb
						}
						seen[node], cache[node] = true, value
						return value
					}
					if bit(row) != evaluate(id) {
						t.Fatalf("node%d cut%v formal assignment%d disagrees", id, leaves, row)
					}
				}
			}
			checked++
		}
		if position != len(data) {
			t.Fatal("cut record length disagrees")
		}
		for i, a := range records {
			for j, b := range records {
				if i == j {
					continue
				}
				contained := true
				for _, leaf := range a {
					if !slices.Contains(b, leaf) {
						contained = false
						break
					}
				}
				if contained {
					t.Fatalf("node%d contains dominated cuts %v %v", id, a, b)
				}
			}
		}
	}
	return checked
}

func TestCutEnumerationMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cut_enumeration_review.c")
	for _, variant := range []struct {
		name        string
		first, last int
		args        []string
	}{
		{"ordinary", 0, 8, nil}, {"typed-XOR", 8, 16, []string{"xor"}}, {"large-IDs", 16, 24, []string{"large"}},
	} {
		t.Run(variant.name, func(t *testing.T) { cutEnumReviewCompare(t, exe, variant.first, variant.last, variant.args...) })
	}
}

func cutEnumReviewCompare(t *testing.T, exe string, first, last int, args ...string) {
	t.Helper()
	native, err := exec.Command(exe, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("native cuts: %v\n%s", err, native)
	}
	var got strings.Builder
	cases, truths := 0, 0
	for seed := first; seed < last; seed++ {
		for size := int32(2); size <= 8; size++ {
			for _, num := range []int32{2, 10, 50} {
				for mode := 0; mode < 3; mode++ {
					g := cutEnumReviewGraph(seed)
					p := cutEnumReviewEnumerate(g, size, num, giaBool(mode > 0), giaBool(mode > 1))
					truths += cutEnumReviewValidate(t, p)
					fmt.Fprintf(&got, "case %d %d %d %d objs%d counts%.0f:%.0f:%.0f:%.0f over%d\n", seed, size, num, mode, g.nObjs, p.CutCount[0], p.CutCount[1], p.CutCount[2], p.CutCount[3], p.nCutsOver)
					for id := int32(0); id < g.nObjs; id++ {
						a, b, _, kind := g.Gate(id)
						fmt.Fprintf(&got, "gate%d %d %d %d\n", id, kind, a, b)
					}
					for id := int32(0); id < g.nObjs; id++ {
						fmt.Fprintf(&got, "node%d", id)
						for _, v := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(id))) {
							fmt.Fprintf(&got, " %d", v)
						}
						got.WriteByte('\n')
					}
					if mode > 0 {
						for id := 0; id < vec.Vec_MemEntryNum(p.vTtMem); id++ {
							fmt.Fprintf(&got, "truth%d", id)
							for _, word := range vec.Vec_MemReadEntry(p.vTtMem, id) {
								fmt.Fprintf(&got, " %016x", word)
							}
							got.WriteByte('\n')
						}
					}
					Gia_StoFree(p)
					Gia_ManStop(g)
					cases++
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(got.String(), "\n"), strings.Split(string(native), "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("cut source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("cut snapshot lengths differ Go%d C%d", got.Len(), len(native))
	}
	t.Logf("%d exact native configurations and %d independently checked cut truths", cases, truths)
}

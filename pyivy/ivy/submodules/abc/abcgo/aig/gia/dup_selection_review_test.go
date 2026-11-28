package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strings"
	"testing"
)

var selectionReviewInputs = [][]int32{
	{0, 1, 2, 3}, {3, 2, 1, 0}, {1, 0, 3, 2}, {2, 3, 0, 1}, {1, 3, 0, 2}, {3, 0, 2, 1},
}

// Boolean formulas independent of GIA evaluation and copy mapping. Internal
// values follow the source graph's physical gate order, including dangling gates.
func selectionReviewTruth(signs int32, typed bool, inputs []bool) (cos, nodes []bool) {
	a, b, c, d := inputs[0], inputs[1], inputs[2], inputs[3]
	if !typed {
		x := (a != (signs&1 != 0)) && b
		y := x && c
		z := y && d
		u := a && (c != (signs&4 != 0))
		v := (b != (signs&8 != 0)) && d
		w := u && v
		return []bool{z != (signs&2 != 0), w, u != (signs&16 != 0), d, y, true}, []bool{x, y, z, u, v, w}
	}
	x := a && b
	y := x && c
	b0 := d != (signs&1 != 0)
	b1 := b0 != (signs&2 != 0)
	z := b1 != (a != (signs&4 != 0))
	b2 := x != (signs&8 != 0)
	mux := b2
	if c != (signs&16 != 0) {
		mux = z
	}
	root := mux && y
	return []bool{y, root, !b1, z, b2}, []bool{x, y, b0, b1, z, b2, mux, root, b && d}
}

func selectionReviewRecords(t *testing.T, typed, dfs bool) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		// Complete, reversed, repeated and empty CO selections. Four registers
		// are valid even when a repeated CO supplies multiple next-state inputs.
		base, _ := levelReviewGraph(signs, 0, typed)
		count := Gia_ManCoNum(base)
		selections := [][]int32{nil, nil, {count - 1, 0, 0, 3}, nil}
		for i := int32(0); i < count; i++ {
			selections[0] = append(selections[0], i)
			selections[1] = append(selections[1], count-1-i)
		}
		for pi, perm := range selectionReviewInputs {
			for ordering := 0; ordering < 2; ordering++ {
				for si, selected := range selections {
					for _, regs := range []int32{0, 2, 4} {
						if regs > int32(len(selected)) {
							continue
						}
						src, levelOrder := levelReviewGraph(signs, 2, typed)
						if dfs {
							src = Gia_ManDupOrderDfs(src)
						}
						src.nConstrs = 1 // Explicit-vector copying defines a new interface, as in C.
						cis := make([]int32, 4)
						for i, index := range perm {
							cis[i] = Gia_ManCi(src, index).id
						}
						var ands []int32
						for _, obj := range src.pObjs {
							if Gia_ObjIsAnd(obj) != 0 {
								ands = append(ands, obj.id)
							}
						}
						if ordering == 1 {
							if !dfs {
								ands = nil
								for _, lit := range levelOrder {
									ands = append(ands, lit>>1)
								}
							} else {
								Gia_ManLevelNum(src)
								slices.SortStableFunc(ands, func(a, b int32) int { return int(Gia_ObjLevelId(src, a) - Gia_ObjLevelId(src, b)) })
							}
						}
						cos := make([]int32, len(selected))
						for i, index := range selected {
							cos[i] = Gia_ManCo(src, index).id
						}
						// Poison old scratch mappings; omitted objects must not be read using
						// mappings left by an earlier copy.
						for _, obj := range src.pObjs {
							obj.Value = 42
						}
						out := Gia_ManDupFromVecs(src, permutation(cis...), permutation(ands...), permutation(cos...), regs)
						name, spec := out.Identity()
						if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != int32(len(selected)) || out.nRegs != regs || out.nConstrs != 0 || out.nObjs != int32(5+len(ands)+len(cos)) || name != "levels" || spec != "levels.aig" {
							t.Fatal("explicit copy changed the requested interface/identity")
						}
						if Gia_ManIsNormalized(out) == 0 {
							t.Fatal("explicit copy not normalized")
						}
						for i, id := range cis {
							if src.pObjs[id].Value != uint32((i+1)<<1) {
								t.Fatal("CI order ignored")
							}
						}
						for i, id := range ands {
							if src.pObjs[id].Value != uint32((i+5)<<1) {
								t.Fatal("gate order ignored")
							}
						}
						fmt.Fprintf(&records, "0 %d %d %d %d %d %d %d %d %d", signs, pi, ordering, si, regs, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
						for mask := 0; mask < 16; mask++ {
							inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
							original := make([]bool, 4)
							for i, index := range perm {
								original[index] = inputs[i]
							}
							sourceOutputs, _ := selectionReviewTruth(signs, typed, original)
							want := make([]bool, len(selected))
							for i, index := range selected {
								want[i] = sourceOutputs[index]
							}
							checkQuantificationOutputs(t, out, inputs, want)
							packed := 0
							for i, value := range want {
								if value {
									packed |= 1 << uint(i)
								}
							}
							fmt.Fprintf(&records, " %d", packed)
						}
						records.WriteByte('\n')
					}
				}
			}
		}
	}
	return records.String()
}

func outputReviewRecords(t *testing.T, typed, dfs bool) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		for _, regs := range []int32{0, 2, 4} {
			src, _ := levelReviewGraph(signs, regs, typed)
			if dfs {
				src = Gia_ManDupOrderDfs(src)
			}
			src.nConstrs = 1
			// Record source node values in physical order: DFS can reorder nodes and
			// remove dangling logic. Evaluate those roots independently below.
			var ids []int32
			for _, obj := range src.pObjs {
				if Gia_ObjIsAnd(obj) != 0 {
					ids = append(ids, obj.id)
				}
			}
			out := Gia_ManDupOuts(src)
			name, spec := out.Identity()
			if out.nObjs != src.nObjs+int32(len(ids)) || Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != Gia_ManCoNum(src)+int32(len(ids)) || out.nRegs != regs || out.nConstrs != 0 || name != "levels" || spec != "levels.aig" || Gia_ManIsNormalized(out) == 0 {
				t.Fatal("all-node output copy changed source conventions")
			}
			fmt.Fprintf(&records, "1 %d %d %d %d %d %d", signs, regs, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				cos, nodes := selectionReviewTruth(signs, typed, inputs)
				if dfs {
					// The source evaluators were checked against formulas above; here the
					// physical root order is intentionally the DFS layout rather than the
					// fixture's original creation order.
					values, _, err := src.evaluateObjects(inputs)
					if err != nil {
						t.Fatal(err)
					}
					nodes = nil
					for _, id := range ids {
						nodes = append(nodes, values[id])
					}
				}
				split := len(cos) - int(regs)
				want := append(slices.Clone(cos[:split]), nodes...)
				want = append(want, cos[split:]...)
				checkQuantificationOutputs(t, out, inputs, want)
				packed := 0
				for i, value := range want {
					if value {
						packed |= 1 << uint(i)
					}
				}
				fmt.Fprintf(&records, " %d", packed)
			}
			records.WriteByte('\n')
		}
	}
	return records.String()
}

func TestExplicitVectorsAndAllNodeOutputs(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, dfs := range []bool{false, true} {
			t.Run(fmt.Sprintf("typed%v_dfs%v", typed, dfs), func(t *testing.T) { selectionReviewRecords(t, typed, dfs); outputReviewRecords(t, typed, dfs) })
		}
	}
}

func TestExplicitVectorsAndAllNodeOutputsMatchSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "copy_vectors_outputs.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C copy: %v: %s", err, native)
	}
	got := strings.Split(strings.TrimSpace(selectionReviewRecords(t, false, false)+outputReviewRecords(t, false, false)), "\n")
	want := strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases %d, C cases %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, got[i], want[i])
		}
	}
}

func TestExplicitVectorSubsetAndValidation(t *testing.T) {
	for regs := int32(0); regs <= 2; regs++ {
		src := Gia_ManStart(1)
		a, b, c, d := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
		x := Gia_ManAppendAnd(src, a, c)
		y := Gia_ManAppendAnd(src, b, d)
		o := Gia_ManAppendCo(src, x^1)
		k := Gia_ManAppendCo(src, 1)
		Gia_ManAppendCo(src, y)
		out := Gia_ManDupFromVecs(src, permutation(c>>1, a>>1), permutation(x>>1), permutation(o>>1, k>>1, o>>1), regs)
		if Gia_ManCiNum(out) != 2 || Gia_ManCoNum(out) != 3 || out.nRegs != regs || Gia_ManAndNum(out) != 1 {
			t.Fatal("subset copy retained omitted logic")
		}
		if src.pObjs[b>>1].Value != ^uint32(0) || src.pObjs[y>>1].Value != ^uint32(0) {
			t.Fatal("subset copy retained stale mappings")
		}
		for mask := 0; mask < 4; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0}
			value := !(inputs[0] && inputs[1])
			checkQuantificationOutputs(t, out, inputs, []bool{value, true, value})
		}
	}
	cases := []struct {
		name           string
		cis, ands, cos []int32
		regs           int32
	}{
		{"negative regs", []int32{1, 2}, []int32{3, 4}, []int32{5}, -1},
		{"too many input regs", []int32{1, 2}, []int32{3, 4}, []int32{5, 5, 5}, 3},
		{"too many output regs", []int32{1, 2}, []int32{3, 4}, []int32{5}, 2},
		{"negative CI", []int32{-1}, nil, nil, 0},
		{"out of range CI", []int32{6}, nil, nil, 0},
		{"constant CI", []int32{0}, nil, nil, 0},
		{"gate CI", []int32{3}, nil, nil, 0},
		{"output CI", []int32{5}, nil, nil, 0},
		{"repeated CI", []int32{1, 1}, nil, nil, 0},
		{"CI as gate", []int32{1, 2}, []int32{1}, nil, 0},
		{"CO as gate", []int32{1, 2}, []int32{5}, nil, 0},
		{"repeated gate", []int32{1, 2}, []int32{3, 3}, nil, 0},
		{"missing gate input", []int32{1}, []int32{3}, nil, 0},
		{"non topological gates", []int32{1, 2}, []int32{4, 3}, nil, 0},
		{"missing output gate", []int32{1, 2}, []int32{3}, []int32{5}, 0},
		{"CI as CO", []int32{1, 2}, []int32{3, 4}, []int32{1}, 0},
		{"negative CO", nil, nil, []int32{-1}, 0},
		{"out of range CO", nil, nil, []int32{6}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := Gia_ManAppendAnd(src, a, b)
			y := Gia_ManAppendAnd(src, x, a)
			Gia_ManAppendCo(src, y)
			for _, obj := range src.pObjs {
				obj.Value = 0
			} // Invalid dependencies must still fail.
			defer func() {
				if recover() == nil {
					t.Fatal("invalid explicit copy accepted")
				}
			}()
			Gia_ManDupFromVecs(src, permutation(tc.cis...), permutation(tc.ands...), permutation(tc.cos...), tc.regs)
		})
	}
}

func TestExplicitAndAllNodeCopiesAtBoundaries(t *testing.T) {
	for _, mode := range []string{"empty", "constants", "wire", "simple"} {
		t.Run(mode, func(t *testing.T) {
			src := Gia_ManStart(1)
			switch mode {
			case "constants":
				Gia_ManAppendCo(src, 0)
				Gia_ManAppendCo(src, 1)
			case "wire":
				Gia_ManAppendCo(src, Gia_ManAppendCi(src)^1)
			case "simple":
				src.fGiaSimple = 1
				a := Gia_ManAppendCi(src)
				Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
			}
			var cis, ands, cos []int32
			for _, obj := range src.pObjs {
				switch {
				case Gia_ObjIsCi(obj) != 0:
					cis = append(cis, obj.id)
				case Gia_ObjIsCo(obj) != 0:
					cos = append(cos, obj.id)
				case Gia_ObjIsAnd(obj) != 0:
					ands = append(ands, obj.id)
				}
			}
			for _, out := range []*Gia_Man_t{Gia_ManDupFromVecs(src, permutation(cis...), permutation(ands...), permutation(cos...), 0), Gia_ManDupOuts(src)} {
				if out.fGiaSimple != src.fGiaSimple || Gia_ManIsNormalized(out) == 0 {
					t.Fatal("copy lost simple mode or normalization")
				}
				rows := 1
				if len(cis) > 0 {
					rows = 2
				}
				for mask := 0; mask < rows; mask++ {
					var inputs, want []bool
					if len(cis) > 0 {
						inputs = []bool{mask != 0}
					}
					switch mode {
					case "constants":
						want = []bool{false, true}
					case "wire":
						want = []bool{mask == 0}
					case "simple":
						want = []bool{false}
						if Gia_ManCoNum(out) == 2 {
							want = append(want, false)
						}
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			}
		})
	}
}

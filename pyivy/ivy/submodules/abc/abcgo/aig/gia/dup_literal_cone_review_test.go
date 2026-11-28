package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"math/bits"
	"slices"
	"strings"
	"testing"
)

func TestDfsNodeRejectsForeignRoot(t *testing.T) {
	src, other := Gia_ManStart(1), Gia_ManStart(1)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendAnd(src, a, b)
	x, y := Gia_ManAppendCi(other), Gia_ManAppendCi(other)
	root := Gia_ManObj(other, Gia_ManAppendAnd(other, x^1, y)>>1)
	defer func() {
		if recover() == nil {
			t.Fatal("foreign root silently produced a constant instead of rejecting a manager mismatch")
		}
	}()
	Gia_ManDupDfsNode(src, root)
}

type literalConeReviewLeaf struct {
	id             int32
	logical        int // -1 constant; 0..4 CIs; 5+ original logical gate index.
	support, nodes uint
}

func literalConeReviewGraph(signs, regs int32, typed, layout bool) (*Gia_Man_t, []literalConeReviewLeaf) {
	src, _ := levelReviewGraph(signs, 0, typed)
	Gia_ManAppendCi(src)
	Gia_ManSetRegNum(src, regs)
	src.nConstrs = 1
	supports, nodes := []uint{3, 7, 15, 5, 10, 15}, []uint{1, 3, 7, 8, 16, 56}
	if typed {
		supports = []uint{3, 7, 8, 8, 9, 3, 15, 15, 10}
		nodes = []uint{1, 3, 4, 12, 28, 33, 125, 255, 256}
	}
	var oldIDs []int32
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			oldIDs = append(oldIDs, obj.id)
		}
	}
	old := src
	if layout {
		src = Gia_ManDupOrderDfsReverse(src)
	}
	leaves := []literalConeReviewLeaf{{id: 0, logical: -1}}
	for i := int32(0); i < 5; i++ {
		leaves = append(leaves, literalConeReviewLeaf{id: Gia_ManCi(src, i).id, logical: int(i), support: 1 << uint(i)})
	}
	for i, id := range oldIDs {
		if layout {
			value := old.pObjs[id].Value
			if value == ^uint32(0) {
				continue
			}
			id = int32(value >> 1)
		}
		leaves = append(leaves, literalConeReviewLeaf{id: id, logical: 5 + i, support: supports[i], nodes: nodes[i]})
	}
	return src, leaves
}

func literalConeReviewRecords(t *testing.T, typed, layout bool) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		for _, regs := range []int32{0, 2, 4} {
			for mode := 0; mode < 2; mode++ {
				src, leaves := literalConeReviewGraph(signs, regs, typed, layout)
				var cases [][]int
				if mode == 0 {
					for i, leaf := range leaves {
						if leaf.logical >= 5 {
							cases = append(cases, []int{i << 1})
						}
					}
				} else {
					for i := range leaves {
						cases = append(cases, []int{i << 1}, []int{(i << 1) | 1})
					}
					last := (len(leaves) - 1) << 1
					cases = append(cases, nil, []int{1, 10, last | 1, 12, last, 3}, []int{last, last | 1, last}, []int{12, 14, 12 | 1})
				}
				for index, selected := range cases {
					var literals []int32
					support, dependencies := uint(0), uint(0)
					for _, token := range selected {
						leaf := leaves[token>>1]
						literals = append(literals, (leaf.id<<1)|int32(token&1))
						support |= leaf.support
						dependencies |= leaf.nodes
					}
					var out *Gia_Man_t
					if mode == 0 {
						out = Gia_ManDupDfsNode(src, Gia_ManObj(src, literals[0]>>1))
					} else {
						args := permutation(literals...)
						out = Gia_ManDupDfsLitArray(src, args)
						if !slices.Equal(Vec_IntArray(args), literals) {
							t.Fatal("literal array was modified")
						}
					}
					expectedSupport := support
					if mode == 1 {
						expectedSupport = 31
					}
					name, spec := out.Identity()
					if Gia_ManCiNum(out) != int32(bits.OnesCount(expectedSupport)) || Gia_ManAndNum(out) != int32(bits.OnesCount(dependencies)) || Gia_ManCoNum(out) != int32(len(selected)) || out.nRegs != 0 || out.nConstrs != 0 || name != "levels" || spec != "levels.aig" {
						t.Fatalf("mode %d case %d: wrong extracted logic/interface", mode, index)
					}
					originalOrder := make([]int, Gia_ManCiNum(out))
					for i := int32(0); i < 5; i++ {
						value := Gia_ManCi(src, i).Value
						if expectedSupport&(1<<uint(i)) == 0 {
							if value != ^uint32(0) {
								t.Fatal("node extraction retained an unused input")
							}
							continue
						}
						if value == ^uint32(0) || value&1 != 0 {
							t.Fatal("invalid extracted input mapping")
						}
						obj := Gia_ManObj(out, int32(value>>1))
						ordinal := Gia_ObjCioId(obj)
						if Gia_ManCi(out, ordinal) != obj {
							t.Fatal("input vector/ordinal mismatch")
						}
						originalOrder[ordinal] = int(i)
						if mode == 1 && ordinal != i {
							t.Fatal("literal extraction reordered original inputs")
						}
					}
					for _, leaf := range leaves {
						if leaf.logical >= 5 {
							copied := src.pObjs[leaf.id].Value != ^uint32(0)
							if copied != (dependencies&(1<<uint(leaf.logical-5)) != 0) {
								t.Fatal("wrong copied/dead source node mapping")
							}
						}
					}
					for i := int32(0); i < Gia_ManCoNum(src); i++ {
						if Gia_ManCo(src, i).Value != ^uint32(0) {
							t.Fatal("literal/node extraction unexpectedly copied a source CO")
						}
					}
					fmt.Fprintf(&records, "%d %d %d %d %d %d %d", mode, signs, regs, index, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out))
					for _, obj := range src.pObjs {
						fmt.Fprintf(&records, " %d", obj.Value)
					}
					for assignment := 0; assignment < 1<<uint(Gia_ManCiNum(out)); assignment++ {
						inputs := make([]bool, Gia_ManCiNum(out))
						original := make([]bool, 5)
						for i, sourceIndex := range originalOrder {
							inputs[i] = assignment&(1<<uint(i)) != 0
							original[sourceIndex] = inputs[i]
						}
						_, nodeValues := selectionReviewTruth(signs, typed, original)
						want := make([]bool, len(selected))
						for i, token := range selected {
							leaf := leaves[token>>1]
							var value bool
							if leaf.logical >= 5 {
								value = nodeValues[leaf.logical-5]
							} else if leaf.logical >= 0 {
								value = original[leaf.logical]
							}
							want[i] = value != (token&1 != 0)
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
	return records.String()
}

func TestDfsNodeAndLiteralConeSemantics(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			t.Run(fmt.Sprintf("typed%v_layout%v", typed, layout), func(t *testing.T) { literalConeReviewRecords(t, typed, layout) })
		}
	}
}

func literalConeReviewOldRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 32; signs++ {
		for _, regs := range []int32{0, 2, 4} {
			src, unused := dfsReviewRepresentativeGraph(signs, regs)
			var selected []int32
			for i := int32(0); i < Gia_ManCoNum(src); i++ {
				lit, _, _, _ := src.Gate(Gia_ManCo(src, i).id)
				selected = append(selected, lit)
			}
			selected = append(selected, 0, 1, selected[2]^1, selected[0])
			args := permutation(selected...)
			out := Gia_ManDupDfsLitArray(src, args)
			if Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != 9 || out.nRegs != 0 || Gia_ManAndNum(out) != 1 || src.pObjs[unused].Value != ^uint32(0) || !slices.Equal(selected, Vec_IntArray(args)) {
				t.Fatal("literal extraction ignored representative substitution or changed its interface/arguments")
			}
			fmt.Fprintf(&records, "%d %d %d", signs, regs, out.nObjs)
			for _, obj := range src.pObjs {
				fmt.Fprintf(&records, " %d", obj.Value)
			}
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				x := (inputs[0] != (signs&1 != 0)) && (inputs[1] != (signs&2 != 0))
				want := []bool{x != (signs&4 != 0), (!x) != (signs&8 != 0), (!x) != (signs&16 != 0), false, true}
				want = append(want, false, true, !want[2], want[0])
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

func TestDfsLiteralConesUseOldRepresentatives(t *testing.T) { literalConeReviewOldRecords(t) }

func TestDfsNodeAndLiteralConesMatchOriginalSource(t *testing.T) {
	exe := compileGiaSourceOracle(t, "dfs_literal_cones.c")
	for _, old := range []bool{false, true} {
		var args []string
		if old {
			args = []string{"old"}
		}
		native, err := testoracle.Command(t, exe, args...).CombinedOutput()
		if err != nil {
			t.Fatalf("C literal cones: %v: %s", err, native)
		}
		var records string
		if old {
			records = literalConeReviewOldRecords(t)
		} else {
			records = literalConeReviewRecords(t, false, false)
		}
		got, want := strings.Split(strings.TrimSpace(records), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
		if len(got) != len(want) {
			t.Fatalf("old%v: Go cases %d, C cases %d", old, len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("old%v case %d:\nGo %s\nC %s", old, i, got[i], want[i])
			}
		}
	}
}

func TestDfsLiteralConeBoundariesAndRejections(t *testing.T) {
	for _, kind := range []string{"empty", "unused-input", "constant-buffer"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(1)
			var selected []int32
			var inputs, want []bool
			switch kind {
			case "unused-input":
				a := Gia_ManAppendCi(src)
				selected = []int32{0, 1, a ^ 1}
				inputs = []bool{true}
				want = []bool{false, true, false}
			case "constant-buffer":
				buf := Gia_ManAppendBuf(src, 1)
				selected = []int32{buf, buf ^ 1, 1, 0}
				want = []bool{true, false, true, false}
				out := Gia_ManDupDfsNode(src, Gia_ManObj(src, buf>>1))
				checkQuantificationOutputs(t, out, nil, []bool{true})
				if Gia_ManCiNum(out) != 0 || Gia_ManAndNum(out) != 1 {
					t.Fatal("constant buffer node extraction changed graph")
				}
			}
			out := Gia_ManDupDfsLitArray(src, permutation(selected...))
			checkQuantificationOutputs(t, out, inputs, want)
			if out.nRegs != 0 || Gia_ManCiNum(out) != Gia_ManCiNum(src) {
				t.Fatal("boundary literal extraction changed inputs/state")
			}
		})
	}
	for _, kind := range []string{"nil", "constant", "input", "output", "inverse"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			x := Gia_ManAppendAnd(src, a, b)
			co := Gia_ManAppendCo(src, x)
			var root *Gia_Obj_t
			switch kind {
			case "constant":
				root = src.pObjs[0]
			case "input":
				root = src.pObjs[a>>1]
			case "output":
				root = src.pObjs[co>>1]
			case "inverse":
				root = Gia_Not(src.pObjs[x>>1])
			}
			for _, obj := range src.pObjs {
				obj.Value = 99
			}
			defer func() {
				if recover() == nil {
					t.Error("invalid node root accepted")
				}
				for _, obj := range src.pObjs {
					if obj.Value != 99 {
						t.Error("root rejection changed source mappings")
					}
				}
			}()
			Gia_ManDupDfsNode(src, root)
		})
	}
	for _, lit := range []int32{-1, 1<<31 - 1, 2, 3} {
		t.Run(fmt.Sprintf("invalidLit%d", lit), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid literal accepted")
				}
			}()
			Gia_ManDupDfsLitArray(Gia_ManStart(1), permutation(lit))
		})
	}
	src := Gia_ManStart(1)
	co := Gia_ManAppendCo(src, 1)
	t.Run("CO literal", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("CO used as a logic literal")
			}
		}()
		Gia_ManDupDfsLitArray(src, permutation(co))
	})
}

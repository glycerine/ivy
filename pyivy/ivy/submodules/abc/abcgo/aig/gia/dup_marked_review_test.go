package gia

import (
	"fmt"
	"math/bits"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestMarkedDuplicationRetainsInterleavedPortOrder(t *testing.T) {
	for _, regs := range []int32{1, 2} {
		t.Run(map[int32]string{1: "one_register", 2: "two_registers"}[regs], func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					t.Fatalf("complete retained register pairs rejected: %v", err)
				}
			}()
			src := Gia_ManStart(1)
			a, b, r0 := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			if regs == 2 {
				Gia_ManAppendCi(src)
			}
			root := Gia_ManAppendAnd(src, Gia_ManAppendAnd(src, a, b), r0^1)
			Gia_ManAppendCo(src, root)
			Gia_ManAppendCo(src, a^1)
			Gia_ManAppendCo(src, b)
			if regs == 2 {
				Gia_ManAppendCo(src, a)
			}
			Gia_ManSetRegNum(src, regs)
			src.nConstrs = 1
			src = Gia_ManDupOrderDfsReverse(src)
			out := Gia_ManDupMarked(src)
			if out.nRegs != regs || out.nConstrs != 1 {
				t.Fatal("register/constraint roles changed")
			}
			for mask := 0; mask < 1<<uint(Gia_ManCiNum(src)); mask++ {
				inputs := make([]bool, Gia_ManCiNum(src))
				for i := range inputs {
					inputs[i] = mask&(1<<uint(i)) != 0
				}
				want := []bool{inputs[0] && inputs[1] && !inputs[2], !inputs[0], inputs[1]}
				if regs == 2 {
					want = append(want, inputs[0])
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}

type markedReviewFixture struct {
	graph      *Gia_Man_t
	branch     [3][]int32
	properties [3]int32
	constants  [2]int32
}

func markedReviewGraph(signs int, typed, layout bool) markedReviewFixture {
	g := Gia_ManStart(1)
	f := markedReviewFixture{graph: g}
	inputs := make([]int32, 7)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
	}
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	var x [3]int32
	if typed {
		x[0] = Gia_ManAppendXorReal(g, inputs[0]^phase(0), inputs[4]^phase(1))
		g.EnableMuxes()
		x[1] = Gia_ManAppendMuxReal(g, inputs[5]^phase(2), inputs[1]^phase(3), inputs[3]^phase(4))
	} else {
		x[0] = Gia_ManAppendAnd(g, inputs[0]^phase(0), inputs[4]^phase(1))
		x[1] = Gia_ManAppendAnd(g, inputs[1]^phase(3), inputs[5]^phase(2))
	}
	x[2] = Gia_ManAppendAnd(g, inputs[2]^phase(5), inputs[6]^phase(6))
	for j := range x {
		f.branch[j] = []int32{x[j] >> 1}
	}
	if typed {
		x[2] = Gia_ManAppendBuf(g, x[2]^1)
		f.branch[2] = append(f.branch[2], x[2]>>1)
	}
	f.properties[0] = Gia_ManAppendCo(g, x[0]^phase(0)) >> 1
	f.constants[0] = Gia_ManAppendCo(g, 0) >> 1
	f.constants[1] = Gia_ManAppendCo(g, 1) >> 1
	f.properties[1] = Gia_ManAppendCo(g, x[1]^phase(1)) >> 1
	f.properties[2] = Gia_ManAppendCo(g, x[2]^phase(2)) >> 1
	for j := range x {
		Gia_ManAppendCo(g, x[j]^1)
	}
	Gia_ManSetRegNum(g, 3)
	g.nConstrs = 1
	g.SetIdentity("marked", "marked.aig")
	if layout {
		next := Gia_ManDupOrderDfsReverse(g)
		remap := func(id int32) int32 { return int32(g.pObjs[id].Value >> 1) }
		for j := range f.branch {
			for i, id := range f.branch[j] {
				f.branch[j][i] = remap(id)
			}
			f.properties[j] = remap(f.properties[j])
		}
		for i, id := range f.constants {
			f.constants[i] = remap(id)
		}
		f.graph = next
		g = next
	}
	g.aigerBadOutputs = 1
	g.VNamesIn = Vec_PtrAlloc(7)
	for _, name := range []string{"a0", "a1", "a2", "anchor", "r0", "r1", "r2"} {
		Vec_PtrPush(g.VNamesIn, name)
	}
	g.VNamesOut = Vec_PtrAlloc(8)
	for _, name := range []string{"ordinary", "false", "true", "bad", "constraint", "next0", "next1", "next2"} {
		Vec_PtrPush(g.VNamesOut, name)
	}
	g.vRegClasses = permutation(10, 20, 30)
	g.vRegInits = permutation(0, 1, 0)
	g.VFlopClasses = permutation(100, 200, 300)
	g.aigerLatchInits = []int32{0, 1, 0}
	g.aigerComment = []byte("marked comment")
	g.pCexSeq = util.AbcCexSentinel // Derived interfaces intentionally omit traces.
	return f
}

func markedReviewRecords(t *testing.T, typed, layout bool) string {
	t.Helper()
	var records strings.Builder
	for _, signs := range []int{0, 37, 90, 127} {
		for regs := 0; regs < 8; regs++ {
			for pis := 0; pis < 8; pis++ {
				if pis&regs != regs {
					continue
				}
				for props := 0; props < 8; props++ {
					if props&regs != props {
						continue
					}
					for constants := 0; constants < 4; constants++ {
						for anchor := 0; anchor < 2; anchor++ {
							if typed && regs&2 != 0 && anchor == 0 {
								continue
							}
							f := markedReviewGraph(signs, typed, layout)
							src := f.graph
							mark := func(id int32) { src.pObjs[id].fMark0 = 1 }
							var inIndices, outIndices, registers []int32
							for j := 0; j < 3; j++ {
								if pis&(1<<uint(j)) == 0 {
									mark(Gia_ManPi(src, int32(j)).id)
								} else {
									inIndices = append(inIndices, int32(j))
								}
								if regs&(1<<uint(j)) == 0 {
									mark(Gia_ManRo(src, int32(j)).id)
									mark(Gia_ManRi(src, int32(j)).id)
									for _, id := range f.branch[j] {
										mark(id)
									}
								} else {
									registers = append(registers, int32(j))
								}
								if props&(1<<uint(j)) == 0 {
									mark(f.properties[j])
								}
							}
							if anchor == 0 {
								mark(Gia_ManPi(src, 3).id)
							} else {
								inIndices = append(inIndices, 3)
							}
							for _, j := range registers {
								inIndices = append(inIndices, 4+j)
							}
							for j, id := range f.constants {
								if constants&(1<<uint(j)) == 0 {
									mark(id)
								}
							}
							for i := int32(0); i < Gia_ManCoNum(src); i++ {
								if Gia_ManCo(src, i).fMark0 == 0 {
									outIndices = append(outIndices, i)
								}
							}
							removed := make([]bool, src.nObjs)
							mapped := make([]int32, src.nObjs)
							kept := int32(0)
							for id, obj := range src.pObjs {
								removed[id] = obj.fMark0 != 0
								mapped[id] = -1
								if !removed[id] {
									mapped[id] = kept
									kept++
								}
								obj.Value = uint32(id + 55)
								obj.fMark1 = uint8(id & 1)
							}
							inputNames, outputNames := src.InputNames(), src.OutputNames()
							out := Gia_ManDupMarked(src)
							name, spec := out.Identity()
							if out.nObjs != kept || out.nRegs != int32(bits.OnesCount(uint(regs))) || out.nConstrs != int32((props>>2)&1) || out.aigerBadOutputs != int32((props>>1)&1) || name != "marked" || spec != "marked.aig" || out.pCexSeq != nil {
								t.Fatal("marked graph changed retained shape, roles, identity or trace policy")
							}
							for id, obj := range src.pObjs {
								expected := ^uint32(0)
								if mapped[id] >= 0 {
									expected = uint32(mapped[id] << 1)
								}
								if obj.Value != expected || obj.fMark0 != 0 || obj.fMark1 != uint8(id&1) {
									t.Fatal("removed identity map, consumed mark0 or independent mark1 was lost")
								}
							}
							wantedNames := func(names []string, indices []int32) []string {
								result := make([]string, 0, len(indices))
								for _, i := range indices {
									result = append(result, names[i])
								}
								return result
							}
							if !slices.Equal(out.InputNames(), wantedNames(inputNames, inIndices)) || !slices.Equal(out.OutputNames(), wantedNames(outputNames, outIndices)) {
								t.Fatal("retained port names lost logical order")
							}
							for i, index := range inIndices {
								obj := Gia_ManCi(out, int32(i))
								if obj.id != mapped[Gia_ManCi(src, index).id] || Gia_ObjCioId(obj) != int32(i) {
									t.Fatal("retained input ordinal not rebuilt")
								}
							}
							for i, index := range outIndices {
								obj := Gia_ManCo(out, int32(i))
								if obj.id != mapped[Gia_ManCo(src, index).id] || Gia_ObjCioId(obj) != int32(i) {
									t.Fatal("retained output ordinal not rebuilt")
								}
							}
							for _, pair := range []struct{ src, dst *Vec_Int_t }{{src.vRegClasses, out.vRegClasses}, {src.vRegInits, out.vRegInits}, {src.VFlopClasses, out.VFlopClasses}} {
								want := make([]int32, 0, len(registers))
								for _, index := range registers {
									want = append(want, Vec_IntEntry(pair.src, int(index)))
								}
								if pair.src == pair.dst || !slices.Equal(Vec_IntArray(pair.dst), want) {
									t.Fatal("register metadata subset/order/ownership lost")
								}
								if len(want) > 0 {
									Vec_IntWriteEntry(pair.dst, 0, -999)
									if Vec_IntEntry(pair.src, int(registers[0])) != want[0] {
										t.Fatal("register metadata aliases source")
									}
								}
							}
							resets := make([]int32, 0, len(registers))
							for _, i := range registers {
								resets = append(resets, src.aigerLatchInits[i])
							}
							if !slices.Equal(out.aigerLatchInits, resets) || string(out.aigerComment) != "marked comment" {
								t.Fatal("reset/comment subset lost")
							}
							emit := !typed && !layout && props&4 != 0 // C retains stale constraint counts when that PO is removed.
							if emit {
								fmt.Fprintf(&records, "%d %d %d %d %d %d result %d %d %d %d %d values", signs, regs, pis, props, constants, anchor, out.nObjs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, out.nConstrs)
								for _, obj := range src.pObjs {
									fmt.Fprintf(&records, " %d", obj.Value)
								}
								fmt.Fprint(&records, " truth")
							}
							for mask := 0; mask < 1<<uint(len(inIndices)); mask++ {
								inputs := make([]bool, len(inIndices))
								original := make([]bool, 7)
								for i, index := range inIndices {
									inputs[i] = mask&(1<<uint(i)) != 0
									original[index] = inputs[i]
								}
								phase := func(bit uint) bool { return signs&(1<<bit) != 0 }
								x0 := (original[0] != phase(0)) && (original[4] != phase(1))
								if typed {
									x0 = (original[0] != phase(0)) != (original[4] != phase(1))
								}
								x1 := (original[1] != phase(3)) && (original[5] != phase(2))
								if typed {
									x1 = original[3] != phase(4)
									if original[5] != phase(2) {
										x1 = original[1] != phase(3)
									}
								}
								x2 := (original[2] != phase(5)) && (original[6] != phase(6))
								if typed {
									x2 = !x2
								}
								all := []bool{x0 != phase(0), false, true, x1 != phase(1), x2 != phase(2), !x0, !x1, !x2}
								want := make([]bool, len(outIndices))
								packed := 0
								for i, index := range outIndices {
									want[i] = all[index]
									if want[i] {
										packed |= 1 << uint(i)
									}
								}
								checkQuantificationOutputs(t, out, inputs, want)
								if emit {
									fmt.Fprintf(&records, " %d", packed)
								}
							}
							if emit {
								records.WriteByte('\n')
							}
							if len(resets) > 0 {
								out.aigerLatchInits[0] = 99
								if src.aigerLatchInits[registers[0]] != resets[0] {
									t.Fatal("reset storage aliases source")
								}
							}
							out.aigerComment[0] = 'X'
							if string(src.aigerComment) != "marked comment" {
								t.Fatal("comment aliases source")
							}
							if len(inIndices) > 0 {
								Vec_PtrWriteEntry(out.VNamesIn, 0, "changed")
								if src.InputNames()[inIndices[0]] != inputNames[inIndices[0]] {
									t.Fatal("input names alias source")
								}
							}
							if len(outIndices) > 0 {
								Vec_PtrWriteEntry(out.VNamesOut, 0, "changed")
								if src.OutputNames()[outIndices[0]] != outputNames[outIndices[0]] {
									t.Fatal("output names alias source")
								}
							}
						}
					}
				}
			}
		}
	}
	return records.String()
}

func TestMarkedDuplicationSelectionAndStateMatrix(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, layout := range []bool{false, true} {
			markedReviewRecords(t, typed, layout)
		}
	}
}

func TestMarkedDuplicationMatchesWorkingCRegion(t *testing.T) {
	exe := compileGiaSourceOracle(t, "marked_duplicate.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C marked duplicate: %v: %s", err, native)
	}
	got, want := strings.Split(strings.TrimSpace(markedReviewRecords(t, false, false)), "\n"), strings.Split(strings.TrimSpace(string(native)), "\n")
	if len(got) != len(want) {
		t.Fatalf("Go cases%d C cases%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("case%d\nGo %s\nC %s", i, got[i], want[i])
		}
	}
}

func TestMarkedDuplicationClassAndChoiceSubsets(t *testing.T) {
	for removed := 0; removed < 8; removed++ {
		src := Gia_ManStart(1)
		a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
		nodes := []int32{Gia_ManAppendAnd(src, a, b) >> 1, Gia_ManAppendAnd(src, b, a) >> 1, Gia_ManAppendAnd(src, a, b) >> 1}
		co := Gia_ManAppendCo(src, nodes[2]<<1) >> 1
		src.StartRepresentatives()
		Gia_ObjSetRepr(src, nodes[1], nodes[0])
		Gia_ObjSetRepr(src, nodes[2], nodes[0])
		src.PNexts = Gia_ManDeriveNexts(src)
		src.pSibls = make([]int32, src.nObjs)
		src.pSibls[nodes[1]] = nodes[0]
		src.pSibls[nodes[2]] = nodes[1]
		beforeReprs, beforeNexts, beforeSibls := slices.Clone(src.PReprs), slices.Clone(src.PNexts), slices.Clone(src.pSibls)
		for i, id := range nodes {
			if removed&(1<<uint(i)) != 0 {
				src.pObjs[id].fMark0 = 1
			}
		}
		if removed&4 != 0 {
			src.pObjs[co].fMark0 = 1
		}
		out := Gia_ManDupMarked(src)
		var survivors []int32
		for i, id := range nodes {
			if removed&(1<<uint(i)) == 0 {
				survivors = append(survivors, int32(src.pObjs[id].Value>>1))
			}
		}
		for i, id := range survivors {
			want := int32(GIA_VOID)
			next := int32(0)
			if i > 0 {
				want = survivors[0]
			}
			if i+1 < len(survivors) {
				next = survivors[i+1]
			}
			if Gia_ObjRepr(out, id) != want || Gia_ObjNext(out, id) != next {
				t.Fatal("retained equivalence class was not reheaded/linked")
			}
		}
		for i, id := range nodes {
			if removed&(1<<uint(i)) != 0 {
				continue
			}
			want := int32(0)
			if sibling := beforeSibls[id]; sibling != 0 && src.pObjs[sibling].Value != ^uint32(0) {
				want = int32(src.pObjs[sibling].Value >> 1)
			}
			if Gia_ObjSibl(out, int32(src.pObjs[id].Value>>1)) != want {
				t.Fatal("choice did not retain the source's immediate-sibling rule")
			}
		}
		if !slices.Equal(src.PReprs, beforeReprs) || !slices.Equal(src.PNexts, beforeNexts) || !slices.Equal(src.pSibls, beforeSibls) {
			t.Fatal("marked copy changed source class/choice metadata")
		}
		for mask := 0; mask < 4; mask++ {
			a, b := mask&1 != 0, mask&2 != 0
			want := []bool{}
			if removed&4 == 0 {
				want = append(want, a && b)
			}
			checkQuantificationOutputs(t, out, []bool{a, b}, want)
		}
		if len(survivors) > 0 {
			out.PReprs[survivors[0]].iRepr = 0
			out.PNexts[survivors[0]] = 0
			out.pSibls[survivors[0]] = 0
		}
		if !slices.Equal(src.PReprs, beforeReprs) || !slices.Equal(src.PNexts, beforeNexts) || !slices.Equal(src.pSibls, beforeSibls) {
			t.Fatal("class/choice storage aliases source")
		}
	}
}

func TestMarkedDuplicationRejectsIncompletePairs(t *testing.T) {
	for _, layout := range []bool{false, true} {
		for _, kind := range []string{"missing_ro", "missing_ri", "mismatched"} {
			t.Run(fmt.Sprintf("%v_%s", layout, kind), func(t *testing.T) {
				src := Gia_ManStart(1)
				a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
				Gia_ManAppendCi(src)
				Gia_ManAppendCi(src)
				Gia_ManAppendCo(src, a)
				Gia_ManAppendCo(src, b)
				Gia_ManAppendCo(src, a)
				Gia_ManSetRegNum(src, 2)
				if layout {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				if kind != "missing_ri" {
					Gia_ManRo(src, 1).fMark0 = 1
				}
				if kind != "missing_ro" {
					Gia_ManRi(src, 0).fMark0 = 1
				}
				defer func() {
					if recover() == nil {
						t.Fatal("incomplete/mismatched RO/RI pairs accepted")
					}
				}()
				Gia_ManDupMarked(src)
			})
		}
	}
}

func TestMarkedDuplicationRejectsRetainedRemovedDependencies(t *testing.T) {
	for _, kind := range []string{"and", "mux", "constant"} {
		t.Run(kind, func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			root := int32(0)
			switch kind {
			case "and":
				root = Gia_ManAppendAnd(src, a, b)
				Gia_ManObj(src, a>>1).fMark0 = 1
			case "mux":
				src.EnableMuxes()
				root = Gia_ManAppendMuxReal(src, a, b, c)
				Gia_ManObj(src, a>>1).fMark0 = 1
			case "constant":
				src.pObjs[0].fMark0 = 1
			}
			Gia_ManAppendCo(src, root)
			defer func() {
				if recover() == nil {
					t.Fatal("invalid retained dependency accepted")
				}
			}()
			Gia_ManDupMarked(src)
		})
	}
}

func TestMarkedDuplicationPreservesSimpleMode(t *testing.T) {
	src := Gia_ManStart(1)
	src.fGiaSimple = 1
	a := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, a^1))
	out := Gia_ManDupMarked(src)
	if out.fGiaSimple != 1 {
		t.Fatal("simple construction mode lost")
	}
	for _, a := range []bool{false, true} {
		checkQuantificationOutputs(t, out, []bool{a}, []bool{false})
	}
}

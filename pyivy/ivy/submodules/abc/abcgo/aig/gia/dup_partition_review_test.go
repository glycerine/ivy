package gia

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

type partitionReviewCopy struct {
	name string
	copy func(*Gia_Man_t) *Gia_Man_t
	regs []int32
}

func partitionReviewCopies() []partitionReviewCopy {
	return []partitionReviewCopy{
		{"basic_control", Gia_ManDup, []int32{0, 1}},
		{"attributes_control", Gia_ManDupWithAttributes, []int32{0, 1}},
		{"ci_map_control", func(g *Gia_Man_t) *Gia_Man_t {
			mapping := make([]int32, Gia_ManCiNum(g))
			for i := range mapping {
				mapping[i] = -1
			}
			return Gia_ManDupDfsCiMap(g, mapping, nil)
		}, []int32{0, 1}},
		{"lowering", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupNoMuxes(g, 1) }, []int32{0, 1}},
		{"normalize", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupNormalize(g, 0) }, []int32{0, 1}},
		{"normalize_hash", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupNormalize(g, 1) }, []int32{0, 1}},
		{"levelized", Gia_ManDupLevelized, []int32{0, 1}},
		{"ordered_dfs", Gia_ManDupOrderDfs, []int32{0, 1}},
		{"reverse_dfs", Gia_ManDupOrderDfsReverse, []int32{0, 1}},
		{"aiger_order", Gia_ManDupOrderAiger, []int32{0, 1}},
		{"dfs", Gia_ManDupDfs, []int32{0, 1}},
		{"dfs2", Gia_ManDupDfs2, []int32{0, 1}},
		{"classes", func(g *Gia_Man_t) *Gia_Man_t {
			g.pReprsOld = make([]int32, g.nObjs)
			for i := range g.pReprsOld {
				g.pReprsOld[i] = -1
			}
			return Gia_ManDupDfsClasses(g)
		}, []int32{0, 1}},
		{"abstract_none", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(), permutation(-1, -1)) }, []int32{0, 1}},
		{"abstract_first", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(0), permutation(0, -1)) }, []int32{1}},
		{"abstract_second", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(1), permutation(-1, 0)) }, []int32{0}},
		{"abstract_both_reversed", func(g *Gia_Man_t) *Gia_Man_t { return Gia_ManDupAbs(g, permutation(1, 0), permutation(1, 0)) }, nil},
		{"cut", Gia_ManDupReplaceCut, []int32{0, 1}},
	}
}

func partitionReviewGraph(signs int, typed, layout bool) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, c, r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	phase := func(bit uint) int32 { return int32(signs>>bit) & 1 }
	var x, y, z int32
	if typed {
		x = Gia_ManAppendXorReal(g, a^phase(0), b^phase(1))
		g.EnableMuxes()
		y = Gia_ManAppendMuxReal(g, r0^phase(2), x, c^phase(3))
		z = Gia_ManAppendBuf(g, Gia_ManAppendAnd(g, r1, c)^1)
	} else {
		x = Gia_ManAppendAnd(g, a^phase(0), b^phase(1))
		y = Gia_ManAppendAnd(g, x, r0^phase(2))
		z = Gia_ManAppendAnd(g, r1, c)
	}
	for _, lit := range []int32{x, y, y ^ 1, z, a ^ 1, b, r1 ^ 1} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, 2)
	g.SetIdentity("partition", "partition.aig")
	if layout {
		g = Gia_ManDupOrderDfsReverse(g)
	}
	return g
}

func TestFullOutputCopiesRetainBadPartition(t *testing.T) {
	for _, tc := range partitionReviewCopies() {
		t.Run(tc.name, func(t *testing.T) {
			src := partitionReviewGraph(0, false, false)
			if err := src.SetAigerProperties(2, 2); err != nil {
				t.Fatal(err)
			}
			out := tc.copy(src)
			if out.nConstrs != 2 || out.aigerBadOutputs != 2 {
				t.Errorf("copied roles: constraints%d bad%d, want 2/2", out.nConstrs, out.aigerBadOutputs)
			}
			data, err := MarshalAiger(out, AigerWriteOptions{OmitExtensions: true, OmitSymbols: true})
			if err != nil {
				t.Fatal(err)
			}
			restored, err := ParseAiger(data, AigerReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if restored.nConstrs != 2 || restored.aigerBadOutputs != 2 {
				t.Errorf("AIGER round trip lost bad outputs: header %s", bytes.SplitN(data, []byte{'\n'}, 2)[0])
			}
		})
	}
}

func partitionReviewRoundTrip(t *testing.T, g *Gia_Man_t, bad, constr int32) *Gia_Man_t {
	t.Helper()
	if g.aigerBadOutputs != bad || g.nConstrs != constr {
		t.Fatalf("role counts bad%d constraints%d want%d/%d", g.aigerBadOutputs, g.nConstrs, bad, constr)
	}
	// AIGER deliberately accepts ordinary AND graphs only. Use the public
	// lowering path for real gates and keep the original graph for truth checks.
	wire := g
	for _, obj := range g.pObjs {
		if Gia_ObjIsXor(obj) != 0 || Gia_ObjIsMux(g, obj) != 0 {
			wire = Gia_ManDupNoMuxes(g, 1)
			break
		}
	}
	data, err := MarshalAiger(wire, AigerWriteOptions{OmitExtensions: true, OmitSymbols: true})
	if err != nil {
		t.Fatal(err)
	}
	header := strings.Fields(string(bytes.SplitN(data, []byte{'\n'}, 2)[0]))
	field := func(i int) int32 {
		if i >= len(header) {
			return 0
		}
		n, err := strconv.ParseInt(header[i], 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		return int32(n)
	}
	if field(2) != Gia_ManPiNum(g) || field(3) != g.nRegs || field(4) != Gia_ManPoNum(g)-bad-constr || field(6) != bad || field(7) != constr {
		t.Fatalf("wrong ordinary/bad/constraint header: %v", header)
	}
	parsed, err := ParseAiger(data, AigerReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if parsed.aigerBadOutputs != bad || parsed.nConstrs != constr || parsed.nRegs != g.nRegs || Gia_ManPiNum(parsed) != Gia_ManPiNum(g) || Gia_ManPoNum(parsed) != Gia_ManPoNum(g) {
		t.Fatal("AIGER round trip changed role/port partition")
	}
	return parsed
}

// Source copies retain source-to-copy Value scratch. Lift original assignments
// to any new cut/PPI inputs and project only the surviving RI suffix. This is
// an exact copy check for identity transforms and one-way lifting for cuts.
func partitionReviewLift(t *testing.T, src, out *Gia_Man_t, original []bool) ([]bool, []bool) {
	t.Helper()
	values, outputs, err := src.evaluateObjects(original)
	if err != nil {
		t.Fatal(err)
	}
	inputs := make([]bool, Gia_ManCiNum(out))
	assigned := make([]bool, len(inputs))
	for id, obj := range src.pObjs {
		if obj.Value == ^uint32(0) || int32(obj.Value>>1) >= out.nObjs {
			continue
		}
		copied := out.pObjs[obj.Value>>1]
		if Gia_ObjIsCi(copied) == 0 {
			continue
		}
		index := Gia_ObjCioId(copied)
		value := values[id] != (obj.Value&1 != 0)
		if assigned[index] && inputs[index] != value {
			t.Fatal("inconsistent copied CI aliases")
		}
		inputs[index] = value
		assigned[index] = true
	}
	for _, ok := range assigned {
		if !ok {
			t.Fatal("copy input has no source assignment")
		}
	}
	return inputs, outputs
}

func TestFullCopyPartitionAndAigerMatrix(t *testing.T) {
	for _, tc := range partitionReviewCopies() {
		t.Run(tc.name, func(t *testing.T) {
			for _, typed := range []bool{false, true} {
				for _, layout := range []bool{false, true} {
					for _, signs := range []int{0, 5, 10, 15} {
						for bad := int32(0); bad <= 5; bad++ {
							for constr := int32(0); constr <= 5-bad; constr++ {
								src := partitionReviewGraph(signs, typed, layout)
								if err := src.SetAigerProperties(bad, constr); err != nil {
									t.Fatal(err)
								}
								out := tc.copy(src)
								if out.nRegs != int32(len(tc.regs)) || Gia_ManPoNum(out) != 5 {
									t.Fatal("copy changed declared PO or retained register count")
								}
								if src.aigerBadOutputs != bad || src.nConstrs != constr {
									t.Fatal("copy changed source role metadata")
								}
								parsed := partitionReviewRoundTrip(t, out, bad, constr)
								for mask := 0; mask < 32; mask++ {
									original := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
									all, err := src.Evaluate(original)
									if err != nil {
										t.Fatal(err)
									}
									inputs := original
									switch tc.name {
									case "abstract_second", "abstract_both_reversed":
										inputs = []bool{original[0], original[1], original[2], original[4], original[3]}
									case "cut":
										inputs, all = partitionReviewLift(t, src, out, original)
									}
									want := append([]bool(nil), all[:5]...)
									for _, reg := range tc.regs {
										want = append(want, all[5+reg])
									}
									checkQuantificationOutputs(t, out, inputs, want)
									checkQuantificationOutputs(t, parsed, inputs, want)
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestFullCopyEmptyPropertyPartition(t *testing.T) {
	for _, tc := range partitionReviewCopies() {
		t.Run(tc.name, func(t *testing.T) {
			src := Gia_ManStart(1)
			r0, r1 := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, r1)
			Gia_ManAppendCo(src, r0)
			Gia_ManSetRegNum(src, 2)
			out := tc.copy(src)
			if Gia_ManPoNum(out) != 0 {
				t.Fatal("copy invented property outputs")
			}
			partitionReviewRoundTrip(t, out, 0, 0)
		})
	}
}

func TestExplicitOutputCopiesKeepDerivedPartition(t *testing.T) {
	for _, mode := range []string{"selected", "presence", "group", "literals", "vectors", "ci_map"} {
		t.Run(mode, func(t *testing.T) {
			src := partitionReviewGraph(0, false, false)
			src.nRegs = 0 // All seven COs are explicit combinational outputs in this fixture.
			if err := src.SetAigerProperties(2, 2); err != nil {
				t.Fatal(err)
			}
			selected, present, cos, lits, cis, nodes := Vec_IntAlloc(7), Vec_IntAlloc(7), Vec_IntAlloc(7), Vec_IntAlloc(7), Vec_IntAlloc(5), Vec_IntAlloc(4)
			for i := int32(0); i < Gia_ManCoNum(src); i++ {
				obj := Gia_ManCo(src, i)
				Vec_IntPush(selected, i)
				Vec_IntPush(present, 1)
				Vec_IntPush(cos, obj.id)
				Vec_IntPush(lits, Gia_ObjFaninLit0(obj, obj.id))
			}
			for i := int32(0); i < Gia_ManCiNum(src); i++ {
				Vec_IntPush(cis, Gia_ManCi(src, i).id)
			}
			for _, obj := range src.pObjs {
				if Gia_ObjIsAnd(obj) != 0 {
					Vec_IntPush(nodes, obj.id)
				}
			}
			var out *Gia_Man_t
			switch mode {
			case "selected":
				out = Gia_ManDupSelectedOutputs(src, selected)
			case "presence":
				out = Gia_ManDupOutputVec(src, present)
			case "group":
				out = Gia_ManDupOutputGroup(src, 0, Gia_ManCoNum(src))
			case "literals":
				out = Gia_ManDupDfsLitArray(src, lits)
			case "vectors":
				out = Gia_ManDupFromVecs(src, cis, nodes, cos, 0)
			case "ci_map":
				out = Gia_ManDupDfsCiMap(src, []int32{-1, -1, -1, -1, -1}, lits)
			}
			parsed := partitionReviewRoundTrip(t, out, 0, 0)
			for mask := 0; mask < 32; mask++ {
				original := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				inputs, want := partitionReviewLift(t, src, out, original)
				checkQuantificationOutputs(t, out, inputs, want)
				checkQuantificationOutputs(t, parsed, inputs, want)
			}
		})
	}
}

func TestNoMuxCopyRetainsLogicalPortOrder(t *testing.T) {
	for _, layout := range []bool{false, true} {
		t.Run(strconv.FormatBool(layout), func(t *testing.T) {
			src := partitionReviewGraph(5, true, layout)
			if err := src.SetAigerProperties(2, 2); err != nil {
				t.Fatal(err)
			}
			out := Gia_ManDupNoMuxes(src, 1)
			if out.nRegs != 2 || out.nConstrs != 2 || out.aigerBadOutputs != 2 {
				t.Fatal("lowering changed role partition")
			}
			for mask := 0; mask < 32; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				want, err := src.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}

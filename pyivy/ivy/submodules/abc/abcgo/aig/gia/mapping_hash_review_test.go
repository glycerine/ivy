package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestDupHashMappingPreservesTypedGateLogic(t *testing.T) {
	for _, kind := range []string{"and", "xor", "mux", "buffer"} {
		for signs := int32(0); signs < 16; signs++ {
			t.Run(fmt.Sprintf("%s/%d", kind, signs), func(t *testing.T) {
				g := Gia_ManStart(1)
				a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
				g.EnableMuxes()
				var root int32
				switch kind {
				case "and":
					root = Gia_ManAppendAnd(g, a^(signs&1), b^((signs>>1)&1))
				case "xor":
					root = Gia_ManAppendXorReal(g, a^(signs&1), b^((signs>>1)&1))
				case "mux":
					root = Gia_ManAppendMuxReal(g, c^((signs>>2)&1), a^(signs&1), b^((signs>>1)&1))
				case "buffer":
					root = Gia_ManAppendBuf(g, a^(signs&1))
				}
				Gia_ManAppendCo(g, root^((signs>>3)&1))
				Gia_ManAppendCo(g, root^1)
				Gia_ManSetRegNum(g, 1)
				g.VMapping = Vec_IntStart(int(g.nObjs))
				Vec_IntWriteEntry(g.VMapping, int(root>>1), int32(Vec_IntSize(g.VMapping)))
				for _, value := range []int32{3, a >> 1, b >> 1, c >> 1, root >> 1} {
					Vec_IntPush(g.VMapping, value)
				}
				q := Gia_ManDupHashMapping(g)
				if Gia_ManCiNum(q) != 3 || Gia_ManCoNum(q) != 2 || Gia_ManRegNum(q) != 1 {
					t.Fatal("terminal/register counts changed")
				}
				for mask := 0; mask < 8; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					av, bv, cv := inputs[0] != (signs&1 != 0), inputs[1] != (signs&2 != 0), inputs[2] != (signs&4 != 0)
					value := av
					switch kind {
					case "and":
						value = av && bv
					case "xor":
						value = av != bv
					case "mux":
						if !cv {
							value = bv
						}
					}
					want := []bool{value != (signs&8 != 0), !value}
					got, err := q.Evaluate(inputs)
					if err != nil || !slices.Equal(got, want) {
						t.Fatalf("inputs %v: got %v (%v), want %v", inputs, got, err, want)
					}
					if kind == "buffer" {
						if Gia_ManLutNum(q) != 0 {
							t.Fatal("forwarded buffer left a LUT record")
						}
						continue
					}
					copied := int32(Gia_ManObj(g, root>>1).Value)
					leaves, truth, ok := q.LUT(copied >> 1)
					if !ok || len(leaves) != 3 || len(truth) != 1 {
						t.Fatal("copied root has no complete LUT record")
					}
					index := 0
					for i, leaf := range leaves {
						if inputs[Gia_ObjCioId(Gia_ManObj(q, leaf))] {
							index |= 1 << uint(i)
						}
					}
					if physical := truth[0]&(uint64(1)<<uint(index)) != 0; physical != (value != (copied&1 != 0)) {
						t.Fatalf("LUT truth lost root phase at inputs %v", inputs)
					}
				}
			})
		}
	}
}

func TestDupHashMappingPreservesDependentLUTCuts(t *testing.T) {
	for signs := int32(0); signs < 32; signs++ {
		g := Gia_ManStart(1)
		a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		g.EnableMuxes()
		x := Gia_ManAppendXorReal(g, a^(signs&1), b^((signs>>1)&1))
		buffer := Gia_ManAppendBuf(g, x^((signs>>2)&1))
		mux := Gia_ManAppendMuxReal(g, buffer, c^((signs>>3)&1), d)
		root := Gia_ManAppendAnd(g, mux^((signs>>4)&1), x)
		Gia_ManAppendCo(g, root)
		Gia_ManAppendCo(g, mux^1)
		g.VMapping = Vec_IntStart(int(g.nObjs))
		for _, cut := range []struct {
			lit    int32
			leaves []int32
		}{{x, []int32{a >> 1, b >> 1}}, {mux, []int32{x >> 1, c >> 1, d >> 1}}, {root, []int32{x >> 1, mux >> 1}}} {
			Vec_IntWriteEntry(g.VMapping, int(cut.lit>>1), int32(Vec_IntSize(g.VMapping)))
			Vec_IntPush(g.VMapping, int32(len(cut.leaves)))
			for _, id := range cut.leaves {
				Vec_IntPush(g.VMapping, id)
			}
			Vec_IntPush(g.VMapping, cut.lit>>1)
		}
		q := Gia_ManDupHashMapping(g)
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			want, err := g.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			values, got, err := q.evaluateObjects(inputs)
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("signs %d inputs %v: got %v (%v), want %v", signs, inputs, got, err, want)
			}
			for id := int32(1); id < q.nObjs; id++ {
				if Gia_ObjIsLut(q, id) == 0 {
					continue
				}
				leaves, truth, ok := q.LUT(id)
				if !ok {
					t.Fatal("missing LUT truth")
				}
				index := 0
				for i, leaf := range leaves {
					if values[leaf] {
						index |= 1 << uint(i)
					}
				}
				if (truth[0]&(uint64(1)<<uint(index)) != 0) != values[id] {
					t.Fatalf("signs %d LUT %d: cut lost logic at %v", signs, id, inputs)
				}
			}
		}
	}
}

func TestDupHashMappingRetainsLastMergedCut(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y^1)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, row := range []struct {
		root   int32
		leaves []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{a >> 1, b >> 1, c >> 1}}} {
		Vec_IntWriteEntry(p.VMapping, int(row.root), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(row.leaves)))
		for _, leaf := range row.leaves {
			Vec_IntPush(p.VMapping, leaf)
		}
		Vec_IntPush(p.VMapping, row.root)
	}
	q := Gia_ManDupHashMapping(p)
	defer Gia_ManStop(q)
	root := Gia_ObjFaninId0p(q, Gia_ManCo(q, 0))
	if got := Gia_ObjLutFanins(q, root); !slices.Equal(got, []int32{a >> 1, b >> 1, c >> 1}) {
		t.Fatalf("merged cut%v want source's last cut [%d %d %d]", got, a>>1, b>>1, c>>1)
	}
	if Gia_ManLutNum(q) != 1 {
		t.Fatal("compaction should retain one live record")
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		got, err := q.Evaluate(inputs)
		value := inputs[0] && inputs[1]
		if err != nil || !slices.Equal(got, []bool{value, !value}) {
			t.Fatalf("assignment%d got%v err%v", mask, got, err)
		}
	}
}

func TestDupHashMappingRejectsCollapsedSelfCut(t *testing.T) {
	p := Gia_ManStart(16)
	defer Gia_ManStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, z := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, z)
	Gia_ManAppendCo(p, y)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, row := range []struct {
		root   int32
		leaves []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {z >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, z >> 1}}} {
		Vec_IntWriteEntry(p.VMapping, int(row.root), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(row.leaves)))
		for _, leaf := range row.leaves {
			Vec_IntPush(p.VMapping, leaf)
		}
		Vec_IntPush(p.VMapping, row.root)
	}
	q := Gia_ManDupHashMapping(p)
	defer Gia_ManStop(q)
	root := Gia_ObjFaninId0p(q, Gia_ManCo(q, 0))
	leaves := Gia_ObjLutFanins(q, root)
	for _, leaf := range leaves {
		if leaf == root {
			t.Fatalf("hashing introduced a cyclic LUT cut: root%d leaves%v", root, leaves)
		}
	}
	if !slices.Equal(leaves, []int32{a >> 1, b >> 1}) {
		t.Fatalf("last usable cut lost: %v", leaves)
	}
}

func TestDupHashMappingRejectsCollapsedForwardCut(t *testing.T) {
	p := Gia_ManStart(20)
	defer Gia_ManStop(p)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, duplicate := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	later := Gia_ManAppendAnd(p, c, d)
	y := Gia_ManAppendAnd(p, x, duplicate)
	Gia_ManAppendCo(p, y)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, row := range []struct {
		root   int32
		leaves []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {duplicate >> 1, []int32{a >> 1, b >> 1}}, {later >> 1, []int32{c >> 1, d >> 1}}, {y >> 1, []int32{a >> 1, b >> 1, later >> 1}}} {
		Vec_IntWriteEntry(p.VMapping, int(row.root), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(row.leaves)))
		for _, leaf := range row.leaves {
			Vec_IntPush(p.VMapping, leaf)
		}
		Vec_IntPush(p.VMapping, row.root)
	}
	q := Gia_ManDupHashMapping(p)
	defer Gia_ManStop(q)
	root := Gia_ObjFaninId0p(q, Gia_ManCo(q, 0))
	leaves := Gia_ObjLutFanins(q, root)
	for _, leaf := range leaves {
		if leaf >= root {
			t.Fatalf("hashing introduced a non-topological cut: root%d leaves%v", root, leaves)
		}
	}
	if !slices.Equal(leaves, []int32{a >> 1, b >> 1}) {
		t.Fatalf("last usable cut lost: %v", leaves)
	}
}

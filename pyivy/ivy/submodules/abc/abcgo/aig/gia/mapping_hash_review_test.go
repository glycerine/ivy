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

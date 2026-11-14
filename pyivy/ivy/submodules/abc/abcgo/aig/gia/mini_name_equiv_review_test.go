package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestMiniNameAndEquivalenceMappingPhases(t *testing.T) {
	var records strings.Builder
	for signs := int32(0); signs < 8; signs++ {
		for phase := int32(0); phase < 8; phase++ {
			combined, a, b := Gia_ManStart(1), Gia_ManStart(1), Gia_ManStart(1)
			a0, a1, a2 := Gia_ManAppendCi(a), Gia_ManAppendCi(a), Gia_ManAppendCi(a)
			ax := Gia_ManAppendAnd(a, a0^(signs&1), a1^((signs>>1)&1))
			ay := Gia_ManAppendAnd(a, ax, a2^((signs>>2)&1))
			ad := Gia_ManAppendAnd(a, a0^(signs&1), a1^((signs>>1)&1))
			ao, an := Gia_ManAppendCo(a, ay), Gia_ManAppendCo(a, ay^1)
			b0, b1, b2 := Gia_ManAppendCi(b), Gia_ManAppendCi(b), Gia_ManAppendCi(b)
			bx := Gia_ManAppendAnd(b, b1^((signs>>1)&1), b2^((signs>>2)&1))
			by := Gia_ManAppendAnd(b, b0^(signs&1), bx)
			unused := Gia_ManAppendAnd(b, b0, b2)
			bd := Gia_ManAppendAnd(b, b0^(signs&1), b1^((signs>>1)&1))
			bo, bn := Gia_ManAppendCo(b, by), Gia_ManAppendCo(b, by^1)
			amap := []int32{0, a0, a1, a2, -1, ax ^ (phase & 1), ay ^ ((phase >> 1) & 1), ao, an, ad ^ ((phase & 1) ^ 1)}
			bmap := []int32{0, 1, b0, b1, b2, bx, by ^ ((phase >> 2) & 1), unused, -1, bo, bn, bd}
			va, vb := Vec_IntAlloc(len(amap)), Vec_IntAlloc(len(bmap))
			for _, entry := range amap {
				Gia_ManAppendCi(combined)
				Vec_IntPush(va, entry)
			}
			for _, entry := range bmap {
				Vec_IntPush(vb, entry)
			}
			combined.StartRepresentatives()
			a.StartRepresentatives()
			// Combined IDs identify functional classes; the two ABC cones
			// associate their ANDs differently and do not share object IDs.
			Gia_ObjSetRepr(combined, 7, 5)
			Gia_ObjSetRepr(combined, 9, 4)
			Gia_ObjSetRepr(combined, 10, 4)
			Gia_ObjSetRepr(a, ad>>1, ax>>1)
			Gia_ManFillValue(a)
			Gia_ManFillValue(b)
			a.pObjs[0].Value, b.pObjs[0].Value = 0, 0
			for id := int32(1); id <= 3; id++ {
				a.pObjs[id].Value, b.pObjs[id].Value = uint32(2*id), uint32(2*id)
			}
			a.pObjs[ax>>1].Value, a.pObjs[ay>>1].Value, a.pObjs[ad>>1].Value = 8, 10, 18
			b.pObjs[bx>>1].Value, b.pObjs[by>>1].Value, b.pObjs[bd>>1].Value = 12, 14, 20
			names := Gia_ManMapMiniLut2MiniAig(combined, a, b, va, vb)
			equivs := Vec_IntArray(Gia_ManMapEquivAfterScorr(a, va))
			if names[5] != -1 || names[7] != -1 || names[8] != -1 {
				t.Fatal("unmatched class, unmapped object or omitted entry acquired a name")
			}
			if names[9] != 14 || names[10] != 16 || names[11]>>1 != 9 {
				t.Fatal("output correspondence or last-name preference changed")
			}
			if equivs[9] != 11 || equivs[4] != -1 {
				t.Fatal("equivalence did not choose the first mapped member with opposite phase")
			}
			for mask := 0; mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				av, _, err := a.evaluateObjects(inputs)
				if err != nil {
					t.Fatal(err)
				}
				bv, _, err := b.evaluateObjects(inputs)
				if err != nil {
					t.Fatal(err)
				}
				read := func(values []bool, lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
				for id, mapped := range names {
					if mapped >= 0 && (read(av, amap[mapped>>1]) != (mapped&1 != 0)) != read(bv, bmap[id]) {
						t.Fatalf("signs %d phase %d input %d name %d changes function", signs, phase, mask, id)
					}
				}
				for id, mapped := range equivs {
					if mapped >= 0 && (read(av, amap[mapped>>1]) != (mapped&1 != 0)) != read(av, amap[id]) {
						t.Fatalf("signs %d phase %d input %d equivalence %d changes function", signs, phase, mask, id)
					}
				}
			}
			fmt.Fprintf(&records, "%d %d names", signs, phase)
			for _, entry := range names {
				fmt.Fprintf(&records, " %d", entry)
			}
			records.WriteString(" equiv")
			for _, entry := range equivs {
				fmt.Fprintf(&records, " %d", entry)
			}
			records.WriteByte('\n')
		}
	}
	exe := compileGiaSourceOracle(t, "mini_name_equiv.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C mapping: %v: %s", err, want)
	}
	if records.String() != string(want) {
		t.Fatalf("Go mappings:\n%sC mappings:\n%s", records.String(), want)
	}
}

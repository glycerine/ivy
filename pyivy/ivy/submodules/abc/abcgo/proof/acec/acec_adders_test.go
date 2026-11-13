package acec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestInsertedFullAdderTruthTable(t *testing.T) {
	p := gia.Gia_ManStart(32)
	inputs := []int32{gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)}
	out := make([]int32, 2)
	Acec_InsertFadd(p, inputs, out)
	gia.Gia_ManAppendCo(p, out[0])
	gia.Gia_ManAppendCo(p, out[1])
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		values, err := p.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		sum := 0
		for _, bit := range in {
			if bit {
				sum++
			}
		}
		if values[0] != (sum&1 != 0) || values[1] != (sum&2 != 0) {
			t.Fatalf("%03b outputs %v", mask, values)
		}
	}
}

func TestInsertedAdderTreeCountsFiveBits(t *testing.T) {
	p := gia.Gia_ManStart(64)
	levels := vec.Vec_WecStart(1)
	for i := 0; i < 5; i++ {
		vec.Vec_IntPush(vec.Vec_WecEntry(levels, 0), gia.Gia_ManAppendCi(p))
	}
	roots := Acec_InsertTree(p, levels)
	for _, root := range vec.Vec_IntArray(roots) {
		gia.Gia_ManAppendCo(p, root)
	}
	for mask := 0; mask < 32; mask++ {
		in := make([]bool, 5)
		sum := 0
		for i := range in {
			in[i] = mask&(1<<i) != 0
			if in[i] {
				sum++
			}
		}
		values, err := p.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		got := 0
		for i, bit := range values {
			if bit {
				got |= 1 << i
			}
		}
		if got != sum {
			t.Fatalf("%05b got %d want %d", mask, got, sum)
		}
	}
}

func TestInsertedBoxPreservesFullAdder(t *testing.T) {
	for _, all := range []int32{0, 1} {
		src := gia.Gia_ManStart(64)
		a, b, c := gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src)
		sum := gia.Gia_ManAppendXor(src, gia.Gia_ManAppendXor(src, a, b), c)
		carry := gia.Gia_ManAppendMaj(src, a, b, c)
		gia.Gia_ManAppendCo(src, sum)
		gia.Gia_ManAppendCo(src, carry)
		leaves, roots := vec.Vec_WecStart(2), vec.Vec_WecStart(2)
		for _, lit := range []int32{a, b, c} {
			vec.Vec_WecPush(leaves, 0, lit)
		}
		vec.Vec_WecPush(roots, 0, sum)
		vec.Vec_WecPush(roots, 1, carry)
		box := &Acec_Box_t{pGia: src, vLeafLits: leaves, vRootLits: roots, vShared: leaves, vUnique: vec.Vec_WecStart(2)}
		dst := Acec_InsertBox(box, all)
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			want, err := src.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			got, err := dst.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
				t.Fatalf("all=%d mask=%03b got=%v want=%v", all, mask, got, want)
			}
		}
	}
}

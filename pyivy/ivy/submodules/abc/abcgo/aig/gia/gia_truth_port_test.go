package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"testing"
)

func truthTestLeaves(p *Gia_Man_t, n int) ([]int32, *Vec_Int_t) {
	l := make([]int32, n)
	v := vec.Vec_IntAlloc(n)
	for i := range l {
		l[i] = Gia_ManAppendCi(p)
		vec.Vec_IntPush(v, l[i]>>1)
	}
	return l, v
}
func TestGiaTruthExhaustiveSynthesis(t *testing.T) {
	for _, hash := range []int32{0, 1} {
		for bits := uint64(0); bits < 256; bits++ {
			p := Gia_ManStart(100)
			l, s := truthTestLeaves(p, 3)
			want := bits * 0x0101010101010101
			lit := Gia_Truth6ToGia(p, l, 3, want, hash)
			Gia_ManAppendCo(p, lit)
			if got := Gia_LutComputeTruth6Simple(p, 0); got != want {
				t.Fatalf("hash=%d truth=%x got=%x", hash, want, got)
			}
			temp := vec.Vec_WrdStart(int(p.nObjs))
			if got := Gia_ObjComputeTruth6(p, lit>>1, s, temp); got^(uint64(0)-uint64(lit&1)) != want {
				t.Fatal("cut truth mismatch")
			}
			root := Gia_ManCo(p, 0)
			if got := Gia_ObjComputeTruthTable(p, root)[0]; got != want {
				t.Fatal("full truth mismatch")
			}
			Gia_ObjComputeTruthTableStop(p)
			if got := Gia_ObjComputeTruthTable6(p, root, s, vec.Vec_WrdAlloc(10)); got != want {
				t.Fatal("compact truth mismatch")
			}
			Gia_ObjComputeTruthTableStart(p, 8)
			if Gia_ObjIsAnd(p.pObjs[lit>>1]) != 0 {
				a := Gia_ObjComputeTruthTableCut(p, p.pObjs[lit>>1], s)
				for _, x := range a {
					if x^(uint64(0)-uint64(lit&1)) != want {
						t.Fatal("multiword cut mismatch")
					}
				}
			}
			Gia_ObjComputeTruthTableStop(p)
		}
	}
}
func TestGiaTruthRandomSixAndMultiword(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for j := 0; j < 80; j++ {
		p := Gia_ManStart(1000)
		l, _ := truthTestLeaves(p, 6)
		want := r.Uint64()
		lit := Gia_Truth6ToGia(p, l, 6, want, 1)
		Gia_ManAppendCo(p, lit)
		if got := Gia_LutComputeTruth6Simple(p, 0); got != want {
			t.Fatalf("got %x want %x", got, want)
		}
	}
	p := Gia_ManStart(100)
	l, _ := truthTestLeaves(p, 9)
	lit := l[0]
	for _, b := range l[1:] {
		lit = giaTruthAnd(p, lit, b, true)
	}
	Gia_ManAppendCo(p, lit)
	got := Gia_ObjComputeTruthTable(p, Gia_ManCo(p, 0))
	if len(got) != 8 {
		t.Fatal(len(got))
	}
	for i, w := range got {
		want := uint64(0)
		if i == 7 {
			want = 1 << 63
		}
		if w != want {
			t.Fatal(i, w)
		}
	}
}
func TestGiaTruthNpnClassesAndRegisters(t *testing.T) {
	p := Gia_ManStart(100)
	l, _ := truthTestLeaves(p, 4)
	and := giaTruthAnd(p, l[0], l[1], true)
	or := giaTruthAnd(p, l[2]^1, l[3]^1, true) ^ 1
	for _, v := range []int32{0, 1, l[0], l[2] ^ 1, and, or} {
		Gia_ManAppendCo(p, v)
	}
	var eq *Vec_Ptr_t
	q := Gia_ManIsoNpnReduce(p, &eq, 0)
	if Gia_ManPoNum(q) != 3 || Gia_ManPiNum(q) != 4 {
		t.Fatal("bad reduction shape")
	}
	for i := 0; i < 3; i++ {
		v := vec.Vec_PtrEntry(eq, i).(*Vec_Int_t)
		a := vec.Vec_IntArray(v)
		if len(a) != 2 || a[0] != int32(2*i) || a[1] != int32(2*i+1) {
			t.Fatal(a)
		}
	}
	p = Gia_ManStart(100)
	l, _ = truthTestLeaves(p, 3)
	Gia_ManAppendCo(p, l[1])
	Gia_ManAppendCo(p, l[2])
	Gia_ManAppendCo(p, l[0])
	p.nRegs = 2
	q = Gia_ManIsoNpnReduce(p, nil, 0)
	if q.nRegs != 2 || Gia_ManPiNum(q) != 1 || Gia_ManPoNum(q) != 1 {
		t.Fatalf("register closure lost: %+v", q)
	}
}
func TestGiaTruthLutMapAndCofactors(t *testing.T) {
	p := Gia_ManStart(20)
	l, s := truthTestLeaves(p, 3)
	lit := giaTruthAnd(p, l[0], l[2]^1, true)
	Gia_ManAppendCo(p, lit)
	m := vec.Vec_IntStart(3)
	vec.Vec_IntWriteEntry(m, 0, 2)
	vec.Vec_IntWriteEntry(m, 1, 1)
	vec.Vec_IntWriteEntry(m, 2, 0)
	want := s_Truth6[2] &^ s_Truth6[0]
	if Gia_LutComputeTruth6Map(p, 0, m) != want {
		t.Fatal("mapped variable order")
	}
	p.VMapping = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(p.VMapping, int(lit>>1), int32(p.nObjs))
	vec.Vec_IntPush(p.VMapping, 3)
	for _, id := range vec.Vec_IntArray(s) {
		vec.Vec_IntPush(p.VMapping, id)
	}
	temp := vec.Vec_WrdStart(int(p.nObjs))
	if Gia_LutComputeTruth6(p, lit>>1, temp) != s_Truth6[0]&^s_Truth6[2] {
		t.Fatal("LUT truth")
	}
	for v := int32(0); v < 5; v++ {
		x := uint32(s_Truth6[v])
		if Abc_Tt5HasVar(x, v) != 1 || Abc_Tt5Cofactor0(x, v) != 0 || Abc_Tt5Cofactor1(x, v) != ^uint32(0) {
			t.Fatal(v)
		}
	}
}

func TestGiaTruthWorkspaceMetadataAndUnrelatedCis(t *testing.T) {
	p := Gia_ManStart(100)
	l, s := truthTestLeaves(p, 3)
	a := giaTruthAnd(p, l[0], l[1], true)
	b := giaTruthAnd(p, a, l[2], true)
	Gia_ManAppendCo(p, b)
	for i := 0; i < 5; i++ {
		Gia_ManAppendCi(p)
	}
	if got := Gia_LutComputeTruth6Simple(p, 0); got != s_Truth6[0]&s_Truth6[1]&s_Truth6[2] {
		t.Fatal("unrelated CIs changed a six-input cone")
	}
	Gia_ObjComputeTruthTableStart(p, 8)
	got := Gia_ObjComputeTruthTableCut(p, p.pObjs[b>>1], s)
	if len(got) != 4 {
		t.Fatal(len(got))
	}
	ids := vec.Vec_IntArray(p.vTtNodes)
	if len(ids) != 3 || ids[0] != -1 || ids[1] != a>>1 || ids[2] != b>>1 {
		t.Fatal("postorder cut metadata", ids)
	}
	for _, v := range vec.Vec_IntArray(p.vTtNums) {
		if v != giaTruthUnnumbered {
			t.Fatal("cut numbering leaked")
		}
	}
	Gia_ObjComputeTruthTableStop(p)
	Gia_ObjComputeTruthTable(p, Gia_ManCo(p, 0))
	ids = vec.Vec_IntArray(p.vTtNodes)
	for i, id := range ids {
		if vec.Vec_IntEntry(p.vTtNums, int(id)) != int32(i) {
			t.Fatal("inconsistent full-table numbering")
		}
		if Gla_ObjTruthNode(p, p.pObjs[id])[0] == 0 {
			t.Fatal("lost internal table")
		}
	}
	// The elementary-table helpers preserve their six-entry minimum contract.
	q := Gia_ManStart(10)
	Gia_ObjComputeTruthTableStart(q, 0)
	if Gla_ObjTruthElem(q, 5)[0] != s_Truth6[5] {
		t.Fatal("elementary truth table missing")
	}
}

package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestJfManGenCnfTruthTables(t *testing.T) {
	for truth := 0; truth < 256; truth++ {
		var bits uint64
		for i := 0; i < 64; i++ {
			if truth&(1<<uint(i%8)) != 0 {
				bits |= 1 << uint(i)
			}
		}
		leaves := vec.Vec_IntAlloc(3)
		for i := int32(0); i < 3; i++ {
			vec.Vec_IntPush(leaves, 2*i)
		}
		lits, clauses, cover := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
		Jf_ManGenCnf(bits, 6, leaves, lits, clauses, cover)
		ls, cs := vec.Vec_IntArray(lits), vec.Vec_IntArray(clauses)
		for assignment := 0; assignment < 16; assignment++ {
			satisfied := true
			for i, start := range cs {
				end := len(ls)
				if i+1 < len(cs) {
					end = int(cs[i+1])
				}
				clause := false
				for _, lit := range ls[start:end] {
					clause = clause || ((assignment>>uint(lit>>1))&1 != int(lit&1))
				}
				satisfied = satisfied && clause
			}
			expected := ((assignment >> 3) & 1) == ((truth >> uint(assignment&7)) & 1)
			if satisfied != expected {
				t.Fatalf("truth=%02x assignment=%x literals=%v clauses=%v", truth, assignment, ls, cs)
			}
		}
	}
}

func TestJfManGenCnfSixLeavesAndAppends(t *testing.T) {
	const truth uint64 = 0x96e8174b3cd25a81
	leaves := vec.Vec_IntAlloc(6)
	for i := int32(0); i < 6; i++ {
		vec.Vec_IntPush(leaves, 2*i+(i&1))
	}
	lits, clauses, cover := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	vec.Vec_IntPush(lits, 12)
	vec.Vec_IntPush(clauses, 0)
	Jf_ManGenCnf(truth, 13, leaves, lits, clauses, cover)
	ls, cs := vec.Vec_IntArray(lits), vec.Vec_IntArray(clauses)
	for assignment := 0; assignment < 128; assignment++ {
		satisfied := true
		for i := 1; i < len(cs); i++ {
			end := len(ls)
			if i+1 < len(cs) {
				end = int(cs[i+1])
			}
			clause := false
			for _, lit := range ls[cs[i]:end] {
				clause = clause || ((assignment>>uint(lit>>1))&1 != int(lit&1))
			}
			satisfied = satisfied && clause
		}
		input := 0
		for i := 0; i < 6; i++ {
			input |= (((assignment >> uint(i)) & 1) ^ (i & 1)) << uint(i)
		}
		output := ((assignment >> 6) & 1) ^ 1
		expected := output == int((truth>>uint(input))&1)
		if satisfied != expected {
			t.Fatalf("assignment=%x input=%x", assignment, input)
		}
	}
	if cs[0] != 0 || ls[0] != 12 || cs[1] != 1 {
		t.Fatalf("prior clauses overwritten: %v %v", cs, ls)
	}
}

func TestJfMappedCNF(t *testing.T) {
	p := Gia_ManStart(10)
	a := Gia_ManAppendCi(p)
	b := Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	newVectors := func() (*Vec_Int_t, *Vec_Int_t) {
		lits, offsets := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
		for _, v := range []int32{x ^ 1, a, x, b ^ 1} {
			vec.Vec_IntPush(lits, v)
		}
		for _, v := range []int32{0, 2} {
			vec.Vec_IntPush(offsets, v)
		}
		return lits, offsets
	}
	lits, offs := newVectors()
	direct := Jf_ManCreateCnf(p, lits, offs)
	if direct.Variables != int(Gia_ManObjNum(p)) || len(direct.Clauses) != 2 || direct.Obj2Clause[x>>1] != 0 || direct.Obj2Count[x>>1] != 2 || direct.Obj2Clause[a>>1] != -1 {
		t.Fatalf("direct CNF: %+v", direct)
	}
	if vec.Vec_IntSize(lits) != 0 {
		t.Fatal("literal vector not released")
	}
	if err := direct.Validate(); err != nil {
		t.Fatal(err)
	}
	lits, offs = newVectors()
	compact := Jf_ManCreateCnfRemap(p, lits, offs, 1)
	if compact.Variables != 4 || len(compact.Clauses) != 3 {
		t.Fatalf("compact CNF: %+v", compact)
	}
	outID := Gia_ObjId(p, Gia_ManPo(p, 0))
	if compact.VarNums[outID] != 0 || compact.VarNums[x>>1] != 1 || compact.VarNums[b>>1] != 2 || compact.VarNums[a>>1] != 3 || compact.VarNums[0] != -1 {
		t.Fatalf("mapping: %v", compact.VarNums)
	}
	if got := compact.Clauses[2]; len(got) != 1 || got[0] != 0 {
		t.Fatalf("output clause: %v", got)
	}
	if err := compact.Validate(); err != nil {
		t.Fatal(err)
	}
}

package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDemiterPartitions(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c^1)
	Gia_ManAppendCo(p, y)
	out := Gia_ManDupDemiter(p, 0)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, e := out.Evaluate(in)
		all := true
		for _, v := range got {
			all = all && v
		}
		if e != nil || all != (in[0] && in[1] && !in[2]) {
			t.Fatalf("demiter %v %v", got, e)
		}
	}
	nodes := permutation(y>>1, x>>1)
	Gia_ManDupDemiterOrderXors2(p, nodes)
	if vec.Vec_IntEntry(nodes, 0) != x>>1 {
		t.Fatal("support order")
	}
	Gia_ManDupDemiterOrderXors(p, nodes)
	Gia_ManCleanMark01(p)
	Gia_ManSetMark0Dfs_rec(p, y>>1)
	if Gia_ManCountMark0Dfs(p, y>>1) != 5 {
		t.Fatal("marked cone count")
	}
}
func TestDemiterSplitOutputOrders(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, b)
	Gia_ManAppendCo(p, a^1)
	Gia_ManAppendCo(p, b^1)
	for _, halves := range []bool{false, true} {
		var x, y *Gia_Man_t
		if halves {
			Gia_ManDemiterTwoWords(p, &x, &y)
		} else {
			Gia_ManDemiterDual(p, &x, &y)
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			first, e := x.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			second, e := y.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			if halves {
				if first[0] != in[0] || first[1] != in[1] || second[0] == in[0] || second[1] == in[1] {
					t.Fatal("halves")
				}
			} else {
				if first[0] != in[0] || first[1] == in[0] || second[0] != in[1] || second[1] == in[1] {
					t.Fatal("interleaved")
				}
			}
		}
	}
}
func TestDupHalfSeq(t *testing.T) {
	p := Gia_ManStart(12)
	a, r0, r1 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, r0)
	Gia_ManAppendCo(p, r1)
	Gia_ManSetRegNum(p, 2)
	left, right := Gia_ManDupHalfSeq(p, 0), Gia_ManDupHalfSeq(p, 1)
	if Gia_ManPiNum(left) != 2 || Gia_ManRegNum(left) != 1 || Gia_ManPiNum(right) != 2 || Gia_ManRegNum(right) != 1 {
		t.Fatal("half interface")
	}
	got, e := left.Evaluate([]bool{true, false, true})
	if e != nil || !got[0] || !got[1] {
		t.Fatalf("left %v %v", got, e)
	}
	got, e = right.Evaluate([]bool{true, false, true})
	if e != nil || !got[0] || !got[1] {
		t.Fatalf("right %v %v", got, e)
	}
}

func TestDemiterToDualMatchesXor(t *testing.T) {
	p := Gia_ManStart(20)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManHashAlloc(p)
	x := Gia_ManHashXor(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManHashStop(p)
	out := Gia_ManDemiterToDual(p)
	if out == nil || Gia_ManPoNum(out) != 2 {
		t.Fatal("dual interface")
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		got, e := out.Evaluate(in)
		if e != nil || (got[0] != got[1]) != (in[0] != in[1]) {
			t.Fatalf("dual xor %v %v", got, e)
		}
	}
}

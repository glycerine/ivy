package extra

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func bddFromTruth(p *Abc_BddMan, truth uint32, v int32) int32 {
	if v == p.nVars {
		return int32(truth & 1)
	}
	var low, high uint32
	for m := uint32(0); m < 1<<uint(p.nVars-v-1); m++ {
		low |= (truth >> (2 * m) & 1) << m
		high |= (truth >> (2*m + 1) & 1) << m
	}
	a := bddFromTruth(p, low, v+1)
	b := bddFromTruth(p, high, v+1)
	return Abc_BddUniqueCreate(p, v, b, a)
}
func bddEval(p *Abc_BddMan, lit int32, assignment uint32) bool {
	for lit >= 2 {
		v := Abc_BddVar(p, lit)
		if assignment>>v&1 != 0 {
			lit = Abc_BddThen(p, lit)
		} else {
			lit = Abc_BddElse(p, lit)
		}
	}
	return lit == 1
}
func TestBddOriginalExhaustive(t *testing.T) {
	p := Abc_BddManAlloc(3, 10000)
	var roots [256]int32
	for i := range roots {
		roots[i] = bddFromTruth(p, uint32(i), 0)
	}
	hash := sha256.New()
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			x, y := Abc_BddAnd(p, roots[a], roots[b]), Abc_BddOr(p, roots[a], roots[b])
			if x != roots[a&b] || y != roots[a|b] {
				t.Fatalf("noncanonical operation %d %d", a, b)
			}
			for _, v := range []int32{x, y, Abc_BddCountNodes(p, x), Abc_BddCountNodes(p, y)} {
				var data [4]byte
				binary.LittleEndian.PutUint32(data[:], uint32(v))
				hash.Write(data[:])
			}
		}
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "ddf6c982094073310b2ad4621532ef2cd85979e691ac3b634bcfbccb86ada9d5" {
		t.Fatalf("original C nodes/count mismatch: %s", got)
	}
	for truth, root := range roots {
		for assignment := uint32(0); assignment < 8; assignment++ {
			if bddEval(p, root, assignment) != (uint32(truth)>>assignment&1 != 0) {
				t.Fatal("evaluation mismatch")
			}
		}
	}
	for _, node := range p.nodes {
		if node.mark {
			t.Fatal("count leaked a mark")
		}
		if node.otherwise&1 != 0 {
			t.Fatal("stored else complemented")
		}
	}
}
func TestBddSharedCountAndPrinter(t *testing.T) {
	p := Abc_BddManAlloc(4, 100)
	a := Abc_BddAnd(p, 2, 4)
	b := Abc_BddAnd(p, 6, 8)
	root := Abc_BddOr(p, a, b)
	roots := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(roots, root)
	vec.Vec_IntPush(roots, root^1)
	vec.Vec_IntPush(roots, root)
	count := Abc_BddCountNodes(p, root)
	if Abc_BddCountNodesArray(p, roots) != count || Abc_BddCountNodesArray2(p, roots) != 3*count {
		t.Fatal("shared versus independent root count")
	}
	var out bytes.Buffer
	path := []int32{-1, -1, -1, -1}
	bddPrint(&out, p, a, path)
	if out.String() != "+0+1 " {
		t.Fatalf("printed %q", out.String())
	}
	for _, v := range path {
		if v != -1 {
			t.Fatal("printer retained path mark")
		}
	}
	if Abc_BddAnd(p, root, root^1) != 0 || Abc_BddOr(p, root, root^1) != 1 {
		t.Fatal("complement reduction")
	}
}

type bddTestGraph struct{}

func (bddTestGraph) ObjectCount() int        { return 5 }
func (bddTestGraph) InputObjectIDs() []int32 { return []int32{1, 2} }
func (bddTestGraph) AndFaninLiterals(id int32) (int32, int32, bool) {
	switch id {
	case 3:
		return 2, 5, true
	case 4:
		return 6, 1, true
	}
	return 0, 0, false
}
func (bddTestGraph) OutputFaninLiterals() []int32 { return []int32{8, 9} }
func TestBddGiaBoundary(t *testing.T) {
	p, roots := bddGiaRoots(bddTestGraph{}, 100)
	for assignment := uint32(0); assignment < 4; assignment++ {
		want := assignment&1 != 0 && assignment&2 == 0
		if bddEval(p, vec.Vec_IntEntry(roots, 0), assignment) != want || bddEval(p, vec.Vec_IntEntry(roots, 1), assignment) == want {
			t.Fatal("AIG complements lost")
		}
	}
}
func TestBddLimits(t *testing.T) {
	for _, check := range []func(){func() { Abc_BddManAlloc(255, 1000) }, func() { Abc_BddManAlloc(2, 2) }, func() { p := Abc_BddManAlloc(2, 3); Abc_BddAnd(p, 2, 4) }, func() { p := Abc_BddManAlloc(2, 4); Abc_BddAnd(p, 2, -1) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid BDD accepted")
				}
			}()
			check()
		}()
	}
	p := Abc_BddManAlloc(0, 1)
	if Abc_BddAnd(p, 1, 1) != 1 {
		t.Fatal("constant manager")
	}
	Abc_BddManFree(p)
	Abc_BddManFree(p)
}

package bdc

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func bdcStretch(t uint32, n int32) uint32 {
	var out uint32
	mask := (1 << uint(n)) - 1
	for a := 0; a < 32; a++ {
		if t&(1<<uint(a&mask)) != 0 {
			out |= 1 << uint(a)
		}
	}
	return out
}
func bdcEvaluate(p *Bdc_Man_t, assignment int, divisors [][]uint32) bool {
	values := make([]bool, len(p.nodes))
	read := func(f *Bdc_Fun_t) bool {
		value := values[Bdc_Regular(f).id]
		return value != (Bdc_IsComplement(f) != 0)
	}
	for _, node := range p.nodes {
		switch node.Type {
		case BDC_TYPE_CONST1:
			values[node.id] = true
		case BDC_TYPE_PI:
			if node.id <= p.nVars {
				values[node.id] = assignment&(1<<uint(node.id-1)) != 0
			} else {
				values[node.id] = divisors[node.id-p.nVars-1][assignment>>5]&(1<<uint(assignment&31)) != 0
			}
		case BDC_TYPE_AND:
			if Bdc_Regular(node.Fan0).id >= node.id || Bdc_Regular(node.Fan1).id >= node.id {
				panic("non-topological BDC graph")
			}
			values[node.id] = read(node.Fan0) && read(node.Fan1)
		}
	}
	return read(p.root)
}
func checkBdc(t *testing.T, p *Bdc_Man_t, truth, care []uint32, n int32, divisors [][]uint32) {
	t.Helper()
	if result := Bdc_ManDecompose(p, truth, care, n, divisors, 1000); result < 0 {
		t.Fatal("unexpected budget failure", n)
	}
	for a := 0; a < 1<<uint(n); a++ {
		if care != nil && care[a>>5]&(1<<uint(a&31)) == 0 {
			continue
		}
		want := truth[a>>5]&(1<<uint(a&31)) != 0
		if got := bdcEvaluate(p, a, divisors); got != want {
			t.Fatalf("truth mismatch n=%d assignment=%d", n, a)
		}
	}
}
func TestAllThreeVariableFunctions(t *testing.T) {
	p := Bdc_ManAlloc(&Bdc_Par_t{NVarsMax: 8})
	for f := 0; f < 256; f++ {
		checkBdc(t, p, []uint32{bdcStretch(uint32(f), 3)}, nil, 3, nil)
	}
}
func TestRandomFunctionsAndCare(t *testing.T) {
	p := Bdc_ManAlloc(&Bdc_Par_t{NVarsMax: 9})
	r := rand.New(rand.NewSource(919))
	for n := int32(0); n <= 8; n++ {
		for trial := 0; trial < 20; trial++ {
			truth, care := make([]uint32, truthWords(n)), make([]uint32, truthWords(n))
			for i := range truth {
				truth[i], care[i] = r.Uint32(), r.Uint32()
			}
			if n < 5 {
				truth[0], care[0] = bdcStretch(truth[0], n), bdcStretch(care[0], n)
			}
			checkBdc(t, p, truth, nil, n, nil)
			checkBdc(t, p, truth, care, n, nil)
		}
	}
}
func TestBudgetDivisorsAndCopies(t *testing.T) {
	p := Bdc_ManAlloc(&Bdc_Par_t{NVarsMax: 6})
	truth := []uint32{0x88888888}
	if Bdc_ManDecompose(p, truth, nil, 2, nil, 0) != -1 || Bdc_ManRoot(p) != nil {
		t.Fatal("budget limit ignored")
	}
	checkBdc(t, p, truth, nil, 2, [][]uint32{truth})
	if Bdc_Regular(p.root).Type != BDC_TYPE_PI || p.nNodesNew != 0 {
		t.Fatal("divisor not reused")
	}
	node := Bdc_ManFunc(p, 1)
	Bdc_FuncSetCopy(node, "edge")
	if Bdc_FuncCopy(Bdc_Not(node)) != "edge" {
		t.Fatal("opaque copy")
	}
	Bdc_FuncSetCopyInt(node, 12)
	if Bdc_FunObjCopy(Bdc_Not(node)) != 13 || Bdc_Not(Bdc_Not(node)) != node {
		t.Fatal("complement/copy")
	}
	zeroCare := []uint32{0}
	if Bdc_ManDecompose(p, truth, zeroCare, 2, nil, 0) != 0 || Bdc_IsComplement(p.root) != 1 {
		t.Fatal("empty care")
	}
}

func TestCopyStorageReplacesOtherRepresentation(t *testing.T) {
	p := Bdc_ManAlloc(&Bdc_Par_t{NVarsMax: 3})
	Bdc_ManDecompose(p, []uint32{0xaaaaaaaa}, nil, 1, nil, 0)
	f := Bdc_ManFunc(p, 1)
	Bdc_FuncSetCopyInt(f, 42)
	Bdc_FuncSetCopy(f, "edge")
	if Bdc_FuncCopyInt(f) != 0 || Bdc_FuncCopy(f) != "edge" {
		t.Fatal("stale integer copy")
	}
	Bdc_FuncSetCopyInt(Bdc_Not(f), 12)
	if Bdc_FuncCopy(f) != nil || Bdc_FuncCopyInt(f) != 12 {
		t.Fatal("stale opaque copy")
	}
}

// Golden counts were produced by compiling the unchanged original bdcCore.c,
// bdcDec.c, bdcTable.c and kitTruth.c, then decomposing all 256 three-variable
// functions with maxNodes=1000 and no care/divisors. No C compiler is needed to
// run the Go regression test.
func TestOriginalCNodeCounts(t *testing.T) {
	const counts = "0,2,2,1,2,1,4,2,2,4,1,2,1,2,2,0,2,1,4,2,4,2,8,4,7,5,5,3,5,3,6,2,2,4,1,2,7,5,5,3,4,8,2,4,5,6,3,2,1,2,2,0,5,3,6,2,5,6,3,2,3,5,5,1,2,4,7,5,1,2,5,3,4,8,5,6,2,4,3,2,1,2,5,3,2,0,6,2,5,6,3,5,3,2,5,1,4,8,5,6,5,6,3,5,8,6,6,8,6,8,5,4,2,4,3,2,3,2,5,1,6,8,5,4,5,4,7,2,2,7,4,5,4,5,8,6,1,5,2,3,2,3,4,2,4,5,8,6,8,6,6,8,5,3,6,5,6,5,8,4,1,5,2,3,5,3,6,5,2,6,0,2,3,5,2,1,2,3,4,2,6,5,8,4,3,5,2,1,5,7,4,2,1,5,5,3,2,3,6,5,2,6,3,5,0,2,2,1,2,3,6,5,4,2,8,4,3,5,5,7,2,1,4,2,2,6,3,5,3,5,5,7,4,8,2,4,2,4,1,2,0,2,2,1,2,1,4,2,2,4,1,2,1,2,2,0"
	p := Bdc_ManAlloc(&Bdc_Par_t{NVarsMax: 8})
	for f, text := range strings.Split(counts, ",") {
		want, err := strconv.Atoi(text)
		if err != nil {
			t.Fatal(err)
		}
		got := Bdc_ManDecompose(p, []uint32{bdcStretch(uint32(f), 3)}, nil, 3, nil, 1000)
		if int(got) != want {
			t.Fatalf("function %d: Go nodes %d, C nodes %d", f, got, want)
		}
	}
}

func TestNewestDivisorReuse(t *testing.T) {
	p := Bdc_ManAlloc(&Bdc_Par_t{NVarsMax: 6})
	truth := []uint32{0x88888888}
	if Bdc_ManDecompose(p, truth, nil, 2, [][]uint32{truth, truth}, 0) != 0 {
		t.Fatal("divisor reuse failed")
	}
	if Bdc_Regular(p.root).id != 4 {
		t.Fatal("table lookup did not prefer newest candidate")
	}
}

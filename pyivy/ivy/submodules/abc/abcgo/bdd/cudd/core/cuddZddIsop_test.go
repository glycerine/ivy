package core

import "testing"

func bddFromTruth(dd *DdManager, table uint8, variables int) *DdNode {
	var build func(int, uint) *DdNode
	build = func(index int, assignment uint) *DdNode {
		if index == variables {
			if table&(1<<assignment) != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		return Cudd_bddIte(dd, Cudd_bddIthVar(dd, index), build(index+1, assignment|1<<index), build(index+1, assignment))
	}
	return build(0, 0)
}
func bddTruthTable(t *testing.T, dd *DdManager, f *DdNode) uint8 {
	t.Helper()
	var table uint8
	for assignment := 0; assignment < 8; assignment++ {
		if bddTruth(t, dd, f, assignment) {
			table |= 1 << assignment
		}
	}
	return table
}
func coverCubeTruth(cube uint) uint8 {
	var table uint8
	for assignment := uint(0); assignment < 8; assignment++ {
		match := true
		for index := uint(0); index < 3; index++ {
			positive, negative := cube&(1<<(2*index)) != 0, cube&(1<<(2*index+1)) != 0
			if positive && assignment&(1<<index) == 0 || negative && assignment&(1<<index) != 0 {
				match = false
			}
		}
		if match {
			table |= 1 << assignment
		}
	}
	return table
}
func nativeCoverTruth(cover uint64) uint8 {
	var table uint8
	for cube := uint(0); cube < 64; cube++ {
		if cover&(uint64(1)<<cube) != 0 {
			table |= coverCubeTruth(cube)
		}
	}
	return table
}
func intervalTables(code int) (lower, upper uint8) {
	for assignment := 0; assignment < 8; assignment++ {
		state := code % 3
		code /= 3
		if state > 0 {
			upper |= 1 << assignment
		}
		if state == 2 {
			lower |= 1 << assignment
		}
	}
	return
}
func TestISOPAllThreeVariableIntervals(t *testing.T) {
	dd := Cudd_Init(3, 6, 256, 256, 0)
	functions := make([]*DdNode, 256)
	for i := range functions {
		functions[i] = bddFromTruth(dd, uint8(i), 3)
	}
	for code := 0; code < 6561; code++ {
		lower, upper := intervalTables(code)
		var cover *DdNode
		chosen := Cudd_zddIsop(dd, functions[lower], functions[upper], &cover)
		if chosen == nil || cover == nil {
			t.Fatal("ISOP failed", code, lower, upper)
		}
		table := bddTruthTable(t, dd, chosen)
		if lower&^table != 0 || table&^upper != 0 {
			t.Fatal("bounds violated", code, lower, table, upper)
		}
		coverBits := coverBits64(dd, cover)
		if nativeCoverTruth(coverBits) != table || Cudd_MakeBddFromZddCover(dd, cover) != chosen {
			t.Fatal("cover/function mismatch", code)
		}
		if Cudd_bddIsop(dd, functions[lower], functions[upper]) != chosen {
			t.Fatal("BDD-only recursion diverged", code)
		}
		for cube := uint(0); cube < 64; cube++ {
			if coverBits&(uint64(1)<<cube) == 0 {
				continue
			}
			if cube&3 == 3 || (cube>>2)&3 == 3 || (cube>>4)&3 == 3 {
				t.Fatal("inconsistent cube", code, cube)
			}
			without := nativeCoverTruth(coverBits &^ (uint64(1) << cube))
			if lower&^without == 0 {
				t.Fatal("redundant cube", code, cube)
			}
		}
	}
}
func TestZDDLogicalComplementPipeline(t *testing.T) {
	dd := Cudd_Init(3, 6, 256, 256, 0)
	a, notA := Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 1)
	if Cudd_zddComplement(dd, a) != notA || Cudd_zddComplement(dd, dd.One) != dd.Zero || Cudd_zddComplement(dd, dd.Zero) != dd.One {
		t.Fatal("logical complement terminal/literal cases")
	}
	if Cudd_zddDiff(dd.Univ[0].manager, dd.Univ[0], a) == notA {
		t.Fatal("fixture did not distinguish logical and set complement")
	}
	redundant := Cudd_zddUnion(dd, a, Cudd_zddChange(dd, a, 2))
	twice := Cudd_zddComplement(dd, Cudd_zddComplement(dd, redundant))
	if Cudd_MakeBddFromZddCover(dd, twice) != Cudd_MakeBddFromZddCover(dd, redundant) {
		t.Fatal("double logical complement changed function")
	}
	if twice == redundant {
		t.Fatal("redundant input unexpectedly retained as an ISOP")
	}
}
func TestISOPInvalidDomainsAndBDDOnlyIsolation(t *testing.T) {
	dd := NewBDDManager(3)
	lower := dd.Vars[0]
	upper := Cudd_bddOr(dd, lower, dd.Vars[1])
	before := len(dd.zddUnique)
	if Cudd_bddIsop(dd, lower, upper) == nil || dd.ZddSize != 0 || len(dd.zddUnique) != before {
		t.Fatal("BDD-only ISOP created ZDDs")
	}
	sentinel := dd.One
	out := sentinel
	if Cudd_zddIsop(dd, dd.One, dd.LogicZero, &out) != nil || out != sentinel {
		t.Fatal("invalid interval wrote output")
	}
	if Cudd_zddIsop(dd, lower, upper, nil) != nil || Cudd_bddIsop(dd, lower, NewBDDManager(0).One) != nil {
		t.Fatal("invalid domain accepted")
	}
	inconsistent := Cudd_zddChange(dd, Cudd_zddChange(dd, dd.One, 0), 1)
	if Cudd_MakeBddFromZddCover(dd, inconsistent) != nil || Cudd_zddComplement(dd, inconsistent) != nil {
		t.Fatal("inconsistent cover accepted")
	}
}

func TestISOPOriginalCUDDOracle(t *testing.T) {
	for _, tc := range []struct {
		name               string
		bddOrder, zddOrder bool
	}{
		{"identity", false, false}, {"BDD order", true, false}, {"ZDD order", false, true}, {"independent BDD and ZDD orders", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dd *DdManager
			if tc.zddOrder {
				dd = NewBDDManager(3)
				dd.ZddSize = 6
				dd.PermZ = []int{2, 3, 4, 5, 0, 1}
				dd.InvPermZ = []int{4, 5, 0, 1, 2, 3}
				if InitializeZDDUniverse(dd) != 1 {
					t.Fatal("invalid fixture order")
				}
			} else {
				dd = Cudd_Init(3, 6, 256, 256, 0)
			}
			if tc.bddOrder && Cudd_ShuffleHeap(dd, []int{2, 0, 1}) != 1 {
				t.Fatal("BDD fixture shuffle failed")
			}
			checkISOPOriginalOracle(t, dd, tc.bddOrder)
		})
	}
}
func checkISOPOriginalOracle(t *testing.T, dd *DdManager, reorderedBDD bool) {
	t.Helper()
	functions := make([]*DdNode, 256)
	for i := range functions {
		functions[i] = bddFromTruth(dd, uint8(i), 3)
	}
	hash := uint64(14695981039346656037)
	each := [6]uint64{hash, hash, hash, hash, hash, hash}
	for code := 0; code < 6561; code++ {
		lower, upper := intervalTables(code)
		var cover *DdNode
		chosen := Cudd_zddIsop(dd, functions[lower], functions[upper], &cover)
		if chosen == nil || cover == nil {
			t.Fatal("ISOP failed", code)
		}
		converted := Cudd_MakeBddFromZddCover(dd, cover)
		bddOnly := Cudd_bddIsop(dd, functions[lower], functions[upper])
		complemented := Cudd_zddComplement(dd, cover)
		complementBDD := Cudd_MakeBddFromZddCover(dd, complemented)
		if converted == nil || bddOnly == nil || complemented == nil || complementBDD == nil {
			t.Fatal("conversion/complement failed", code)
		}
		values := [6]uint64{uint64(bddTruthTable(t, dd, chosen)), coverBits64(dd, cover), uint64(bddTruthTable(t, dd, converted)), uint64(bddTruthTable(t, dd, bddOnly)), coverBits64(dd, complemented), uint64(bddTruthTable(t, dd, complementBDD))}
		for i, value := range values {
			hash ^= value
			hash *= 1099511628211
			each[i] ^= value
			each[i] *= 1099511628211
		}
	}
	original := uint64(0x23c2e19ddcb24343)
	expected := [6]uint64{0x51ce9de0d876d35f, 0x006b0a50abaed6be, 0x51ce9de0d876d35f, 0x51ce9de0d876d35f, 0x58fae6796c5b0469, 0xf96686de28bbe666}
	if reorderedBDD {
		original = 0xf1ff6f7e29aa5823
		expected = [6]uint64{0x2ae4db1389fd602f, 0x4f97e0598223ed7a, 0x2ae4db1389fd602f, 0x2ae4db1389fd602f, 0xf60df4c8fff5f389, 0x8551c85193264e5a}
	}
	if hash != original {
		t.Errorf("ISOP checksum: got %016x, want original CUDD %016x", hash, original)
	}
	for i, want := range expected {
		if each[i] != want {
			t.Errorf("ISOP result %d: got %016x, want %016x", i, each[i], want)
		}
	}
}

func TestISOPReversedAdjacentLiteralPairs(t *testing.T) {
	dd := NewBDDManager(3)
	dd.ZddSize = 6
	dd.PermZ = []int{1, 0, 3, 2, 5, 4}
	dd.InvPermZ = []int{1, 0, 3, 2, 5, 4}
	InitializeZDDUniverse(dd)
	var cover *DdNode
	if Cudd_zddIsop(dd, dd.Vars[0], dd.Vars[0], &cover) != dd.Vars[0] || cover == nil {
		t.Fatal("valid reversed adjacent literals rejected")
	}
	if Cudd_MakeBddFromZddCover(dd, cover) != dd.Vars[0] {
		t.Fatal("reversed literal cover conversion failed")
	}
	checkISOPOriginalOracle(t, dd, false)
	if Cudd_ShuffleHeap(dd, []int{2, 0, 1}) != 1 {
		t.Fatal("BDD order fixture failed")
	}
	checkISOPOriginalOracle(t, dd, true)
}

func TestISOPCoverAfterBDDHeapShuffle(t *testing.T) {
	dd := Cudd_Init(3, 6, 256, 256, 0)
	f := Cudd_bddIte(dd, dd.Vars[0], dd.Vars[1], dd.Vars[2])
	Cudd_Ref(f)
	var cover *DdNode
	if Cudd_zddIsop(dd, f, f, &cover) != f {
		t.Fatal("exact ISOP changed function")
	}
	Cudd_Ref(cover)
	before := coverBits64(dd, cover)
	if Cudd_ShuffleHeap(dd, []int{2, 1, 0}) != 1 {
		t.Fatal("heap shuffle failed")
	}
	if coverBits64(dd, cover) != before || Cudd_MakeBddFromZddCover(dd, cover) != f {
		t.Fatal("cover/BDD correspondence lost after reorder")
	}
	complement := Cudd_zddComplement(dd, cover)
	if Cudd_MakeBddFromZddCover(dd, complement) != Cudd_Not(f) {
		t.Fatal("logical complement lost after reorder")
	}
	Cudd_RecursiveDeref(dd, f)
	Cudd_RecursiveDeref(dd, cover)
}

func TestISOPPairValidationAfterOrientationFix(t *testing.T) {
	reversed := NewBDDManager(3)
	reversed.ZddSize = 6
	reversed.PermZ = []int{1, 0, 3, 2, 5, 4}
	reversed.InvPermZ = []int{1, 0, 3, 2, 5, 4}
	InitializeZDDUniverse(reversed)
	literal := Cudd_zddChange(reversed, reversed.One, 0)
	p, n, d, ok := ZDDCoverCofactors(reversed, literal, 1)
	if !ok || p != reversed.One || n != reversed.Zero || d != reversed.Zero {
		t.Fatal("reversed cofactors")
	}
	inconsistent := Cudd_zddChange(reversed, literal, 1)
	if Cudd_MakeBddFromZddCover(reversed, inconsistent) != nil {
		t.Fatal("reversed contradictory cube accepted")
	}
	bad := NewBDDManager(3)
	bad.ZddSize = 6
	bad.PermZ = []int{0, 2, 1, 3, 4, 5}
	bad.InvPermZ = []int{0, 2, 1, 3, 4, 5}
	InitializeZDDUniverse(bad)
	sentinel := bad.One
	out := sentinel
	if Cudd_zddIsop(bad, bad.Vars[0], bad.Vars[0], &out) != nil || out != sentinel {
		t.Fatal("nonadjacent pair failure changed output")
	}
}

func TestISOPClosedAndNilManagers(t *testing.T) {
	if Cudd_bddIsop(nil, nil, nil) != nil || Cudd_MakeBddFromZddCover(nil, nil) != nil || Cudd_zddComplement(nil, nil) != nil {
		t.Fatal("nil manager accepted")
	}
	dd := Cudd_Init(1, 2, 256, 256, 0)
	f := dd.Vars[0]
	cover := dd.One
	out := cover
	Cudd_Quit(dd)
	if Cudd_zddIsop(dd, f, f, &out) != nil || out != cover || Cudd_MakeBddFromZddCover(dd, cover) != nil || Cudd_zddComplement(dd, cover) != nil {
		t.Fatal("closed manager accepted or output changed")
	}
}

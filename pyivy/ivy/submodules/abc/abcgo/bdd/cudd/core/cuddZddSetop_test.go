package core

import "testing"

func zddTable(dd *DdManager, f *DdNode) uint16 {
	var table uint16
	var visit func(*DdNode, uint)
	visit = func(n *DdNode, set uint) {
		if n == dd.Zero {
			return
		}
		if n == dd.One {
			table |= 1 << set
			return
		}
		visit(n.E, set)
		visit(n.T, set|1<<n.Index)
	}
	visit(f, 0)
	return table
}
func zddFromTable(dd *DdManager, table uint16) *DdNode {
	f := dd.Zero
	for set := 0; set < 8; set++ {
		if table&(1<<set) == 0 {
			continue
		}
		combination := dd.One
		for variable := 0; variable < 3; variable++ {
			if set&(1<<variable) != 0 {
				combination = Cudd_zddChange(dd, combination, variable)
			}
		}
		f = Cudd_zddUnion(dd, f, combination)
	}
	return f
}
func TestZDDSetOperationsCanonicalExhaustive(t *testing.T) {
	dd := Cudd_Init(0, 3, 256, 256, 0)
	families := make([]*DdNode, 256)
	for table := range families {
		families[table] = zddFromTable(dd, uint16(table))
		if got := zddTable(dd, families[table]); got != uint16(table) {
			t.Fatal(table, got)
		}
	}
	for a, f := range families {
		for b, g := range families {
			if Cudd_zddUnion(dd, f, g) != families[a|b] {
				t.Fatal("union", a, b)
			}
			if Cudd_zddIntersect(dd, f, g) != families[a&b] {
				t.Fatal("intersection", a, b)
			}
			if Cudd_zddDiff(dd, f, g) != families[a&^b] {
				t.Fatal("difference", a, b)
			}
			before := len(dd.zddUnique)
			query := Cudd_zddDiffConst(dd, f, g)
			if (query == dd.Zero) != (a&^b == 0) || len(dd.zddUnique) != before {
				t.Fatal("nonallocating inclusion", a, b)
			}
		}
	}
	for _, n := range dd.zddUnique {
		if n.T == dd.Zero || n.family != 1 || Cudd_IsComplement(n.T) || Cudd_IsComplement(n.E) || dd.zddLevel(n.T) <= dd.zddLevel(n) || dd.zddLevel(n.E) <= dd.zddLevel(n) {
			t.Fatal("ZDD invariant")
		}
	}
}
func TestZDDIteSubsetsAndChange(t *testing.T) {
	dd := Cudd_Init(0, 3, 256, 256, 0)
	for table := 0; table < 256; table++ {
		f := zddFromTable(dd, uint16(table))
		for variable := 0; variable < 3; variable++ {
			var positive, negative, changed uint16
			for set := 0; set < 8; set++ {
				if table&(1<<set) == 0 {
					continue
				}
				if set&(1<<variable) != 0 {
					positive |= 1 << (set &^ (1 << variable))
				} else {
					negative |= 1 << set
				}
				changed |= 1 << (set ^ (1 << variable))
			}
			if zddTable(dd, Cudd_zddSubset1(dd, f, variable)) != positive || zddTable(dd, Cudd_zddSubset0(dd, f, variable)) != negative || zddTable(dd, Cudd_zddChange(dd, f, variable)) != changed {
				t.Fatal("cofactor/change", table, variable)
			}
			if Cudd_zddChange(dd, Cudd_zddChange(dd, f, variable), variable) != f {
				t.Fatal("toggle involution")
			}
		}
		g, h := zddFromTable(dd, uint16(table^0x5a)), zddFromTable(dd, uint16(table^0xc3))
		got := zddTable(dd, Cudd_zddIte(dd, f, g, h))
		want := uint16(table&(table^0x5a) | ((table ^ 0xc3) &^ table))
		if got != want {
			t.Fatal("ITE", table, got, want)
		}
	}
}
func TestZDDProjectionGrowthAndIsolation(t *testing.T) {
	dd := Cudd_Init(1, 3, 256, 256, 0)
	for i := 0; i < 3; i++ {
		var expected uint16
		for set := 0; set < 8; set++ {
			if set&(1<<i) != 0 {
				expected |= 1 << set
			}
		}
		if zddTable(dd, Cudd_zddIthVar(dd, i)) != expected {
			t.Fatal("projection", i)
		}
	}
	if Cudd_zddChange(dd, dd.One, 5) == nil || dd.ZddSize != 6 || len(dd.Univ) != 6 {
		t.Fatal("ZDD growth")
	}
	if len(dd.Vars) != 1 {
		t.Fatal("ZDD growth modified BDD inventory")
	}
	before := dd.ZddSize
	if Cudd_zddChange(dd, dd.One, -1) != nil || Cudd_zddChange(dd, dd.One, int(CUDD_MAXINDEX)) != nil || dd.ZddSize != before {
		t.Fatal("invalid index mutated manager")
	}
	if Cudd_zddUnion(dd, dd.LogicZero, dd.One) != nil || Cudd_zddUnion(dd, dd.Vars[0], dd.One) != nil || Cudd_zddUnion(dd, NewBDDManager(0).Zero, dd.Zero) != nil {
		t.Fatal("invalid edge family/manager accepted")
	}
	if Cudd_ReadZddOne(dd, dd.ZddSize) != dd.One || Cudd_ReadPermZdd(dd, 5) != 5 || Cudd_ReadInvPermZdd(dd, 5) != 5 {
		t.Fatal("ZDD accessors")
	}
	Cudd_Quit(dd)
	if Cudd_zddUnion(dd, dd.One, dd.Zero) != nil || len(dd.PermZ) != 0 {
		t.Fatal("closed manager accepted")
	}
}

func TestZDDNonidentityLevelOrderAndInitValidation(t *testing.T) {
	dd := NewBDDManager(0)
	dd.ZddSize = 3
	dd.PermZ = []int{2, 1, 0}
	dd.InvPermZ = []int{2, 1, 0}
	if InitializeZDDUniverse(dd) != 1 || dd.Univ[0].Index != 2 || zddTable(dd, dd.Univ[0]) != 255 {
		t.Fatal("universe ignored level order")
	}
	for a := 0; a < 256; a += 7 {
		f := zddFromTable(dd, uint16(a))
		g := zddFromTable(dd, uint16(a^0xaa))
		if zddTable(dd, Cudd_zddUnion(dd, f, g)) != uint16(a|(a^0xaa)) || zddTable(dd, Cudd_zddDiff(dd, f, g)) != uint16(a&^(a^0xaa)) {
			t.Fatal("set operations ignored level order", a)
		}
	}
	previous := dd.Univ[0]
	dd.InvPermZ[0] = 0
	if InitializeZDDUniverse(dd) != 0 || dd.Univ[0] != previous {
		t.Fatal("invalid permutation mutated universe")
	}
	if Cudd_Init(CUDD_MAXINDEX, 0, 0, 0, 0) != nil || Cudd_Init(0, CUDD_MAXINDEX, 0, 0, 0) != nil {
		t.Fatal("invalid variable counts accepted")
	}
}

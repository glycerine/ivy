package core

import "testing"

var pairedCubes2 = []uint{0, 1, 2, 4, 5, 6, 8, 9, 10}

func zddCover(dd *DdManager, mask uint) *DdNode {
	f := dd.Zero
	for term, cube := range pairedCubes2 {
		if mask&(1<<term) == 0 {
			continue
		}
		n := dd.One
		for i := 0; i < 4; i++ {
			if cube&(1<<i) != 0 {
				n = Cudd_zddChange(dd, n, i)
			}
		}
		f = Cudd_zddUnion(dd, f, n)
	}
	return f
}

// The original C oracle covers all512 valid two-variable covers, all511
// nonempty divisors, and all six original APIs. No C runtime is used here.
func TestZDDCoverOriginalCUDDOracle(t *testing.T) {
	dd := Cudd_Init(0, 4, 256, 256, 0)
	covers := make([]*DdNode, 512)
	for i := range covers {
		covers[i] = zddCover(dd, uint(i))
	}
	ops := []struct {
		name     string
		apply    func(*DdManager, *DdNode, *DdNode) *DdNode
		checksum uint64
	}{
		{"unate product", Cudd_zddUnateProduct, 0xa78a582914a5d325},
		{"paired product", Cudd_zddProduct, 0xd16e4808d4109325},
		{"divide", Cudd_zddDivide, 0x5a4ccbe787e78e30},
		{"divideF", Cudd_zddDivideF, 0x5a4ccbe787e78e30},
		{"weak divide", Cudd_zddWeakDiv, 0x5a4ccbe787e78e30},
		{"weak divideF", Cudd_zddWeakDivF, 0x5a4ccbe787e78e30},
	}
	hash := uint64(14695981039346656037)
	each := make([]uint64, len(ops))
	for i := range each {
		each[i] = hash
	}
	for a := 0; a < 512; a++ {
		for b := 1; b < 512; b++ {
			for i, op := range ops {
				result := op.apply(dd, covers[a], covers[b])
				if result == nil {
					t.Fatalf("%s failed for %d/%d", op.name, a, b)
				}
				value := uint64(zddTable(dd, result))
				hash ^= value
				hash *= 1099511628211
				each[i] ^= value
				each[i] *= 1099511628211
			}
		}
	}
	const original uint64 = 0x1c45d75814740e31
	if hash != original {
		t.Errorf("cover results differ from original CUDD: got %016x, want %016x", hash, original)
	}
	for i, op := range ops {
		if each[i] != op.checksum {
			t.Errorf("%s: got %016x, want %016x", op.name, each[i], op.checksum)
		}
	}
}
func TestZDDProductLiteralConflictsAndDivision(t *testing.T) {
	dd := Cudd_Init(0, 4, 256, 256, 0)
	a, notA := Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 1)
	b := Cudd_zddChange(dd, dd.One, 2)
	if Cudd_zddProduct(dd, a, notA) != dd.Zero {
		t.Fatal("opposite literals were multiplied")
	}
	if Cudd_zddUnateProduct(dd, a, notA) == dd.Zero {
		t.Fatal("unate product incorrectly enforces paired literals")
	}
	product := Cudd_zddProduct(dd, a, b)
	if Cudd_zddWeakDiv(dd, product, b) != a || Cudd_zddWeakDivF(dd, product, b) != a {
		t.Fatal("earlier dividend pair lost")
	}
	if Cudd_zddDivide(dd, product, b) != a || Cudd_zddDivideF(dd, product, b) != a {
		t.Fatal("unate quotient")
	}
	for _, op := range []func(*DdManager, *DdNode, *DdNode) *DdNode{Cudd_zddDivide, Cudd_zddDivideF, Cudd_zddWeakDiv, Cudd_zddWeakDivF} {
		if op(dd, product, dd.Zero) != nil {
			t.Fatal("division by empty family accepted")
		}
	}
	inconsistent := Cudd_zddChange(dd, a, 1)
	if Cudd_zddProduct(dd, inconsistent, b) != nil || Cudd_zddWeakDiv(dd, inconsistent, b) != nil {
		t.Fatal("inconsistent literal cube accepted")
	}
}
func TestZDDPairedOrderConstraintsAndDomainIsolation(t *testing.T) {
	dd := Cudd_Init(1, 4, 256, 256, 0)
	a := Cudd_zddChange(dd, dd.One, 0)
	if Cudd_zddProduct(dd, a, dd.Vars[0]) != nil || Cudd_zddWeakDivF(dd, a, NewBDDManager(0).One) != nil {
		t.Fatal("foreign domain accepted")
	}
	reordered := NewBDDManager(0)
	reordered.ZddSize = 4
	reordered.PermZ = []int{2, 3, 0, 1}
	reordered.InvPermZ = []int{2, 3, 0, 1}
	InitializeZDDUniverse(reordered)
	pa := Cudd_zddChange(reordered, reordered.One, 0)
	pb := Cudd_zddChange(reordered, reordered.One, 2)
	product := Cudd_zddProduct(reordered, pa, pb)
	if product == nil || Cudd_zddWeakDivF(reordered, product, pb) != pa {
		t.Fatal("valid pair-group order rejected")
	}
	bad := NewBDDManager(0)
	bad.ZddSize = 4
	bad.PermZ = []int{0, 2, 1, 3}
	bad.InvPermZ = []int{0, 2, 1, 3}
	InitializeZDDUniverse(bad)
	f, g := Cudd_zddChange(bad, bad.One, 0), Cudd_zddChange(bad, bad.One, 2)
	if Cudd_zddProduct(bad, f, g) != nil || Cudd_zddWeakDiv(bad, f, g) != nil {
		t.Fatal("nonadjacent literal pairs accepted")
	}
	if Cudd_zddUnateProduct(bad, f, g) == nil {
		t.Fatal("unate product rejected arbitrary variable order")
	}
}

func coverCubes3() []uint {
	var cubes []uint
	for set := uint(0); set < 64; set++ {
		if set&3 != 3 && (set>>2)&3 != 3 && (set>>4)&3 != 3 {
			cubes = append(cubes, set)
		}
	}
	return cubes
}
func coverFromCubes(dd *DdManager, mask uint, cubes []uint) *DdNode {
	f := dd.Zero
	for term, cube := range cubes {
		if mask&(1<<term) == 0 {
			continue
		}
		n := dd.One
		for i := 0; i < 6; i++ {
			if cube&(1<<i) != 0 {
				n = Cudd_zddChange(dd, n, i)
			}
		}
		f = Cudd_zddUnion(dd, f, n)
	}
	return f
}
func coverBits64(dd *DdManager, f *DdNode) uint64 {
	var table uint64
	var visit func(*DdNode, uint)
	visit = func(n *DdNode, set uint) {
		if n == dd.Zero {
			return
		}
		if n == dd.One {
			table |= uint64(1) << set
			return
		}
		visit(n.E, set)
		visit(n.T, set|1<<n.Index)
	}
	visit(f, 0)
	return table
}

// Native mask multiplication constructs structured input fixtures independently
// of the graph algorithms being checked against the original C implementation.
func nativeCoverProduct(a, b uint, cubes []uint) uint {
	lookup := map[uint]int{}
	for i, cube := range cubes {
		lookup[cube] = i
	}
	var result uint
	for i, x := range cubes {
		if a&(1<<i) == 0 {
			continue
		}
		for j, y := range cubes {
			if b&(1<<j) == 0 {
				continue
			}
			if index, ok := lookup[x|y]; ok {
				result |= 1 << index
			}
		}
	}
	return result
}
func TestZDDCoverThreeVariableOriginalOracle(t *testing.T) {
	t.Run("identity", func(t *testing.T) { checkZDDCoverThreeVariableOracle(t, Cudd_Init(0, 6, 256, 256, 0)) })
	t.Run("pair group permutation", func(t *testing.T) {
		dd := NewBDDManager(0)
		dd.ZddSize = 6
		dd.PermZ = []int{2, 3, 4, 5, 0, 1}
		dd.InvPermZ = []int{4, 5, 0, 1, 2, 3}
		if InitializeZDDUniverse(dd) != 1 {
			t.Fatal("invalid fixture order")
		}
		checkZDDCoverThreeVariableOracle(t, dd)
	})
}
func checkZDDCoverThreeVariableOracle(t *testing.T, dd *DdManager) {
	t.Helper()
	cubes := coverCubes3()
	ops := []struct {
		apply    func(*DdManager, *DdNode, *DdNode) *DdNode
		checksum uint64
	}{
		{Cudd_zddUnateProduct, 0x663614f4e05e0ffb}, {Cudd_zddProduct, 0x4188ee05273f6323},
		{Cudd_zddDivide, 0x7289456388a26e57}, {Cudd_zddDivideF, 0x7289456388a26e57},
		{Cudd_zddWeakDiv, 0x7289456388a26e57}, {Cudd_zddWeakDivF, 0x7289456388a26e57},
	}
	hash := uint64(14695981039346656037)
	each := make([]uint64, 6)
	for i := range each {
		each[i] = hash
	}
	seed := uint32(0x9e3779b9)
	for iteration := 0; iteration < 2000; iteration++ {
		seed = seed*1664525 + 1013904223
		a := uint(seed^(seed>>16)) & 0x7ffffff
		seed = seed*1664525 + 1013904223
		b := uint(seed^(seed>>16)) & 0x7ffffff
		if b == 0 {
			b = 1
		}
		if iteration%3 == 0 {
			a = nativeCoverProduct(a, b, cubes)
		}
		if iteration%3 == 1 {
			literal := uint(1) << ((iteration / 3) % 6)
			for i, cube := range cubes {
				if cube == literal {
					b = 1 << i
				}
			}
		}
		f, g := coverFromCubes(dd, a, cubes), coverFromCubes(dd, b, cubes)
		for i, op := range ops {
			result := op.apply(dd, f, g)
			if result == nil {
				t.Fatalf("operation %d failed at iteration %d", i, iteration)
			}
			value := coverBits64(dd, result)
			hash ^= value
			hash *= 1099511628211
			each[i] ^= value
			each[i] *= 1099511628211
		}
	}
	const original uint64 = 0x5eb39e097b2af27d
	if hash != original {
		t.Errorf("three-variable cover checksum: got %016x, want %016x", hash, original)
	}
	for i, op := range ops {
		if each[i] != op.checksum {
			t.Errorf("operation %d: got %016x, want %016x", i, each[i], op.checksum)
		}
	}
}
func TestZDDCoverDecompositionUtilities(t *testing.T) {
	dd := Cudd_Init(0, 4, 256, 256, 0)
	a, notA, b := Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 1), Cudd_zddChange(dd, dd.One, 2)
	f := Cudd_zddUnion(dd, Cudd_zddProduct(dd, a, b), Cudd_zddUnion(dd, notA, b))
	positive, negative, absent, ok := ZDDCoverCofactors(dd, f, 1)
	if !ok || positive != b || negative != dd.One || absent != b {
		t.Fatal("three-way decomposition")
	}
	p, n, ok := ZDDCofactors(dd, f, 0)
	if !ok || p != b || n != Cudd_zddUnion(dd, notA, b) {
		t.Fatal("two-way decomposition")
	}
	if _, _, _, ok := ZDDCoverCofactors(dd, f, -1); ok {
		t.Fatal("negative index accepted")
	}
	if _, _, ok := ZDDCofactors(dd, f, -1); ok {
		t.Fatal("negative cofactor index accepted")
	}
}

func TestZDDPositiveFirstLiteralConstraint(t *testing.T) {
	dd := NewBDDManager(0)
	dd.ZddSize = 4
	dd.PermZ = []int{1, 0, 2, 3}
	dd.InvPermZ = []int{1, 0, 2, 3}
	InitializeZDDUniverse(dd)
	a, b := Cudd_zddChange(dd, dd.One, 0), Cudd_zddChange(dd, dd.One, 2)
	if Cudd_zddProduct(dd, a, b) != nil || Cudd_zddWeakDivF(dd, a, b) != nil {
		t.Fatal("reversed literal pair accepted")
	}
	if Cudd_zddUnateProduct(dd, a, b) == nil {
		t.Fatal("unate API inherited paired constraint")
	}
}

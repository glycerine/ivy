package abc

// Stand-alone model tests for the ZDD ports from src/bdd/cudd/cuddZdd*.c.

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func cuddZDDTestSets(mask uint64, universe int) []uint64 {
	sets := make([]uint64, 0)
	for set := 0; set < universe; set++ {
		if mask&(uint64(1)<<set) != 0 {
			sets = append(sets, uint64(set))
		}
	}
	return sets
}

func cuddZDDTestMap(sets []uint64) map[uint64]bool {
	r := make(map[uint64]bool, len(sets))
	for _, set := range sets {
		r[set] = true
	}
	return r
}

func cuddZDDTestSorted(model map[uint64]bool) []uint64 {
	r := make([]uint64, 0, len(model))
	for set := range model {
		r = append(r, set)
	}
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return r
}

func cuddZDDTestRead(t testing.TB, manager *CuddZDDManager, root CuddZDDEdge) []uint64 {
	t.Helper()
	sets, err := manager.ToBitsets(root, CuddZDDOptions{})
	if err != nil {
		t.Fatalf("ToBitsets: %v", err)
	}
	return sets
}

func TestCuddZDDExhaustiveThreeVariableFamilies(t *testing.T) {
	manager, err := NewCuddZDDManager(3, 1<<18)
	if err != nil {
		t.Fatal(err)
	}
	root := make([]CuddZDDEdge, 256)
	for family := 0; family < 256; family++ {
		root[family], err = manager.FromBitsets(cuddZDDTestSets(uint64(family), 8), CuddZDDOptions{})
		if err != nil {
			t.Fatalf("family %d: %v", family, err)
		}
		again, makeErr := manager.FromBitsets(cuddZDDTestSets(uint64(family), 8), CuddZDDOptions{})
		if makeErr != nil || again != root[family] {
			t.Fatalf("noncanonical family %d: %d %d %v", family, root[family], again, makeErr)
		}
	}
	for left := 0; left < 256; left++ {
		for right := 0; right < 256; right++ {
			a, b := cuddZDDTestMap(cuddZDDTestSets(uint64(left), 8)), cuddZDDTestMap(cuddZDDTestSets(uint64(right), 8))
			union, intersection, difference, product := map[uint64]bool{}, map[uint64]bool{}, map[uint64]bool{}, map[uint64]bool{}
			for set := range a {
				union[set] = true
				if b[set] {
					intersection[set] = true
				} else {
					difference[set] = true
				}
				for other := range b {
					product[set|other] = true
				}
			}
			for set := range b {
				union[set] = true
			}
			checks := []struct {
				name string
				want map[uint64]bool
				call func() (CuddZDDEdge, error)
			}{
				{"union", union, func() (CuddZDDEdge, error) { return manager.Union(root[left], root[right], CuddZDDOptions{}) }},
				{"intersection", intersection, func() (CuddZDDEdge, error) { return manager.Intersect(root[left], root[right], CuddZDDOptions{}) }},
				{"difference", difference, func() (CuddZDDEdge, error) { return manager.Difference(root[left], root[right], CuddZDDOptions{}) }},
				{"product", product, func() (CuddZDDEdge, error) { return manager.Product(root[left], root[right], CuddZDDOptions{}) }},
			}
			for _, check := range checks {
				gotRoot, callErr := check.call()
				if callErr != nil {
					t.Fatalf("%s(%d,%d): %v", check.name, left, right, callErr)
				}
				if got := cuddZDDTestRead(t, manager, gotRoot); !reflect.DeepEqual(got, cuddZDDTestSorted(check.want)) {
					t.Fatalf("%s(%d,%d)=%v want %v", check.name, left, right, got, cuddZDDTestSorted(check.want))
				}
			}
		}
	}
}

func TestCuddZDDChangeSubsetsDivisionCountSupport(t *testing.T) {
	manager, err := NewCuddZDDManager(4, 8192)
	if err != nil {
		t.Fatal(err)
	}
	root, err := manager.FromBitsets([]uint64{0, 1, 3, 6, 9, 15}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	condition, _ := manager.FromBitsets([]uint64{0, 1}, CuddZDDOptions{})
	thenFamily, _ := manager.FromBitsets([]uint64{0, 2}, CuddZDDOptions{})
	elseFamily, _ := manager.FromBitsets([]uint64{1, 2}, CuddZDDOptions{})
	selected, err := manager.Ite(condition, thenFamily, elseFamily, CuddZDDOptions{})
	if err != nil || !reflect.DeepEqual(cuddZDDTestRead(t, manager, selected), []uint64{0, 2}) {
		t.Fatalf("ite=%v %v", cuddZDDTestRead(t, manager, selected), err)
	}
	changed, err := manager.Change(root, 1, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cuddZDDTestRead(t, manager, changed), []uint64{1, 2, 3, 4, 11, 13}; !reflect.DeepEqual(got, want) {
		t.Fatalf("change=%v want %v", got, want)
	}
	one, err := manager.Subset1(root, 0, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cuddZDDTestRead(t, manager, one), []uint64{0, 2, 8, 14}; !reflect.DeepEqual(got, want) {
		t.Fatalf("subset1=%v want %v", got, want)
	}
	zero, err := manager.Subset0(root, 0, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cuddZDDTestRead(t, manager, zero), []uint64{0, 6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("subset0=%v want %v", got, want)
	}
	dividend, _ := manager.FromBitsets([]uint64{3, 5, 6, 7}, CuddZDDOptions{})
	divisor, _ := manager.FromBitsets([]uint64{1, 2}, CuddZDDOptions{})
	quotient, err := manager.Divide(dividend, divisor, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cuddZDDTestRead(t, manager, quotient), []uint64{4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("divide=%v want %v", got, want)
	}
	count, err := manager.Count(root, CuddZDDOptions{})
	if err != nil || count != 6 {
		t.Fatalf("count=%d %v", count, err)
	}
	minterms, err := manager.CountMinterms(root, 1, CuddZDDOptions{})
	if err != nil || minterms != 3 {
		t.Fatalf("minterms=%g %v", minterms, err)
	}
	empty, err := manager.DifferenceEmpty(root, root, CuddZDDOptions{})
	if err != nil || !empty {
		t.Fatalf("difference inclusion=%v %v", empty, err)
	}
	support, err := manager.Support(root, CuddZDDOptions{})
	if err != nil || !reflect.DeepEqual(support, []int{0, 1, 2, 3}) {
		t.Fatalf("support=%v %v", support, err)
	}
	size, err := manager.DagSize(root)
	if err != nil || size < 3 {
		t.Fatalf("dag size=%d %v", size, err)
	}
}

func TestCuddZDDPortsAndISOPTruth(t *testing.T) {
	bdd, err := NewExtraBDDManager(3, 4096)
	if err != nil {
		t.Fatal(err)
	}
	x0, x1, x2 := ExtraBDDLit(2), ExtraBDDLit(4), ExtraBDDLit(6)
	a, err := bdd.And(x0, x1.Not(), ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	f, err := bdd.Or(a, x2, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	zdd, err := NewCuddZDDManager(3, 4096)
	if err != nil {
		t.Fatal(err)
	}
	cover, err := zdd.PortFromBDD(bdd, f, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	round, err := zdd.PortToBDD(cover, bdd, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, _ := bdd.Evaluate(f, values)
		got, _ := bdd.Evaluate(round, values)
		if got != want {
			t.Fatalf("assignment %d: got %v want %v", assignment, got, want)
		}
	}
	result, isop, err := zdd.Isop(bdd, f, ExtraBDDTrue, CuddZDDOptions{})
	if err != nil || result != f {
		t.Fatalf("isop result=%d cover=%d err=%v", result, isop, err)
	}
	isopBDD, err := zdd.PortToBDD(isop, bdd, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		a, _ := bdd.Evaluate(f, values)
		b, _ := bdd.Evaluate(isopBDD, values)
		if a != b {
			t.Fatalf("isop mismatch %d", assignment)
		}
	}
}

func TestCuddZDDLimitsCancellationRollbackAndValidation(t *testing.T) {
	manager, _ := NewCuddZDDManager(5, 128)
	root, _ := manager.FromBitsets([]uint64{1, 2, 4, 8, 16}, CuddZDDOptions{})
	before, _ := manager.Statistics()
	if result, err := manager.Product(root, root, CuddZDDOptions{MaxOperations: 1}); result != 0 || !errors.Is(err, ErrLimit) {
		t.Fatalf("limit result=%d err=%v", result, err)
	}
	after, _ := manager.Statistics()
	if before.Nodes != after.Nodes {
		t.Fatalf("rollback nodes %d -> %d", before.Nodes, after.Nodes)
	}
	if result, err := manager.Change(root, 0, CuddZDDOptions{Stop: func() bool { return true }}); result != 0 || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel result=%d err=%v", result, err)
	}
	if _, err := manager.ToBitsets(999, CuddZDDOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid edge: %v", err)
	}
	if _, err := manager.Change(root, 5, CuddZDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("invalid variable: %v", err)
	}
	if _, err := manager.FromBitsets([]uint64{32}, CuddZDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("invalid set: %v", err)
	}
	if _, err := manager.FromBitsets([]uint64{1}, CuddZDDOptions{MaxBytes: 15}); !errors.Is(err, ErrLimit) {
		t.Fatalf("family memory limit: %v", err)
	}
	var nilManager *CuddZDDManager
	if _, err := nilManager.Change(0, 0, CuddZDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil manager: %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Statistics(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed: %v", err)
	}
}

func TestCuddZDDCompactCountSupportAndDagSize(t *testing.T) {
	manager, err := NewCuddZDDManager(20, 128)
	if err != nil {
		t.Fatal(err)
	}
	work := &cuddZDDWork{options: CuddZDDOptions{MaxOperations: 1000, MaxBytes: 1 << 20}}
	root := CuddZDDBase
	for variable := 19; variable >= 0; variable-- {
		root, err = manager.makeNode(variable, root, root, work)
		if err != nil {
			t.Fatal(err)
		}
	}
	count, err := manager.Count(root, CuddZDDOptions{MaxOperations: 100})
	if err != nil || count != 1<<20 {
		t.Fatalf("compact count=%d err=%v", count, err)
	}
	support, err := manager.Support(root, CuddZDDOptions{MaxOperations: 100})
	if err != nil || len(support) != 20 || support[0] != 0 || support[19] != 19 {
		t.Fatalf("compact support=%v err=%v", support, err)
	}
	size, err := manager.DagSize(root)
	if err != nil || size != 20 {
		t.Fatalf("compact DAG size=%d err=%v", size, err)
	}
	if size, err = manager.DagSize(CuddZDDBase); err != nil || size != 0 {
		t.Fatalf("terminal DAG size=%d err=%v", size, err)
	}
	if _, err = manager.ToBitsets(root, CuddZDDOptions{MaxOperations: 100}); !errors.Is(err, ErrLimit) {
		t.Fatalf("bounded expansion: %v", err)
	}
}

func TestCuddZDDPortsUseCumulativeBudgetAndRollback(t *testing.T) {
	zdd, err := NewCuddZDDManager(5, 4096)
	if err != nil {
		t.Fatal(err)
	}
	sets := []uint64{0, 3, 5, 9, 14, 17, 22, 31}
	root, err := zdd.FromBitsets(sets, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	bdd, err := NewExtraBDDManager(5, 4096)
	if err != nil {
		t.Fatal(err)
	}
	before, err := bdd.Statistics()
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	_, err = zdd.PortToBDD(root, bdd, CuddZDDOptions{Stop: func() bool {
		calls++
		return calls > 220
	}})
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("port cancellation: calls=%d err=%v", calls, err)
	}
	after, err := bdd.Statistics()
	if err != nil {
		t.Fatal(err)
	}
	if after.Nodes != before.Nodes {
		t.Fatalf("failed port retained BDD nodes: %d -> %d", before.Nodes, after.Nodes)
	}
	trueBDD := ExtraBDDTrue
	beforeZDD, _ := zdd.Statistics()
	if result, portErr := zdd.PortFromBDD(bdd, trueBDD, CuddZDDOptions{MaxOperations: 10}); result != 0 || !errors.Is(portErr, ErrLimit) {
		t.Fatalf("bounded BDD enumeration result=%d err=%v", result, portErr)
	}
	afterZDD, _ := zdd.Statistics()
	if afterZDD.Nodes != beforeZDD.Nodes {
		t.Fatalf("failed BDD import retained ZDD nodes: %d -> %d", beforeZDD.Nodes, afterZDD.Nodes)
	}
}

func TestCuddZDDRandomModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5dd))
	manager, _ := NewCuddZDDManager(6, 1<<18)
	for iteration := 0; iteration < 5000; iteration++ {
		leftModel, rightModel := map[uint64]bool{}, map[uint64]bool{}
		for set := uint64(0); set < 64; set++ {
			if rng.Intn(5) == 0 {
				leftModel[set] = true
			}
			if rng.Intn(5) == 0 {
				rightModel[set] = true
			}
		}
		left, _ := manager.FromBitsets(cuddZDDTestSorted(leftModel), CuddZDDOptions{})
		right, _ := manager.FromBitsets(cuddZDDTestSorted(rightModel), CuddZDDOptions{})
		got, err := manager.Product(left, right, CuddZDDOptions{})
		if err != nil {
			t.Fatalf("iteration %d: %v", iteration, err)
		}
		want := map[uint64]bool{}
		for a := range leftModel {
			for b := range rightModel {
				want[a|b] = true
			}
		}
		if sets := cuddZDDTestRead(t, manager, got); !reflect.DeepEqual(sets, cuddZDDTestSorted(want)) {
			t.Fatalf("iteration %d: got %v want %v", iteration, sets, cuddZDDTestSorted(want))
		}
	}
}

func FuzzCuddZDDSetAlgebra(f *testing.F) {
	f.Add(uint64(0x96), uint64(0x69), uint8(1))
	f.Add(uint64(0xffff), uint64(0x8001), uint8(2))
	f.Fuzz(func(t *testing.T, leftMask, rightMask uint64, variable uint8) {
		manager, _ := NewCuddZDDManager(3, 4096)
		leftMask &= 255
		rightMask &= 255
		left, _ := manager.FromBitsets(cuddZDDTestSets(leftMask, 8), CuddZDDOptions{})
		right, _ := manager.FromBitsets(cuddZDDTestSets(rightMask, 8), CuddZDDOptions{})
		product, err := manager.Product(left, right, CuddZDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := map[uint64]bool{}
		for a := range cuddZDDTestMap(cuddZDDTestSets(leftMask, 8)) {
			for b := range cuddZDDTestMap(cuddZDDTestSets(rightMask, 8)) {
				want[a|b] = true
			}
		}
		if got := cuddZDDTestRead(t, manager, product); !reflect.DeepEqual(got, cuddZDDTestSorted(want)) {
			t.Fatalf("product=%v want %v", got, cuddZDDTestSorted(want))
		}
		changed, err := manager.Change(left, int(variable%3), CuddZDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		toggled := map[uint64]bool{}
		for set := range cuddZDDTestMap(cuddZDDTestSets(leftMask, 8)) {
			toggled[set^(uint64(1)<<uint(variable%3))] = true
		}
		if got := cuddZDDTestRead(t, manager, changed); !reflect.DeepEqual(got, cuddZDDTestSorted(toggled)) {
			t.Fatalf("change=%v want %v", got, cuddZDDTestSorted(toggled))
		}
	})
}

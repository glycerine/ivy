package abc

// Stand-alone tests for the src/bdd/extrab port.

import (
	"errors"
	"slices"
	"testing"
)

func extraBDDTestRoot(t testing.TB, manager *ExtraBDDManager, table uint64) ExtraBDDLit {
	t.Helper()
	count := 1 << uint(manager.variables)
	values := make([]bool, count)
	for assignment := range values {
		values[assignment] = table&(uint64(1)<<uint(assignment)) != 0
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: ExtraBDDOptions{MaxOperations: 1 << 20, MaxBytes: 1 << 20}, transaction: &transaction}
	root, err := manager.analysisBuild(values, &work)
	if err != nil {
		manager.rollback(&transaction)
		t.Fatalf("build table %#x: %v", table, err)
	}
	return root
}

func extraBDDTestTable(t testing.TB, manager *ExtraBDDManager, root ExtraBDDLit) uint64 {
	t.Helper()
	var table uint64
	assignment := make([]bool, manager.variables)
	for word := 0; word < 1<<uint(manager.variables); word++ {
		for variable := range assignment {
			assignment[variable] = word&(1<<uint(variable)) != 0
		}
		value, err := manager.Evaluate(root, assignment)
		if err != nil {
			t.Fatalf("evaluate: %v", err)
		}
		if value {
			table |= uint64(1) << uint(word)
		}
	}
	return table
}

func extraBDDExpectedSupport(table uint64, variables int) []int {
	result := make([]int, 0, variables)
	for variable := 0; variable < variables; variable++ {
		dependent := false
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			if assignment&(1<<uint(variable)) == 0 && (table>>uint(assignment)&1) != (table>>uint(assignment|(1<<uint(variable)))&1) {
				dependent = true
				break
			}
		}
		if dependent {
			result = append(result, variable)
		}
	}
	return result
}

func TestExtraBDDClusterInventory(t *testing.T) {
	want := []string{"extraBddAuto.c", "extraBddCas.c", "extraBddImage.c", "extraBddKmap.c", "extraBddMaxMin.c", "extraBddMisc.c", "extraBddSet.c", "extraBddSymm.c", "extraBddThresh.c", "extraBddTime.c", "extraBddUnate.c"}
	if got := ExtraBDDClusterBuildSources(); !slices.Equal(got, want) {
		t.Fatalf("inventory %v", got)
	}
	got := ExtraBDDClusterBuildSources()
	got[0] = "changed"
	if ExtraBDDClusterBuildSources()[0] != want[0] {
		t.Fatal("inventory aliases package state")
	}
}

func TestExtraBDDExhaustiveSupportUnateSymmetryAndCube(t *testing.T) {
	manager, err := NewExtraBDDManager(3, 2048)
	if err != nil {
		t.Fatal(err)
	}
	options := ExtraBDDOptions{MaxOperations: 1 << 20, MaxBytes: 1 << 20}
	for table := uint64(0); table < 256; table++ {
		root := extraBDDTestRoot(t, manager, table)
		support, supportErr := manager.ExtraSupport(root, options)
		if supportErr != nil || !slices.Equal(support, extraBDDExpectedSupport(table, 3)) {
			t.Fatalf("table %#x support %v err %v", table, support, supportErr)
		}
		unate, unateErr := manager.ExtraUnateness(root, options)
		if unateErr != nil {
			t.Fatalf("table %#x unate: %v", table, unateErr)
		}
		if len(unate) != len(support) {
			t.Fatalf("table %#x unate support %v %v", table, unate, support)
		}
		for _, item := range unate {
			positive, negative := true, true
			for assignment := 0; assignment < 8; assignment++ {
				if assignment&(1<<uint(item.Variable)) != 0 {
					continue
				}
				low := table>>uint(assignment)&1 != 0
				high := table>>uint(assignment|(1<<uint(item.Variable)))&1 != 0
				positive = positive && (!low || high)
				negative = negative && (!high || low)
			}
			if item.Positive != positive || item.Negative != negative {
				t.Fatalf("table %#x unate variable %d", table, item.Variable)
			}
		}
		pairs, pairErr := manager.ExtraSymmetry(root, options)
		if pairErr != nil {
			t.Fatalf("table %#x symmetry: %v", table, pairErr)
		}
		for _, pair := range pairs {
			for assignment := 0; assignment < 8; assignment++ {
				first := assignment >> uint(pair.First) & 1
				second := assignment >> uint(pair.Second) & 1
				swapped := assignment
				if first != second {
					swapped ^= 1<<uint(pair.First) | 1<<uint(pair.Second)
				}
				if table>>uint(assignment)&1 != table>>uint(swapped)&1 {
					t.Fatalf("table %#x false symmetry %+v", table, pair)
				}
			}
		}
		cube, cubeErr := manager.ExtraOneCube(root, options)
		if cubeErr != nil {
			t.Fatalf("table %#x cube: %v", table, cubeErr)
		}
		if (table == 0) != !cube.Exists {
			t.Fatalf("table %#x cube existence %+v", table, cube)
		}
		if cube.Exists {
			cubeTable := extraBDDTestTable(t, manager, cube.Root)
			if cubeTable&^table != 0 {
				t.Fatalf("table %#x cube %#x escapes", table, cubeTable)
			}
		}
		space, spaceErr := manager.ExtraTranslationSpace(root, options)
		if spaceErr != nil {
			t.Fatalf("table %#x translation space: %v", table, spaceErr)
		}
		var expectedSpace uint64
		for shift := 0; shift < 8; shift++ {
			invariant := true
			for assignment := 0; assignment < 8; assignment++ {
				invariant = invariant && table>>uint(assignment)&1 == table>>uint(assignment^shift)&1
			}
			if invariant {
				expectedSpace |= 1 << uint(shift)
			}
		}
		if got := extraBDDTestTable(t, manager, space); got != expectedSpace {
			t.Fatalf("table %#x translation space %#x want %#x", table, got, expectedSpace)
		}
	}
}

func TestExtraBDDCubesNaryAndTransfer(t *testing.T) {
	manager, err := NewExtraBDDManager(3, 1024)
	if err != nil {
		t.Fatal(err)
	}
	options := ExtraBDDOptions{}
	rangeCube, err := manager.ExtraRangeCube(0, 3, options)
	if err != nil || extraBDDTestTable(t, manager, rangeCube) != 0x80 {
		t.Fatalf("range cube %#x %v", extraBDDTestTable(t, manager, rangeCube), err)
	}
	emptyRange, err := manager.ExtraRangeCube(2, 2, options)
	if err != nil || emptyRange != ExtraBDDTrue {
		t.Fatalf("empty half-open range %d %v", emptyRange, err)
	}
	bitsCube, err := manager.ExtraBitsCube(5, []int{0, 1, 2}, false, options)
	if err != nil || extraBDDTestTable(t, manager, bitsCube) != 1<<5 {
		t.Fatalf("bits cube %#x %v", extraBDDTestTable(t, manager, bitsCube), err)
	}
	and, _ := manager.ExtraNaryAnd(3, options)
	or, _ := manager.ExtraNaryOr(3, options)
	xor, _ := manager.ExtraNaryXor(3, options)
	if extraBDDTestTable(t, manager, and) != 0x80 || extraBDDTestTable(t, manager, or) != 0xfe || extraBDDTestTable(t, manager, xor) != 0x96 {
		t.Fatalf("nary tables %#x %#x %#x", extraBDDTestTable(t, manager, and), extraBDDTestTable(t, manager, or), extraBDDTestTable(t, manager, xor))
	}
	profile, width, err := manager.ExtraProfileWidth(and, -1, options)
	if err != nil || !slices.Equal(profile, []int{1, 2, 2, 1}) || width != 2 {
		t.Fatalf("profile %v width %d err %v", profile, width, err)
	}
	destination, err := NewExtraBDDManager(3, 1024)
	if err != nil {
		t.Fatal(err)
	}
	permutation := []int{2, 0, 1}
	for table := uint64(0); table < 256; table++ {
		root := extraBDDTestRoot(t, manager, table)
		transferred, transferErr := ExtraBDDTransferPermute(manager, destination, root, permutation, ExtraBDDOptions{MaxOperations: 1 << 16})
		if transferErr != nil {
			t.Fatalf("table %#x transfer: %v", table, transferErr)
		}
		var expected uint64
		for assignment := 0; assignment < 8; assignment++ {
			old := 0
			for variable, next := range permutation {
				if assignment&(1<<uint(next)) != 0 {
					old |= 1 << uint(variable)
				}
			}
			if table&(1<<uint(old)) != 0 {
				expected |= 1 << uint(assignment)
			}
		}
		if got := extraBDDTestTable(t, destination, transferred); got != expected {
			t.Fatalf("table %#x transfer got %#x want %#x", table, got, expected)
		}
	}
}

func TestExtraBDDImageExhaustive(t *testing.T) {
	manager, err := NewExtraBDDManager(2, 1024)
	if err != nil {
		t.Fatal(err)
	}
	for careTable := uint64(0); careTable < 16; careTable++ {
		care := extraBDDTestRoot(t, manager, careTable)
		for partTable := uint64(0); partTable < 16; partTable++ {
			part := extraBDDTestRoot(t, manager, partTable)
			image, imageErr := manager.ExtraImage(care, []ExtraBDDLit{part}, []int{0}, ExtraBDDOptions{MaxOperations: 1 << 16, MaxBytes: 1 << 16})
			if imageErr != nil {
				t.Fatalf("care %#x part %#x: %v", careTable, partTable, imageErr)
			}
			var expected uint64
			for assignment := 0; assignment < 4; assignment++ {
				base := assignment &^ 1
				enabled := false
				for value := 0; value < 2; value++ {
					full := base | value
					enabled = enabled || careTable&(1<<uint(full)) != 0 && partTable&(1<<uint(full)) != 0
				}
				if enabled {
					expected |= 1 << uint(assignment)
				}
			}
			if got := extraBDDTestTable(t, manager, image); got != expected {
				t.Fatalf("care %#x part %#x got %#x want %#x", careTable, partTable, got, expected)
			}
		}
	}
}

func extraZDDTestFamily(mask int) []uint64 {
	result := make([]uint64, 0, 4)
	for set := 0; set < 4; set++ {
		if mask&(1<<uint(set)) != 0 {
			result = append(result, uint64(set))
		}
	}
	return result
}

func extraZDDModelFilter(first, second []uint64, subset, negate, maximal bool) []uint64 {
	result := make([]uint64, 0)
	for _, candidate := range first {
		matched := false
		for _, reference := range second {
			if subset {
				matched = matched || extraSetSubset(candidate, reference)
			} else {
				matched = matched || extraSetSubset(reference, candidate)
			}
		}
		if matched != negate {
			result = append(result, candidate)
		}
	}
	if maximal {
		filtered := result[:0]
		for _, candidate := range result {
			keep := true
			for _, other := range result {
				if candidate != other && extraSetSubset(candidate, other) {
					keep = false
				}
			}
			if keep {
				filtered = append(filtered, candidate)
			}
		}
		result = filtered
	}
	slices.Sort(result)
	return result
}

func extraZDDModelReduce(values []uint64, maximal bool) []uint64 {
	unique := make(map[uint64]struct{}, len(values))
	for _, value := range values {
		unique[value] = struct{}{}
	}
	result := make([]uint64, 0, len(unique))
	for candidate := range unique {
		keep := true
		for other := range unique {
			if candidate == other {
				continue
			}
			if maximal && extraSetSubset(candidate, other) || !maximal && extraSetSubset(other, candidate) {
				keep = false
			}
		}
		if keep {
			result = append(result, candidate)
		}
	}
	slices.Sort(result)
	return result
}

func extraZDDAssertSets(t testing.TB, manager *CuddZDDManager, root CuddZDDEdge, want []uint64) {
	t.Helper()
	got, err := manager.ToBitsets(root, CuddZDDOptions{})
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("sets %v want %v err %v", got, want, err)
	}
}

func TestExtraZDDSetAlgebraExhaustive(t *testing.T) {
	manager, err := NewCuddZDDManager(2, 4096)
	if err != nil {
		t.Fatal(err)
	}
	roots := make([]CuddZDDEdge, 16)
	for mask := range roots {
		roots[mask], err = manager.FromBitsets(extraZDDTestFamily(mask), CuddZDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
	}
	for firstMask, firstRoot := range roots {
		first := extraZDDTestFamily(firstMask)
		maximal, maxErr := manager.ExtraMaximal(firstRoot, CuddZDDOptions{})
		minimal, minErr := manager.ExtraMinimal(firstRoot, CuddZDDOptions{})
		if maxErr != nil || minErr != nil {
			t.Fatalf("reduce %d: %v %v", firstMask, maxErr, minErr)
		}
		extraZDDAssertSets(t, manager, maximal, extraZDDModelReduce(first, true))
		extraZDDAssertSets(t, manager, minimal, extraZDDModelReduce(first, false))
		for secondMask, secondRoot := range roots {
			second := extraZDDTestFamily(secondMask)
			cases := []struct {
				root CuddZDDEdge
				want []uint64
				err  error
			}{}
			root, operationErr := manager.ExtraSubsets(firstRoot, secondRoot, CuddZDDOptions{})
			cases = append(cases, struct {
				root CuddZDDEdge
				want []uint64
				err  error
			}{root, extraZDDModelFilter(first, second, true, false, false), operationErr})
			root, operationErr = manager.ExtraSupersets(firstRoot, secondRoot, CuddZDDOptions{})
			cases = append(cases, struct {
				root CuddZDDEdge
				want []uint64
				err  error
			}{root, extraZDDModelFilter(first, second, false, false, false), operationErr})
			root, operationErr = manager.ExtraNotSubsets(firstRoot, secondRoot, CuddZDDOptions{})
			cases = append(cases, struct {
				root CuddZDDEdge
				want []uint64
				err  error
			}{root, extraZDDModelFilter(first, second, true, true, false), operationErr})
			root, operationErr = manager.ExtraNotSupersets(firstRoot, secondRoot, CuddZDDOptions{})
			cases = append(cases, struct {
				root CuddZDDEdge
				want []uint64
				err  error
			}{root, extraZDDModelFilter(first, second, false, true, false), operationErr})
			root, operationErr = manager.ExtraMaxNotSupersets(firstRoot, secondRoot, CuddZDDOptions{})
			cases = append(cases, struct {
				root CuddZDDEdge
				want []uint64
				err  error
			}{root, extraZDDModelFilter(first, second, false, true, true), operationErr})
			for _, test := range cases {
				if test.err != nil {
					t.Fatalf("pair %d %d: %v", firstMask, secondMask, test.err)
				}
				extraZDDAssertSets(t, manager, test.root, test.want)
			}
			unionWant := make(map[uint64]struct{})
			crossWant := make(map[uint64]struct{})
			for _, left := range first {
				for _, right := range second {
					unionWant[left|right] = struct{}{}
					crossWant[left&right] = struct{}{}
				}
			}
			mapSets := func(values map[uint64]struct{}) []uint64 {
				sets := make([]uint64, 0, len(values))
				for set := range values {
					sets = append(sets, set)
				}
				slices.Sort(sets)
				return sets
			}
			dot, dotErr := manager.ExtraDotProduct(firstRoot, secondRoot, CuddZDDOptions{})
			cross, crossErr := manager.ExtraCrossProduct(firstRoot, secondRoot, CuddZDDOptions{})
			maxDot, maxDotErr := manager.ExtraMaxDotProduct(firstRoot, secondRoot, CuddZDDOptions{})
			maxUnion, maxUnionErr := manager.ExtraMaxUnion(firstRoot, secondRoot, CuddZDDOptions{})
			minUnion, minUnionErr := manager.ExtraMinUnion(firstRoot, secondRoot, CuddZDDOptions{})
			if dotErr != nil || crossErr != nil || maxDotErr != nil || maxUnionErr != nil || minUnionErr != nil {
				t.Fatalf("products %d %d: %v %v %v %v %v", firstMask, secondMask, dotErr, crossErr, maxDotErr, maxUnionErr, minUnionErr)
			}
			dotWant := mapSets(unionWant)
			extraZDDAssertSets(t, manager, dot, dotWant)
			extraZDDAssertSets(t, manager, cross, mapSets(crossWant))
			extraZDDAssertSets(t, manager, maxDot, extraZDDModelReduce(dotWant, true))
			combined := append(append([]uint64(nil), first...), second...)
			extraZDDAssertSets(t, manager, maxUnion, extraZDDModelReduce(combined, true))
			extraZDDAssertSets(t, manager, minUnion, extraZDDModelReduce(combined, false))
		}
	}
}

func TestExtraZDDTuplesSelectionAndMembership(t *testing.T) {
	manager, err := NewCuddZDDManager(5, 1024)
	if err != nil {
		t.Fatal(err)
	}
	tuples, err := manager.ExtraTuples([]int{0, 2, 4}, 2, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	extraZDDAssertSets(t, manager, tuples, []uint64{5, 17, 20})
	one, err := manager.ExtraSelectOne(tuples, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	extraZDDAssertSets(t, manager, one, []uint64{5})
	empty, err := manager.FromBitsets([]uint64{0, 3}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	belongs, err := manager.ExtraEmptyBelongs(empty, CuddZDDOptions{})
	if err != nil || !belongs {
		t.Fatalf("empty belongs %v %v", belongs, err)
	}
	single, err := manager.ExtraIsOneSubset(one, CuddZDDOptions{})
	if err != nil || !single {
		t.Fatalf("single %v %v", single, err)
	}
	emptyFamily, err := manager.FromBitsets(nil, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	single, err = manager.ExtraIsOneSubset(emptyFamily, CuddZDDOptions{})
	if err != nil || single {
		t.Fatalf("empty family is not one subset: %v %v", single, err)
	}
}

func TestExtraBDDThresholdGrayAndKMap(t *testing.T) {
	values := make([]bool, 8)
	for assignment := range values {
		values[assignment] = bitsCount(uint(assignment)) >= 2
	}
	constant, coefficients, err := ExtraBDDThresholdChow(values, 3, ExtraBDDOptions{})
	if err != nil || constant != 0 || !slices.Equal(coefficients, []int{2, 2, 2}) {
		t.Fatalf("chow %d %v %v", constant, coefficients, err)
	}
	valid, err := ExtraBDDThresholdCheck(values, []int{1, 1, 1}, 2, ExtraBDDOptions{})
	if err != nil || !valid {
		t.Fatalf("threshold %v %v", valid, err)
	}
	for value := uint64(0); value < 1024; value++ {
		if got := ExtraBDDBinaryCode(ExtraBDDGrayCode(value)); got != value {
			t.Fatalf("gray %d -> %d", value, got)
		}
	}
	manager, _ := NewExtraBDDManager(2, 64)
	root := extraBDDTestRoot(t, manager, 0x8)
	text, err := manager.ExtraKMap(root, ExtraBDDOptions{})
	if err != nil || text != "0 0\n0 1\n" {
		t.Fatalf("kmap %q %v", text, err)
	}
}

func TestExtraBDDAtomicLimitsCancellationAndValidation(t *testing.T) {
	manager, _ := NewExtraBDDManager(4, 256)
	root := extraBDDTestRoot(t, manager, 0xa55a)
	before := len(manager.nodes)
	if result, err := manager.ExtraImage(root, []ExtraBDDLit{root}, []int{0, 1}, ExtraBDDOptions{MaxOperations: 1}); result != 0 || !errors.Is(err, ErrLimit) || len(manager.nodes) != before {
		t.Fatalf("image rollback %d %v %d/%d", result, err, len(manager.nodes), before)
	}
	if result, err := manager.ExtraNaryXor(4, ExtraBDDOptions{Stop: func() bool { return true }}); result != 0 || !errors.Is(err, ErrCancelled) || len(manager.nodes) != before {
		t.Fatalf("xor rollback %d %v", result, err)
	}
	if _, err := manager.ExtraBitsCube(0, []int{0, 0}, false, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("duplicate variables: %v", err)
	}
	if _, err := ExtraBDDTransferPermute(manager, manager, root, []int{0, 0, 2, 3}, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad permutation: %v", err)
	}
	if _, err := (*ExtraBDDManager)(nil).ExtraNaryAnd(0, ExtraBDDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil manager: %v", err)
	}
	zdd, _ := NewCuddZDDManager(4, 256)
	family, _ := zdd.FromBitsets([]uint64{1, 2, 3, 7}, CuddZDDOptions{})
	zddBefore := len(zdd.nodes)
	if result, err := zdd.ExtraMaximal(family, CuddZDDOptions{MaxOperations: 1}); result != 0 || !errors.Is(err, ErrLimit) || len(zdd.nodes) != zddBefore {
		t.Fatalf("zdd rollback %d %v %d/%d", result, err, len(zdd.nodes), zddBefore)
	}
}

func TestExtraBDDRejectsMalformedGraphsAndWideCubeCounts(t *testing.T) {
	bdd, err := NewExtraBDDManager(2, 32)
	if err != nil {
		t.Fatal(err)
	}
	root := ExtraBDDLit(1 << 1)
	bdd.nodes[root.ID()].then = ExtraBDDLit(999 << 1)
	if _, err := bdd.ExtraSupport(root, ExtraBDDOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("malformed BDD child: %v", err)
	}

	wide, err := NewExtraBDDManager(64, 128)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wide.ExtraOneCube(ExtraBDDTrue, ExtraBDDOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("wide minterm count: %v", err)
	}

	zdd, err := NewCuddZDDManager(3, 64)
	if err != nil {
		t.Fatal(err)
	}
	family, err := zdd.FromBitsets([]uint64{1, 3}, CuddZDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if family <= CuddZDDBase {
		t.Fatal("expected nonterminal family")
	}
	zdd.nodes[family].low = CuddZDDEdge(999)
	if _, err := zdd.ExtraMaximal(family, CuddZDDOptions{}); !errors.Is(err, ErrInvariant) && !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("malformed ZDD child: %v", err)
	}
}

func TestExtraBDDDeterministicModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 { state ^= state << 7; state ^= state >> 9; state ^= state << 8; return state }
	manager, _ := NewExtraBDDManager(4, 40000)
	for trial := 0; trial < 5000; trial++ {
		table := next() & 0xffff
		root := extraBDDTestRoot(t, manager, table)
		support, err := manager.ExtraSupport(root, ExtraBDDOptions{})
		if err != nil || !slices.Equal(support, extraBDDExpectedSupport(table, 4)) {
			t.Fatalf("trial %d support %v %v", trial, support, err)
		}
	}
}

func FuzzExtraBDDContracts(f *testing.F) {
	f.Add(uint16(0xa55a), uint8(0x39), uint8(0xc3))
	f.Add(uint16(0), uint8(0), uint8(0xff))
	f.Fuzz(func(t *testing.T, truth uint16, firstMask, secondMask uint8) {
		manager, err := NewExtraBDDManager(4, 512)
		if err != nil {
			t.Fatal(err)
		}
		root := extraBDDTestRoot(t, manager, uint64(truth))
		support, err := manager.ExtraSupport(root, ExtraBDDOptions{MaxOperations: 1 << 16})
		if err != nil || !slices.Equal(support, extraBDDExpectedSupport(uint64(truth), 4)) {
			t.Fatalf("support %v %v", support, err)
		}
		unate, err := manager.ExtraUnateness(root, ExtraBDDOptions{MaxOperations: 1 << 16})
		if err != nil || len(unate) != len(support) {
			t.Fatalf("unate %v %v", unate, err)
		}
		zdd, err := NewCuddZDDManager(3, 512)
		if err != nil {
			t.Fatal(err)
		}
		toFamily := func(mask uint8) []uint64 {
			sets := make([]uint64, 0, 8)
			for set := 0; set < 8; set++ {
				if mask&(1<<uint(set)) != 0 {
					sets = append(sets, uint64(set))
				}
			}
			return sets
		}
		first, err := zdd.FromBitsets(toFamily(firstMask), CuddZDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		second, err := zdd.FromBitsets(toFamily(secondMask), CuddZDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		got, err := zdd.ExtraDotProduct(first, second, CuddZDDOptions{MaxOperations: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		sets, err := zdd.ToBitsets(got, CuddZDDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for _, set := range sets {
			found := false
			for _, left := range toFamily(firstMask) {
				for _, right := range toFamily(secondMask) {
					found = found || set == left|right
				}
			}
			if !found {
				t.Fatalf("unexpected dot set %d", set)
			}
		}
	})
}

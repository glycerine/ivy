package abc

// Go-only semantic tests for ABC names, latches, fanin ordering and SOPs.

import (
	"errors"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

func abcMetadataIndependentSOP(t *testing.T, text string, assignment uint64) bool {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	covered := false
	phase := byte(0)
	for _, line := range lines {
		if len(line) < 2 {
			t.Fatalf("bad independent SOP %q", text)
		}
		cube := line[:len(line)-2]
		phase = line[len(line)-1]
		match := true
		for variable := range cube {
			value := byte('0' + ((assignment >> uint(variable)) & 1))
			if cube[variable] != '-' && cube[variable] != value {
				match = false
			}
		}
		covered = covered || match
	}
	if phase == '0' || phase == 'n' {
		return !covered
	}
	return covered
}

func checkABCSOPIndependent(t *testing.T, sop ABCSOP) {
	t.Helper()
	text, err := sop.Format(ABCMetadataOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(sop.Variables); assignment++ {
		got, err := sop.Evaluate(assignment)
		if err != nil {
			t.Fatal(err)
		}
		if want := abcMetadataIndependentSOP(t, text, assignment); got != want {
			t.Fatalf("SOP %q assignment %d = %v, want %v", text, assignment, got, want)
		}
	}
}

func TestABCSOPPrimitivesParsingAndTruth(t *testing.T) {
	and, _ := ABCSOPAnd(3, []bool{false, true, false}, false)
	nand, _ := ABCSOPAnd(3, nil, true)
	or, _ := ABCSOPOr(3, []bool{false, true, false}, false)
	nor, _ := ABCSOPOr(3, nil, true)
	cases := []ABCSOP{ABCSOPConst(false), ABCSOPConst(true), and, nand, or, nor, ABCSOPXOR(), ABCSOPXNOR(), ABCSOPMux()}
	for _, sop := range cases {
		text, err := sop.Format(ABCMetadataOptions{})
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseABCSOP(text, ABCMetadataOptions{})
		if err != nil || parsed.Variables != sop.Variables || parsed.Phase != sop.Phase || !slices.Equal(parsed.Cubes, sop.Cubes) {
			t.Fatalf("roundtrip %+v -> %q -> %+v, %v", sop, text, parsed, err)
		}
		checkABCSOPIndependent(t, parsed)
	}
	if !ABCSOPConst(false).IsConst0() || !ABCSOPConst(true).IsConst1() || !(ABCSOP{Variables: 1, Cubes: []string{"1"}, Phase: '1'}).IsBuffer() || !(ABCSOP{Variables: 1, Cubes: []string{"0"}, Phase: '1'}).IsInverter() {
		t.Fatal("SOP recognizers")
	}
	if !and.IsANDType() || !(ABCSOP{Variables: 2, Cubes: []string{"1-", "-1"}, Phase: '1'}).IsORType() {
		t.Fatal("AND/OR recognition")
	}
	xorMarker, err := ParseABCSOP("-- x\n", ABCMetadataOptions{})
	if err != nil || !xorMarker.IsXORType() {
		t.Fatalf("special XOR marker: %+v, %v", xorMarker, err)
	}
	nxorMarker := xorMarker.Complement()
	if nxorMarker.Phase != 'n' {
		t.Fatalf("special XOR complement phase %q", nxorMarker.Phase)
	}
	for assignment := uint64(0); assignment < 4; assignment++ {
		x, _ := xorMarker.Evaluate(assignment)
		n, _ := nxorMarker.Evaluate(assignment)
		if x == n {
			t.Fatalf("special phases did not complement at assignment %d", assignment)
		}
	}
}

func TestABCSOPExhaustiveTruthISOPAndTransforms(t *testing.T) {
	for variables := 0; variables <= 3; variables++ {
		functions := 1 << uint(1<<uint(variables))
		for function := 0; function < functions; function++ {
			words := []uint64{uint64(function)}
			sop, err := ABCSOPFromTruth(variables, words, ABCMetadataOptions{})
			if err != nil {
				t.Fatalf("vars=%d function=%x: %v", variables, function, err)
			}
			for assignment := 0; assignment < 1<<uint(variables); assignment++ {
				evalAssignment := uint64(assignment)
				if sop.Variables == 0 {
					evalAssignment = 0
				}
				got, evalErr := sop.Evaluate(evalAssignment)
				want := function&(1<<uint(assignment)) != 0
				if evalErr != nil || got != want {
					t.Fatalf("vars=%d function=%x assignment=%d got=%v/%v want=%v cover=%+v", variables, function, assignment, got, evalErr, want, sop)
				}
			}
			if sop.Variables > 0 {
				isop, _ := sop.ISOP()
				rebuilt, rebuildErr := ABCSOPFromISOP(sop.Variables, isop, sop.Phase == '0')
				if rebuildErr != nil || !slices.Equal(rebuilt.Cubes, sop.Cubes) || rebuilt.Phase != sop.Phase {
					t.Fatalf("ISOP roundtrip %+v -> %v -> %+v/%v", sop, isop, rebuilt, rebuildErr)
				}
			}
		}
	}
	sop := ABCSOP{Variables: 3, Cubes: []string{"10-", "-01"}, Phase: '1'}
	complemented := sop.Complement()
	flipped, err := sop.ComplementVariable(1)
	if err != nil {
		t.Fatal(err)
	}
	permuted, err := sop.Permute([]int{2, 0, 1})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		original, _ := sop.Evaluate(assignment)
		inverse, _ := complemented.Evaluate(assignment)
		if inverse == original {
			t.Fatal("cover complement")
		}
		flippedValue, _ := flipped.Evaluate(assignment)
		if source, _ := sop.Evaluate(assignment ^ 2); flippedValue != source {
			t.Fatal("variable complement")
		}
		// New variables [old2,old0,old1].
		oldAssignment := ((assignment >> 1) & 1) | ((assignment >> 2) & 1 << 1) | ((assignment & 1) << 2)
		permutedValue, _ := permuted.Evaluate(assignment)
		oldValue, _ := sop.Evaluate(oldAssignment)
		if permutedValue != oldValue {
			t.Fatalf("permutation assignment %d old %d", assignment, oldAssignment)
		}
	}
}

func TestABCSOPNormalizeLimitsAndMalformed(t *testing.T) {
	original := ABCSOP{Variables: 3, Cubes: []string{"100", "101", "1--", "-10"}, Phase: '1'}
	legit, changed, err := original.Legit(ABCMetadataOptions{})
	if err != nil || !changed || !slices.Equal(legit.Cubes, []string{"1--", "-10"}) {
		t.Fatalf("legit %+v changed=%v err=%v", legit, changed, err)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		before, _ := original.Evaluate(assignment)
		after, _ := legit.Evaluate(assignment)
		if before != after {
			t.Fatalf("normalization changed assignment %d", assignment)
		}
	}
	weighted := ABCSOP{Variables: 3, Cubes: []string{"---", "1--", "101"}, Phase: '1'}.SortCubes(true)
	if !slices.Equal(weighted.Cubes, []string{"101", "1--", "---"}) {
		t.Fatal(weighted.Cubes)
	}
	for _, text := range []string{"", "1 1", "1 1\n0 0\n", "2 1\n", "1 x\n1 n\n", "1 1\n\n"} {
		if _, err := ParseABCSOP(text, ABCMetadataOptions{}); err == nil {
			t.Fatalf("accepted malformed %q", text)
		}
	}
	if _, err := ParseABCSOP("1 1\n", ABCMetadataOptions{MaxBytes: 3}); !errors.Is(err, ErrFormat) && !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit %v", err)
	}
	if _, _, err := original.Legit(ABCMetadataOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("operation limit %v", err)
	}
	if _, _, err := original.Legit(ABCMetadataOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	maxInt := int(^uint(0) >> 1)
	if err := (ABCSOP{Variables: maxInt, Cubes: []string{""}, Phase: '1'}).Check(ABCMetadataOptions{MaxVariables: maxInt, MaxBytes: maxInt}); !errors.Is(err, ErrLimit) {
		t.Fatalf("overflowing SOP size = %v", err)
	}
}

func TestABCNamesOrderingTransferAndShortening(t *testing.T) {
	network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	p0, _ := network.CreatePI()
	p1, _ := network.CreatePI()
	po, _ := network.CreatePO()
	network.SetObjectName(p0, "zeta")
	network.SetObjectName(p1, "alpha")
	network.SetObjectName(po, "out")
	ordered, err := ABCOrderedObjectsByName(network, []int{p0, p1})
	if err != nil || !slices.Equal(ordered, []int{p1, p0}) || !slices.Equal(network.PIs(), []int{p0, p1}) {
		t.Fatalf("order=%v err=%v PIs=%v", ordered, err, network.PIs())
	}
	fake, _ := ABCDummyName("pi", 3, 2)
	if fake != "pi03" {
		t.Fatal(fake)
	}
	destination, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	d0, _ := destination.CreatePI()
	d1, _ := destination.CreatePI()
	if err = ABCTransferNames(network, destination, []int{p0, p1}, []int{d0, d1}, "_copy"); err != nil {
		t.Fatal(err)
	}
	if d, _ := destination.Object(d0); d.Name != "zeta_copy" {
		t.Fatal(d.Name)
	}
	before := destination.Clone()
	if err = ABCTransferNames(network, destination, []int{p0, p0}, []int{d0, d1}, "_same"); err == nil {
		t.Fatal("expected duplicate transfer failure")
	}
	if destination.objects[d0].Name != before.objects[d0].Name || destination.objects[d1].Name != before.objects[d1].Name {
		t.Fatal("transfer was not atomic")
	}
	if err = ABCShortNames(network); err != nil {
		t.Fatal(err)
	}
	if a, _ := network.Object(p0); a.Name != "pi0" {
		t.Fatal(a.Name)
	}
	if b, _ := network.Object(po); b.Name != "po0" {
		t.Fatal(b.Name)
	}

	withLatch, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	wpi, _ := withLatch.CreatePI()
	triple, _ := ABCAddNamedLatch(withLatch, ABCObjectEdge{}, false, ABCInitZero, "state")
	wpo, _ := withLatch.CreatePO()
	withLatch.SetObjectName(wpi, "ci0")
	withLatch.SetObjectName(triple.Output, "ci1")
	withLatch.SetObjectName(triple.Input, "co0")
	withLatch.SetObjectName(wpo, "co1")
	withoutLatch, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	for _, name := range []string{"ci0", "ci1"} {
		id, _ := withoutLatch.CreatePI()
		withoutLatch.SetObjectName(id, name)
	}
	for _, name := range []string{"co0", "co1"} {
		id, _ := withoutLatch.CreatePO()
		withoutLatch.SetObjectName(id, name)
	}
	if same, compareErr := ABCCompareTerminalNames(withLatch, withoutLatch); compareErr != nil || same {
		t.Fatalf("partition-mismatched terminals same=%v err=%v", same, compareErr)
	}
}

func TestABCLatchValuesNamesSelfFeedAndRemoval(t *testing.T) {
	network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	first, err := ABCAddNamedLatch(network, ABCObjectEdge{}, false, ABCInitZero, "state0")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ABCAddNamedLatch(network, ABCObjectEdge{}, false, ABCInitOne, "state1")
	if err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(first.Input, ABCObjectEdge{ID: second.Output}); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(second.Input, ABCObjectEdge{ID: first.Output}); err != nil {
		t.Fatal(err)
	}
	self, err := ABCSelfFeedLatches(network)
	if err != nil || !slices.Equal(self, []int{first.Latch, second.Latch}) {
		t.Fatalf("self=%v err=%v", self, err)
	}
	if values, _ := ABCLatchValueString(network); values != "01" {
		t.Fatal(values)
	}
	if err = ABCSetLatchValues(network, []ABCLatchInit{ABCInitDC, ABCInitOne}); err != nil {
		t.Fatal(err)
	}
	if values, _ := ABCLatchValueString(network); values != "x1" {
		t.Fatal(values)
	}
	if count, removeErr := ABCRemoveSelfFeedLatches(network); removeErr != nil || count != 1 {
		t.Fatalf("removed=%d err=%v", count, removeErr)
	}
	if self, _ = ABCSelfFeedLatches(network); len(self) != 0 {
		t.Fatal(self)
	}
	firstInput, _ := network.Object(first.Input)
	firstDriver, _ := network.Object(firstInput.Fanins[0].ID)
	if firstDriver.Function != " 1\n" {
		t.Fatalf("first driver %+v", firstDriver)
	}
	secondInput, _ := network.Object(second.Input)
	if secondInput.Fanins[0].ID != first.Output {
		t.Fatalf("second driver changed: %+v", secondInput.Fanins)
	}
	if err = network.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestABCMetadataCancellationAndNetworkAtomicity(t *testing.T) {
	calls := 0
	network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{Stop: func() bool {
		calls++
		return calls >= 3
	}})
	// Construct directly while the stop callback is temporarily disabled.
	stop := network.limits.Stop
	network.limits.Stop = nil
	_, latch0, _, _ := network.CreateLatchBox(ABCInitZero)
	_, latch1, _, _ := network.CreateLatchBox(ABCInitOne)
	network.limits.Stop = stop
	before0, before1 := network.objects[latch0].Init, network.objects[latch1].Init
	if err := ABCSetLatchValues(network, []ABCLatchInit{ABCInitOne, ABCInitZero}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel=%v", err)
	}
	if network.objects[latch0].Init != before0 || network.objects[latch1].Init != before1 {
		t.Fatal("cancelled latch update was not atomic")
	}

	shortCalls := 0
	short, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{Stop: func() bool {
		shortCalls++
		return shortCalls >= 4
	}})
	short.limits.Stop = nil
	shortPI, _ := short.CreatePI()
	shortPO, _ := short.CreatePO()
	short.SetObjectName(shortPI, "input")
	short.SetObjectName(shortPO, "output")
	short.limits.Stop = func() bool {
		shortCalls++
		return shortCalls >= 4
	}
	if err := ABCShortNames(short); !errors.Is(err, ErrCancelled) {
		t.Fatalf("short-name cancellation = %v", err)
	}
	if short.objects[shortPI].Name != "input" || short.objects[shortPO].Name != "output" {
		t.Fatal("cancelled short names were not atomic")
	}

	ordered, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	p0, _ := ordered.CreatePI()
	p1, _ := ordered.CreatePI()
	first, _ := ordered.CreateLogicNode("10 1\n", ABCObjectEdge{ID: p1}, ABCObjectEdge{ID: p0})
	_, _ = ordered.CreateLogicNode("broken", ABCObjectEdge{ID: p1}, ABCObjectEdge{ID: p0})
	firstBefore, _ := ordered.Object(first)
	if err := ABCOrderNetworkFanins(ordered, ABCFaninByID); err == nil {
		t.Fatal("expected malformed second SOP")
	}
	firstAfter, _ := ordered.Object(first)
	if !slices.Equal(firstAfter.Fanins, firstBefore.Fanins) || firstAfter.Function != firstBefore.Function {
		t.Fatal("network fanin ordering was not atomic")
	}
}

func abcMetadataEvalNode(t *testing.T, network *ABCNetwork, node int, values map[int]bool) bool {
	t.Helper()
	object, err := network.Object(node)
	if err != nil {
		t.Fatal(err)
	}
	sop, err := ParseABCSOP(object.Function, ABCMetadataOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assignment := uint64(0)
	for variable, edge := range object.Fanins {
		value := values[edge.ID] != edge.Complemented
		if value {
			assignment |= uint64(1) << uint(variable)
		}
	}
	result, _ := sop.Evaluate(assignment)
	return result
}

func TestABCFaninOrderingAndSOPNormalization(t *testing.T) {
	network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	p0, _ := network.CreatePI()
	p1, _ := network.CreatePI()
	p2, _ := network.CreatePI()
	node, err := network.CreateLogicNode("1-0 1\n-11 1\n", ABCObjectEdge{ID: p2}, ABCObjectEdge{ID: p0}, ABCObjectEdge{ID: p1})
	if err != nil {
		t.Fatal(err)
	}
	want := make([]bool, 8)
	for assignment := 0; assignment < 8; assignment++ {
		values := map[int]bool{p0: assignment&1 != 0, p1: assignment&2 != 0, p2: assignment&4 != 0}
		want[assignment] = abcMetadataEvalNode(t, network, node, values)
	}
	if err = ABCOrderNodeFanins(network, node, ABCFaninByID); err != nil {
		t.Fatal(err)
	}
	object, _ := network.Object(node)
	if ids := []int{object.Fanins[0].ID, object.Fanins[1].ID, object.Fanins[2].ID}; !slices.Equal(ids, []int{p0, p1, p2}) {
		t.Fatal(ids)
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := map[int]bool{p0: assignment&1 != 0, p1: assignment&2 != 0, p2: assignment&4 != 0}
		if got := abcMetadataEvalNode(t, network, node, values); got != want[assignment] {
			t.Fatalf("assignment %d got=%v want=%v", assignment, got, want[assignment])
		}
	}
	duplicate, _ := network.CreateLogicNode("100 1\n101 1\n1-- 1\n", ABCObjectEdge{ID: p0}, ABCObjectEdge{ID: p1}, ABCObjectEdge{ID: p2})
	changed, err := ABCNormalizeNodeSOP(network, duplicate, false, ABCMetadataOptions{})
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	normalized, _ := network.Object(duplicate)
	if normalized.Function != "1-- 1\n" {
		t.Fatalf("normalized %q", normalized.Function)
	}
}

func TestABCFaninColumnAndLiteralOrderingPreserveFunction(t *testing.T) {
	for _, test := range []struct {
		mode ABCFaninOrder
		want []int
	}{
		{mode: ABCFaninByColumn, want: []int{1, 2, 0}},
		{mode: ABCFaninByLiteralCount, want: []int{1, 0, 2}},
	} {
		network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
		pis := make([]int, 3)
		for i := range pis {
			pis[i], _ = network.CreatePI()
		}
		node, err := network.CreateLogicNode("1-0 1\n-11 1\n",
			ABCObjectEdge{ID: pis[0], Complemented: true},
			ABCObjectEdge{ID: pis[1]}, ABCObjectEdge{ID: pis[2]})
		if err != nil {
			t.Fatal(err)
		}
		before := make([]bool, 8)
		for assignment := range before {
			values := map[int]bool{
				pis[0]: assignment&1 != 0,
				pis[1]: assignment&2 != 0,
				pis[2]: assignment&4 != 0,
			}
			before[assignment] = abcMetadataEvalNode(t, network, node, values)
		}
		if err := ABCOrderNodeFanins(network, node, test.mode); err != nil {
			t.Fatal(err)
		}
		object, _ := network.Object(node)
		for i, old := range test.want {
			if object.Fanins[i].ID != pis[old] || object.Fanins[i].Complemented != (old == 0) {
				t.Fatalf("mode %d fanin %d = %+v, want old %d", test.mode, i, object.Fanins[i], old)
			}
		}
		for assignment := range before {
			values := map[int]bool{
				pis[0]: assignment&1 != 0,
				pis[1]: assignment&2 != 0,
				pis[2]: assignment&4 != 0,
			}
			if got := abcMetadataEvalNode(t, network, node, values); got != before[assignment] {
				t.Fatalf("mode %d assignment %d = %v, want %v", test.mode, assignment, got, before[assignment])
			}
		}
	}
}

func TestABCMetadataDeterministicModel(t *testing.T) {
	random := rand.New(rand.NewSource(0xabc7))
	for iteration := 0; iteration < 5000; iteration++ {
		variables := 1 + random.Intn(5)
		cubes := 1 + random.Intn(8)
		cover := ABCSOP{Variables: variables, Cubes: make([]string, cubes), Phase: byte("01"[random.Intn(2)])}
		for i := range cover.Cubes {
			cube := make([]byte, variables)
			for j := range cube {
				cube[j] = "01-"[random.Intn(3)]
			}
			cover.Cubes[i] = string(cube)
		}
		text, err := cover.Format(ABCMetadataOptions{})
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseABCSOP(text, ABCMetadataOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
			got, _ := parsed.Evaluate(assignment)
			if want := abcMetadataIndependentSOP(t, text, assignment); got != want {
				t.Fatalf("iteration %d assignment %d", iteration, assignment)
			}
		}
		legit, _, err := parsed.Legit(ABCMetadataOptions{MaxOperations: 10000})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
			before, _ := parsed.Evaluate(assignment)
			after, _ := legit.Evaluate(assignment)
			if before != after {
				t.Fatalf("legit iteration %d assignment %d", iteration, assignment)
			}
		}
	}
}

func FuzzABCMetadataSOP(f *testing.F) {
	f.Add([]byte{2, 2, 0, 1, 2, 1, 0})
	f.Add([]byte{3, 1, 1, 1, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		variables := 1 + int(data[0]%5)
		cubeCount := 1 + int(data[1]%8)
		cover := ABCSOP{Variables: variables, Cubes: make([]string, cubeCount), Phase: '1'}
		if data[0]&0x80 != 0 {
			cover.Phase = '0'
		}
		position := 2
		for i := range cover.Cubes {
			cube := make([]byte, variables)
			for j := range cube {
				value := byte(0)
				if position < len(data) {
					value = data[position]
					position++
				}
				cube[j] = "01-"[value%3]
			}
			cover.Cubes[i] = string(cube)
		}
		text, err := cover.Format(ABCMetadataOptions{MaxCubes: 8, MaxVariables: 5, MaxOperations: 10000})
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseABCSOP(text, ABCMetadataOptions{MaxCubes: 8, MaxVariables: 5, MaxOperations: 10000})
		if err != nil {
			t.Fatal(err)
		}
		legit, _, err := parsed.Legit(ABCMetadataOptions{MaxCubes: 8, MaxVariables: 5, MaxOperations: 10000})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
			value, _ := parsed.Evaluate(assignment)
			if value != abcMetadataIndependentSOP(t, text, assignment) {
				t.Fatal("model disagreement")
			}
			reduced, _ := legit.Evaluate(assignment)
			if reduced != value {
				t.Fatal("normalization changed function")
			}
		}
	})
}

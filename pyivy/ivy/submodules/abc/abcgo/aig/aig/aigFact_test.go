package aig

import "testing"

func factValue(p *Graph, e Edge, inputs []bool) bool {
	return aigUtilEdgeValue(e, aigSimulate(p, aigUtilOrder(p), inputs))
}
func TestAlgebraicFactoringImplications(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	condition := p.AND(a, b)
	function := p.AND(a, c)
	p.AddOutput(condition)
	p.AddOutput(function)
	result := Aig_ManFactorAlgebraic(p, 0, 1, 1)
	if result != c {
		t.Fatal("condition implications not propagated", result, c)
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		if factValue(p, condition, inputs) && factValue(p, result, inputs) != factValue(p, function, inputs) {
			t.Fatal("factoring changed cared function")
		}
	}
	nested := p.AND(condition, c)
	p.AddOutput(nested)
	zero := Aig_ManFactorAlgebraic(p, 0, 2, 0)
	if zero != p.Const0() {
		t.Fatal("zero condition was not substituted", zero)
	}
	p.AddOutput(p.Const1())
	if Aig_ManFactorAlgebraic(p, 3, 1, 1) != InvalidEdge {
		t.Fatal("constant condition accepted")
	}
}
func TestSupportMinimizationSourceExamples(t *testing.T) {
	first, result, status := Aig_ManSupportMinimizationTest()
	if status != 0 || result == InvalidEdge {
		t.Fatal("first source example", status, result)
	}
	function := first.Mux(Aig_IthVar(first, 3), Aig_IthVar(first, 1), Aig_IthVar(first, 0))
	function = first.Mux(Aig_IthVar(first, 4), Aig_IthVar(first, 2), function)
	condition := first.OR(Aig_IthVar(first, 3), Aig_IthVar(first, 4))
	for mask := 0; mask < 32; mask++ {
		inputs := make([]bool, 5)
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		if factValue(first, condition, inputs) && factValue(first, function, inputs) != factValue(first, result, inputs) {
			t.Fatalf("source example 1 changed care assignment %d", mask)
		}
	}
	second, result, status := Aig_ManSupportMinimizationTest2()
	if status != 0 || result == InvalidEdge {
		t.Fatal("second source example", status, result)
	}
	a, b, c := Aig_IthVar(second, 0), Aig_IthVar(second, 1), Aig_IthVar(second, 2)
	function = second.AND(second.AND(second.AND(a, b.Not()).Not(), c.Not()), b)
	condition = second.OR(second.OR(b, c), a)
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		if factValue(second, condition, inputs) && factValue(second, function, inputs) != factValue(second, result, inputs) {
			t.Fatalf("source example 2 changed care assignment %d", mask)
		}
	}
}
func TestSupportCofactorsCheckedProgressively(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	care := p.AND(a, b)
	function := p.AND(a.Not(), b.Not())
	nodes := []Edge{care, function}
	support := []Edge{a, b}
	cofactors := Aig_SuppMinPerform(p, []Edge{care}, nodes, support)
	result := Aig_SuppMinReconstruct(p, cofactors, nodes, support)
	// Both variables can individually be fixed to zero on care=a&&b, but
	// simultaneously doing so changes !a&&!b from zero to one on the care point.
	if factValue(p, result, []bool{true, true}) != factValue(p, function, []bool{true, true}) {
		t.Fatal("individually legal cofactors were combined unsafely", cofactors)
	}
}
func TestSupportMinimizationStatusesAndPolarity(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	function := p.Mux(b, a, c)
	if e, status := Aig_ManSupportMinimization(p, a, function); e != InvalidEdge || status != -1 {
		t.Fatal(e, status)
	}
	condition := p.OR(b, c)
	if e, status := Aig_ManSupportMinimization(p, condition, a); e != InvalidEdge || status != -2 {
		t.Fatal(e, status)
	}
	// A complement on the function must survive cofactoring and reconstruction.
	result, status := Aig_ManSupportMinimization(p, condition, function.Not())
	if status == 0 {
		for mask := 0; mask < 8; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			if factValue(p, condition, inputs) && factValue(p, result, inputs) != factValue(p, function.Not(), inputs) {
				t.Fatal("function complement lost")
			}
		}
	}
	independent := p.NewInput()
	p.AddOutput(independent)
	if Aig_ManFindConeOverlap(p, []Edge{a}, independent) {
		t.Fatal("disjoint cone overlaps")
	}
}

func TestSupportMinimizationNativeXorAndDontCareInput(t *testing.T) {
	p := NewGraph()
	a, b, c, d := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	condition := p.OR(a, b)
	function := p.Mux(condition, p.NewXor(a, c), d)
	result, status := Aig_ManSupportMinimization(p, condition, function)
	if status != 0 {
		t.Fatal("native XOR support minimization failed", status)
	}
	for mask := 0; mask < 16; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		if factValue(p, condition, inputs) && factValue(p, result, inputs) != factValue(p, function, inputs) {
			t.Fatal("XOR reconstruction changed cared assignment", mask)
		}
	}
	// Input d can be removed: it is selected only outside the care set.
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, false}
		zero := factValue(p, result, inputs)
		inputs[3] = true
		if factValue(p, result, inputs) != zero {
			t.Fatal("don't-care input not removed")
		}
	}
}
func TestFactoringHelpersRespectTraversalAndCopies(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(a, p.AND(b, c))
	if !Aig_ManFindConeOverlap(p, []Edge{b}, root) {
		t.Fatal("overlap missed")
	}
	result := Aig_ManDeriveNewCone(p, []Edge{a.Not()}, root)
	if result != p.Const0() {
		t.Fatal("negative implication not applied")
	}
	before := p.travID
	if !Aig_SuppMinGateIsInSupport(p, []Edge{a.Not(), b}, []Edge{a, b, c}) || p.travID != before+1 || p.node(a).TravID != p.travID {
		t.Fatal("public traversal protocol changed")
	}
}

func TestFactoringTraversalSeedsSurviveGenerationRollover(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(a, p.AND(b, c))
	p.travID = (1 << 30) - 2
	if !Aig_ManFindConeOverlap(p, []Edge{b}, root) {
		t.Fatal("rollover discarded overlap seeds")
	}
	p.travID = (1 << 30) - 3
	if !Aig_SuppMinHighlightCone(p, root, []Edge{a, b}) {
		t.Fatal("rollover lost care frontier")
	}
	nodes := Aig_SuppMinCollectCone(p, root)
	if len(nodes) != 2 {
		t.Fatal("rollover discarded highlighted internal nodes", nodes)
	}
}

func TestSupportMinimizationExhaustiveRandomCones(t *testing.T) {
	random := NewAigRandomSource()
	successes := 0
	for iteration := 0; iteration < 100; iteration++ {
		p := NewGraph()
		inputs := []Edge{p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()}
		pool := append([]Edge(nil), inputs...)
		for i := 0; i < 12; i++ {
			a, b := pool[int(random.Uint32())%len(pool)], pool[int(random.Uint32())%len(pool)]
			if random.Uint32()&1 != 0 {
				a = a.Not()
			}
			if random.Uint32()&1 != 0 {
				b = b.Not()
			}
			var e Edge
			switch random.Uint32() % 3 {
			case 0:
				e = p.AND(a, b)
			case 1:
				e = p.OR(a, b)
			case 2:
				e = p.NewXor(a, b)
			}
			pool = append(pool, e)
		}
		condition := p.OR(inputs[0], inputs[1])
		function := pool[len(pool)-1]
		result, status := Aig_ManSupportMinimization(p, condition, function)
		if status != 0 {
			continue
		}
		successes++
		for mask := 0; mask < 16; mask++ {
			assignment := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			if factValue(p, condition, assignment) && factValue(p, result, assignment) != factValue(p, function, assignment) {
				t.Fatalf("care mismatch cone %d assignment %d", iteration, mask)
			}
		}
		if err := p.Check(); err != nil {
			t.Fatal("reconstructed graph invalid", err)
		}
	}
	if successes == 0 {
		t.Fatal("randomized suite did not exercise any successful reductions")
	}
}

func TestSupportTruthRejectsInvalidCuts(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	for _, support := range [][]Edge{{a, a, b}, {p.Const1(), a, b}, {a.Not(), b}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid cut accepted", support)
				}
			}()
			supportTruth(p, []Edge{root}, support)
		}()
	}
	large := NewGraph()
	var support []Edge
	for i := 0; i < 17; i++ {
		support = append(support, large.NewInput())
	}
	condition := support[0]
	for _, e := range support[1:] {
		condition = large.OR(condition, e)
	}
	result, status := Aig_ManSupportMinimization(large, condition, large.AND(support[0], support[1]))
	if result != InvalidEdge || status != SupportMinInapplicable {
		t.Fatal("large support accepted", result, status)
	}
}

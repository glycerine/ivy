package abc

// Go-only model and fuzz tests for src/map/if/ifCheck.c.

import (
	"errors"
	"math/rand"
	"testing"
)

func ifCheckFixture(t *testing.T) (*AigManager, []AigLit) {
	t.Helper()
	graph := NewAigManager()
	inputs := make([]AigLit, 3)
	for index := range inputs {
		var err error
		inputs[index], err = graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	return graph, inputs
}

func TestIfCutCheckStructuralFunctionalAndStatistics(t *testing.T) {
	graph, inputs := ifCheckFixture(t)
	path, err := graph.And(inputs[1], inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	root, err := graph.And(inputs[0], path)
	if err != nil {
		t.Fatal(err)
	}
	leaves := []AigID{inputs[0].ID(), inputs[1].ID(), inputs[2].ID()}
	checker := new(IfCutChecker)
	result, err := checker.Analyze(graph, root.ID(), leaves, IfCutCheckOptions{})
	if err != nil || !result.DirectLeaf || !result.Structural || result.StructuralConflict || !result.Functional || result.FunctionalLeaf != inputs[0].ID() || result.FunctionalCofactor != 0 {
		t.Fatal(result, err)
	}
	if statistics := checker.Statistics(); statistics != (IfCutCheckStatistics{Calls: 1, Structural: 1, Functional: 1}) {
		t.Fatal(statistics)
	}

	conflictPath, err := graph.And(inputs[0].Not(), inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	conflictRoot, err := graph.And(inputs[0], conflictPath)
	if err != nil {
		t.Fatal(err)
	}
	result, err = checker.Analyze(graph, conflictRoot.ID(), leaves, IfCutCheckOptions{})
	if err != nil || !result.DirectLeaf || result.Structural || !result.StructuralConflict || !result.Functional {
		t.Fatal(result, err)
	}
	if statistics := checker.Statistics(); statistics.Calls != 2 || statistics.Structural != 1 || statistics.StructuralConflicts != 1 || statistics.Functional != 2 {
		t.Fatal(statistics)
	}
}

func TestIfCutCheckMajorityNoConstantCofactor(t *testing.T) {
	graph, inputs := ifCheckFixture(t)
	root, err := graph.Majority(inputs[0], inputs[1], inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	checker := new(IfCutChecker)
	result, err := checker.Analyze(graph, root.ID(), []AigID{inputs[0].ID(), inputs[1].ID(), inputs[2].ID()}, IfCutCheckOptions{})
	if err != nil || result.Functional || result.FunctionalLeaf != InvalidAigID || result.FunctionalCofactor != -1 {
		t.Fatal(result, err)
	}
}

func TestIfCutCheckBoundaryErrorsLimitsAndCancellation(t *testing.T) {
	graph, inputs := ifCheckFixture(t)
	path, _ := graph.And(inputs[1], inputs[2])
	root, _ := graph.And(inputs[0], path)
	leaves := []AigID{inputs[0].ID(), inputs[1].ID(), inputs[2].ID()}
	if reached, err := graph.IfCutReach(inputs[0].ID(), inputs[0].ID(), leaves, IfCutCheckOptions{}); err != nil || !reached {
		t.Fatal(reached, err)
	}
	if reached, err := graph.IfCutReach(path.ID(), inputs[0].ID(), leaves, IfCutCheckOptions{}); err != nil || reached {
		t.Fatal(reached, err)
	}
	for _, options := range []IfCutCheckOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxWords: -1}, {MaxObjects: 1}, {MaxMappingEntries: 1}} {
		if _, err := new(IfCutChecker).Analyze(graph, root.ID(), leaves, options); err == nil {
			t.Fatal(options)
		}
	}
	for _, bad := range [][]AigID{nil, make([]AigID, 8), {inputs[0].ID(), inputs[0].ID()}, {0, inputs[1].ID()}} {
		if _, err := new(IfCutChecker).Analyze(graph, root.ID(), bad, IfCutCheckOptions{}); err == nil {
			t.Fatal(bad)
		}
	}
	if _, err := (*IfCutChecker)(nil).Analyze(graph, root.ID(), leaves, IfCutCheckOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	before := new(IfCutChecker)
	calls := 0
	_, err := before.Analyze(graph, root.ID(), leaves, IfCutCheckOptions{Stop: func() bool { calls++; return calls == 3 }})
	if !errors.Is(err, ErrCancelled) || before.Statistics().Calls != 0 {
		t.Fatal(err, before.Statistics())
	}
	if _, err = graph.IfCutCofactorConstant(root.ID(), leaves, len(leaves), false, IfCutCheckOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestIfCutReachDeepIterative(t *testing.T) {
	graph, inputs := ifCheckFixture(t)
	chain, err := graph.And(inputs[0], inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10000; index++ {
		chain, err = graph.And(chain, inputs[0].NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	leaves := []AigID{inputs[0].ID(), inputs[1].ID(), inputs[2].ID()}
	if reached, reachErr := graph.IfCutReach(chain.ID(), inputs[1].ID(), leaves, IfCutCheckOptions{}); reachErr != nil || !reached {
		t.Fatal(reached, reachErr)
	}
	if _, reachErr := graph.IfCutReach(chain.ID(), inputs[1].ID(), leaves, IfCutCheckOptions{MaxObjects: 10}); !errors.Is(reachErr, ErrLimit) {
		t.Fatal(reachErr)
	}
}

func ifCheckEvaluateRegular(graph *AigManager, root AigID, inputs []bool) bool {
	values := make([]bool, len(graph.objects))
	values[0] = true
	for index, id := range graph.cis {
		values[id] = inputs[index]
	}
	read := func(literal AigLit) bool { return values[literal.ID()] != literal.Complemented() }
	for id, object := range graph.objects {
		if object.Type == AigAND {
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		}
	}
	return values[root]
}

func ifCheckFunctionalModel(graph *AigManager, root AigID, inputCount int) (bool, int, int) {
	for leaf := 0; leaf < inputCount; leaf++ {
		for cofactor := 0; cofactor < 2; cofactor++ {
			initialized, first, constant := false, false, true
			for assignment := 0; assignment < 1<<uint(inputCount); assignment++ {
				if assignment>>uint(leaf)&1 != cofactor {
					continue
				}
				inputs := make([]bool, inputCount)
				for index := range inputs {
					inputs[index] = assignment>>uint(index)&1 != 0
				}
				value := ifCheckEvaluateRegular(graph, root, inputs)
				if !initialized {
					first, initialized = value, true
				} else if value != first {
					constant = false
					break
				}
			}
			if constant {
				return true, leaf, cofactor
			}
		}
	}
	return false, -1, -1
}

func FuzzIfCutCheck(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 99, 1001} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		inputCount := 2 + random.Intn(5)
		graph := NewAigManager()
		literals := make([]AigLit, inputCount)
		leaves := make([]AigID, inputCount)
		for index := range literals {
			literal, err := graph.CreateCI()
			if err != nil {
				t.Fatal(err)
			}
			literals[index], leaves[index] = literal, literal.ID()
		}
		var roots []AigID
		for gate := 0; gate < 24; gate++ {
			left := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			right := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			literal, err := graph.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			literals = append(literals, literal)
			if graph.objects[literal.ID()].Type == AigAND {
				roots = append(roots, literal.ID())
			}
		}
		if len(roots) == 0 {
			return
		}
		root := roots[random.Intn(len(roots))]
		actual, err := new(IfCutChecker).Analyze(graph, root, leaves, IfCutCheckOptions{})
		if err != nil {
			t.Fatal(err)
		}
		functional, leafIndex, cofactor := ifCheckFunctionalModel(graph, root, inputCount)
		if actual.Functional != functional || functional && (actual.FunctionalLeaf != leaves[leafIndex] || actual.FunctionalCofactor != cofactor) {
			t.Fatal(seed, root, actual, functional, leafIndex, cofactor)
		}
		object := graph.objects[root]
		directIndex := -1
		alternate := InvalidAigID
		for index, leaf := range leaves {
			if object.Fanin0.ID() == leaf {
				directIndex = index
				alternate = object.Fanin1.ID()
				break
			}
			if object.Fanin1.ID() == leaf {
				directIndex = index
				alternate = object.Fanin0.ID()
				break
			}
		}
		if actual.DirectLeaf != (directIndex >= 0) {
			t.Fatal(seed, actual, directIndex)
		}
		reachable := false
		if directIndex >= 0 {
			stack := []AigID{alternate}
			seen := make([]bool, len(graph.objects))
			for len(stack) != 0 {
				id := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if id == leaves[directIndex] {
					reachable = true
					break
				}
				if seen[id] || graph.objects[id].Type == AigCI {
					continue
				}
				seen[id] = true
				object := graph.objects[id]
				stack = append(stack, object.Fanin0.ID(), object.Fanin1.ID())
			}
		}
		if actual.Structural != (directIndex >= 0 && !reachable) || actual.StructuralConflict != (directIndex >= 0 && reachable) {
			t.Fatal(seed, actual, directIndex, reachable)
		}
	})
}

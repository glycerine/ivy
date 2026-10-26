package abc

// Standalone semantic, SAT, sequential, limit, concurrency and fuzz tests for saigOutDec.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func saigPrimeValues(graph *AigManager, inputs []bool) []bool {
	values := make([]bool, graph.ObjectCount())
	values[0] = true
	for index, id := range graph.cis {
		values[id] = inputs[index]
	}
	read := func(literal AigLit) bool { return values[literal.ID()] != literal.Complemented() }
	for id, object := range graph.objects {
		switch object.Type {
		case AigAND:
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[id] = read(object.Fanin0)
		}
	}
	return values
}

func saigPrimeCubeValue(cube SAIGPrimeCube, values []bool) bool {
	for _, literal := range cube.Literals {
		if values[literal.ID()] == literal.Complemented() {
			return false
		}
	}
	return true
}

func saigExpectedPrimes(t testing.TB, graph *AigManager, literalCount int) []SAIGPrimeCube {
	t.Helper()
	root := graph.objects[graph.cos[0]].Fanin0.ID()
	cone, err := graph.DFSNodes([]AigID{root}, AigTraversalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assignments := make([][]bool, 1<<graph.InputCount())
	values := make([][]bool, len(assignments))
	for assignment := range assignments {
		assignments[assignment] = make([]bool, graph.InputCount())
		for variable := range assignments[assignment] {
			assignments[assignment][variable] = assignment&(1<<variable) != 0
		}
		values[assignment] = saigPrimeValues(graph, assignments[assignment])
	}
	impliesProperty := func(cube SAIGPrimeCube) bool {
		for _, row := range values {
			if saigPrimeCubeValue(cube, row) && !row[graph.cos[0]] {
				return false
			}
		}
		return true
	}
	result := make([]SAIGPrimeCube, 0)
	marked := make([]bool, len(cone))
	for index, id := range cone {
		if id == root {
			continue
		}
		cube := SAIGPrimeCube{Literals: []AigLit{saigPrimeLiteral(id, graph.objects[id].Phase)}}
		if impliesProperty(cube) {
			result = append(result, cube)
			marked[index] = true
		}
	}
	if literalCount == 2 {
		for first, firstID := range cone {
			if firstID == root {
				continue
			}
			for second := first + 1; second < len(cone); second++ {
				secondID := cone[second]
				if secondID == root || marked[first] || marked[second] {
					continue
				}
				for mode := 0; mode < 3; mode++ {
					cube := SAIGPrimeCube{Literals: []AigLit{
						saigPrimeLiteral(firstID, graph.objects[firstID].Phase != (mode&1 != 0)),
						saigPrimeLiteral(secondID, graph.objects[secondID].Phase != (mode&2 != 0)),
					}}
					if impliesProperty(cube) {
						result = append(result, cube)
						break
					}
				}
			}
		}
	}
	return result
}

func equalSAIGPrimeCubes(left, right []SAIGPrimeCube) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if !slices.Equal(left[index].Literals, right[index].Literals) {
			return false
		}
	}
	return true
}

func saigPairFixture(t testing.TB, complemented bool) (*AigManager, AigLit, AigLit) {
	t.Helper()
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	root, err := graph.And(a, b.NotIf(complemented))
	if err != nil {
		t.Fatal(err)
	}
	graph.CreateCO(root)
	return graph, a, b
}

func TestSAIGFindPropertyPrimesExactOrderAndPolarity(t *testing.T) {
	graph, a, b := saigPairFixture(t, true)
	result, err := SAIGFindPropertyPrimes(graph, 2, SAIGPropertyPrimeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []SAIGPrimeCube{{Literals: []AigLit{a, b.Not()}}}
	if !equalSAIGPrimeCubes(result.Cubes, want) || result.Single != 0 || result.Pair != 1 || result.Unknown != 0 {
		t.Fatalf("result=%+v want=%v", result, want)
	}
	model := saigExpectedPrimes(t, graph, 2)
	if !equalSAIGPrimeCubes(result.Cubes, model) {
		t.Fatalf("result=%v model=%v", result.Cubes, model)
	}

	orGraph := NewAigManager()
	x, _ := orGraph.CreateCI()
	y, _ := orGraph.CreateCI()
	z, _ := orGraph.CreateCI()
	xy, _ := orGraph.And(x, y)
	root, _ := orGraph.And(xy.Not(), z.Not())
	orGraph.CreateCO(root.Not())
	singles, err := SAIGFindPropertyPrimes(orGraph, 1, SAIGPropertyPrimeOptions{})
	if err != nil || singles.Single != 2 || singles.Pair != 0 || !equalSAIGPrimeCubes(singles.Cubes, saigExpectedPrimes(t, orGraph, 1)) {
		t.Fatalf("singles=%+v err=%v", singles, err)
	}
}

func TestSAIGDecomposePropertyOutputSemanticsSequentialAndOwnership(t *testing.T) {
	graph, _, _ := saigPairFixture(t, false)
	graph.Name, graph.Spec = "property", "property.aig"
	result, primes, err := SAIGDecomposePropertyOutput(graph, 2, SAIGPropertyPrimeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != graph.Name || result.Spec != graph.Spec || result.InputCount() != 2 || result.PrimaryOutputCount() != 2 || len(primes.Cubes) != 1 {
		t.Fatalf("result inputs/outputs=%d/%d primes=%+v", result.InputCount(), result.PrimaryOutputCount(), primes)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		sourceOutputs, _ := graph.Evaluate(inputs)
		outputs, _ := result.Evaluate(inputs)
		values := saigPrimeValues(graph, inputs)
		if outputs[0] != sourceOutputs[0] || outputs[1] != saigPrimeCubeValue(primes.Cubes[0], values) {
			t.Fatalf("assignment=%d source=%v result=%v", assignment, sourceOutputs, outputs)
		}
	}
	primes.Cubes[0].Literals[0] = AigConst0
	again, againPrimes, againErr := SAIGDecomposePropertyOutput(graph, 2, SAIGPropertyPrimeOptions{})
	if againErr != nil || again == result || againPrimes.Cubes[0].Literals[0] == AigConst0 {
		t.Fatal("returned graph or cubes do not own their state")
	}

	sequential := NewAigManager()
	pi, _ := sequential.CreateCI()
	lo, _ := sequential.CreateCI()
	property, _ := sequential.And(pi, lo)
	sequential.CreateCO(property)
	sequential.CreateCO(pi)
	sequential.SetRegisterCount(1)
	seqResult, seqPrimes, err := SAIGDecomposePropertyOutput(sequential, 2, SAIGPropertyPrimeOptions{})
	if err != nil || seqResult.RegisterCount() != 1 || seqResult.PrimaryInputCount() != 1 || seqResult.PrimaryOutputCount() != 2 || len(seqPrimes.Cubes) != 1 {
		t.Fatalf("sequential result=%v primes=%+v err=%v", seqResult, seqPrimes, err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		outputs, _ := seqResult.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
		want := assignment == 3
		if outputs[0] != want || outputs[1] != want || outputs[2] != (assignment&1 != 0) {
			t.Fatalf("sequential assignment=%d outputs=%v", assignment, outputs)
		}
	}
}

func TestSAIGPropertyPrimesLimitsDepthAndConcurrency(t *testing.T) {
	graph, _, _ := saigPairFixture(t, false)
	if result, err := SAIGFindPropertyPrimes(nil, 2, SAIGPropertyPrimeOptions{}); result.Cubes != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil result=%+v err=%v", result, err)
	}
	if result, err := SAIGFindPropertyPrimes(graph, 3, SAIGPropertyPrimeOptions{}); result.Cubes != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("literal count result=%+v err=%v", result, err)
	}
	empty := NewAigManager()
	if result, err := SAIGFindPropertyPrimes(empty, 1, SAIGPropertyPrimeOptions{}); result.Cubes != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("empty result=%+v err=%v", result, err)
	}
	for _, options := range []SAIGPropertyPrimeOptions{
		{MaxObjects: 1}, {MaxVisits: 1}, {MaxCandidates: 1}, {MaxQueries: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {MaxWork: 1}, {Stop: func() bool { return true }},
	} {
		if result, err := SAIGFindPropertyPrimes(graph, 2, options); result.Cubes != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("options=%+v result=%+v err=%v", options, result, err)
		}
	}
	orGraph := NewAigManager()
	a, _ := orGraph.CreateCI()
	b, _ := orGraph.CreateCI()
	c, _ := orGraph.CreateCI()
	ab, _ := orGraph.And(a, b)
	root, _ := orGraph.And(ab.Not(), c.Not())
	orGraph.CreateCO(root.Not())
	if result, err := SAIGFindPropertyPrimes(orGraph, 1, SAIGPropertyPrimeOptions{MaxCubes: 1}); result.Cubes != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("cube limit result=%+v err=%v", result, err)
	}
	three := NewAigManager()
	x, _ := three.CreateCI()
	y, _ := three.CreateCI()
	z, _ := three.CreateCI()
	xy, _ := three.And(x, y)
	xyz, _ := three.And(xy, z)
	three.CreateCO(xyz)
	if result, err := SAIGFindPropertyPrimes(three, 2, SAIGPropertyPrimeOptions{MaxPairs: 1}); result.Cubes != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("pair limit result=%+v err=%v", result, err)
	}

	deep := NewAigManager()
	p, _ := deep.CreateCI()
	q, _ := deep.CreateCI()
	deepRoot := p
	for index := 0; index < 10000; index++ {
		deepRoot, _ = deep.And(deepRoot.NotIf(index%3 == 0), q.NotIf(index%5 == 0))
	}
	deep.CreateCO(deepRoot)
	if deep.ObjectCount() < 10000 {
		t.Fatalf("deep fixture collapsed to %d", deep.ObjectCount())
	}
	if result, err := SAIGFindPropertyPrimes(deep, 1, SAIGPropertyPrimeOptions{MaxObjects: 20000, MaxVisits: 100000, MaxCandidates: 20000, MaxQueries: 1}); result.Cubes != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("deep result=%+v err=%v", result, err)
	}

	baseline, _ := SAIGFindPropertyPrimes(graph, 2, SAIGPropertyPrimeOptions{})
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, callErr := SAIGFindPropertyPrimes(graph, 2, SAIGPropertyPrimeOptions{})
			if callErr != nil || !equalSAIGPrimeCubes(got.Cubes, baseline.Cubes) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzSAIGPropertyPrimes(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4}, uint8(2))
	f.Add([]byte{0xff, 9, 17}, uint8(1))
	f.Fuzz(func(t *testing.T, data []byte, rawLiterals uint8) {
		if len(data) == 0 {
			return
		}
		graph := NewAigManager()
		pool := make([]AigLit, 3)
		for index := range pool {
			pool[index], _ = graph.CreateCI()
		}
		gates := 1 + int(data[0]%4)
		for gate := 0; gate < gates; gate++ {
			left := pool[int(data[(2*gate+1)%len(data)])%len(pool)].NotIf(data[(gate+1)%len(data)]&1 != 0)
			right := pool[int(data[(2*gate+2)%len(data)])%len(pool)].NotIf(data[(gate+2)%len(data)]&2 != 0)
			root, err := graph.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, root)
		}
		graph.CreateCO(pool[len(pool)-1].NotIf(data[0]&4 != 0))
		literalCount := 1 + int(rawLiterals&1)
		result, err := SAIGFindPropertyPrimes(graph, literalCount, SAIGPropertyPrimeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := saigExpectedPrimes(t, graph, literalCount)
		if !equalSAIGPrimeCubes(result.Cubes, want) {
			t.Fatalf("literals=%d got=%v want=%v", literalCount, result.Cubes, want)
		}
	})
}

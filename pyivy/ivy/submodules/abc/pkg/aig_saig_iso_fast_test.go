package abc

// Independent packed-signature, class, sequential, limit, concurrency and fuzz tests for saigIsoFast.c.

import (
	"errors"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
)

type saigIsoReferenceData struct {
	faninNegative, fanoutNegative, fanoutPositive uint8
	firstLevel, secondLevel, level                uint8
	visited                                       bool
}

func (data *saigIsoReferenceData) fanout(complement bool) {
	if complement {
		data.fanoutNegative = (data.fanoutNegative + 1) % 4
	} else {
		data.fanoutPositive = (data.fanoutPositive + 1) % 4
	}
}

func (data saigIsoReferenceData) word() uint16 {
	return uint16(data.faninNegative%8) |
		uint16(data.fanoutNegative%4)<<3 |
		uint16(data.fanoutPositive%4)<<5 |
		uint16(data.firstLevel%8)<<7 |
		uint16(data.secondLevel%8)<<10 |
		uint16(data.level%8)<<13
}

func saigIsoReferencePermutation(costs []int) []int {
	result := make([]int, len(costs))
	for index := range result {
		result[index] = index
	}
	scratch := make([]int, len(result))
	var order func([]int, []int)
	order = func(input, output []int) {
		if len(input) < 8 {
			for i := 0; i+1 < len(input); i++ {
				best := i
				for j := i + 1; j < len(input); j++ {
					if costs[input[j]] < costs[input[best]] {
						best = j
					}
				}
				input[i], input[best] = input[best], input[i]
			}
			return
		}
		middle := len(input) / 2
		order(input[:middle], output[:middle])
		order(input[middle:], output[middle:])
		left, right, out := 0, middle, 0
		for left < middle && right < len(input) {
			switch {
			case costs[input[left]] == costs[input[right]]:
				output[out], output[out+1] = input[left], input[right]
				left, right, out = left+1, right+1, out+2
			case costs[input[left]] < costs[input[right]]:
				output[out] = input[left]
				left, out = left+1, out+1
			default:
				output[out] = input[right]
				right, out = right+1, out+1
			}
		}
		out += copy(output[out:], input[left:middle])
		copy(output[out:], input[right:])
		copy(input, output)
	}
	order(result, scratch)
	return result
}

func saigIsoReferenceSignature(graph *AigManager, output int) []uint32 {
	data := make([]saigIsoReferenceData, graph.ObjectCount())
	visited := make([]AigID, 0)
	ciOrdinal := make([]int, graph.ObjectCount())
	for index := range ciOrdinal {
		ciOrdinal[index] = -1
	}
	for index, id := range graph.cis {
		ciOrdinal[id] = index
	}
	roots := []AigID{graph.cos[output]}
	var collect func(AigID, bool)
	collect = func(id AigID, complement bool) {
		entry := &data[id]
		if entry.visited {
			entry.fanout(complement)
			return
		}
		entry.visited = true
		entry.fanout(complement)
		object := graph.objects[id]
		entry.level = uint8(object.Level % 8)
		switch object.Type {
		case AigAND:
			if object.Fanin0.Complemented() {
				entry.faninNegative++
			}
			if object.Fanin1.Complemented() {
				entry.faninNegative++
			}
			first := object.Level - graph.objects[object.Fanin0.ID()].Level
			second := object.Level - graph.objects[object.Fanin1.ID()].Level
			c0, c1 := object.Fanin0.Complemented(), object.Fanin1.Complemented()
			if !(c0 != c1 && !c0) && !(c0 == c1 && graph.objects[object.Fanin0.ID()].Level < graph.objects[object.Fanin1.ID()].Level) {
				first, second = second, first
			}
			entry.firstLevel, entry.secondLevel = uint8(first%8), uint8(second%8)
			collect(object.Fanin0.ID(), object.Fanin0.Complemented())
			collect(object.Fanin1.ID(), object.Fanin1.Complemented())
		case AigCI:
			if ciOrdinal[id] >= graph.PrimaryInputCount() {
				entry.firstLevel = 1
				register := ciOrdinal[id] - graph.PrimaryInputCount()
				roots = append(roots, graph.cos[graph.PrimaryOutputCount()+register])
			}
		}
		visited = append(visited, id)
	}
	for root := 0; root < len(roots); root++ {
		literal := graph.objects[roots[root]].Fanin0
		if literal.ID() != 0 {
			collect(literal.ID(), literal.Complemented())
		}
	}
	counts := make(map[uint16]int)
	places := make([]uint16, 0)
	for _, id := range visited {
		word := data[id].word()
		if counts[word] == 0 {
			places = append(places, word)
		}
		counts[word]++
	}
	frequencies := make([]int, len(places))
	for index, place := range places {
		frequencies[index] = counts[place]
	}
	permutation := saigIsoReferencePermutation(frequencies)
	result := make([]uint32, 0, len(places))
	for index := len(permutation) - 1; index >= 0; index-- {
		position := permutation[index]
		result = append(result, uint32(frequencies[position])<<16|uint32(places[position]))
	}
	return result
}

func saigIsoReferenceResult(graph *AigManager) SAIGIsoFastResult {
	result := SAIGIsoFastResult{Outputs: make([]SAIGIsoFastOutput, graph.PrimaryOutputCount())}
	combined := make([][]uint32, graph.PrimaryOutputCount())
	for output := range result.Outputs {
		signature := saigIsoReferenceSignature(graph, output)
		result.Outputs[output] = SAIGIsoFastOutput{Output: output, Signature: signature}
		combined[output] = append(append([]uint32(nil), signature...), uint32(output))
	}
	sort.Slice(combined, func(i, j int) bool { return saigIsoFastCompare(combined[i], combined[j]) < 0 })
	for _, info := range combined {
		output := int(info[len(info)-1])
		signature := info[:len(info)-1]
		if len(result.Classes) == 0 || saigIsoFastCompare(result.Outputs[result.Classes[len(result.Classes)-1][0]].Signature, signature) != 0 {
			result.Classes = append(result.Classes, []int{output})
		} else {
			result.Classes[len(result.Classes)-1] = append(result.Classes[len(result.Classes)-1], output)
		}
	}
	for _, class := range result.Classes {
		if len(class) == 1 {
			result.Unique++
		}
	}
	return result
}

func saigIsoFastFixture(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	c, _ := graph.CreateCI()
	d, _ := graph.CreateCI()
	state0, _ := graph.CreateCI()
	state1, _ := graph.CreateCI()
	first, _ := graph.And(a, state0)
	second, _ := graph.And(b, state1)
	different, _ := graph.And(a.Not(), state0)
	graph.CreateCO(first)
	graph.CreateCO(second)
	graph.CreateCO(different)
	graph.CreateCO(AigConst0)
	graph.CreateCO(AigConst1)
	graph.CreateCO(c)
	graph.CreateCO(d)
	if err := graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestSAIGIsoFastExactPackedSignaturesAndClasses(t *testing.T) {
	simple := NewAigManager()
	a, _ := simple.CreateCI()
	b, _ := simple.CreateCI()
	root, _ := simple.And(a, b)
	simple.CreateCO(root)
	result, err := SAIGDetectIsomorphicOutputsFast(simple, SAIGIsoFastOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantSignature := []uint32{0x00020020, 0x000124a0}
	if !reflect.DeepEqual(result.Outputs[0].Signature, wantSignature) || !reflect.DeepEqual(result.Classes, [][]int{{0}}) || result.Unique != 1 {
		t.Fatalf("result=%+v want=%x", result, wantSignature)
	}

	graph := saigIsoFastFixture(t)
	before := graph.Clone()
	result, err = SAIGDetectIsomorphicOutputsFast(graph, SAIGIsoFastOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := saigIsoReferenceResult(graph)
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("got=%+v want=%+v", result, want)
	}
	classOf := func(output int) int {
		for class, members := range result.Classes {
			for _, member := range members {
				if member == output {
					return class
				}
			}
		}
		return -1
	}
	if classOf(0) != classOf(1) || classOf(0) == classOf(2) || classOf(3) != classOf(4) {
		t.Fatalf("unexpected classes %v", result.Classes)
	}
	result.Outputs[0].Signature[0] ^= 1
	result.Classes[0][0] = 99
	again, againErr := SAIGDetectIsomorphicOutputsFast(graph, SAIGIsoFastOptions{})
	if againErr != nil || reflect.DeepEqual(result, again) {
		t.Fatal("result storage is not owned")
	}
	if !reflect.DeepEqual(graph.objects, before.objects) || !reflect.DeepEqual(graph.cis, before.cis) || !reflect.DeepEqual(graph.cos, before.cos) {
		t.Fatal("analysis mutated source")
	}
}

func TestSAIGIsoFastLimitsDepthAndConcurrency(t *testing.T) {
	graph := saigIsoFastFixture(t)
	if result, err := SAIGDetectIsomorphicOutputsFast(nil, SAIGIsoFastOptions{}); result.Outputs != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil result=%+v err=%v", result, err)
	}
	empty := NewAigManager()
	if result, err := SAIGDetectIsomorphicOutputsFast(empty, SAIGIsoFastOptions{}); result.Outputs != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("empty result=%+v err=%v", result, err)
	}
	for _, options := range []SAIGIsoFastOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxRoots: -1}, {MaxEntries: -1}, {MaxWork: -1}} {
		if result, err := SAIGDetectIsomorphicOutputsFast(graph, options); result.Outputs != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("options=%+v result=%+v err=%v", options, result, err)
		}
	}
	for _, options := range []SAIGIsoFastOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxRoots: 1}, {MaxEntries: 1}, {MaxWork: 1}, {Stop: func() bool { return true }}} {
		if result, err := SAIGDetectIsomorphicOutputsFast(graph, options); result.Outputs != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("options=%+v result=%+v err=%v", options, result, err)
		}
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(buffer)
	if result, err := SAIGDetectIsomorphicOutputsFast(buffered, SAIGIsoFastOptions{}); result.Outputs != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("buffer result=%+v err=%v", result, err)
	}

	deep := NewAigManager()
	deepRoot, _ := deep.CreateCI()
	for index := 0; index < 10000; index++ {
		input, _ := deep.CreateCI()
		deepRoot, _ = deep.And(deepRoot.NotIf(index&1 != 0), input.NotIf(index&2 != 0))
	}
	deep.CreateCO(deepRoot)
	deepResult, err := SAIGDetectIsomorphicOutputsFast(deep, SAIGIsoFastOptions{MaxObjects: 50000, MaxVisits: 100000, MaxEntries: 50000, MaxWork: 1000000})
	if err != nil || len(deepResult.Outputs) != 1 || len(deepResult.Outputs[0].Signature) == 0 {
		t.Fatalf("deep result=%+v err=%v", deepResult, err)
	}

	baseline, err := SAIGDetectIsomorphicOutputsFast(graph, SAIGIsoFastOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, callErr := SAIGDetectIsomorphicOutputsFast(graph, SAIGIsoFastOptions{})
			if callErr != nil || !reflect.DeepEqual(got, baseline) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzSAIGIsoFastSignatures(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{0xff, 0, 17})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		at := func(index int) byte { return data[index%len(data)] }
		graph := NewAigManager()
		pool := make([]AigLit, 5)
		for index := range pool {
			pool[index], _ = graph.CreateCI()
		}
		for gate := 0; gate < 1+int(at(0)%6); gate++ {
			left := pool[int(at(2*gate+1))%len(pool)].NotIf(at(gate+1)&1 != 0)
			right := pool[int(at(2*gate+2))%len(pool)].NotIf(at(gate+2)&2 != 0)
			root, err := graph.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, root)
		}
		outputs := 1 + int(at(3)%4)
		for output := 0; output < outputs; output++ {
			literal := pool[int(at(output+4))%len(pool)].NotIf(at(output+5)&4 != 0)
			graph.CreateCO(literal)
		}
		graph.CreateCO(pool[int(at(9))%len(pool)].NotIf(at(10)&8 != 0))
		if err := graph.SetRegisterCount(1); err != nil {
			t.Fatal(err)
		}
		got, err := SAIGDetectIsomorphicOutputsFast(graph, SAIGIsoFastOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := saigIsoReferenceResult(graph)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got=%+v want=%+v", got, want)
		}
	})
}

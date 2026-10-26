package abc

// Go-only model and fuzz tests for src/aig/gia/giaSupMin.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func giaSupportExperimentFixture(t *testing.T) *GiaManager {
	t.Helper()
	manager := NewGiaManager()
	inputs := make([]GiaLit, 4)
	for index := range inputs {
		var err error
		inputs[index], err = manager.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	first, err := manager.AppendAnd(inputs[0], inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	data, err := manager.AppendAnd(first, inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	care, err := manager.AppendAnd(inputs[0].Not(), inputs[3])
	if err != nil {
		t.Fatal(err)
	}
	for _, driver := range []GiaLit{data.Not(), care, inputs[0], data} {
		if _, err = manager.AppendCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	return manager
}

func TestGiaSupportExperimentPairsAndInvalidCare(t *testing.T) {
	manager := giaSupportExperimentFixture(t)
	before := slices.Clone(manager.objects)
	pairs := [][2]int{{0, 1}, {0, 2}, {1, 3}}
	result, err := manager.AnalyzeSupportPairs(pairs, GiaSupportExperimentOptions{})
	if err != nil || len(result) != 3 || !slices.Equal(before, manager.objects) {
		t.Fatal(result, err)
	}
	if result[0] != (GiaSupportPairAnalysis{DataOutput: 0, CareOutput: 1, CareIsAnd: true, Data: GiaConeStatistics{Inputs: 3, Gates: 2}, Care: GiaConeStatistics{Inputs: 2, Gates: 1}}) {
		t.Fatal(result[0])
	}
	if result[1] != (GiaSupportPairAnalysis{DataOutput: 0, CareOutput: 2}) {
		t.Fatal(result[1])
	}
	if !result[2].CareIsAnd || result[2].Data != (GiaConeStatistics{Inputs: 2, Gates: 1}) || result[2].Care != (GiaConeStatistics{Inputs: 3, Gates: 2}) {
		t.Fatal(result[2])
	}
}

func TestGiaSupportExperimentErrorsLimitsCancellationAndDeep(t *testing.T) {
	manager := giaSupportExperimentFixture(t)
	for _, options := range []GiaSupportExperimentOptions{{MaxObjects: -1}, {MaxPairs: -1}, {MaxObjects: 1}, {MaxPairs: 1}} {
		if result, err := manager.AnalyzeSupportPairs([][2]int{{0, 1}, {1, 3}}, options); err == nil || result != nil {
			t.Fatal(options, result, err)
		}
	}
	for _, pair := range [][2]int{{-1, 1}, {0, 4}} {
		if result, err := manager.AnalyzeSupportPairs([][2]int{pair}, GiaSupportExperimentOptions{}); !errors.Is(err, ErrDimension) || result != nil {
			t.Fatal(pair, result, err)
		}
	}
	if result, err := (*GiaManager)(nil).AnalyzeSupportPairs(nil, GiaSupportExperimentOptions{}); !errors.Is(err, ErrDimension) || result != nil {
		t.Fatal(result, err)
	}
	calls := 0
	if result, err := manager.AnalyzeSupportPairs([][2]int{{0, 1}}, GiaSupportExperimentOptions{Stop: func() bool { calls++; return calls == 5 }}); !errors.Is(err, ErrCancelled) || result != nil {
		t.Fatal(result, err, calls)
	}

	deep := NewGiaManager()
	a, _ := deep.AppendCI()
	b, _ := deep.AppendCI()
	chain, _ := deep.AppendAnd(a, b)
	for index := 0; index < 10000; index++ {
		var err error
		chain, err = deep.AppendAnd(chain, a.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	deep.AppendCO(chain)
	result, err := deep.AnalyzeSupportPairs([][2]int{{0, 0}}, GiaSupportExperimentOptions{})
	if err != nil || result[0].Data != (GiaConeStatistics{Inputs: 2, Gates: 10001}) || result[0].Care != result[0].Data {
		t.Fatal(result, err)
	}
}

func giaSupportExperimentModel(manager *GiaManager, root GiaID) GiaConeStatistics {
	result := GiaConeStatistics{}
	seen := make([]bool, len(manager.objects))
	stack := []GiaID{root}
	for len(stack) != 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] || id == 0 {
			continue
		}
		seen[id] = true
		object := manager.objects[id]
		if object.IsCI() {
			result.Inputs++
			continue
		}
		if !object.IsAnd() {
			continue
		}
		result.Gates++
		first, _ := object.Fanin0(id)
		second, _ := object.Fanin1(id)
		stack = append(stack, first.ID())
		if !object.IsBuffer() {
			stack = append(stack, second.ID())
		}
		if control, found := manager.MuxControl(id); found {
			stack = append(stack, control.ID())
		}
	}
	return result
}

func FuzzGiaSupportExperiment(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 17, 1001} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		manager := NewGiaManager()
		literals := make([]GiaLit, 2+random.Intn(6))
		for index := range literals {
			literal, err := manager.AppendCI()
			if err != nil {
				t.Fatal(err)
			}
			literals[index] = literal
		}
		var gates []GiaLit
		for gate := 0; gate < 32; gate++ {
			left := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			right := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
			if left.ID() == right.ID() {
				right = GiaConst0
			}
			literal, err := manager.AppendAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			literals = append(literals, literal)
			gates = append(gates, literal)
		}
		outputCount := 1 + random.Intn(8)
		drivers := make([]GiaID, outputCount)
		for output := range drivers {
			driver := gates[random.Intn(len(gates))].NotIf(random.Intn(2) != 0)
			drivers[output] = driver.ID()
			if _, err := manager.AppendCO(driver); err != nil {
				t.Fatal(err)
			}
		}
		pairs := make([][2]int, 1+random.Intn(12))
		for index := range pairs {
			pairs[index] = [2]int{random.Intn(outputCount), random.Intn(outputCount)}
		}
		actual, err := manager.AnalyzeSupportPairs(pairs, GiaSupportExperimentOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(actual) != len(pairs) {
			t.Fatal(len(actual), len(pairs))
		}
		for index, pair := range pairs {
			wantData := giaSupportExperimentModel(manager, drivers[pair[0]])
			wantCare := giaSupportExperimentModel(manager, drivers[pair[1]])
			if !actual[index].CareIsAnd || actual[index].Data != wantData || actual[index].Care != wantCare || !slices.Equal([]int{actual[index].DataOutput, actual[index].CareOutput}, pair[:]) {
				t.Fatal(seed, index, pair, actual[index], wantData, wantCare)
			}
		}
	})
}

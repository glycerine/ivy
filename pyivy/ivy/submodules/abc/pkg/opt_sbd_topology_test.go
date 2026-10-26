package abc

// Standalone model, limits, concurrency and fuzz tests for sbd.c.

import (
	"errors"
	"math"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSBDCountConfigurationVariables(t *testing.T) {
	tests := []struct {
		levels         []int
		inputs, degree int
		want           int
	}{
		{[]int{1}, 2, 2, 4},
		{[]int{1, 1}, 3, 2, 10},
		{[]int{1, 2}, 3, 2, 12},
		{[]int{1}, 4, 2, 0},
		{[]int{0}, 1, 3, 0},
	}
	for _, test := range tests {
		before := slices.Clone(test.levels)
		got, err := SBDCountConfigurationVariables(test.levels, test.inputs, test.degree, SBDTopologyLimits{})
		if err != nil || got != test.want || !slices.Equal(test.levels, before) {
			t.Fatalf("levels=%v inputs=%d degree=%d got=%d want=%d err=%v", test.levels, test.inputs, test.degree, got, test.want, err)
		}
	}
	for _, levels := range [][]int{nil, {}, {1, 3}, {1, -1}} {
		if _, err := SBDCountConfigurationVariables(levels, 2, 2, SBDTopologyLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("levels=%v err=%v", levels, err)
		}
	}
	if _, err := SBDCountConfigurationVariables([]int{math.MaxInt}, 2, 2, SBDTopologyLimits{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("overflow err=%v", err)
	}
}

func TestSBDCountTopologiesExactOrderAndOwnership(t *testing.T) {
	result, err := SBDCountTopologies(2, 3, 2, SBDTopologyLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := SBDTopologyResult{Inputs: 2, Nodes: 3, Degree: 2, Stages: []SBDTopologyStage{
		{Nodes: 1, New: 1, Total: 2, Configurations: []SBDTopologyConfiguration{{Levels: []int{1}, ConfigurationVariables: 4}}},
		{Nodes: 2, New: 1, Total: 3, Configurations: []SBDTopologyConfiguration{{Levels: []int{1, 1}, ConfigurationVariables: 7}}},
		{Nodes: 3, New: 2, Total: 5, Configurations: []SBDTopologyConfiguration{
			{Levels: []int{1, 2}, ConfigurationVariables: 8},
			{Levels: []int{1, 1, 1}, ConfigurationVariables: 11},
		}},
	}}
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("got=%+v want=%+v", result, want)
	}
	result.Stages[0].Configurations[0].Levels[0] = 99
	again, err := SBDCountTopologies(2, 3, 2, SBDTopologyLimits{})
	if err != nil || again.Stages[0].Configurations[0].Levels[0] != 1 {
		t.Fatalf("retained result state: %+v err=%v", again, err)
	}
	empty, err := SBDCountTopologies(3, 0, 4, SBDTopologyLimits{})
	if err != nil || empty.Inputs != 3 || empty.Nodes != 0 || empty.Degree != 4 || len(empty.Stages) != 0 {
		t.Fatalf("zero-node result=%+v err=%v", empty, err)
	}
}

func TestSBDCountTopologiesNativeExperimentTotals(t *testing.T) {
	result, err := SBDCountTopologies(9, 10, 3, SBDTopologyLimits{})
	if err != nil {
		t.Fatal(err)
	}
	wantNew := []int{1, 1, 2, 4, 7, 13, 25, 48, 92, 176}
	wantTotal := []int{2, 3, 5, 9, 16, 29, 54, 102, 194, 370}
	if len(result.Stages) != len(wantNew) {
		t.Fatalf("stage count=%d", len(result.Stages))
	}
	for index, stage := range result.Stages {
		if stage.Nodes != index+1 || stage.New != wantNew[index] || stage.Total != wantTotal[index] || len(stage.Configurations) != stage.New {
			t.Fatalf("stage %d: %+v", index+1, stage)
		}
	}
}

func TestSBDCountTopologiesLimitsCancellationAndConcurrency(t *testing.T) {
	for _, limits := range []SBDTopologyLimits{
		{MaxLevels: -1}, {MaxConfigurations: -1}, {MaxLevelEntries: -1}, {MaxWork: -1},
	} {
		if _, err := SBDCountTopologies(2, 2, 2, limits); !errors.Is(err, ErrDimension) {
			t.Fatalf("limits=%+v err=%v", limits, err)
		}
	}
	for _, limits := range []SBDTopologyLimits{
		{MaxLevels: 1}, {MaxConfigurations: 2}, {MaxLevelEntries: 1}, {MaxWork: 1},
	} {
		if _, err := SBDCountTopologies(2, 3, 2, limits); !errors.Is(err, ErrLimit) {
			t.Fatalf("limits=%+v err=%v", limits, err)
		}
	}
	if _, err := SBDCountTopologies(2, 2, 2, SBDTopologyLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel err=%v", err)
	}
	if _, err := SBDCountTopologies(0, 2, 2, SBDTopologyLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("input err=%v", err)
	}

	want, err := SBDCountTopologies(4, 6, 3, SBDTopologyLimits{})
	if err != nil {
		t.Fatal(err)
	}
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, runErr := SBDCountTopologies(4, 6, 3, SBDTopologyLimits{})
			if runErr != nil || !reflect.DeepEqual(got, want) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func sbdIndependentCount(levels []int, inputs, degree int) (int, bool) {
	if len(levels) == 0 || inputs < 1 || degree < 1 {
		return 0, false
	}
	count, total := 0, 0
	previous := levels[0]
	for _, entry := range levels {
		if entry < 0 {
			return 0, false
		}
		total += entry
	}
	for index := 1; index < len(levels); index++ {
		entry := levels[index]
		maximum := degree * previous
		if maximum < entry {
			return 0, false
		}
		if maximum != entry {
			below := inputs
			for _, suffix := range levels[index:] {
				below += suffix
			}
			count += below * (maximum - 1)
		}
		previous = entry
	}
	count += inputs * degree * previous
	if total < inputs-1 {
		return 0, true
	}
	return count, true
}

func sbdIndependentTopologies(inputs, nodes, degree int) SBDTopologyResult {
	configurations := [][]int{nil}
	start := 0
	result := SBDTopologyResult{Inputs: inputs, Nodes: nodes, Degree: degree, Stages: make([]SBDTopologyStage, 0, nodes)}
	add := func(candidate []int) {
		for _, existing := range configurations {
			if slices.Equal(existing, candidate) {
				return
			}
		}
		configurations = append(configurations, slices.Clone(candidate))
	}
	for node := 1; node <= nodes; node++ {
		stop := len(configurations)
		for entry := start; entry < stop; entry++ {
			working := slices.Clone(configurations[entry])
			for level := range working {
				if level == 0 && working[0] > 0 || level > 0 && working[level] == degree*working[level-1] {
					continue
				}
				working[level]++
				add(working)
				working[level]--
			}
			add(append(working, 1))
		}
		stage := SBDTopologyStage{Nodes: node, New: len(configurations) - stop, Total: len(configurations)}
		for configuration := stop; configuration < len(configurations); configuration++ {
			count, _ := sbdIndependentCount(configurations[configuration], inputs, degree)
			stage.Configurations = append(stage.Configurations, SBDTopologyConfiguration{Levels: slices.Clone(configurations[configuration]), ConfigurationVariables: count})
		}
		result.Stages = append(result.Stages, stage)
		start = stop
	}
	return result
}

func FuzzSBDCountTopologies(f *testing.F) {
	f.Add(uint8(2), uint8(3), uint8(2))
	f.Add(uint8(5), uint8(5), uint8(3))
	f.Fuzz(func(t *testing.T, inputByte, nodeByte, degreeByte uint8) {
		inputs := 1 + int(inputByte%6)
		nodes := int(nodeByte % 8)
		degree := 1 + int(degreeByte%4)
		got, err := SBDCountTopologies(inputs, nodes, degree, SBDTopologyLimits{MaxConfigurations: 10000, MaxLevelEntries: 100000, MaxWork: 1000000})
		if err != nil {
			t.Fatal(err)
		}
		want := sbdIndependentTopologies(inputs, nodes, degree)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("inputs=%d nodes=%d degree=%d got=%+v want=%+v", inputs, nodes, degree, got, want)
		}
	})
}

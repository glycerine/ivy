package abc

// Stand-alone tests for SOP expansion from bmcExpand.c.

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"testing"
)

func sopTestMinterm(word, variables int) string {
	result := make([]byte, variables)
	for index := range result {
		result[index] = '0' + byte(word>>index&1)
	}
	return string(result)
}

func sopTestMatches(cube string, word int) bool {
	for variable := range cube {
		if cube[variable] == '-' {
			continue
		}
		if (word>>variable&1 != 0) != (cube[variable] == '1') {
			return false
		}
	}
	return true
}

func sopTestExpected(cubes []string, variables int, truth uint64) []string {
	all := 1 << variables
	expanded := slices.Clone(cubes)
	for cubeIndex, original := range cubes {
		candidate := []byte(original)
		for variable := variables - 1; variable >= 0; variable-- {
			if candidate[variable] == '-' {
				continue
			}
			saved := candidate[variable]
			candidate[variable] = '-'
			implicant := true
			for word := 0; word < all; word++ {
				if sopTestMatches(string(candidate), word) && truth>>word&1 == 0 {
					implicant = false
					break
				}
			}
			if !implicant {
				candidate[variable] = saved
			}
		}
		expanded[cubeIndex] = string(candidate)
	}
	kept := make([]bool, len(expanded))
	for index := range kept {
		kept[index] = true
	}
	for current, cube := range expanded {
		redundant := true
		for word := 0; word < all && redundant; word++ {
			if !sopTestMatches(cube, word) {
				continue
			}
			covered := false
			for other, otherCube := range expanded {
				if other != current && kept[other] && sopTestMatches(otherCube, word) {
					covered = true
					break
				}
			}
			if !covered {
				redundant = false
			}
		}
		if redundant {
			kept[current] = false
		}
	}
	result := make([]string, 0, len(expanded))
	for index, cube := range expanded {
		if kept[index] {
			result = append(result, cube)
		}
	}
	return result
}

func sopTestGraph(t testing.TB, variables int, truth uint64) *GiaManager {
	t.Helper()
	manager := NewGiaManager()
	if err := manager.HashStart(); err != nil {
		t.Fatal(err)
	}
	inputs := make([]GiaLit, variables)
	for index := range inputs {
		inputs[index] = giaTestCI(t, manager)
	}
	result := GiaConst0
	for word := 0; word < 1<<variables; word++ {
		if truth>>word&1 == 0 {
			continue
		}
		product := GiaConst1
		for variable, input := range inputs {
			literal := input.NotIf(word>>variable&1 == 0)
			var err error
			product, err = manager.HashAnd(product, literal)
			if err != nil {
				t.Fatal(err)
			}
		}
		var err error
		result, err = manager.HashOr(result, product)
		if err != nil {
			t.Fatal(err)
		}
	}
	giaTestCO(t, manager, result)
	giaTestCheck(t, manager)
	return manager
}

func TestSOPCoverParseFormatAndEvaluate(t *testing.T) {
	for _, text := range []string{"1- 1\n-1 1\n", "11 x\n", "11 n\n", " 0\n", " 1\n"} {
		cover, err := ParseSOPCover(text, SOPExpansionOptions{})
		if err != nil {
			t.Fatal(text, err)
		}
		formatted, err := FormatSOPCover(cover)
		if err != nil || formatted != text {
			t.Fatal(text, formatted, err)
		}
	}
	xor, _ := ParseSOPCover("11 x\n", SOPExpansionOptions{})
	xnor, _ := ParseSOPCover("11 n\n", SOPExpansionOptions{})
	for word := 0; word < 4; word++ {
		inputs := []bool{word&1 != 0, word&2 != 0}
		x, err := xor.Evaluate(inputs)
		n, nErr := xnor.Evaluate(inputs)
		if err != nil || nErr != nil || x == n || x != (inputs[0] != inputs[1]) {
			t.Fatal(word, x, n, err, nErr)
		}
	}
	for _, malformed := range []string{"", "1 1", "1 2\n", "1 1\n0 0\n", "1z 1\n"} {
		if cover, err := ParseSOPCover(malformed, SOPExpansionOptions{}); len(cover.Cubes) != 0 || err == nil {
			t.Fatal(malformed, cover, err)
		}
	}
}

func TestExpandSOPCubesWithSolverReverseOrder(t *testing.T) {
	solver, err := NewSatSolverWithLimit(2)
	if err != nil {
		t.Fatal(err)
	}
	if err = solver.SetVariableCount(2); err != nil {
		t.Fatal(err)
	}
	a := boundedCoreLit(0, true)
	b := boundedCoreLit(1, true)
	if ok, addErr := solver.AddClause([]SatLit{a}); addErr != nil || !ok {
		t.Fatal(ok, addErr)
	}
	if ok, addErr := solver.AddClause([]SatLit{b}); addErr != nil || !ok {
		t.Fatal(ok, addErr)
	}
	cover := SOPCover{Variables: 2, Cubes: []string{"10"}, Output: '1'}
	result, err := ExpandSOPCubesWithSolver(cover, solver, []SatVar{0, 1}, SOPExpansionOptions{})
	if err != nil || result.Status != SatSatisfiable || !slices.Equal(result.Cover.Cubes, []string{"1-"}) || !result.Changed {
		t.Fatal(result, err)
	}
	if !slices.Equal(cover.Cubes, []string{"10"}) {
		t.Fatal("input cover mutated", cover)
	}
}

func TestExpandGiaSOPCoverExactAndConstants(t *testing.T) {
	truth := uint64(0)
	for word := 0; word < 8; word++ {
		a, b, c := word&1 != 0, word&2 != 0, word&4 != 0
		if a && b || c {
			truth |= 1 << word
		}
	}
	manager := sopTestGraph(t, 3, truth)
	cubes := make([]string, 0)
	for word := 0; word < 8; word++ {
		if truth>>word&1 != 0 {
			cubes = append(cubes, sopTestMinterm(word, 3))
		}
	}
	cover := SOPCover{Variables: 3, Cubes: cubes, Output: '1'}
	result, err := manager.ExpandGiaSOPCover(0, []int{0, 1, 2}, cover, SOPExpansionOptions{})
	want := sopTestExpected(cubes, 3, truth)
	if err != nil || result.Status != SatSatisfiable || !slices.Equal(result.Cover.Cubes, want) {
		t.Fatal(result.Cover.Cubes, want, result, err)
	}
	for word := 0; word < 8; word++ {
		got, evalErr := result.Cover.Evaluate([]bool{word&1 != 0, word&2 != 0, word&4 != 0})
		if evalErr != nil || got != (truth>>word&1 != 0) {
			t.Fatal(word, got, evalErr, result.Cover)
		}
	}
	for _, constant := range []struct {
		truth uint64
		out   byte
	}{
		{0, '0'}, {15, '1'},
	} {
		graph := sopTestGraph(t, 2, constant.truth)
		constantResult, expandErr := graph.ExpandGiaSOPCover(0, []int{0, 1}, SOPCover{Variables: 2, Cubes: []string{"00"}, Output: '1'}, SOPExpansionOptions{})
		if expandErr != nil || !constantResult.Constant || constantResult.Cover.Variables != 0 || constantResult.Cover.Output != constant.out || !slices.Equal(constantResult.Cover.Cubes, []string{""}) {
			t.Fatal(constant, constantResult, expandErr)
		}
	}
}

func TestExpandSOPNetworkAndExactSorting(t *testing.T) {
	manager := sopTestGraph(t, 3, 0xfe)
	network := SOPNetwork{
		InputCount: 3,
		Nodes: []SOPNode{
			{Fanins: []int{0, 1, 2}, Cover: SOPCover{Variables: 3, Cubes: []string{"1--", "-1-", "--1"}, Output: '1'}},
			{Fanins: []int{2, 0, 1}, Cover: SOPCover{Variables: 3, Cubes: []string{"1-0", "-11", "00-"}, Output: '1'}},
		},
		Outputs: []int{0},
	}
	result, err := manager.ExpandSOPNetwork(network, SOPExpansionOptions{})
	if err != nil || result.Outputs[0].Status != SatSatisfiable {
		t.Fatal(result, err)
	}
	if !slices.Equal(result.Network.Nodes[0].Fanins, []int{2, 1, 0}) || !slices.Equal(result.Network.Nodes[0].Cover.Cubes, []string{"--1", "-1-", "1--"}) {
		t.Fatal(result.Network.Nodes[0])
	}
	if !slices.Equal(result.Network.Nodes[1].Fanins, []int{2, 1, 0}) || !slices.Equal(result.Network.Nodes[1].Cover.Cubes, []string{"-11", "0-0", "10-"}) {
		t.Fatal(result.Network.Nodes[1])
	}
	if !slices.Equal(network.Nodes[0].Fanins, []int{0, 1, 2}) || !slices.Equal(network.Nodes[1].Cover.Cubes, []string{"1-0", "-11", "00-"}) {
		t.Fatal("source network mutated", network)
	}
}

func TestExpandSOPLimitsCancellationUnknownAndDeep(t *testing.T) {
	manager := sopTestGraph(t, 2, 8)
	cover := SOPCover{Variables: 2, Cubes: []string{"11"}, Output: '1'}
	bad := []struct {
		order   []int
		cover   SOPCover
		options SOPExpansionOptions
	}{
		{[]int{0}, cover, SOPExpansionOptions{}},
		{[]int{0, 0}, cover, SOPExpansionOptions{}},
		{[]int{0, 1}, SOPCover{Variables: 2, Cubes: []string{"1"}, Output: '1'}, SOPExpansionOptions{}},
		{[]int{0, 1}, cover, SOPExpansionOptions{MaxVariables: -1}},
		{[]int{0, 1}, cover, SOPExpansionOptions{MaxVariables: 1}},
		{[]int{0, 1}, cover, SOPExpansionOptions{MaxClauses: 1}},
		{[]int{0, 1}, cover, SOPExpansionOptions{MaxLiterals: 1}},
		{[]int{0, 1}, cover, SOPExpansionOptions{MaxCubes: 1, MaxCubeEntries: 1}},
	}
	for index, test := range bad {
		if result, err := manager.ExpandGiaSOPCover(0, test.order, test.cover, test.options); len(result.Cover.Cubes) != 0 || err == nil {
			t.Fatal(index, result, err)
		}
	}
	if result, err := manager.ExpandGiaSOPCover(0, []int{0, 1}, cover, SOPExpansionOptions{Stop: func() bool { return true }}); len(result.Cover.Cubes) != 0 || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	unknown, err := manager.ExpandGiaSOPCover(0, []int{0, 1}, cover, SOPExpansionOptions{SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || unknown.Status != SatUnknown || unknown.Reason != SatCancelled || !slices.Equal(unknown.Cover.Cubes, cover.Cubes) {
		t.Fatal(unknown, err)
	}
	irredundantOptions, _ := checkSOPExpansionOptions(SOPExpansionOptions{SAT: SatLimits{Stop: func() bool { return true }}})
	irredundantInput := SOPCover{Variables: 2, Cubes: []string{"00", "0-"}, Output: '1'}
	irredundant, _, status, reason, irredundantErr := sopCoverIrredundant(irredundantInput, irredundantOptions)
	if irredundantErr != nil || status != SatUnknown || reason != SatCancelled || !slices.Equal(irredundant.Cubes, irredundantInput.Cubes) {
		t.Fatal(irredundant, status, reason, irredundantErr)
	}
	deep := NewGiaManager()
	if err = deep.HashStart(); err != nil {
		t.Fatal(err)
	}
	a, b := giaTestCI(t, deep), giaTestCI(t, deep)
	current, err := deep.HashAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10000; index++ {
		current, err = deep.AppendAnd(current, a.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	giaTestCO(t, deep, current)
	deepResult, err := deep.ExpandGiaSOPCover(0, []int{0, 1}, cover, SOPExpansionOptions{MaxVariables: 20000, MaxClauses: 40000, MaxLiterals: 100000})
	if err != nil || deepResult.Status != SatSatisfiable {
		t.Fatal(deepResult.Status, err)
	}
}

func FuzzExpandGiaSOPCover(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x6996, 0xffff, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		variables := 1 + int(seed%4)
		mask := uint64(1)<<(1<<variables) - 1
		truth := (seed >> 8) & mask
		manager := sopTestGraph(t, variables, truth)
		cubes := make([]string, 0)
		for word := 0; word < 1<<variables; word++ {
			if truth>>word&1 != 0 {
				cubes = append(cubes, sopTestMinterm(word, variables))
			}
		}
		if len(cubes) == 0 {
			cubes = []string{sopTestMinterm(0, variables)}
		}
		order := make([]int, variables)
		for index := range order {
			order[index] = index
		}
		cover := SOPCover{Variables: variables, Cubes: cubes, Output: '1'}
		result, err := manager.ExpandGiaSOPCover(0, order, cover, SOPExpansionOptions{MaxVariables: 1000, MaxClauses: 10000, MaxLiterals: 30000})
		if err != nil || result.Status != SatSatisfiable {
			t.Fatal(seed, variables, truth, result, err)
		}
		if truth == 0 || truth == mask {
			want := byte('0')
			if truth == mask {
				want = '1'
			}
			if !result.Constant || result.Cover.Output != want {
				t.Fatal(seed, result)
			}
			return
		}
		want := sopTestExpected(cubes, variables, truth)
		if !slices.Equal(result.Cover.Cubes, want) {
			t.Fatal(seed, truth, result.Cover.Cubes, want)
		}
		for word := 0; word < 1<<variables; word++ {
			inputs := make([]bool, variables)
			for variable := range inputs {
				inputs[variable] = word>>variable&1 != 0
			}
			got, evalErr := result.Cover.Evaluate(inputs)
			if evalErr != nil || got != (truth>>word&1 != 0) {
				t.Fatal(seed, word, got, evalErr, result.Cover)
			}
		}
	})
}

func TestSOPExpectedModelSanity(t *testing.T) {
	// Keep the independent oracle honest on a cover where direct-order removal
	// matters: the first cube is removed before the third is queried.
	cubes := []string{"00", "0-", "-0"}
	truth := uint64(0)
	for word := 0; word < 4; word++ {
		if word&1 == 0 || word&2 == 0 {
			truth |= 1 << word
		}
	}
	want := sopTestExpected(cubes, 2, truth)
	if len(want) == 0 {
		t.Fatal(fmt.Sprint(want))
	}
	sorted := slices.Clone(want)
	sort.Strings(sorted)
	if !slices.Equal(sorted, []string{"-0", "0-"}) {
		t.Fatal(want)
	}
}

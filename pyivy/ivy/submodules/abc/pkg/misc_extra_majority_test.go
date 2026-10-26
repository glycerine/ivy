package abc

// Go-only exhaustive, model, lifecycle, and fuzz tests for extraUtilMaj.c.

import (
	"errors"
	"math/bits"
	"slices"
	"strings"
	"testing"
)

func gemTestTruth(t *testing.T, variables int, value func(int) bool) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for bit := 0; bit < len(table.words)*64; bit++ {
		assignment := bit
		if variables < 6 {
			assignment &= (1 << uint(variables)) - 1
		}
		if value(assignment) {
			table.words[bit>>6] |= uint64(1) << uint(bit&63)
		}
	}
	return table
}

func TestGemGroupOperationsAndDerivation(t *testing.T) {
	removed, err := GemGroupVarRemove(0b1011, 1)
	inserted, err2 := GemGroupVarsInsert1(removed, 0, true)
	three, err3 := GemGroupVarsInsert3(1, 0)
	front, err4 := GemGroupVarsInsert3(1, -1)
	if err != nil || err2 != nil || err3 != nil || err4 != nil || removed != 0b0101 || inserted != 0b1011 || three != 0b1001 || front != 0b1100 {
		t.Fatal(removed, inserted, three, front, err, err2, err3, err4)
	}
	if got := GemGroupUnpack(0x8109); !slices.Equal(got, []int{0, 3, 8, 15}) {
		t.Fatal(got)
	}
	if _, err = GemGroupVarRemove(1, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = GemGroupVarsInsert1(1, 0, true); err != nil {
		t.Fatal(err)
	}
	if _, err = GemGroupVarsInsert1(1, 1, true); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = GemGroupVarsInsert3(1|1<<13, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = GemGroupVarsInsert1(1|1<<15, 0, true); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}

	majority := gemTestTruth(t, 3, func(a int) bool { return a&1+a>>1&1+a>>2&1 >= 2 })
	groups, err := GemGroupsDerive(majority)
	if err != nil || groups != 0b100 {
		t.Fatal(groups, err)
	}
	nonsymmetric := gemTestTruth(t, 3, func(a int) bool { return a&1 != 0 && a&4 != 0 })
	groups, err = GemGroupsDerive(nonsymmetric)
	if err != nil || groups != 0b111 {
		t.Fatal(groups, err)
	}
}

func TestGemFunctionFindPlace(t *testing.T) {
	variable, _ := NewVariableTruthTable(4, 0, 0)
	before := variable.Words()
	best, place, err := GemFunctionFindPlace(variable, 1, true)
	want, _ := NewVariableTruthTable(4, 1, 0)
	if err != nil || place != -1 || !best.Equal(want) || !slices.Equal(variable.Words(), before) {
		t.Fatal(place, err, best.Words(), want.Words())
	}
	wide, _ := NewVariableTruthTable(6, 0, 0)
	best, place, err = GemFunctionFindPlace(wide, 1, false)
	if err != nil || place != 0 || !best.Equal(wide) {
		t.Fatal(place, err)
	}
	if _, _, err = GemFunctionFindPlace(variable, 1<<4, true); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestGemManagerExpandReduceAndMajority(t *testing.T) {
	manager, err := NewGemManager(5, true, GemLimits{})
	if err != nil {
		t.Fatal(err)
	}
	statistics, err := manager.Statistics()
	if err != nil || statistics.Objects != 2 || statistics.Variables != 5 {
		t.Fatal(statistics, err)
	}
	objects, _ := manager.Objects()
	objects[1].Groups = 0
	again, _ := manager.Objects()
	if again[1].Groups != 1 {
		t.Fatal("object snapshot aliases manager")
	}
	index, inserted, err := manager.Expand(1, 0)
	if err != nil || !inserted || index != 2 {
		t.Fatal(index, inserted, err)
	}
	object := again[0]
	objects, _ = manager.Objects()
	object = objects[index]
	if object != (GemObject{Variables: 3, Nodes: 1, History: 0, Groups: 4, Predecessor: 1}) {
		t.Fatal(object)
	}
	report, err := manager.CheckMajority(index)
	if err != nil || report != (GemSymmetryReport{Symmetric: true, Majority: true, Polar: 12}) {
		t.Fatal(report, err)
	}
	xor := gemTestTruth(t, 3, func(a int) bool { return bits.OnesCount(uint(a))&1 != 0 })
	xorIndex, _, err := manager.insert(xor, GemObject{Variables: 3, Groups: 4, Predecessor: -1})
	report, reportErr := manager.CheckMajority(xorIndex)
	if err != nil || reportErr != nil || report != (GemSymmetryReport{Symmetric: true, Polar: 10}) {
		t.Fatal(report, err, reportErr)
	}
	nonsymmetric := gemTestTruth(t, 3, func(a int) bool { return a&1 != 0 && a&4 != 0 })
	nonIndex, _, err := manager.insert(nonsymmetric, GemObject{Variables: 3, Groups: 7, Predecessor: -1})
	report, reportErr = manager.CheckMajority(nonIndex)
	if err != nil || reportErr != nil || report != (GemSymmetryReport{}) {
		t.Fatal(report, err, reportErr)
	}
	for assignment := 0; assignment < 8; assignment++ {
		table, _ := manager.TruthTable(index)
		got, _ := table.Get(uint64(assignment))
		want := assignment&1+assignment>>1&1+assignment>>2&1 >= 2
		if got != want {
			t.Fatal(assignment, got, want)
		}
	}
	line, err := manager.FormatObject(index, "Expand  ", true)
	if err != nil || !strings.Contains(line, "Node      2 : Expand    Pred =      1  Vars = 3  Nodes = 1  History = 00  Profile: 00100 *\n") {
		t.Fatal(line, err)
	}
	reduced, added, err := manager.Reduce(index, 0, 1)
	if err != nil || added || reduced != 1 {
		t.Fatal(reduced, added, err)
	}
	if _, _, err = manager.Reduce(index, 1, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if !manager.Close() || manager.Close() {
		t.Fatal("close lifecycle")
	}
	if _, err = manager.Objects(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestGemEnumerationAndLimits(t *testing.T) {
	result, err := EnumerateGemMajority(3, GemEnumerationOptions{DumpTables: true, Verbose: true})
	if err != nil || len(result.Objects) != 3 || len(result.Tables) != 3 || len(result.Stages) != 1 || len(result.Diagnostics) != 3 {
		t.Fatal(len(result.Objects), len(result.Tables), result.Stages, len(result.Diagnostics), err)
	}
	if result.Stages[0] != (GemStage{Variable: 1, Before: 2, AfterExpand: 3, AfterReduce: 3}) {
		t.Fatal(result.Stages)
	}
	result.Tables[1].words[0] = 0
	again, err := EnumerateGemMajority(3, GemEnumerationOptions{DumpTables: true})
	if err != nil || again.Tables[1].words[0] == 0 {
		t.Fatal("dump ownership", err)
	}
	if _, err = NewGemManager(0, false, GemLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = NewGemManager(3, false, GemLimits{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = NewGemManager(3, false, GemLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	limited, _ := NewGemManager(3, false, GemLimits{MaxObjects: 2})
	if _, _, err = limited.Expand(1, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	objects, _ := limited.Objects()
	if len(objects) != 2 {
		t.Fatal("failed insert changed objects")
	}
	workLimited, _ := NewGemManager(3, false, GemLimits{MaxWork: 7})
	if _, _, err = workLimited.Expand(1, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = EnumerateGemMajority(3, GemEnumerationOptions{Verbose: true, Limits: GemLimits{MaxOutputBytes: 1}}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	calls := 0
	cancelled, _ := NewGemManager(5, false, GemLimits{Stop: func() bool { calls++; return calls > 2 }})
	if _, _, err = cancelled.Expand(1, 0); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func gemModelExpand(t *TruthTable, variable int) *TruthTable {
	oldVariables := t.variables
	return gemTestTruthNoTest(oldVariables+2, func(assignment int) bool {
		last := assignment >> uint(oldVariables-1) & 7
		replacement := last&1+last>>1&1+last>>2&1 >= 2
		oldAssignment := 0
		for old := 0; old < oldVariables; old++ {
			value := replacement
			if old != variable {
				position := old
				if old > variable {
					position--
				}
				value = assignment&(1<<uint(position)) != 0
			}
			if value {
				oldAssignment |= 1 << uint(old)
			}
		}
		value, _ := t.Get(uint64(oldAssignment))
		return value
	})
}

func gemModelReduce(t *TruthTable, first, second int) *TruthTable {
	newVariables := t.variables - 1
	return gemTestTruthNoTest(newVariables, func(assignment int) bool {
		replacement := assignment&(1<<uint(newVariables-1)) != 0
		oldAssignment, position := 0, 0
		for old := 0; old < t.variables; old++ {
			value := replacement
			if old != first && old != second {
				value = assignment&(1<<uint(position)) != 0
				position++
			}
			if value {
				oldAssignment |= 1 << uint(old)
			}
		}
		value, _ := t.Get(uint64(oldAssignment))
		return value
	})
}

func gemTestTruthNoTest(variables int, value func(int) bool) *TruthTable {
	table, _ := NewConstantTruthTable(variables, false, 0)
	for bit := 0; bit < len(table.words)*64; bit++ {
		assignment := bit
		if variables < 6 {
			assignment &= (1 << uint(variables)) - 1
		}
		if value(assignment) {
			table.words[bit>>6] |= uint64(1) << uint(bit&63)
		}
	}
	return table
}

func gemPermutationEquivalent(first, second *TruthTable) bool {
	if first.variables != second.variables || first.variables > 7 {
		return false
	}
	permutation := make([]int, first.variables)
	used := make([]bool, first.variables)
	var visit func(int) bool
	visit = func(position int) bool {
		if position == len(permutation) {
			for assignment := 0; assignment < 1<<uint(first.variables); assignment++ {
				mapped := 0
				for target, source := range permutation {
					if assignment&(1<<uint(target)) != 0 {
						mapped |= 1 << uint(source)
					}
				}
				a, _ := first.Get(uint64(mapped))
				b, _ := second.Get(uint64(assignment))
				if a != b {
					return false
				}
			}
			return true
		}
		for source := range permutation {
			if used[source] {
				continue
			}
			used[source] = true
			permutation[position] = source
			if visit(position + 1) {
				return true
			}
			used[source] = false
		}
		return false
	}
	return visit(0)
}

func FuzzGemMajority(f *testing.F) {
	for _, seed := range []struct{ operations, choices uint8 }{{1, 0}, {2, 1}, {5, 0xff}} {
		f.Add(seed.operations, seed.choices)
	}
	f.Fuzz(func(t *testing.T, operations, choices uint8) {
		manager, err := NewGemManager(5, false, GemLimits{MaxObjects: 128, MaxWork: 1 << 18})
		if err != nil {
			t.Fatal(err)
		}
		current := 1
		for step := 0; step < int(operations%8)+1; step++ {
			source, _ := manager.TruthTable(current)
			object := manager.objects[current]
			if object.Variables+2 <= 5 && (choices>>uint(step&7))&1 == 0 {
				variable := int(choices) % object.Variables
				model := gemModelExpand(source, variable)
				next, inserted, expandErr := manager.Expand(current, variable)
				if expandErr != nil {
					t.Fatal(expandErr)
				}
				got, _ := manager.TruthTable(next)
				if got.variables < model.variables {
					got, _ = got.Stretch(model.variables)
				}
				if !gemPermutationEquivalent(model, got) {
					t.Fatal("expand orbit", step, inserted)
				}
				if inserted {
					groups, _ := GemGroupsDerive(got)
					if groups != manager.objects[next].Groups {
						t.Fatal("expand groups")
					}
				}
				current = next
				continue
			}
			if object.Variables < 2 {
				continue
			}
			first := int(choices) % (object.Variables - 1)
			second := first + 1 + int(operations)%((object.Variables-1)-first)
			model := gemModelReduce(source, first, second)
			next, inserted, reduceErr := manager.Reduce(current, first, second)
			if reduceErr != nil {
				t.Fatal(reduceErr)
			}
			got, _ := manager.TruthTable(next)
			if got.variables < model.variables {
				got, _ = got.Stretch(model.variables)
			}
			if !gemPermutationEquivalent(model, got) {
				t.Fatal("reduce orbit", step, inserted)
			}
			current = next
		}
		objects, _ := manager.Objects()
		for i := range objects {
			table, _ := manager.TruthTable(i)
			key, _ := gemTruthKey(table, 5)
			if manager.unique[key] != i {
				t.Fatal("unique table", i)
			}
		}
	})
}

package abc

// Stand-alone tests for src/bool/rpo/literal.h, rpo.h, and rpo.c.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func rpoTruth(t *testing.T, variables int, logical uint64) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
		if logical&(uint64(1)<<assignment) != 0 {
			if err = table.Set(assignment, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func rpoLogical(table *TruthTable) uint64 {
	var result uint64
	for assignment := uint64(0); assignment < uint64(1)<<uint(table.Variables()); assignment++ {
		value, _ := table.Get(assignment)
		if value {
			result |= uint64(1) << assignment
		}
	}
	return result
}

func rpoTransitionModel(logical uint64, variables, variable int, positive bool) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		zero := assignment &^ (1 << uint(variable))
		one := zero | 1<<uint(variable)
		z := logical&(uint64(1)<<uint(zero)) != 0
		o := logical&(uint64(1)<<uint(one)) != 0
		value := !z && o
		if !positive {
			value = z && !o
		}
		if value {
			result |= uint64(1) << uint(assignment)
		}
	}
	return result
}

// rpoReadOnceModel independently enumerates every function expressible with
// constants or an AND/OR tree whose leaves use each variable polarity at most
// once. The production search and its transition/grouping predicates are not
// used.
func rpoReadOnceModel(variables int) map[uint64]bool {
	assignments := 1 << uint(variables)
	full := uint64(1)<<uint(assignments) - 1
	resources := 2 * variables
	bySupport := make([]map[uint64]bool, 1<<uint(resources))
	for variable := 0; variable < variables; variable++ {
		positive := truthVariablePattern(variable) & full
		bySupport[1<<uint(2*variable)] = map[uint64]bool{positive: true}
		bySupport[1<<uint(2*variable+1)] = map[uint64]bool{(^positive) & full: true}
	}
	for size := 2; size <= resources; size++ {
		for support := 1; support < len(bySupport); support++ {
			if bits.OnesCount(uint(support)) != size {
				continue
			}
			functions := make(map[uint64]bool)
			for left := (support - 1) & support; left != 0; left = (left - 1) & support {
				right := support ^ left
				if right == 0 || left > right {
					continue
				}
				for leftFunction := range bySupport[left] {
					for rightFunction := range bySupport[right] {
						functions[leftFunction&rightFunction] = true
						functions[leftFunction|rightFunction] = true
					}
				}
			}
			bySupport[support] = functions
		}
	}
	all := map[uint64]bool{0: true, full: true}
	for _, functions := range bySupport {
		for function := range functions {
			all[function] = true
		}
	}
	return all
}

func TestRPOTransitionsExhaustive(t *testing.T) {
	for variables := 1; variables <= 3; variables++ {
		functions := 1 << uint(1<<uint(variables))
		for logical := 0; logical < functions; logical++ {
			table := rpoTruth(t, variables, uint64(logical))
			for variable := 0; variable < variables; variable++ {
				positive, err := RPOPositiveTransition(table, variable)
				if err != nil || rpoLogical(positive) != rpoTransitionModel(uint64(logical), variables, variable, true) {
					t.Fatalf("positive vars=%d logical=%x variable=%d got=%x err=%v", variables, logical, variable, rpoLogical(positive), err)
				}
				negative, err := RPONegativeTransition(table, variable)
				if err != nil || rpoLogical(negative) != rpoTransitionModel(uint64(logical), variables, variable, false) {
					t.Fatalf("negative vars=%d logical=%x variable=%d got=%x err=%v", variables, logical, variable, rpoLogical(negative), err)
				}
			}
		}
	}
}

func TestRPOLiteralsGroupsAndOwnership(t *testing.T) {
	target := rpoTruth(t, 3, truthVariablePattern(0)&truthVariablePattern(1)&0xff)
	a, present, err := NewRPOLiteral(target, 0, true)
	if err != nil || !present {
		t.Fatal(present, err)
	}
	b, present, err := NewRPOLiteral(target, 1, true)
	if err != nil || !present {
		t.Fatal(present, err)
	}
	and, err := RPOCheckANDGroup(a, b)
	if err != nil || !and {
		t.Fatal(and, err)
	}
	or, err := RPOCheckORGroup(a, b)
	if err != nil || or {
		t.Fatal(or, err)
	}
	grouped, err := GroupRPOLiterals(a, b, RPOOperatorAND, 0)
	if err != nil {
		t.Fatal(err)
	}
	if expression, _ := grouped.Expression(); expression != "(a*b)" {
		t.Fatal(expression)
	}
	function, _ := grouped.Function()
	if rpoLogical(function) != rpoLogical(target) {
		t.Fatalf("%x %x", rpoLogical(function), rpoLogical(target))
	}
	_ = function.Set(0, true)
	again, _ := grouped.Function()
	if rpoLogical(again) != rpoLogical(target) {
		t.Fatal("returned truth table aliased literal")
	}
	clone, err := grouped.Clone()
	if err != nil || clone == grouped || clone.function == grouped.function || clone.transition == grouped.transition {
		t.Fatal(err)
	}
	if _, err = GroupRPOLiterals(a, b, RPOOperatorXOR, 0); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err = GroupRPOLiterals(a, b, RPOOperatorAND, 3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func TestRPOEdgesOrderingAndFormatting(t *testing.T) {
	degrees := make([]int, 4)
	edges := make([]RPOEdge, 0)
	for _, row := range []struct {
		op            RPOOperator
		first, second int
	}{{RPOOperatorAND, 1, 3}, {RPOOperatorOR, 0, 2}, {RPOOperatorAND, 0, 1}} {
		edge, err := NewRPOEdge(row.op, row.first, row.second, degrees)
		if err != nil {
			t.Fatal(err)
		}
		edges = append(edges, edge)
	}
	index, err := rpoMinimumEdge(edges, degrees)
	if err != nil || index != 1 {
		t.Fatal(index, degrees, err)
	}
	text, err := FormatRPOEdge(edges[1])
	if err != nil || text != "Edge (0,2)/2\n" {
		t.Fatal(text, err)
	}
	if _, err = NewRPOEdge(RPOOperatorAND, 1, 1, degrees); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	tiedDegrees := []int{2, 0, 1, 1}
	tied := []RPOEdge{{Operator: RPOOperatorAND, First: 0, Second: 2}, {Operator: RPOOperatorOR, First: 0, Second: 3}}
	if index, err = rpoMinimumEdge(tied, tiedDegrees); err != nil || index != 0 {
		t.Fatalf("equal ties must retain source insertion order: index=%d err=%v", index, err)
	}
	truthText, err := FormatRPOTruth(rpoTruth(t, 2, 8))
	if err != nil || truthText != "88888888" {
		t.Fatal(truthText, err)
	}
	wide, err := TruthTableFromWords(7, []uint64{0x0123456789abcdef, 0xfedcba9876543210}, 0)
	if err != nil {
		t.Fatal(err)
	}
	truthText, err = FormatRPOTruth(wide)
	if err != nil || truthText != "FEDCBA98765432100123456789ABCDEF" {
		t.Fatal(truthText, err)
	}
}

func rpoEvalExpression(expression string, assignment uint64) (bool, bool) {
	position := 0
	var parse func() (bool, bool)
	parse = func() (bool, bool) {
		if position >= len(expression) {
			return false, false
		}
		if expression[position] == '(' {
			position++
			left, ok := parse()
			if !ok || position >= len(expression) {
				return false, false
			}
			op := expression[position]
			position++
			right, ok := parse()
			if !ok || position >= len(expression) || expression[position] != ')' {
				return false, false
			}
			position++
			if op == '*' {
				return left && right, true
			}
			if op == '+' {
				return left || right, true
			}
			return false, false
		}
		negated := false
		if expression[position] == '!' {
			negated = true
			position++
		}
		if position >= len(expression) {
			return false, false
		}
		if expression[position] == '0' || expression[position] == '1' {
			value := expression[position] == '1'
			position++
			if negated {
				value = !value
			}
			return value, true
		}
		variable := int(expression[position] - 'a')
		if variable < 0 || variable >= 64 {
			return false, false
		}
		position++
		value := assignment&(uint64(1)<<uint(variable)) != 0
		if negated {
			value = !value
		}
		return value, true
	}
	value, ok := parse()
	return value, ok && position == len(expression)
}

func rpoCheckFactor(t *testing.T, target *TruthTable, literal *RPOLiteral) {
	t.Helper()
	if literal == nil {
		return
	}
	expression, err := literal.Expression()
	if err != nil {
		t.Fatal(err)
	}
	function, err := literal.Function()
	if err != nil || !function.Equal(target) {
		t.Fatalf("function mismatch %s %v", expression, err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(target.Variables()); assignment++ {
		got, ok := rpoEvalExpression(expression, assignment)
		want, _ := target.Get(assignment)
		if !ok || got != want {
			t.Fatalf("expression %s assignment=%d got=%v/%v want=%v", expression, assignment, got, ok, want)
		}
	}
}

func TestRPOFactorizationKnownAndModels(t *testing.T) {
	for variables := 1; variables <= 4; variables++ {
		for _, logical := range []uint64{0, ^uint64(0), truthVariablePattern(0), ^truthVariablePattern(0), truthVariablePattern(0) & truthVariablePattern(min(1, variables-1)), truthVariablePattern(0) | truthVariablePattern(min(1, variables-1))} {
			maskBits := 1 << uint(variables)
			if maskBits < 64 {
				logical &= uint64(1)<<uint(maskBits) - 1
			}
			target := rpoTruth(t, variables, logical)
			literal, stats, err := FactorizeRPO(target, RPOFactorOptions{MaxSteps: 100000})
			if err != nil {
				t.Fatalf("variables=%d logical=%x stats=%+v err=%v", variables, logical, stats, err)
			}
			if literal == nil {
				t.Fatalf("expected factorization variables=%d logical=%x stats=%+v", variables, logical, stats)
			}
			rpoCheckFactor(t, target, literal)
		}
	}
	// Exhaust every Boolean function through three variables. A second run
	// independently checks deterministic search order, including failures.
	for variables := 1; variables <= 3; variables++ {
		model := rpoReadOnceModel(variables)
		functionCount := 1 << uint(1<<uint(variables))
		for logical := 0; logical < functionCount; logical++ {
			target := rpoTruth(t, variables, uint64(logical))
			first, firstStats, err := FactorizeRPO(target, RPOFactorOptions{MaxSteps: 100000})
			if err != nil {
				t.Fatalf("exhaustive variables=%d logical=%x stats=%+v: %v", variables, logical, firstStats, err)
			}
			rpoCheckFactor(t, target, first)
			if (first != nil) != model[uint64(logical)] {
				t.Fatalf("completeness mismatch variables=%d logical=%x found=%v model=%v", variables, logical, first != nil, model[uint64(logical)])
			}
			second, secondStats, err := FactorizeRPO(target, RPOFactorOptions{MaxSteps: 100000})
			if err != nil || firstStats != secondStats || (first == nil) != (second == nil) {
				t.Fatalf("nondeterministic variables=%d logical=%x first=%+v second=%+v err=%v", variables, logical, firstStats, secondStats, err)
			}
			if first != nil {
				firstExpression, _ := first.Expression()
				secondExpression, _ := second.Expression()
				if firstExpression != secondExpression {
					t.Fatalf("nondeterministic expression variables=%d logical=%x: %q != %q", variables, logical, firstExpression, secondExpression)
				}
			}
		}
	}
	var state uint64 = 0x9e3779b97f4a7c15
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		variables := 1 + int(state%4)
		logical := state
		if variables < 4 {
			logical &= uint64(1)<<uint(1<<uint(variables)) - 1
		}
		target := rpoTruth(t, variables, logical)
		literal, _, err := FactorizeRPO(target, RPOFactorOptions{Threshold: 200, MaxSteps: 1000})
		if err != nil && !errors.Is(err, ErrLimit) {
			t.Fatalf("execution=%d variables=%d logical=%x: %v", execution, variables, logical, err)
		}
		if err == nil {
			rpoCheckFactor(t, target, literal)
		}
	}
}

func TestRPOSourceSearchOrder(t *testing.T) {
	patterns := []struct {
		logical    uint64
		expression string
	}{
		{truthVariablePattern(0) & truthVariablePattern(1) & truthVariablePattern(2) & 0xff, "((c*b)*a)"},
		{(truthVariablePattern(0) | truthVariablePattern(1) | truthVariablePattern(2)) & 0xff, "((c+b)+a)"},
		{(^truthVariablePattern(0)) & (^truthVariablePattern(1)) & 0xff, "(!b*!a)"},
	}
	for _, pattern := range patterns {
		target := rpoTruth(t, 3, pattern.logical)
		literal, _, err := FactorizeRPO(target, RPOFactorOptions{})
		if err != nil || literal == nil {
			t.Fatalf("logical=%x literal=%v err=%v", pattern.logical, literal, err)
		}
		expression, err := literal.Expression()
		if err != nil || expression != pattern.expression {
			t.Fatalf("logical=%x expression=%q want=%q err=%v", pattern.logical, expression, pattern.expression, err)
		}
	}
}

func TestRPOErrorsLimitsCancellationAndAtomicity(t *testing.T) {
	target := rpoTruth(t, 3, 0x96)
	if _, _, err := FactorizeRPO(nil, RPOFactorOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err := FactorizeRPO(target, RPOFactorOptions{MaxVariables: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := FactorizeRPO(target, RPOFactorOptions{MaxLiterals: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := FactorizeRPO(target, RPOFactorOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	andTarget := rpoTruth(t, 3, truthVariablePattern(0)&truthVariablePattern(1)&0xff)
	if _, _, err := FactorizeRPO(andTarget, RPOFactorOptions{MaxSteps: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	threeWayAND := rpoTruth(t, 3, truthVariablePattern(0)&truthVariablePattern(1)&truthVariablePattern(2)&0xff)
	if _, _, err := FactorizeRPO(threeWayAND, RPOFactorOptions{MaxEdges: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, _, err := FactorizeRPO(andTarget, RPOFactorOptions{MaxExpressionBytes: 4}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	stopCalls := 0
	if _, _, err := FactorizeRPO(threeWayAND, RPOFactorOptions{Stop: func() bool {
		stopCalls++
		return stopCalls == 4
	}}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	_, stats, err := FactorizeRPO(target, RPOFactorOptions{Threshold: 1})
	if err != nil || !stats.ThresholdReached || stats.Steps != 1 {
		t.Fatal(stats, err)
	}
	before := threeWayAND.Words()
	_, _, _ = FactorizeRPO(threeWayAND, RPOFactorOptions{MaxEdges: 1})
	if !slices.Equal(before, threeWayAND.Words()) {
		t.Fatal("failed factorization mutated target")
	}
}

func FuzzRPOFactorization(f *testing.F) {
	f.Add(uint8(2), uint64(8), uint16(200))
	f.Add(uint8(3), uint64(0x96), uint16(50))
	f.Add(uint8(4), uint64(0x8000), uint16(500))
	f.Fuzz(func(t *testing.T, rawVariables uint8, logical uint64, rawThreshold uint16) {
		variables := 1 + int(rawVariables%4)
		if variables < 4 {
			logical &= uint64(1)<<uint(1<<uint(variables)) - 1
		}
		target := rpoTruth(t, variables, logical)
		threshold := 1 + int(rawThreshold%1000)
		literal, _, err := FactorizeRPO(target, RPOFactorOptions{Threshold: threshold, MaxSteps: 2000, MaxEdges: 256})
		if err != nil {
			if errors.Is(err, ErrLimit) {
				return
			}
			t.Fatal(err)
		}
		rpoCheckFactor(t, target, literal)
		// Independent transition population sanity: every transition bit is a
		// directed Boolean difference and therefore occurs in variable pairs.
		for variable := 0; variable < variables; variable++ {
			positive, _ := RPOPositiveTransition(target, variable)
			negative, _ := RPONegativeTransition(target, variable)
			if bits.OnesCount64(rpoLogical(positive))+bits.OnesCount64(rpoLogical(negative)) > 1<<uint(variables) {
				t.Fatal("transition population")
			}
		}
	})
}

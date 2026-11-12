package rpo

import (
	"fmt"
	"slices"
	"testing"
)

func tableFromFunction(vars int32, fn func(int) bool) []uint32 {
	table := make([]uint32, rpoWordNum(vars))
	for i := 0; i < len(table)*32; i++ {
		if fn(i) {
			rpoTruthSet(table, i)
		}
	}
	return table
}
func evalExpression(t *testing.T, expression string, assignment int) bool {
	t.Helper()
	at := 0
	var parse func() bool
	parse = func() bool {
		if at >= len(expression) {
			t.Fatal("truncated expression")
		}
		c := expression[at]
		at++
		switch c {
		case '!':
			return !parse()
		case '0':
			return false
		case '1':
			return true
		case '(':
			a := parse()
			if at >= len(expression) {
				t.Fatal(expression)
			}
			op := expression[at]
			at++
			b := parse()
			if at >= len(expression) || expression[at] != ')' {
				t.Fatal(expression)
			}
			at++
			if op == '*' {
				return a && b
			}
			if op == '+' {
				return a || b
			}
			t.Fatal("invalid operator", expression)
		default:
			if c >= 'a' && c <= 'z' {
				return assignment&(1<<uint(c-'a')) != 0
			}
			t.Fatal("invalid token", expression)
		}
		return false
	}
	value := parse()
	if at != len(expression) {
		t.Fatal("unparsed suffix", expression)
	}
	return value
}
func TestFactorizeKnownFunctions(t *testing.T) {
	cases := []struct {
		name string
		vars int32
		fn   func(int) bool
	}{
		{"zero", 2, func(int) bool { return false }},
		{"one", 2, func(int) bool { return true }},
		{"variable", 2, func(i int) bool { return i&1 != 0 }},
		{"negated variable", 2, func(i int) bool { return i&2 == 0 }},
		{"and", 2, func(i int) bool { return i&3 == 3 }},
		{"or", 2, func(i int) bool { return i&3 != 0 }},
		{"nested", 3, func(i int) bool { return i&1 != 0 && (i&2 != 0 || i&4 == 0) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target := tableFromFunction(tc.vars, tc.fn)
			before := slices.Clone(target)
			result := Rpo_Factorize(target, tc.vars, -1, 0)
			if result == nil {
				t.Fatal("no factorization")
			}
			if !slices.Equal(before, target) {
				t.Fatal("target mutated")
			}
			for i := 0; i < len(target)*32; i++ {
				if evalExpression(t, result.Expression, i) != tc.fn(i) {
					t.Fatal(result.Expression, i)
				}
			}
			if !slices.Equal(result.Function, target) {
				t.Fatal(result.Function, target)
			}
		})
	}
}
func TestFactorizeAllThreeVariableTables(t *testing.T) {
	found := 0
	for bits := 0; bits < 256; bits++ {
		target := tableFromFunction(3, func(i int) bool { return bits&(1<<uint(i&7)) != 0 })
		result := Rpo_Factorize(target, 3, 20000, 0)
		if result == nil {
			continue
		}
		found++
		for i := 0; i < 8; i++ {
			if evalExpression(t, result.Expression, i) != (bits&(1<<uint(i)) != 0) {
				t.Fatalf("truth %02x: %s", bits, result.Expression)
			}
		}
	}
	if found < 30 {
		t.Fatal("unexpectedly few factorizations", found)
	}
}
func TestTransitionsAndGrouping(t *testing.T) {
	target := []uint32{0x88888888}
	a := Lit_Alloc(target, 2, 0, '+')
	b := Lit_Alloc(target, 2, 1, '+')
	if a == nil || b == nil || a.Transition[0] != 0xcccccccc || b.Transition[0] != 0xaaaaaaaa || !Rpo_CheckANDGroup(a, b, 2) || Rpo_CheckORGroup(a, b, 2) {
		t.Fatal(a, b)
	}
	if Lit_Alloc(target, 2, 0, '-') != nil {
		t.Fatal("spurious negative literal")
	}
	group := Lit_GroupLiterals(a, b, LIT_AND, 2)
	if group.Function[0] != target[0] {
		t.Fatal(group)
	}
}
func TestEdgeCostTieBreakAndThreshold(t *testing.T) {
	degree := make([]int32, 4)
	edges := []*Rpo_LCI_Edge_t{Rpo_CreateEdge(LIT_AND, 2, 3, degree), Rpo_CreateEdge(LIT_AND, 0, 1, degree)}
	if index := Rpo_computeMinEdgeCost(edges, 2, degree); index != 1 {
		t.Fatal(index)
	}
	edges[1].Visited = true
	if index := Rpo_computeMinEdgeCost(edges, 2, degree); index != 0 {
		t.Fatal(index)
	}
	target := []uint32{0x88888888}
	if Rpo_Factorize(target, 2, 1, 0) != nil {
		t.Fatal("threshold not enforced")
	}
	if Rpo_Factorize(target, 2, 0, 0) == nil {
		t.Fatal("zero means unlimited")
	}
}
func TestRecursionRestoresClusters(t *testing.T) {
	target := []uint32{0x88888888}
	a, b := Lit_Alloc(target, 2, 0, '+'), Lit_Alloc(target, 2, 1, '+')
	literals := []*Literal_t{a, b}
	var count int32
	if result := Rpo_Recursion(target, literals, 2, 2, 2, &count, -1, 0); result == nil {
		t.Fatal("not found")
	}
	if literals[0] != a || literals[1] != b {
		t.Fatal("clusters not restored")
	}
}
func TestFactorizeMultiword(t *testing.T) {
	target := tableFromFunction(6, func(i int) bool { return i&1 != 0 && i&32 != 0 })
	result := Rpo_Factorize(target, 6, 10000, 0)
	if result == nil {
		t.Fatal("missing multiword factorization")
	}
	for i := 0; i < 64; i++ {
		if evalExpression(t, result.Expression, i) != (i&1 != 0 && i&32 != 0) {
			t.Fatal(fmt.Sprint(result.Expression, i))
		}
	}
}

func TestTransitionSupportsAliasedOutput(t *testing.T) {
	positive := []uint32{0x88888888}
	Lit_TruthPositiveTransition(positive, positive, 2, 0)
	if positive[0] != 0xcccccccc {
		t.Fatalf("alias changed positive transition: %08x", positive[0])
	}
	negative := []uint32{0x77777777}
	Lit_TruthNegativeTransition(negative, negative, 2, 0)
	if negative[0] != 0xcccccccc {
		t.Fatalf("alias changed negative transition: %08x", negative[0])
	}
}
func TestConstantLiteralTransitionAndOwnership(t *testing.T) {
	input := []uint32{^uint32(0)}
	literal := Lit_CreateLiteralConst(input, 2, 1)
	if literal.Transition[0] != 0 {
		t.Fatal("constant has no transition", literal.Transition)
	}
	literal.Function[0] = 0
	if input[0] != ^uint32(0) {
		t.Fatal("constant literal aliases caller's table")
	}
}

func TestTransitionsWordParallelAgainstAssignments(t *testing.T) {
	for vars := int32(1); vars <= 8; vars++ {
		input := tableFromFunction(vars, func(i int) bool { return i%7 == 1 || i%5 == 0 })
		// For fewer than five variables, use the normal repeated KIT layout.
		if vars < 5 {
			input = tableFromFunction(vars, func(i int) bool { i &= (1 << uint(vars)) - 1; return i%7 == 1 || i%5 == 0 })
		}
		for variable := int32(0); variable < vars; variable++ {
			for _, positive := range []bool{false, true} {
				got := make([]uint32, len(input))
				litTransition(input, got, vars, variable, positive)
				mask := 1 << uint(variable)
				for assignment := 0; assignment < len(input)*32; assignment++ {
					zero, one := rpoTruthBit(input, assignment&^mask), rpoTruthBit(input, assignment|mask)
					want := zero && !one
					if positive {
						want = !zero && one
					}
					if rpoTruthBit(got, assignment) != want {
						t.Fatalf("vars=%d variable=%d positive=%v assignment=%d", vars, variable, positive, assignment)
					}
				}
			}
		}
	}
}

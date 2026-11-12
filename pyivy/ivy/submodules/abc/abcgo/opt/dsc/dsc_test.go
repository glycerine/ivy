package dsc

import "testing"

func dscTruth(n int, fn func(int) bool) []uint64 {
	words := int(uint64(1) << uint(max(n-6, 0)))
	truth := make([]uint64, words)
	for assignment := 0; assignment < words*64; assignment++ {
		if fn(assignment) {
			truth[assignment>>6] |= uint64(1) << uint(assignment&63)
		}
	}
	return truth
}

func TestDscDecomposeBasicFunctions(t *testing.T) {
	tests := []struct {
		name string
		n    int32
		fn   func(int) bool
		want string
	}{
		{"variable", 1, func(x int) bool { return x&1 != 0 }, "a"},
		{"negated variable", 1, func(x int) bool { return x&1 == 0 }, "!a"},
		{"and", 2, func(x int) bool { return x&3 == 3 }, "(ba)"},
		{"xor", 2, func(x int) bool { return (x & 1) != ((x >> 1) & 1) }, "[ba]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			truth := dscTruth(int(test.n), test.fn)
			got, status := Dsc_Decompose(truth, test.n, nil)
			if status != 0 || got != test.want {
				t.Fatalf("Dsc_Decompose() = (%q, %d), want (%q, 0)", got, status, test.want)
			}
			for assignment := 0; assignment < len(truth)*64; assignment++ {
				if gotTruth := evalDsc(got, assignment); gotTruth != test.fn(assignment) {
					t.Fatalf("expression %q differs from input at assignment %d", got, assignment)
				}
			}
			if count := Dsc_CountAnds(got); count < 0 {
				t.Fatalf("Dsc_CountAnds(%q) = %d", got, count)
			}
		})
	}
}

func evalDsc(expression string, assignment int) bool {
	position := 0
	value := evalDscNode(expression, &position, assignment)
	if position != len(expression) {
		panic("test expression has trailing text")
	}
	return value
}

func evalDscNode(expression string, position *int, assignment int) bool {
	negated := false
	for *position < len(expression) && expression[*position] == '!' {
		negated = !negated
		*position++
	}
	if *position >= len(expression) {
		panic("test expression ended before operand")
	}
	ch := expression[*position]
	*position++
	var value bool
	switch {
	case ch >= 'a' && ch <= 'z':
		value = assignment&(1<<uint(ch-'a')) != 0
	case ch == '0':
		value = false
	case ch == '1':
		value = true
	case ch == '(' || ch == '[':
		isAnd := ch == '('
		value = isAnd
		children := 0
		for *position < len(expression) && expression[*position] != ')' && expression[*position] != ']' {
			child := evalDscNode(expression, position, assignment)
			if isAnd {
				value = value && child
			} else {
				value = value != child
			}
			children++
		}
		if *position >= len(expression) {
			panic("test expression has unclosed operator")
		}
		*position++
		if children < 2 {
			panic("test expression has too few children")
		}
	default:
		panic("test expression has unsupported operator")
	}
	return value != negated
}

func TestDscConstantsAndPoolReuse(t *testing.T) {
	pool := Dsc_AllocPool(3)
	for _, test := range []struct {
		value uint64
		want  string
	}{{0, "0"}, {^uint64(0), "1"}} {
		got, status := Dsc_Decompose([]uint64{test.value}, 0, pool)
		if status != 0 || got != test.want {
			t.Fatalf("constant %016x decomposed as (%q, %d), want (%q, 0)", test.value, got, status, test.want)
		}
	}
	truth := dscTruth(3, func(x int) bool { return x&1 != 0 })
	for iteration := 0; iteration < 4; iteration++ {
		got, status := Dsc_Decompose(truth, 3, pool)
		if status != 0 || got != "a" {
			t.Fatalf("pool reuse %d: got (%q, %d), want (a, 0)", iteration, got, status)
		}
	}
	Dsc_FreePool(pool)
}

func TestDscDecompositionTruthTableAcrossAllThreeVariableFunctions(t *testing.T) {
	for mask := 0; mask < 1<<8; mask++ {
		fn := func(assignment int) bool { return mask&(1<<uint(assignment&7)) != 0 }
		truth := dscTruth(3, fn)
		expression, status := Dsc_Decompose(truth, 3, nil)
		if status != 0 {
			continue // This bottom-up algorithm intentionally rejects prime functions.
		}
		for assignment := 0; assignment < 1<<3; assignment++ {
			if got := evalDsc(expression, assignment); got != fn(assignment) {
				t.Fatalf("truth mask %#02x produced %q, wrong at assignment %d", mask, expression, assignment)
			}
		}
	}
}

func TestDscCountAnds(t *testing.T) {
	for _, test := range []struct {
		expression string
		want       int32
	}{
		{"a", 0},
		{"(ab)", 1},
		{"(a(bc))", 2},
		{"[ab]", 3},
		{"<abc>", 3},
		{"!(a[b!c])", 4},
	} {
		if got := Dsc_CountAnds(test.expression); got != test.want {
			t.Errorf("Dsc_CountAnds(%q) = %d, want %d", test.expression, got, test.want)
		}
	}
}

func TestDscComputeMatchesAndMalformedExpressions(t *testing.T) {
	expression := "(a[bc])"
	matches := Dsc_ComputeMatches(expression)
	if matches[0] != len(expression)-1 || matches[2] != 5 {
		t.Fatalf("unexpected matches for %q: %v", expression, matches)
	}
	for _, malformed := range []string{"(a]", "(ab", "a)", "()"} {
		t.Run(malformed, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered == nil {
					t.Fatalf("expected malformed expression %q to panic", malformed)
				}
			}()
			Dsc_CountAnds(malformed)
		})
	}
}

func TestDscExpressionHelpersAreConcurrent(t *testing.T) {
	const workers, iterations = 8, 100
	done := make(chan struct{}, workers)
	for worker := 0; worker < workers; worker++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for i := 0; i < iterations; i++ {
				if got := Dsc_CountAnds("(a[b(c!d)])"); got != 5 {
					t.Errorf("Dsc_CountAnds concurrent = %d, want 5", got)
					return
				}
			}
		}()
	}
	for worker := 0; worker < workers; worker++ {
		<-done
	}
}

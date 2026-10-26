package abc

// Go-only unit and fuzz checks for opt_dau_dsd.go.

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestDauDSDMatches(t *testing.T) {
	got, err := DauDSDMatches("!([ab]<abc>{def})", 0)
	if err != nil {
		t.Fatal(err)
	}
	for left, right := range map[int]int{1: 16, 2: 5, 6: 10, 11: 15} {
		if got[left] != right || got[right] != left {
			t.Fatalf("match %d/%d: %v", left, right, got)
		}
	}
	for _, input := range []string{"(", "]", "([)]", "<ab}"} {
		if _, err := DauDSDMatches(input, 0); err == nil {
			t.Fatalf("DauDSDMatches(%q) succeeded", input)
		}
	}
	if _, err := DauDSDMatches("(ab)", 3); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit: %v", err)
	}
	if _, err := DauDSDMatches("a", -1); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative limit: %v", err)
	}
}

func TestDauDSDTruthAndCosts(t *testing.T) {
	tests := []struct {
		formula    string
		vars, cost int
		want       func(uint64) bool
	}{
		{"0", 0, 0, func(uint64) bool { return false }},
		{"1", 0, 0, func(uint64) bool { return true }},
		{"a", 1, 0, func(a uint64) bool { return a&1 != 0 }},
		{"!a", 1, 0, func(a uint64) bool { return a&1 == 0 }},
		{"(ab)", 2, 1, func(a uint64) bool { return a&3 == 3 }},
		{"[ab]", 2, 3, func(a uint64) bool { return (a&1 != 0) != (a&2 != 0) }},
		{"<abc>", 3, 3, func(a uint64) bool {
			if a&1 != 0 {
				return a&2 != 0
			}
			return a&4 != 0
		}},
		{"8{ab}", 2, 3, func(a uint64) bool { return a&3 == 3 }},
		{"E{ab}", 2, 3, func(a uint64) bool { return a&3 != 0 }},
		{"80{abc}", 3, 3, func(a uint64) bool { return a&7 == 7 }},
		{"(a[bc])", 3, 4, func(a uint64) bool { return a&1 != 0 && ((a&2 != 0) != (a&4 != 0)) }},
		{"<abc>{d!e[fg]}", 7, 6, func(a uint64) bool {
			if a&8 != 0 {
				return a&16 == 0
			}
			return (a&32 != 0) != (a&64 != 0)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.formula, func(t *testing.T) {
			f, err := ParseDauDSD(tc.formula, DauDSDOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if f.VariableCount() != tc.vars {
				t.Fatalf("variables=%d want %d", f.VariableCount(), tc.vars)
			}
			if cost, err := f.AndCount(); err != nil || cost != tc.cost {
				t.Fatalf("cost=%d, %v; want %d", cost, err, tc.cost)
			}
			truth, err := f.TruthTable(tc.vars, 0, nil)
			if err != nil {
				t.Fatal(err)
			}
			for a := uint64(0); a < uint64(1)<<uint(tc.vars); a++ {
				got, err := truth.Get(a)
				if err != nil {
					t.Fatal(err)
				}
				if got != tc.want(a) {
					t.Fatalf("assignment %b: got %t", a, got)
				}
			}
			copy, err := ParseDauDSD(f.String(), DauDSDOptions{})
			if err != nil {
				t.Fatalf("reparse %q: %v", f.String(), err)
			}
			other, err := copy.TruthTable(tc.vars, 0, nil)
			if err != nil || !truth.Equal(other) {
				t.Fatalf("round trip: %v", err)
			}
		})
	}
}

func TestDauDSDNormalizeAndReject(t *testing.T) {
	f, err := ParseDauDSD("[c(ab)]", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Normalized(); got != "[(ab)c]" {
		t.Fatalf("normalized=%q", got)
	}
	again, err := ParseDauDSD(f.Normalized(), DauDSDOptions{})
	if err != nil || again.Normalized() != f.Normalized() {
		t.Fatalf("normalization not idempotent: %q, %v", again.Normalized(), err)
	}

	bad := []string{"", "!", "!!a", "(a)", "[]", "<ab>", "<abcd>", "<abc>{}", "<abc>{ab}", "8{a}", "8{abc}", "ABC{ab}", "a!", "g{a}"}
	for _, input := range bad {
		if _, err := ParseDauDSD(input, DauDSDOptions{}); err == nil {
			t.Errorf("ParseDauDSD(%q) succeeded", input)
		}
	}
	limits := []DauDSDOptions{{MaxBytes: 1}, {MaxNodes: 1}, {MaxDepth: 1}, {MaxVariables: 1}, {Stop: func() bool { return true }}}
	for _, o := range limits {
		if _, err := ParseDauDSD("(ab)", o); !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Errorf("options %+v: %v", o, err)
		}
	}
	if _, err := ParseDauDSD("a", DauDSDOptions{MaxBytes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative option: %v", err)
	}
	if _, err := f.TruthTable(2, 0, nil); !errors.Is(err, ErrDimension) {
		t.Fatalf("small domain: %v", err)
	}
	if truth, err := f.TruthTable(3, 0, func() bool { return true }); truth != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v, %v", truth, err)
	}
	if _, err := f.TruthTable(30, 1, nil); !errors.Is(err, ErrLimit) {
		t.Fatalf("word limit: %v", err)
	}
}

func fuzzDauFormula(data []byte) string {
	vars := func(b byte) byte { return 'a' + b%4 }
	if len(data) == 0 {
		return "a"
	}
	parts := make([]string, 0, 4)
	for i, b := range data {
		if i == 4 {
			break
		}
		v := string(vars(b))
		if b&4 != 0 {
			v = "!" + v
		}
		parts = append(parts, v)
	}
	switch data[0] % 5 {
	case 0:
		return "(" + strings.Join(parts, "") + ")"
	case 1:
		return "[" + strings.Join(parts, "") + "]"
	case 2:
		for len(parts) < 3 {
			parts = append(parts, string(byte('a'+len(parts))))
		}
		return "<" + strings.Join(parts[:3], "") + ">"
	case 3:
		for len(parts) < 2 {
			parts = append(parts, string(byte('a'+len(parts))))
		}
		return fmt.Sprintf("%X{%s%s}", data[0]&15, parts[0], parts[1])
	default:
		return parts[0]
	}
}

func FuzzDauDSDParser(f *testing.F) {
	for _, seed := range []string{"a", "(ab)", "<abc>{d!e[fg]}", "80{abc}", "([)]", "!"} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		text := string(data)
		parsed, err := ParseDauDSD(text, DauDSDOptions{MaxBytes: 256, MaxNodes: 128, MaxDepth: 32, MaxVariables: 12})
		if err != nil {
			return
		}
		reparsed, err := ParseDauDSD(parsed.String(), DauDSDOptions{})
		if err != nil {
			t.Fatalf("render %q: %v", parsed.String(), err)
		}
		if reparsed.String() != parsed.String() || reparsed.Normalized() != parsed.Normalized() {
			t.Fatalf("unstable render: %q / %q", parsed.String(), reparsed.String())
		}
	})
}

func FuzzDauDSDTruthRoundTrip(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2}, {2, 3, 4}, {3, 7, 9, 11}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		formula := fuzzDauFormula(data)
		parsed, err := ParseDauDSD(formula, DauDSDOptions{})
		if err != nil {
			return
		}
		normal, err := ParseDauDSD(parsed.Normalized(), DauDSDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		variables := parsed.VariableCount()
		left, err := parsed.TruthTable(variables, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		right, err := normal.TruthTable(variables, 0, nil)
		if err != nil || !left.Equal(right) {
			t.Fatalf("%q => %q differs: %v", formula, parsed.Normalized(), err)
		}
	})
}

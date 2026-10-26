package abc

// Go-only semantic and fuzz checks for DSD variable permutation.

import (
	"errors"
	"slices"
	"testing"
)

func TestDauDSDPermute(t *testing.T) {
	if count, err := DauDSDVariableCount("1", 0); err != nil || count != 1 {
		t.Fatalf("constant count=%d, %v", count, err)
	}
	if count, err := DauDSDVariableCount("(ac)", 0); err != nil || count != 3 {
		t.Fatalf("sparse count=%d, %v", count, err)
	}
	got, err := DauDSDPermute("!(a[bc])", []int{2, 0, 1}, DauDSDOptions{})
	if err != nil || got != "!(c[ab])" {
		t.Fatalf("permuted=%q, %v", got, err)
	}
	for _, permutation := range [][]int{nil, {0, 1}, {0, 0, 2}, {0, 1, 3}, {-1, 1, 2}} {
		if _, err := DauDSDPermute("(abc)", permutation, DauDSDOptions{}); !errors.Is(err, ErrDimension) {
			t.Errorf("permutation %v: %v", permutation, err)
		}
	}
	if _, err := DauDSDPermute("(abc)", []int{0, 1, 2}, DauDSDOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := DauDSDVariableCount("a", -1); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative: %v", err)
	}
}

func TestDauDSDRandomPermutation(t *testing.T) {
	first, permutation, err := DauDSDRandomPermutation("(abcdef)", 123, DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, again, err := DauDSDRandomPermutation("(abcdef)", 123, DauDSDOptions{})
	if err != nil || first != second || !slices.Equal(permutation, again) {
		t.Fatalf("not deterministic: %q %v / %q %v, %v", first, permutation, second, again, err)
	}
	permutation[0] = 99
	_, fresh, err := DauDSDRandomPermutation("(abcdef)", 123, DauDSDOptions{})
	if err != nil || fresh[0] == 99 {
		t.Fatalf("aliased result: %v, %v", fresh, err)
	}
}

func checkDauDSDPermutationSemantics(t *testing.T, formula string, permutation []int) {
	t.Helper()
	original, err := ParseDauDSD(formula, DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	text, err := DauDSDPermute(formula, permutation, DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	permuted, err := ParseDauDSD(text, DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	left, err := original.TruthTable(len(permutation), 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	right, err := permuted.TruthTable(len(permutation), 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(len(permutation)); assignment++ {
		var source uint64
		for old, target := range permutation {
			if assignment&(uint64(1)<<uint(target)) != 0 {
				source |= uint64(1) << uint(old)
			}
		}
		want, _ := left.Get(source)
		got, _ := right.Get(assignment)
		if got != want {
			t.Fatalf("%q -> %q permutation %v assignment %b", formula, text, permutation, assignment)
		}
	}
}

func FuzzDauDSDPermute(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 123, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		formula := "!([ab](c<dab>))"
		_, permutation, err := DauDSDRandomPermutation(formula, seed, DauDSDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkDauDSDPermutationSemantics(t, formula, permutation)
	})
}

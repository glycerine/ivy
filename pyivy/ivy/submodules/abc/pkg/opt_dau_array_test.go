package abc

// Go-only structure, ownership and fuzz checks for opt_dau_array.go.

import (
	"errors"
	"testing"
)

func TestDauDSDArrayConstructFormatAndCount(t *testing.T) {
	a, err := DauDSDArrayFromFormula("[(ab)[cf]]", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := a.Format(); err != nil || got != "[a0(a1ab)[c1cf]]\n" {
		t.Fatalf("format=%q, %v", got, err)
	}
	if cost, err := a.AndCount(); err != nil || cost != 7 {
		t.Fatalf("cost=%d, %v", cost, err)
	}
	nodes := a.Nodes()
	nodes[0].Fanins = 99
	if err := a.Check(); err != nil {
		t.Fatalf("accessor aliased storage: %v", err)
	}

	flat, err := a.WithoutUselessAssociativeNodes()
	if err != nil {
		t.Fatal(err)
	}
	if got, err := flat.Format(); err != nil || got != "[a0(a1ab)cf]\n" {
		t.Fatalf("flat=%q, %v", got, err)
	}
	if cost, err := flat.AndCount(); err != nil || cost != 4 {
		t.Fatalf("flat cost=%d, %v", cost, err)
	}
	if got, _ := a.Format(); got != "[a0(a1ab)[c1cf]]\n" {
		t.Fatalf("receiver mutated: %q", got)
	}
}

func TestDauDSDArrayConstantsDepthsAndLimits(t *testing.T) {
	for formula, want := range map[string]string{"0": "0\n", "1": "1\n"} {
		a, err := DauDSDArrayFromFormula(formula, DauDSDOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got, err := a.Format(); err != nil || got != want {
			t.Fatalf("%s => %q, %v", formula, got, err)
		}
	}
	a, err := DauDSDArrayFromFormula("(a!(bc))", DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	depths, err := a.WithDepths()
	if err != nil {
		t.Fatal(err)
	}
	if got, err := depths.Format(); err != nil || got != "(a0a!(b1bc))\n" {
		t.Fatalf("depths=%q, %v", got, err)
	}
	flat, err := a.WithoutUselessAssociativeNodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(flat.Nodes()) != len(a.Nodes()) {
		t.Fatal("complemented AND was flattened")
	}
	if _, err := DauDSDArrayFromFormula("<abc>{def}", DauDSDOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("local bindings: %v", err)
	}
	if err := (*DauDSDArray)(nil).Check(); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil: %v", err)
	}
	malformed := []*DauDSDArray{
		{},
		{nodes: []DauDSDArrayNode{{Kind: DauDSDAnd, Fanins: 2}}},
		{nodes: []DauDSDArrayNode{{Kind: DauDSDVariable, Variable: 26}}},
		{nodes: []DauDSDArrayNode{{Kind: DauDSDVariable, Depth: 1}}},
		{nodes: []DauDSDArrayNode{{Kind: DauDSDNodeKind(99)}}},
	}
	for i, array := range malformed {
		if err := array.Check(); err == nil {
			t.Errorf("malformed %d accepted", i)
		}
	}
}

func FuzzDauDSDArray(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2}, {2, 3, 4}, {4, 5, 6, 7}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		formula := fuzzDauFormula(data)
		a, err := DauDSDArrayFromFormula(formula, DauDSDOptions{})
		if err != nil {
			return
		}
		before, err := a.Format()
		if err != nil {
			t.Fatal(err)
		}
		depths, err := a.WithDepths()
		if err != nil {
			t.Fatal(err)
		}
		if after, err := depths.Format(); err != nil || after != before {
			t.Fatalf("depth instability: %q / %q, %v", before, after, err)
		}
		flat, err := a.WithoutUselessAssociativeNodes()
		if err != nil || flat.Check() != nil {
			t.Fatalf("flatten: %v / %v", err, flat.Check())
		}
	})
}

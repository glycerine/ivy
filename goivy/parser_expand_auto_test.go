package goivy

import "testing"

func TestGetTypeNamesFromConstantLikeDecls(t *testing.T) {
	cfg := NewAstConfig()
	newTypedAtom := func(name string) *Atom {
		a := cfg.NewAtom(name)
		a.ASort = cfg.NewAtom("arr")
		return a
	}

	cases := []struct {
		name string
		decl Node
	}{
		{name: "constant", decl: cfg.NewConstantDecl(newTypedAtom("c"))},
		{name: "fresh", decl: cfg.NewFreshConstantDecl(*cfg.NewConstantDecl(newTypedAtom("f")))},
		{name: "destructor", decl: cfg.NewDestructorDecl(newTypedAtom("d"))},
		{name: "constructor", decl: cfg.NewConstructorDecl(newTypedAtom("k"))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			names := newTypeNames()
			getTypeNamesFromDecl(tc.decl, names)
			if !names.nameset["arr"] {
				t.Fatalf("getTypeNamesFromDecl(%s) did not collect arr; got %v", tc.name, names.namelist)
			}
		})
	}
}

package cudd

import (
	"bytes"
	"testing"

	core "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestCuddZddSymmProfileUsesManagerWriterAndRange(t *testing.T) {
	dd := core.NewBDDManager(2)
	var out bytes.Buffer
	core.Cudd_SetOutput(dd, &out, nil)
	Cudd_zddSymmProfile(dd, 0, 2)
	if got, want := out.String(), "Total Symmetric = 0\nTotal Groups = 0\n"; got != want {
		t.Fatalf("profile output %q, want %q", got, want)
	}
}

func TestCuddZddSymmetryWrapperBounds(t *testing.T) {
	if cuddZddSymmCheck(nil, 0, 1) != 0 {
		t.Fatal("nil manager reported symmetry")
	}
	dd := core.NewBDDManager(2)
	if cuddZddSymmCheck(dd, 0, 1) != 0 {
		t.Fatal("uninitialized ZDD levels reported symmetry")
	}
	if cuddZddSymmSifting(nil, 0, 1) != 0 || cuddZddSymmSiftingConv(nil, 0, 1) != 0 {
		t.Fatal("nil manager sifting should fail")
	}
}

func TestCuddZddSymmSiftingAndProfileCircularGroup(t *testing.T) {
	dd := core.NewBDDManager(2)
	if core.Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	if got := cuddZddSymmSifting(dd, 0, 1); got != 3 {
		t.Fatalf("symmetric sifting result = %d, want 3", got)
	}
	var out bytes.Buffer
	core.Cudd_SetOutput(dd, &out, nil)
	Cudd_zddSymmProfile(dd, 0, 2)
	if got, want := out.String(), "Group:  0  1\nTotal Symmetric = 2\nTotal Groups = 1\n"; got != want {
		t.Fatalf("full profile output %q, want %q", got, want)
	}
	out.Reset()
	Cudd_zddSymmProfile(dd, 1, 2)
	if got, want := out.String(), "Group:  1  0\nTotal Symmetric = 2\nTotal Groups = 1\n"; got != want {
		t.Fatalf("partial circular profile output %q, want %q", got, want)
	}
	out.Reset()
	Cudd_zddSymmProfile(dd, 2, 2)
	if got, want := out.String(), "Total Symmetric = 0\nTotal Groups = 0\n"; got != want {
		t.Fatalf("empty-range profile output %q, want %q", got, want)
	}
}

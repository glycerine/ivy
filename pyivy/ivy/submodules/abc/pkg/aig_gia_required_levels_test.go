package abc

import "testing"

func TestGiaRequiredLevelsConstantOnlyAndLimits(t *testing.T) {
	g := NewGiaManager()
	buffer, err := g.AppendBuffer(GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	root, err := g.AppendAnd(buffer, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	reverse, err := g.WeightedReverseLevels(0)
	if err != nil || reverse[0] != 1 || reverse[buffer.ID()] != 1 {
		t.Fatal(reverse, err)
	}
	required, err := g.RequiredLevels(0)
	if err != nil || required[0] != -1 || required[buffer.ID()] != -1 || required[root.ID()] != 0 {
		t.Fatal("signed required levels", required, err)
	}
	slacks, err := g.LevelSlacks(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range slacks {
		if value != 0 {
			t.Fatal(slacks)
		}
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.WeightedReverseLevels(cap); err == nil {
			t.Fatal("reverse limit")
		}
		if _, err := g.RequiredLevels(cap); err == nil {
			t.Fatal("required limit")
		}
		if _, err := g.LevelSlacks(cap); err == nil {
			t.Fatal("slack limit")
		}
	}
}

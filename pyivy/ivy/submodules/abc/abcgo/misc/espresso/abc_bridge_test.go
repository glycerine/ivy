package espresso

import "testing"

func TestAbcMinimizePreservesSimpleCube(t *testing.T) {
	ConfigureCube(2, nil)
	on := NewFamily(1, 4)
	row := NewSet(4)
	row.Insert(0) // x = 0
	row.Insert(1) // x = 1
	row.Insert(3) // y = 1
	on.Add(row)
	got := AbcMinimize(on, nil)
	if got == nil || got.Count() != 1 || got.SetSize() != 4 {
		t.Fatalf("minimized cover = %#v", got)
	}
	if !got.Row(0).Contains(3) || got.Row(0).Contains(2) {
		t.Fatalf("minimized cube does not represent y=1: %v", got.Row(0))
	}
}

package mvc

import "testing"

// TestCoverSortBreaksTiesLikeAbc: Mvc_CoverSortMerge takes the cube of the
// second list unless the first compares strictly smaller, so equal cubes
// come out reversed (not in stable order).
func TestCoverSortBreaksTiesLikeAbc(t *testing.T) {
	c := Mvc_CoverAlloc(Mvc_ManagerStart(), 4)
	var cubes []*Mvc_Cube_t
	for i := 0; i < 3; i++ {
		x := Mvc_CubeAlloc(c)
		Mvc_CoverAddCubeTail(c, x)
		cubes = append(cubes, x)
	}
	Mvc_CoverSort(c, nil, func(a, b, _ *Mvc_Cube_t) int { return 0 })
	want := []*Mvc_Cube_t{cubes[2], cubes[1], cubes[0]}
	i := 0
	for x := c.Cubes.Head; x != nil; x = x.Next {
		if x != want[i] {
			t.Fatalf("position %d holds cube %p, want %p (ABC's merge order)", i, x, want[i])
		}
		i++
	}
	if c.Cubes.Tail != cubes[0] {
		t.Fatal("tail not updated")
	}
}

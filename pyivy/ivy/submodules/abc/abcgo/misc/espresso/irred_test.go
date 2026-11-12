package espresso

import (
	"math/rand"
	"testing"
)

func irredGeometry(parts ...int32) {
	cube = cube_struct{num_vars: int32(len(parts)), part_size: append([]int32(nil), parts...)}
	cube_setup()
}
func irredCube(parts ...uint32) pcube {
	p := new_cube()
	for v, mask := range parts {
		for j := int32(0); j < cube.part_size[v]; j++ {
			if mask&(1<<j) != 0 {
				set_insert(p, cube.first_part[v]+j)
			}
		}
	}
	return p
}
func irredMatches(p pcube, point []int32) bool {
	for v, x := range point {
		if !is_in_set(p, cube.first_part[v]+x) {
			return false
		}
	}
	return true
}
func irredPoints(visit func([]int32)) {
	point := make([]int32, cube.num_vars)
	var walk func(int)
	walk = func(v int) {
		if v == len(point) {
			visit(point)
			return
		}
		for x := int32(0); x < cube.part_size[v]; x++ {
			point[v] = x
			walk(v + 1)
		}
	}
	walk(0)
}
func irredCoverMatches(f pcover, point []int32) bool {
	if f != nil {
		for i := int32(0); i < f.count; i++ {
			if irredMatches(GETSET(f, i), point) {
				return true
			}
		}
	}
	return false
}
func TestIrredCoverageAgainstMinterms(t *testing.T) {
	irredGeometry(2, 3, 2)
	r := rand.New(rand.NewSource(744))
	for iteration := 0; iteration < 150; iteration++ {
		f := new_cover(12)
		for n := 0; n < r.Intn(10); n++ {
			sf_addset(f, irredCube(uint32(r.Intn(3)+1), uint32(r.Intn(7)+1), uint32(r.Intn(3)+1)))
		}
		all := cube1list(f)
		expected := true
		irredPoints(func(point []int32) {
			if !irredCoverMatches(f, point) {
				expected = false
			}
		})
		if got := tautology(cube1list(f)); got != expected {
			t.Fatalf("iteration%d tautology=%v expected%v", iteration, got, expected)
		}
		query := irredCube(uint32(r.Intn(3)+1), uint32(r.Intn(7)+1), uint32(r.Intn(3)+1))
		covered := true
		irredPoints(func(point []int32) {
			if irredMatches(query, point) && !irredCoverMatches(f, point) {
				covered = false
			}
		})
		if cube_is_covered(all, query) != covered {
			t.Fatalf("iteration%d cube coverage", iteration)
		}
	}
}
func TestIrredundantPreservesCoverAndRows(t *testing.T) {
	irredGeometry(2, 3)
	r := rand.New(rand.NewSource(99))
	for iteration := 0; iteration < 100; iteration++ {
		f, d := new_cover(12), new_cover(3)
		for n := 0; n < 10; n++ {
			sf_addset(f, irredCube(uint32(r.Intn(3)+1), uint32(r.Intn(7)+1)))
		}
		for n := 0; n < 2; n++ {
			sf_addset(d, irredCube(uint32(r.Intn(3)+1), uint32(r.Intn(7)+1)))
		}
		before := sf_save(f)
		mark_irredundant(f, d)
		if f.count != before.count {
			t.Fatal("marking changed row count")
		}
		for i := int32(0); i < f.count; i++ {
			if !setp_equal(GETSET(f, i), GETSET(before, i)) {
				t.Fatal("marking changed row bits")
			}
		}
		result := sf_inactive(f)
		irredPoints(func(point []int32) {
			if !irredCoverMatches(d, point) && irredCoverMatches(before, point) != irredCoverMatches(result, point) {
				t.Fatalf("iteration%d lost coverage at%v", iteration, point)
			}
		})
	}
}
func TestCofactorBorrowingAndSingletonSimplification(t *testing.T) {
	irredGeometry(2, 2)
	f := new_cover(3)
	sf_addset(f, irredCube(1, 3))
	sf_addset(f, irredCube(2, 3))
	list := cube1list(f)
	child := cofactor(list, GETSET(f, 0))
	if len(child.cubes) != 0 {
		t.Fatal("cofactor retained self or disjoint row")
	}
	single := &cubeList{cofactor: new_cube(), cubes: []pcube{GETSET(f, 0), GETSET(f, 0)}}
	simplify_cubelist(single)
	if len(single.cubes) != 1 || !sameCube(single.cubes[0], GETSET(f, 0)) {
		t.Fatal("simplification lost singleton")
	}
}

func TestIrredClassificationAndEssentialFlags(t *testing.T) {
	irredGeometry(2, 2)
	f, d := new_cover(4), new_cover(0)
	// Two duplicate x=0 cubes compete; x=1,y=0 has a private minterm.
	sf_addset(f, irredCube(1, 3))
	sf_addset(f, irredCube(1, 3))
	sf_addset(f, irredCube(2, 1))
	var e, rt, rp pcover
	irred_split_cover(f, d, &e, &rt, &rp)
	if e.count != 1 || rt.count != 0 || rp.count != 2 {
		t.Fatalf("classification E=%d Rt=%d Rp=%d", e.count, rt.count, rp.count)
	}
	mark_irredundant(f, d)
	if !TESTP(GETSET(f, 2), ACTIVE) || !TESTP(GETSET(f, 2), RELESSEN) {
		t.Fatal("relative essential not marked")
	}
	chosen := 0
	for i := int32(0); i < 2; i++ {
		if TESTP(GETSET(f, i), ACTIVE) {
			chosen++
		}
	}
	if chosen != 1 {
		t.Fatalf("selected %d duplicate rows", chosen)
	}
}
func TestCofactorSplitAndCountMetadata(t *testing.T) {
	irredGeometry(3, 2)
	f := new_cover(2)
	sf_addset(f, irredCube(1, 3))
	sf_addset(f, irredCube(6, 1))
	list := cube1list(f)
	massive_count(list)
	if cdata.part_zeros[0] != 1 || cdata.part_zeros[1] != 1 || cdata.part_zeros[2] != 1 || cdata.parts_active[0] != 3 {
		t.Fatal("multiple-valued column counts")
	}
	l, r := new_cube(), new_cube()
	v := binate_split_select(list, l, r, 0)
	if v != 0 {
		t.Fatalf("split variable %d", v)
	}
	union := set_or(new_cube(), l, r)
	if !setp_equal(union, cube.fullset) {
		t.Fatal("split does not partition domain")
	}
	if cubeVarIntersects(l, r, v) {
		t.Fatal("split branches overlap")
	}
	a, b := scofactor(list, l, v), scofactor(list, r, v)
	free_cubelist(a)
	free_cubelist(b)
	if len(list.cubes) != 2 || !setp_empty(list.cofactor) {
		t.Fatal("child disposal mutated parent")
	}
}

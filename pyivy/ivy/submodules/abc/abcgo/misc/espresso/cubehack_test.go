package espresso

import "testing"

func TestCubeHackGeometryReuse(t *testing.T) {
	cube = cube_struct{}
	cubeHackFullSet = nil
	for _, n := range []int32{0, 1, 16, 17, 100, 3, 0, 101, 150, 2, 99} {
		define_cube_size(n)
		if cube.num_vars != n || cube.num_binary_vars != n || cube.size != 2*n || cube.output != -1 || cube.num_mv_vars != 0 {
			t.Fatalf("incorrect geometry for%d", n)
		}
		if set_ord(cube.fullset) != 2*n || !setp_equal(cube.fullset, cube.binary_mask) {
			t.Fatalf("full-set width%d", n)
		}
		for i := int32(0); i < n; i++ {
			if cube.first_part[i] != 2*i || cube.last_part[i] != 2*i+1 || LOOP(cube.var_mask[i]) != LOOP(cube.fullset) || set_ord(cube.var_mask[i]) != 2 {
				t.Fatalf("variable mask%d/%d", i, n)
			}
		}
		if n <= 100 && len(cube.var_mask) < 100 {
			t.Fatal("small geometry did not retain reusable storage")
		}
	}
	first := cube.fullset
	define_cube_size(2)
	if !sameCube(first, cube.fullset) {
		t.Fatal("small geometry unnecessarily reallocated")
	}
	undefine_cube_size()
	if cube.fullset != nil || cube.part_size != nil {
		t.Fatal("geometry teardown retained buffers")
	}
	define_cube_size(3)
	if cube.size != 6 {
		t.Fatal("geometry restart")
	}
}
func TestCautiousCubeSizeRejectsMixedGeometry(t *testing.T) {
	cube = cube_struct{num_vars: 3, num_binary_vars: 2, part_size: []int32{2, 2, 4}}
	cube_setup()
	cautious_define_cube_size(2)
	if cube.num_vars != 2 || cube.size != 4 || cube.output != -1 {
		t.Fatal("mixed geometry mistaken for matching binary geometry")
	}
}
func TestCubeHackCacheInvalidation(t *testing.T) {
	define_cube_size(4)
	// An independent setup invalidates the cache even if counts happen to match.
	cube = cube_struct{num_vars: 3, part_size: []int32{3, 3, 3}}
	cube_setup()
	define_cube_size(2)
	if cube.size != 4 || len(cube.var_mask) != 100 || cube.first_part[1] != 2 {
		t.Fatal("stale geometry cache reused")
	}
	set_espresso_flags()
	if summary || trace || !remove_essential || !force_irredundant || !unwrap_onset || single_expand || pos || recompute_onset || use_super_gasp || use_random_order {
		t.Fatal("incorrect default espresso flags")
	}
}

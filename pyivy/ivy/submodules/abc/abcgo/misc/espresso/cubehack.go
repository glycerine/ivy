package espresso

// The small-binary-geometry cache is identified by its backing full-set buffer,
// rather than a stale "called_before" flag after an external cube_setup.
var cubeHackFullSet pcube

func cautious_define_cube_size(n int32) {
	if n < 0 || n > BPI*1023/2 {
		panic("espresso: binary variable count outside packed-set range")
	}
	if cube.fullset != nil && cube.num_binary_vars == n && cube.num_vars == n {
		return
	}
	if cube.fullset != nil {
		setdown_cube()
	}
	cube.part_size = make([]int32, n)
	cube.num_binary_vars = n
	cube.num_vars = n
	cube_setup()
	cubeHackFullSet = nil
}
func define_cube_size(n int32) {
	if n < 0 || n > BPI*1023/2 {
		panic("espresso: binary variable count outside packed-set range")
	}
	if cube.fullset != nil && cube.num_binary_vars == n && cube.num_vars == n {
		return
	}
	if n > 100 {
		cautious_define_cube_size(n)
		return
	}
	if !sameCube(cube.fullset, cubeHackFullSet) {
		cautious_define_cube_size(100)
		cubeHackFullSet = cube.fullset
	}
	cube.num_vars = n
	cube.num_binary_vars = n
	cube.num_mv_vars = 0
	cube.output = -1
	cube.size = n * 2
	set_fill(cube.fullset, cube.size)
	set_fill(cube.binary_mask, cube.size)
	header := cube.fullset[0]
	for i := int32(0); i < n; i++ {
		cube.var_mask[i][0] = header
	}
	for _, p := range cube.temp {
		p[0] = header
	}
	cube.emptyset[0] = header
	cube.mv_mask[0] = header
	if n > 0 {
		cube.inword = cube.last_word[n-1]
		cube.inmask = cube.binary_mask[cube.inword] & DISJOINT
	} else {
		cube.inword = -1
		cube.inmask = 0
	}
}
func undefine_cube_size() {
	if cube.num_binary_vars <= 100 {
		cube.num_vars = 100
		cube.num_binary_vars = 100
	}
	if cube.fullset != nil {
		setdown_cube()
		cube.part_size = nil
	}
	cubeHackFullSet = nil
}
func set_espresso_flags() {
	summary = false
	trace = false
	remove_essential = true
	force_irredundant = true
	unwrap_onset = true
	single_expand = false
	pos = false
	recompute_onset = false
	use_super_gasp = false
	use_random_order = false
}

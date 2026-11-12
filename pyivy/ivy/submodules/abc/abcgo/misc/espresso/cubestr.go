package espresso

// Cube geometry retains the source's process-local global context. Operations
// using cube/cdata must share one configuration and are not reentrant.
func cube_setup() {
	if cube.num_binary_vars < 0 || cube.num_vars < cube.num_binary_vars {
		panic("espresso: inconsistent variable counts")
	}
	if int32(len(cube.part_size)) < cube.num_vars {
		panic("espresso: missing variable part sizes")
	}
	cube.num_mv_vars = cube.num_vars - cube.num_binary_vars
	cube.output = -1
	if cube.num_mv_vars > 0 {
		cube.output = cube.num_vars - 1
	}
	cube.size = 0
	cube.first_part = make([]int32, cube.num_vars)
	cube.last_part = make([]int32, cube.num_vars)
	cube.first_word = make([]int32, cube.num_vars)
	cube.last_word = make([]int32, cube.num_vars)
	for variable := int32(0); variable < cube.num_vars; variable++ {
		if variable < cube.num_binary_vars {
			cube.part_size[variable] = 2
		}
		size := int64(cube.part_size[variable])
		if size < 0 {
			size = -size
		}
		if size == 0 || int64(cube.size)+size > int64(BPI*1023) {
			panic("espresso: variable parts exceed packed-set range")
		}
		cube.first_part[variable] = cube.size
		cube.first_word[variable] = WHICH_WORD(cube.size)
		cube.size += int32(size)
		cube.last_part[variable] = cube.size - 1
		cube.last_word[variable] = WHICH_WORD(cube.size - 1)
	}
	cube.var_mask = make([]pset, cube.num_vars)
	cube.sparse = make([]int32, cube.num_vars)
	cube.binary_mask = new_cube()
	cube.mv_mask = new_cube()
	for variable := int32(0); variable < cube.num_vars; variable++ {
		mask := new_cube()
		cube.var_mask[variable] = mask
		for i := cube.first_part[variable]; i <= cube.last_part[variable]; i++ {
			set_insert(mask, i)
		}
		if variable < cube.num_binary_vars {
			set_or(cube.binary_mask, cube.binary_mask, mask)
		} else {
			set_or(cube.mv_mask, cube.mv_mask, mask)
			cube.sparse[variable] = 1
		}
	}
	cube.inword = -1
	cube.inmask = 0
	if cube.num_binary_vars > 0 {
		cube.inword = cube.last_word[cube.num_binary_vars-1]
		cube.inmask = cube.binary_mask[cube.inword] & DISJOINT
	}
	cube.temp = make([]pset, CUBE_TEMP)
	for i := range cube.temp {
		cube.temp[i] = new_cube()
	}
	cube.fullset = set_full(cube.size)
	cube.emptyset = new_cube()
	cdata = cdata_struct{part_zeros: make([]int32, cube.size), var_zeros: make([]int32, cube.num_vars), parts_active: make([]int32, cube.num_vars), is_unate: make([]bool, cube.num_vars)}
}
func setdown_cube() {
	cube.first_part = nil
	cube.last_part = nil
	cube.first_word = nil
	cube.last_word = nil
	cube.sparse = nil
	cube.binary_mask = nil
	cube.mv_mask = nil
	cube.fullset = nil
	cube.emptyset = nil
	cube.var_mask = nil
	cube.temp = nil
	cdata.part_zeros = nil
	cdata.var_zeros = nil
	cdata.parts_active = nil
	cdata.is_unate = nil
}
func save_cube_struct() {
	temp_cube_save = cube
	temp_cdata_save = cdata
	cube.first_part = nil
	cube.last_part = nil
	cube.first_word = nil
	cube.last_word = nil
	cube.part_size = nil
	cube.binary_mask = nil
	cube.mv_mask = nil
	cube.fullset = nil
	cube.emptyset = nil
	cube.var_mask = nil
	cube.temp = nil
	cdata.part_zeros = nil
	cdata.var_zeros = nil
	cdata.parts_active = nil
	cdata.is_unate = nil
}
func restore_cube_struct() { cube = temp_cube_save; cdata = temp_cdata_save }
func ConfigureCube(binaryVariables int32, multiValuedPartSizes []int32) {
	if binaryVariables < 0 {
		panic("espresso: negative binary-variable count")
	}
	parts := make([]int32, int(binaryVariables)+len(multiValuedPartSizes))
	copy(parts[binaryVariables:], multiValuedPartSizes)
	cube = cube_struct{num_vars: int32(len(parts)), num_binary_vars: binaryVariables, part_size: parts}
	cube_setup()
}

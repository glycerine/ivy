package espresso

// Shared Espresso cube/PLA declarations. Packed-set storage and metadata live
// in set_types.go; C forward declarations are unnecessary in a Go package.
type cost_t struct{ cubes, in, out, mv, total, primes int32 }
type pcost = *cost_t
type pair_t struct {
	cnt        int32
	var1, var2 []int32
}
type ppair = *pair_t
type symbolic_list_t struct {
	variable, pos int32
	next          *symbolic_list_t
}
type symbolic_label_t struct {
	label string
	next  *symbolic_label_t
}
type symbolic_t struct {
	symbolic_list         *symbolic_list_t
	symbolic_list_length  int32
	symbolic_label        *symbolic_label_t
	symbolic_label_length int32
	next                  *symbolic_t
}
type PLA_t struct {
	F, D, R                   pcover
	filename                  string
	pla_type                  int32
	phase                     pcube
	pair                      ppair
	label                     []string
	symbolic, symbolic_output *symbolic_t
}
type pPLA = *PLA_t
type pla_types_struct struct {
	key   string
	value int32
}

const (
	FALSE                            = false
	TRUE                             = true
	MAYBE                            = 2
	F_type                           = 1
	D_type                           = 2
	R_type                           = 4
	PLEASURE_type                    = 8
	EQNTOTT_type                     = 16
	KISS_type                        = 128
	CONSTRAINTS_type                 = 256
	SYMBOLIC_CONSTRAINTS_type        = 512
	FD_type                          = F_type | D_type
	FR_type                          = F_type | R_type
	DR_type                          = D_type | R_type
	FDR_type                         = F_type | D_type | R_type
	COMPL                            = 0x0001
	ESSEN                            = 0x0002
	EXPAND                           = 0x0004
	EXPAND1                          = 0x0008
	GASP                             = 0x0010
	IRRED                            = 0x0020
	REDUCE                           = 0x0040
	REDUCE1                          = 0x0080
	SPARSE                           = 0x0100
	TAUT                             = 0x0200
	EXACT                            = 0x0400
	MINCOV                           = 0x0800
	MINCOV1                          = 0x1000
	SHARP                            = 0x2000
	IRRED1                           = 0x4000
	VERSION                          = "UC Berkeley, Espresso Version #2.3, Release date 01/31/88"
	TIME_COUNT                       = 16
	READ_TIME                        = 0
	COMPL_TIME                       = 1
	ONSET_TIME                       = 2
	ESSEN_TIME                       = 3
	EXPAND_TIME                      = 4
	IRRED_TIME                       = 5
	REDUCE_TIME                      = 6
	GEXPAND_TIME                     = 7
	GIRRED_TIME                      = 8
	GREDUCE_TIME                     = 9
	PRIMES_TIME                      = 10
	MINCOV_TIME                      = 11
	MV_REDUCE_TIME                   = 12
	RAISE_IN_TIME                    = 13
	VERIFY_TIME                      = 14
	WRITE_TIME                       = 15
	CUBE_TEMP                        = 10
	DISJOINT                  uint32 = 0x55555555
	TWO                              = 3
	DASH                             = 3
	ONE                              = 2
	ZERO                             = 1
)

type cube_struct struct {
	size, num_vars, num_binary_vars                         int32
	first_part, last_part, part_size, first_word, last_word []int32
	binary_mask, mv_mask                                    pset
	var_mask, temp                                          []pset
	fullset, emptyset                                       pset
	inmask                                                  uint32
	inword                                                  int32
	sparse                                                  []int32
	num_mv_vars, output                                     int32
}
type cdata_struct struct {
	part_zeros, var_zeros, parts_active []int32
	is_unate                            []bool
	vars_active, vars_unate, best       int32
}

// cubeList replaces C's pointer-punned [cofactor,end,...cubes] vector. Cubes
// are borrowed rows; cofactor is an independently owned packed cube.
type cubeList struct {
	cofactor pcube
	cubes    []pcube
}

func CUBELISTSIZE(list *cubeList) int32 { return int32(len(list.cubes)) }
func free_cubelist(list *cubeList) {
	if list != nil {
		list.cofactor = nil
		list.cubes = nil
	}
}
func new_cube() pcube                 { return set_new(cube.size) }
func free_cube(c pcube)               { set_free(c) }
func new_cover(capacity int32) pcover { return sf_new(capacity, cube.size) }
func free_cover(c pcover)             { sf_free(c) }
func equal(a, b string) bool          { return a == b }
func GETINPUT(c pcube, position int32) uint32 {
	return c[WHICH_WORD(2*position)] >> WHICH_BIT(2*position) & 3
}
func GETOUTPUT(c pcube, position int32) bool {
	return is_in_set(c, cube.first_part[cube.output]+position)
}
func PUTINPUT(c pcube, position int32, value uint32) {
	word, shift := WHICH_WORD(2*position), WHICH_BIT(2*position)
	c[word] = c[word]&^(3<<shift) | value<<shift
}

// PUTOUTPUT takes an absolute packed-set position, as in the original macro.
func PUTOUTPUT(c pcube, position int32, value uint32) {
	word, shift := WHICH_WORD(position), WHICH_BIT(position)
	c[word] = c[word]&^(1<<shift) | value<<shift
}
func print_bool(value any) string {
	switch x := value.(type) {
	case bool:
		if x {
			return "TRUE"
		}
		return "FALSE"
	case int:
		if x == 0 {
			return "FALSE"
		}
		if x == 1 {
			return "TRUE"
		}
	case int32:
		if x == 0 {
			return "FALSE"
		}
		if x == 1 {
			return "TRUE"
		}
	}
	return "MAYBE"
}

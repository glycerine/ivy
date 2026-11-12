package espresso

import "math/bits"

var debug uint32
var verbose_debug bool
var total_name [TIME_COUNT]string
var total_time [TIME_COUNT]int64
var total_calls [TIME_COUNT]int32
var echo_comments, echo_unknown_commands, force_irredundant, skip_make_sparse bool
var kiss, pos, print_solution, recompute_onset, remove_essential, single_expand bool
var summary, trace, unwrap_onset, use_random_order, use_super_gasp bool
var debug_exact_minimization bool
var filename string
var pla_types = []pla_types_struct{{"-f", F_type}, {"-r", R_type}, {"-d", D_type}, {"-fd", FD_type}, {"-fr", FR_type}, {"-dr", DR_type}, {"-fdr", FDR_type}, {"-fc", F_type | CONSTRAINTS_type}, {"-rc", R_type | CONSTRAINTS_type}, {"-dc", D_type | CONSTRAINTS_type}, {"-fdc", FD_type | CONSTRAINTS_type}, {"-frc", FR_type | CONSTRAINTS_type}, {"-drc", DR_type | CONSTRAINTS_type}, {"-fdrc", FDR_type | CONSTRAINTS_type}, {"-pleasure", PLEASURE_type}, {"-eqn", EQNTOTT_type}, {"-eqntott", EQNTOTT_type}, {"-kiss", KISS_type}, {"-cons", CONSTRAINTS_type}, {"-scons", SYMBOLIC_CONSTRAINTS_type}, {"", 0}}
var cube, temp_cube_save cube_struct
var cdata, temp_cdata_save cdata_struct
var bit_count = func() [256]int32 {
	var table [256]int32
	for i := range table {
		table[i] = int32(bits.OnesCount8(uint8(i)))
	}
	return table
}()

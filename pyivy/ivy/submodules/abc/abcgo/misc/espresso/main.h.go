package espresso

type commandKey int32

const (
	KEY_ESPRESSO commandKey = iota
	KEY_PLA_verify
	KEY_check
	KEY_contain
	KEY_d1merge
	KEY_disjoint
	KEY_dsharp
	KEY_echo
	KEY_essen
	KEY_exact
	KEY_expand
	KEY_gasp
	KEY_intersect
	KEY_irred
	KEY_lexsort
	KEY_make_sparse
	KEY_map
	KEY_mapdc
	KEY_minterms
	KEY_opo
	KEY_opoall
	KEY_pair
	KEY_pairall
	KEY_primes
	KEY_qm
	KEY_reduce
	KEY_sharp
	KEY_simplify
	KEY_so
	KEY_so_both
	KEY_stats
	KEY_super_gasp
	KEY_taut
	KEY_test
	KEY_equiv
	KEY_union
	KEY_verify
	KEY_MANY_ESPRESSO
	KEY_separate
	KEY_xor
	KEY_d1merge_in
	KEY_fsm
	KEY_unknown
)

type commandOption struct {
	name                      string
	key                       commandKey
	num_plas                  int32
	needs_offset, needs_dcset bool
}

var option_table = []commandOption{
	{"ESPRESSO", KEY_ESPRESSO, 1, true, true},
	{"many", KEY_MANY_ESPRESSO, 1, true, true},
	{"exact", KEY_exact, 1, true, true},
	{"qm", KEY_qm, 1, true, true},
	{"single_output", KEY_so, 1, true, true},
	{"so", KEY_so, 1, true, true},
	{"so_both", KEY_so_both, 1, true, true},
	{"simplify", KEY_simplify, 1, false, false},
	{"echo", KEY_echo, 1, false, false},
	{"opo", KEY_opo, 1, true, true},
	{"opoall", KEY_opoall, 1, true, true},
	{"pair", KEY_pair, 1, true, true},
	{"pairall", KEY_pairall, 1, true, true},
	{"check", KEY_check, 1, true, true},
	{"stats", KEY_stats, 1, false, false},
	{"verify", KEY_verify, 2, false, true},
	{"PLAverify", KEY_PLA_verify, 2, false, true},
	{"equiv", KEY_equiv, 1, true, true},
	{"map", KEY_map, 1, false, false},
	{"mapdc", KEY_mapdc, 1, false, false},
	{"fsm", KEY_fsm, 1, false, true},
	{"contain", KEY_contain, 1, false, false},
	{"d1merge", KEY_d1merge, 1, false, false},
	{"d1merge_in", KEY_d1merge_in, 1, false, false},
	{"disjoint", KEY_disjoint, 1, true, false},
	{"dsharp", KEY_dsharp, 2, false, false},
	{"intersect", KEY_intersect, 2, false, false},
	{"minterms", KEY_minterms, 1, false, false},
	{"primes", KEY_primes, 1, false, true},
	{"separate", KEY_separate, 1, true, true},
	{"sharp", KEY_sharp, 2, false, false},
	{"union", KEY_union, 2, false, false},
	{"xor", KEY_xor, 2, true, true},
	{"essen", KEY_essen, 1, false, true},
	{"expand", KEY_expand, 1, true, false},
	{"gasp", KEY_gasp, 1, true, true},
	{"irred", KEY_irred, 1, false, true},
	{"make_sparse", KEY_make_sparse, 1, true, true},
	{"reduce", KEY_reduce, 1, false, true},
	{"taut", KEY_taut, 1, false, false},
	{"super_gasp", KEY_super_gasp, 1, true, true},
	{"lexsort", KEY_lexsort, 1, false, false},
	{"test", KEY_test, 1, true, true},
}

var debug_table = map[string]uint32{
	"":        EXPAND | ESSEN | IRRED | REDUCE | SPARSE | GASP | SHARP | MINCOV,
	"compl":   COMPL,
	"essen":   ESSEN,
	"expand":  EXPAND,
	"expand1": EXPAND1 | EXPAND,
	"irred":   IRRED,
	"irred1":  IRRED1 | IRRED,
	"reduce":  REDUCE,
	"reduce1": REDUCE1 | REDUCE,
	"mincov":  MINCOV,
	"mincov1": MINCOV1 | MINCOV,
	"sparse":  SPARSE,
	"sharp":   SHARP,
	"taut":    TAUT,
	"gasp":    GASP,
	"exact":   EXACT,
}

type espressoOption struct {
	variable *bool
	value    bool
}

var esp_opt_table = map[string]espressoOption{
	"eat":     {&echo_comments, false},
	"eatdots": {&echo_unknown_commands, false},
	"fast":    {&single_expand, true},
	"kiss":    {&kiss, true},
	"ness":    {&remove_essential, false},
	"nirr":    {&force_irredundant, false},
	"nunwrap": {&unwrap_onset, false},
	"onset":   {&recompute_onset, true},
	"pos":     {&pos, true},
	"random":  {&use_random_order, true},
	"strong":  {&use_super_gasp, true},
}

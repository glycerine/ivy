package abc

import "os"

// Hand-checked skeleton of src/base/abci/abcSaucy.c.

// SaucyStats is struct saucy_stats.
type SaucyStats struct {
	grpsize_base float64
	grpsize_exp  int
	levels       int
	nodes        int
	bads         int
	gens         int
	support      int
}

// SaucyGraph is struct saucy_graph.
type SaucyGraph struct {
	n   int
	e   int
	adj []int // int *
	edg []int // int *
}

// SaucyColoring is struct coloring.
type SaucyColoring struct {
	lab    []int // int * : Labelling of objects
	unlab  []int // int * : Inverse of lab
	cfront []int // int * : Pointer to front of cells
	clen   []int // int * : Length of cells (defined for cfront's)
}

// SaucySimResult is struct sim_result.
type SaucySimResult struct {
	inVec          []int // int *
	outVec         []int // int *
	inVecSignature int
	outVecOnes     int
	activity       float64
}

// Saucy is struct saucy.
type Saucy struct {
	/* Graph data */
	n    int   // Size of domain
	adj  []int // Neighbors of k: edg[adj[k]]..edg[adj[k+1]]
	edg  []int // Actual neighbor data
	dadj []int // Fanin neighbor indices, for digraphs
	dedg []int // Fanin neighbor data, for digraphs
	/* Coloring data */
	left, right SaucyColoring
	nextnon     []int // Forward next-nonsingleton pointers
	prevnon     []int // Backward next-nonsingleton pointers
	/* Refinement: inducers */
	indmark  []int8 // char * : Induce marks
	ninduce  []int  // Nonsingletons that might induce refinement
	sinduce  []int  // Singletons that might induce refinement
	nninduce int    // Size of ninduce stack
	nsinduce int    // Size of sinduce stack
	/* Refinement: marked cells */
	clist []int // List of cells marked for refining
	csize int   // Number of cells in clist
	/* Refinement: workspace */
	stuff    []int8 // char * : Bit vector, but one char per bit
	ccount   []int  // Number of connections to refining cell
	bucket   []int  // Workspace
	count    []int  // Num vertices with same adj count to ref cell
	junk     []int  // More workspace
	gamma    []int  // Working permutation
	conncnts []int  // Connection counts for cell fronts
	/* Search data */
	lev     int   // Current search tree level
	anc     int   // Level of greatest common ancestor with zeta
	anctar  []int // Copy of target cell at anc
	kanctar int   // Location within anctar to iterate from
	start   []int // Location of target at each level
	indmin  int   // Used for group size computation
	match   int   // Have we not diverged from previous left?
	/* Search: orbit partition */
	theta   []int // Running approximation of orbit partition
	thsize  []int // Size of cells in theta, defined for mcrs
	thnext  []int // Next rep in list (circular list)
	thprev  []int // Previous rep in list
	threp   []int // First rep for a given cell front
	thfront []int // The cell front associated with this rep
	/* Search: split record */
	splitvar  []int // The actual value of the splits on the left-most branch
	splitwho  []int // List of where splits occurred
	splitfrom []int // List of cells which were split
	splitlev  []int // Where splitwho/from begins for each level
	nsplits   int   // Number of splits at this point
	/* Search: differences from leftmost */
	diffmark   []int8 // char * : Marked for diff labels
	diffs      []int  // List of diff labels
	difflev    []int  // How many labels diffed at each level
	ndiffs     int    // Current number of diffs
	undifflev  []int  // How many diff labels fixed at each level
	nundiffs   int    // Current number of diffs in singletons (fixed)
	unsupp     []int  // Inverted diff array
	specmin    []int  // Speculated mappings
	pairs      []int  // Not-undiffed diffs that can make two-cycles
	unpairs    []int  // Indices into pairs
	npairs     int    // Number of such pairs
	diffnons   []int  // Diffs that haven't been undiffed
	undiffnons []int  // Inverse of that
	ndiffnons  int    // Number of such diffs
	/* Polymorphic functions */
	split                func(*Saucy, *SaucyColoring, int, int) int
	is_automorphism      func(*Saucy) int
	ref_singleton        func(*Saucy, *SaucyColoring, int) int
	ref_nonsingle        func(*Saucy, *SaucyColoring, int) int
	select_decomposition func(*Saucy, *int, *int, *int)
	/* Statistics structure */
	stats *SaucyStats
	/* New data structures for Boolean formulas */
	pNtk                   *ABCNetwork
	pNtk_permuted          *ABCNetwork
	depAdj                 []int
	depEdg                 []int
	iDep, oDep             []*IntVector
	obs, ctrl              []*IntVector
	topOrder               []*PointerVector[ABCNetworkObject]
	randomVectorArray_sim1 *PointerVector[IntVector]
	randomVectorSplit_sim1 []int
	randomVectorArray_sim2 *PointerVector[IntVector]
	randomVectorSplit_sim2 []int
	marks                  []int8 // char *
	pModel                 []int
	satCounterExamples     *PointerVector[SaucySimResult]
	activityInc            float64
	fBooleanMatching       int
	fPrintTree             int
	fLookForSwaps          int
	gFile                  *os.File
	refineBySim1           func(*Saucy, *SaucyColoring) int
	refineBySim2           func(*Saucy, *SaucyColoring) int
	print_automorphism     func(f *os.File, n int, gamma []int, nsupp int, support []int, marks []int8, pNtk *ABCNetwork) int
}

// ints (static)
// C: int * ints(int n)
func ints(n int) []int {
	panic("TODO, stubbed")
}

// zeros (static)
// C: int * zeros(int n)
func zeros(n int) []int {
	panic("TODO, stubbed")
}

// bits (static)
// C: char * bits(int n)
func bits_saucy(n int) []int8 {
	panic("TODO, stubbed")
}

// print_automorphism_ntk (static)
// C: int print_automorphism_ntk(FILE *f, int n, const int *gamma, int nsupp, const int *support, char * marks, Abc_Ntk_t * pNtk)
func print_automorphism_ntk(f *os.File, n int, gamma []int, nsupp int, support []int, marks []int8, pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// print_automorphism_ntk2 (static)
// C: int print_automorphism_ntk2(FILE *f, int n, const int *gamma, int nsupp, const int *support, char * marks, Abc_Ntk_t * pNtk)
func print_automorphism_ntk2(f *os.File, n int, gamma []int, nsupp int, support []int, marks []int8, pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// print_automorphism_quiet (static)
// C: int print_automorphism_quiet(FILE *f, int n, const int *gamma, int nsupp, const int *support, char * marks, Abc_Ntk_t * pNtk)
func print_automorphism_quiet(f *os.File, n int, gamma []int, nsupp int, support []int, marks []int8, pNtk *ABCNetwork) int {
	panic("TODO, stubbed")
}

// array_find_min (static)
// C: int array_find_min(const int *a, int n)
func array_find_min(a []int, n int) int {
	panic("TODO, stubbed")
}

// swap (static)
// C: void swap(int *a, int x, int y)
func swap_saucy(a []int, x int, y int) {
	panic("TODO, stubbed")
}

// sift_up (static)
// C: void sift_up(int *a, int k)
func sift_up(a []int, k int) {
	panic("TODO, stubbed")
}

// sift_down (static)
// C: void sift_down(int *a, int n)
func sift_down(a []int, n int) {
	panic("TODO, stubbed")
}

// heap_sort (static)
// C: void heap_sort(int *a, int n)
func heap_sort(a []int, n int) {
	panic("TODO, stubbed")
}

// insertion_sort (static)
// C: void insertion_sort(int *a, int n)
func insertion_sort(a []int, n int) {
	panic("TODO, stubbed")
}

// partition (static)
// C: int partition(int *a, int n, int m)
func partition_saucy(a []int, n int, m int) int {
	panic("TODO, stubbed")
}

// log_base2 (static)
// C: int log_base2(int n)
func log_base2(n int) int {
	panic("TODO, stubbed")
}

// median (static)
// C: int median(int a, int b, int c)
func median_saucy(a int, b int, c int) int {
	panic("TODO, stubbed")
}

// introsort_loop (static)
// C: void introsort_loop(int *a, int n, int lim)
func introsort_loop(a []int, n int, lim int) {
	panic("TODO, stubbed")
}

// introsort (static)
// C: void introsort(int *a, int n)
func introsort(a []int, n int) {
	panic("TODO, stubbed")
}

// do_find_min (static)
// C: int do_find_min(struct coloring *c, int t)
func do_find_min(c *SaucyColoring, t int) int {
	panic("TODO, stubbed")
}

// find_min (static)
// C: int find_min(struct saucy *s, int t)
func (s *Saucy) find_min(t int) int {
	panic("TODO, stubbed")
}

// set_label (static)
// C: void set_label(struct coloring *c, int index, int value)
func set_label(c *SaucyColoring, index int, value int) {
	panic("TODO, stubbed")
}

// swap_labels (static)
// C: void swap_labels(struct coloring *c, int a, int b)
func swap_labels(c *SaucyColoring, a int, b int) {
	panic("TODO, stubbed")
}

// move_to_back (static)
// C: void move_to_back(struct saucy *s, struct coloring *c, int k)
func (s *Saucy) move_to_back(c *SaucyColoring, k int) {
	panic("TODO, stubbed")
}

// data_mark (static)
// C: void data_mark(struct saucy *s, struct coloring *c, int k)
func (s *Saucy) data_mark(c *SaucyColoring, k int) {
	panic("TODO, stubbed")
}

// data_count (static)
// C: void data_count(struct saucy *s, struct coloring *c, int k)
func (s *Saucy) data_count(c *SaucyColoring, k int) {
	panic("TODO, stubbed")
}

// check_mapping (static)
// C: int check_mapping(struct saucy *s, const int *adj, const int *edg, int k)
func (s *Saucy) check_mapping(adj []int, edg []int, k int) int {
	panic("TODO, stubbed")
}

// add_conterexample (static)
// C: int add_conterexample(struct saucy *s, struct sim_result * cex)
func (s *Saucy) add_conterexample(cex *SaucySimResult) int {
	panic("TODO, stubbed")
}

// is_undirected_automorphism (static)
// C: int is_undirected_automorphism(struct saucy *s)
func (s *Saucy) is_undirected_automorphism() int {
	panic("TODO, stubbed")
}

// is_directed_automorphism (static)
// C: int is_directed_automorphism(struct saucy *s)
func (s *Saucy) is_directed_automorphism() int {
	panic("TODO, stubbed")
}

// add_induce (static)
// C: void add_induce(struct saucy *s, struct coloring *c, int who)
func (s *Saucy) add_induce(c *SaucyColoring, who int) {
	panic("TODO, stubbed")
}

// fix_fronts (static)
// C: void fix_fronts(struct coloring *c, int cf, int ff)
func fix_fronts(c *SaucyColoring, cf int, ff int) {
	panic("TODO, stubbed")
}

// array_indirect_sort (static)
// C: void array_indirect_sort(int *a, const int *b, int n)
func array_indirect_sort(a []int, b []int, n int) {
	panic("TODO, stubbed")
}

// at_terminal (static)
// C: int at_terminal(struct saucy *s)
func (s *Saucy) at_terminal() int {
	panic("TODO, stubbed")
}

// add_diffnon (static)
// C: void add_diffnon(struct saucy *s, int k)
func (s *Saucy) add_diffnon(k int) {
	panic("TODO, stubbed")
}

// remove_diffnon (static)
// C: void remove_diffnon(struct saucy *s, int k)
func (s *Saucy) remove_diffnon(k int) {
	panic("TODO, stubbed")
}

// add_diff (static)
// C: void add_diff(struct saucy *s, int k)
func (s *Saucy) add_diff(k int) {
	panic("TODO, stubbed")
}

// is_a_pair (static)
// C: int is_a_pair(struct saucy *s, int k)
func (s *Saucy) is_a_pair(k int) int {
	panic("TODO, stubbed")
}

// in_cell_range (static)
// C: int in_cell_range(struct coloring *c, int ff, int cf)
func in_cell_range(c *SaucyColoring, ff int, cf int) int {
	panic("TODO, stubbed")
}

// add_pair (static)
// C: void add_pair(struct saucy *s, int k)
func (s *Saucy) add_pair(k int) {
	panic("TODO, stubbed")
}

// eat_pair (static)
// C: void eat_pair(struct saucy *s, int k)
func (s *Saucy) eat_pair(k int) {
	panic("TODO, stubbed")
}

// pick_all_the_pairs (static)
// C: void pick_all_the_pairs(struct saucy *s)
func (s *Saucy) pick_all_the_pairs() {
	panic("TODO, stubbed")
}

// clear_undiffnons (static)
// C: void clear_undiffnons(struct saucy *s)
func (s *Saucy) clear_undiffnons() {
	panic("TODO, stubbed")
}

// fix_diff_singleton (static)
// C: void fix_diff_singleton(struct saucy *s, int cf)
func (s *Saucy) fix_diff_singleton(cf int) {
	panic("TODO, stubbed")
}

// fix_diff_subtract (static)
// C: void fix_diff_subtract(struct saucy *s, int cf, const int *a, const int *b)
func (s *Saucy) fix_diff_subtract(cf int, a []int, b []int) {
	panic("TODO, stubbed")
}

// fix_diffs (static)
// C: void fix_diffs(struct saucy *s, int cf, int ff)
func (s *Saucy) fix_diffs(cf int, ff int) {
	panic("TODO, stubbed")
}

// split_color (static)
// C: void split_color(struct coloring *c, int cf, int ff)
func split_color(c *SaucyColoring, cf int, ff int) {
	panic("TODO, stubbed")
}

// split_common (static)
// C: void split_common(struct saucy *s, struct coloring *c, int cf, int ff)
func (s *Saucy) split_common(c *SaucyColoring, cf int, ff int) {
	panic("TODO, stubbed")
}

// split_left (static)
// C: int split_left(struct saucy *s, struct coloring *c, int cf, int ff)
func (s *Saucy) split_left(c *SaucyColoring, cf int, ff int) int {
	panic("TODO, stubbed")
}

// split_init (static)
// C: int split_init(struct saucy *s, struct coloring *c, int cf, int ff)
func (s *Saucy) split_init(c *SaucyColoring, cf int, ff int) int {
	panic("TODO, stubbed")
}

// split_other (static)
// C: int split_other(struct saucy *s, struct coloring *c, int cf, int ff)
func (s *Saucy) split_other(c *SaucyColoring, cf int, ff int) int {
	panic("TODO, stubbed")
}

// print_partition (static)
// C: int print_partition(struct coloring *left, struct coloring *right, int n, Abc_Ntk_t * pNtk, int fNames)
func print_partition(left *SaucyColoring, right *SaucyColoring, n int, pNtk *ABCNetwork, fNames int) int {
	panic("TODO, stubbed")
}

// refine_cell (static)
// C: int refine_cell(struct saucy *s, struct coloring *c, int (*refine)(struct saucy *, struct coloring *, int))
func (s *Saucy) refine_cell(c *SaucyColoring, refine func(*Saucy, *SaucyColoring, int) int) int {
	panic("TODO, stubbed")
}

// maybe_split (static)
// C: int maybe_split(struct saucy *s, struct coloring *c, int cf, int ff)
func (s *Saucy) maybe_split(c *SaucyColoring, cf int, ff int) int {
	panic("TODO, stubbed")
}

// ref_single_cell (static)
// C: int ref_single_cell(struct saucy *s, struct coloring *c, int cf)
func (s *Saucy) ref_single_cell(c *SaucyColoring, cf int) int {
	panic("TODO, stubbed")
}

// ref_singleton_directed (static)
// C: int ref_singleton_directed(struct saucy *s, struct coloring *c, int cf)
func (s *Saucy) ref_singleton_directed(c *SaucyColoring, cf int) int {
	panic("TODO, stubbed")
}

// ref_singleton_undirected (static)
// C: int ref_singleton_undirected(struct saucy *s, struct coloring *c, int cf)
func (s *Saucy) ref_singleton_undirected(c *SaucyColoring, cf int) int {
	panic("TODO, stubbed")
}

// ref_nonsingle_cell (static)
// C: int ref_nonsingle_cell(struct saucy *s, struct coloring *c, int cf)
func (s *Saucy) ref_nonsingle_cell(c *SaucyColoring, cf int) int {
	panic("TODO, stubbed")
}

// ref_nonsingle_directed (static)
// C: int ref_nonsingle_directed(struct saucy *s, struct coloring *c, int cf)
func (s *Saucy) ref_nonsingle_directed(c *SaucyColoring, cf int) int {
	panic("TODO, stubbed")
}

// ref_nonsingle_undirected (static)
// C: int ref_nonsingle_undirected(struct saucy *s, struct coloring *c, int cf)
func (s *Saucy) ref_nonsingle_undirected(c *SaucyColoring, cf int) int {
	panic("TODO, stubbed")
}

// ref_singleton (static; _method avoids the clash with the struct field of the same name)
// C: int ref_singleton(struct saucy *s, struct coloring *c, const int *adj, const int *edg, int cf)
func (s *Saucy) ref_singleton_method(c *SaucyColoring, adj []int, edg []int, cf int) int {
	panic("TODO, stubbed")
}

// ref_nonsingle (static; _method avoids the clash with the struct field of the same name)
// C: int ref_nonsingle(struct saucy *s, struct coloring *c, const int *adj, const int *edg, int cf)
func (s *Saucy) ref_nonsingle_method(c *SaucyColoring, adj []int, edg []int, cf int) int {
	panic("TODO, stubbed")
}

// clear_refine (static)
// C: void clear_refine(struct saucy *s)
func (s *Saucy) clear_refine() {
	panic("TODO, stubbed")
}

// refine (static)
// C: int refine(struct saucy *s, struct coloring *c)
func (s *Saucy) refine(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// refineByDepGraph (static)
// C: int refineByDepGraph(struct saucy *s, struct coloring *c)
func (s *Saucy) refineByDepGraph(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// backtrackBysatCounterExamples (static)
// C: int backtrackBysatCounterExamples(struct saucy *s, struct coloring *c)
func (s *Saucy) backtrackBysatCounterExamples(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// refineBySim1_init (static)
// C: int refineBySim1_init(struct saucy *s, struct coloring *c)
func (s *Saucy) refineBySim1_init(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// refineBySim1_left (static)
// C: int refineBySim1_left(struct saucy *s, struct coloring *c)
func (s *Saucy) refineBySim1_left(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// refineBySim1_other (static)
// C: int refineBySim1_other(struct saucy *s, struct coloring *c)
func (s *Saucy) refineBySim1_other(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// refineBySim2_init (static)
// C: int refineBySim2_init(struct saucy *s, struct coloring *c)
func (s *Saucy) refineBySim2_init(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// refineBySim2_left (static)
// C: int refineBySim2_left(struct saucy *s, struct coloring *c)
func (s *Saucy) refineBySim2_left(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// refineBySim2_other (static)
// C: int refineBySim2_other(struct saucy *s, struct coloring *c)
func (s *Saucy) refineBySim2_other(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// check_OPP_only_has_swaps (static)
// C: int check_OPP_only_has_swaps(struct saucy *s, struct coloring *c)
func (s *Saucy) check_OPP_only_has_swaps(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// check_OPP_for_Boolean_matching (static)
// C: int check_OPP_for_Boolean_matching(struct saucy *s, struct coloring *c)
func (s *Saucy) check_OPP_for_Boolean_matching(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// double_check_OPP_isomorphism (static)
// C: int double_check_OPP_isomorphism(struct saucy *s, struct coloring * c)
func (s *Saucy) double_check_OPP_isomorphism(c *SaucyColoring) int {
	panic("TODO, stubbed")
}

// descend (static)
// C: int descend(struct saucy *s, struct coloring *c, int target, int min)
func (s *Saucy) descend(c *SaucyColoring, target int, min int) int {
	panic("TODO, stubbed")
}

// select_smallest_max_connected_cell (static)
// C: int select_smallest_max_connected_cell(struct saucy *s, int start, int end)
func (s *Saucy) select_smallest_max_connected_cell(start int, end int) int {
	panic("TODO, stubbed")
}

// descend_leftmost (static)
// C: int descend_leftmost(struct saucy *s)
func (s *Saucy) descend_leftmost() int {
	panic("TODO, stubbed")
}

// zeta_fixed (static)
// C: int zeta_fixed(struct saucy *s)
func (s *Saucy) zeta_fixed() int {
	panic("TODO, stubbed")
}

// select_dynamically (static)
// C: void select_dynamically(struct saucy *s, int *target, int *lmin, int *rmin)
func (s *Saucy) select_dynamically(target *int, lmin *int, rmin *int) {
	panic("TODO, stubbed")
}

// select_statically (static)
// C: void select_statically(struct saucy *s, int *target, int *lmin, int *rmin)
func (s *Saucy) select_statically(target *int, lmin *int, rmin *int) {
	panic("TODO, stubbed")
}

// descend_left (static)
// C: int descend_left(struct saucy *s)
func (s *Saucy) descend_left() int {
	panic("TODO, stubbed")
}

// find_representative (static)
// C: int find_representative(int k, int *theta)
func find_representative(k int, theta []int) int {
	panic("TODO, stubbed")
}

// update_theta (static)
// C: void update_theta(struct saucy *s)
func (s *Saucy) update_theta() {
	panic("TODO, stubbed")
}

// theta_prune (static)
// C: int theta_prune(struct saucy *s)
func (s *Saucy) theta_prune() int {
	panic("TODO, stubbed")
}

// orbit_prune (static)
// C: int orbit_prune(struct saucy *s)
func (s *Saucy) orbit_prune() int {
	panic("TODO, stubbed")
}

// note_anctar_reps (static)
// C: void note_anctar_reps(struct saucy *s)
func (s *Saucy) note_anctar_reps() {
	panic("TODO, stubbed")
}

// multiply_index (static)
// C: void multiply_index(struct saucy *s, int k)
func (s *Saucy) multiply_index(k int) {
	panic("TODO, stubbed")
}

// backtrack_leftmost (static)
// C: int backtrack_leftmost(struct saucy *s)
func (s *Saucy) backtrack_leftmost() int {
	panic("TODO, stubbed")
}

// backtrack_other (static)
// C: int backtrack_other(struct saucy *s)
func (s *Saucy) backtrack_other() int {
	panic("TODO, stubbed")
}

// rewind_coloring (static)
// C: void rewind_coloring(struct saucy *s, struct coloring *c, int lev)
func (s *Saucy) rewind_coloring(c *SaucyColoring, lev int) {
	panic("TODO, stubbed")
}

// rewind_simulation_vectors (static)
// C: void rewind_simulation_vectors(struct saucy *s, int lev)
func (s *Saucy) rewind_simulation_vectors(lev int) {
	panic("TODO, stubbed")
}

// do_backtrack (static)
// C: int do_backtrack(struct saucy *s)
func (s *Saucy) do_backtrack() int {
	panic("TODO, stubbed")
}

// backtrack_loop (static)
// C: int backtrack_loop(struct saucy *s)
func (s *Saucy) backtrack_loop() int {
	panic("TODO, stubbed")
}

// backtrack (static)
// C: int backtrack(struct saucy *s)
func (s *Saucy) backtrack() int {
	panic("TODO, stubbed")
}

// backtrack_bad (static)
// C: int backtrack_bad(struct saucy *s)
func (s *Saucy) backtrack_bad() int {
	panic("TODO, stubbed")
}

// prepare_permutation_ntk
// C: void prepare_permutation_ntk(struct saucy *s)
func (s *Saucy) Prepare_permutation_ntk() {
	panic("TODO, stubbed")
}

// prepare_permutation (static)
// C: void prepare_permutation(struct saucy *s)
func (s *Saucy) prepare_permutation() {
	panic("TODO, stubbed")
}

// unprepare_permutation_ntk
// C: void unprepare_permutation_ntk(struct saucy *s)
func (s *Saucy) Unprepare_permutation_ntk() {
	panic("TODO, stubbed")
}

// unprepare_permutation (static)
// C: void unprepare_permutation(struct saucy *s)
func (s *Saucy) unprepare_permutation() {
	panic("TODO, stubbed")
}

// do_search (static)
// C: int do_search(struct saucy *s)
func (s *Saucy) do_search() int {
	panic("TODO, stubbed")
}

// saucy_search
// C: void saucy_search( Abc_Ntk_t * pNtk, struct saucy *s, int directed, const int *colors, struct saucy_stats *stats)
func SaucySearch(pNtk *ABCNetwork, s *Saucy, directed int, colors []int, stats *SaucyStats) {
	panic("TODO, stubbed")
}

// saucy_free
// C: void saucy_free(struct saucy *s)
func (s *Saucy) SaucyFree() {
	panic("TODO, stubbed")
}

// saucy_alloc
// C: struct saucy * saucy_alloc(Abc_Ntk_t * pNtk)
func SaucyAlloc(pNtk *ABCNetwork) *Saucy {
	panic("TODO, stubbed")
}

// print_stats (static)
// C: void print_stats(FILE *f, struct saucy_stats stats )
func printStats(f *os.File, stats SaucyStats) {
	panic("TODO, stubbed")
}

// getVertexName (static)
// C: char * getVertexName(Abc_Ntk_t *pNtk, int v)
func getVertexName(pNtk *ABCNetwork, v int) string {
	panic("TODO, stubbed")
}

// findTopologicalOrder (static)
// C: Vec_Ptr_t ** findTopologicalOrder( Abc_Ntk_t * pNtk )
func findTopologicalOrder_saucy(pNtk *ABCNetwork) []*PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// getDependencies (static)
// C: void getDependencies(Abc_Ntk_t *pNtk, Vec_Int_t** iDep, Vec_Int_t** oDep)
func (pNtk *ABCNetwork) getDependencies_saucy(iDep *[]*IntVector, oDep *[]*IntVector) {
	panic("TODO, stubbed")
}

// getDependenciesDummy (static)
// C: void getDependenciesDummy(Abc_Ntk_t *pNtk, Vec_Int_t** iDep, Vec_Int_t** oDep)
func (pNtk *ABCNetwork) getDependenciesDummy(iDep *[]*IntVector, oDep *[]*IntVector) {
	panic("TODO, stubbed")
}

// buildDepGraph (static)
// C: struct saucy_graph * buildDepGraph(Abc_Ntk_t *pNtk, Vec_Int_t ** iDep, Vec_Int_t ** oDep)
func (pNtk *ABCNetwork) buildDepGraph(iDep []*IntVector, oDep []*IntVector) *SaucyGraph {
	panic("TODO, stubbed")
}

// assignRandomBitsToCells (static)
// C: Vec_Int_t * assignRandomBitsToCells(Abc_Ntk_t * pNtk, struct coloring *c)
func (pNtk *ABCNetwork) assignRandomBitsToCells(c *SaucyColoring) *IntVector {
	panic("TODO, stubbed")
}

// generateProperInputVector (static)
// C: int * generateProperInputVector( Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randomVector )
func (pNtk *ABCNetwork) generateProperInputVector(c *SaucyColoring, randomVector *IntVector) []int {
	panic("TODO, stubbed")
}

// ifInputVectorsAreConsistent (static)
// C: int ifInputVectorsAreConsistent( struct saucy * s, int * leftVec, int * rightVec )
func (s *Saucy) ifInputVectorsAreConsistent(leftVec []int, rightVec []int) int {
	panic("TODO, stubbed")
}

// ifOutputVectorsAreConsistent (static)
// C: int ifOutputVectorsAreConsistent( struct saucy * s, int * leftVec, int * rightVec )
func (s *Saucy) ifOutputVectorsAreConsistent(leftVec []int, rightVec []int) int {
	panic("TODO, stubbed")
}

// buildSim1Graph (static)
// C: struct saucy_graph * buildSim1Graph( Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randVec, Vec_Int_t ** iDep, Vec_Int_t ** oDep )
func (pNtk *ABCNetwork) buildSim1Graph(c *SaucyColoring, randVec *IntVector, iDep []*IntVector, oDep []*IntVector) *SaucyGraph {
	panic("TODO, stubbed")
}

// buildSim2Graph (static)
// C: struct saucy_graph * buildSim2Graph( Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randVec, Vec_Int_t ** iDep, Vec_Int_t ** oDep, Vec_Ptr_t ** topOrder, Vec_Int_t ** obs, Vec_Int_t ** ctrl )
func (pNtk *ABCNetwork) buildSim2Graph(c *SaucyColoring, randVec *IntVector, iDep []*IntVector, oDep []*IntVector, topOrder []*PointerVector[ABCNetworkObject], obs *[]*IntVector, ctrl *[]*IntVector) *SaucyGraph {
	panic("TODO, stubbed")
}

// bumpActivity (static)
// C: void bumpActivity( struct saucy * s, struct sim_result * cex )
func (s *Saucy) bumpActivity(cex *SaucySimResult) {
	panic("TODO, stubbed")
}

// reduceDB (static)
// C: void reduceDB( struct saucy * s )
func (s *Saucy) reduceDB() {
	panic("TODO, stubbed")
}

// analyzeConflict (static)
// C: struct sim_result * analyzeConflict( Abc_Ntk_t * pNtk, int * pModel, int fVerbose )
func (pNtk *ABCNetwork) analyzeConflict(pModel []int, fVerbose int) *SaucySimResult {
	panic("TODO, stubbed")
}

// Abc_NtkCecSat_saucy (static)
// C: int Abc_NtkCecSat_saucy( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel )
func (pNtk1 *ABCNetwork) abcNtkCecSat_saucy(pNtk2 *ABCNetwork, pModel []int) int {
	panic("TODO, stubbed")
}

// saucyGateWay
// C: void saucyGateWay( Abc_Ntk_t * pNtkOrig, Abc_Obj_t * pNodePo, FILE * gFile, int fBooleanMatching, int fLookForSwaps, int fFixOutputs, int fFixInputs, int fQuiet, int fPrintTree )
func (pNtkOrig *ABCNetwork) SaucyGateWay(pNodePo *ABCNetworkObject, gFile *os.File, fBooleanMatching int, fLookForSwaps int, fFixOutputs int, fFixInputs int, fQuiet int, fPrintTree int) {
	panic("TODO, stubbed")
}

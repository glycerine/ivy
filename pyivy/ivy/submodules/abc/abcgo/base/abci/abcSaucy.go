package abci

import (
	"fmt"
	"os"
	"time"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sim"
)

/**CFile****************************************************************

  FileName    [abcSaucy.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Symmetry Detection Package.]

  Synopsis    [Finds symmetries under permutation (but not negation) of I/Os.]

  Author      [Hadi Katebi]

  Affiliation [University of Michigan]

  Date        [Ver. 1.0. Started - April, 2012.]

  Revision    [No revisions so far]

  Comments    []

  Debugging   [There are some part of the code that are commented out. Those parts mostly print
               the contents of the data structures to the standard output. Un-comment them if you
               find them useful for debugging.]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "opt/sim/sim.h"

//ABC_NAMESPACE_IMPL_START

/* on/off switches */
const REFINE_BY_SIM_1 = 0
const REFINE_BY_SIM_2 = 0
const BACKTRACK_BY_SAT = 1
const SELECT_DYNAMICALLY = 0

/* number of iterations for sim1 and sim2 refinements */
var NUM_SIM1_ITERATION int32
var NUM_SIM2_ITERATION int32

/* conflict analysis */
const CLAUSE_DECAY = 0.9
const MAX_LEARNTS = 50

/*
 * saucy.c
 *
 * by Paul T. Darga <pdarga@umich.edu>
 * and Mark Liffiton <liffiton@umich.edu>
 * and Hadi Katebi <hadik@eecs.umich.edu>
 *
 * Copyright (C) 2004, The Regents of the University of Michigan
 * See the LICENSE file for details.
 */

type saucy_stats struct {
	grpsize_base float64
	grpsize_exp  int32
	levels       int32
	nodes        int32
	bads         int32
	gens         int32
	support      int32
}

type saucy_graph struct {
	n   int32
	e   int32
	adj []int32
	edg []int32
}

type coloring struct {
	lab    []int32 /* Labelling of objects */
	unlab  []int32 /* Inverse of lab */
	cfront []int32 /* Pointer to front of cells */
	clen   []int32 /* Length of cells (defined for cfront's) */
}

type sim_result struct {
	inVec          []int32
	outVec         []int32
	inVecSignature int32
	outVecOnes     int32
	activity       float64
}

type saucy struct {
	/* Graph data */
	n    int32   /* Size of domain */
	adj  []int32 /* Neighbors of k: edg[adj[k]]..edg[adj[k+1]] */
	edg  []int32 /* Actual neighbor data */
	dadj []int32 /* Fanin neighbor indices, for digraphs */
	dedg []int32 /* Fanin neighbor data, for digraphs */

	/* Coloring data */
	left, right coloring
	nextnon     []int32 /* Forward next-nonsingleton pointers */
	prevnon     []int32 /* Backward next-nonsingleton pointers */

	/* Refinement: inducers */
	indmark  []byte  /* Induce marks */
	ninduce  []int32 /* Nonsingletons that might induce refinement */
	sinduce  []int32 /* Singletons that might induce refinement */
	nninduce int32   /* Size of ninduce stack */
	nsinduce int32   /* Size of sinduce stack */

	/* Refinement: marked cells */
	clist []int32 /* List of cells marked for refining */
	csize int32   /* Number of cells in clist */

	/* Refinement: workspace */
	stuff    []byte  /* Bit vector, but one char per bit */
	ccount   []int32 /* Number of connections to refining cell */
	bucket   []int32 /* Workspace */
	count    []int32 /* Num vertices with same adj count to ref cell */
	junk     []int32 /* More workspace */
	gamma    []int32 /* Working permutation */
	conncnts []int32 /* Connection counts for cell fronts */

	/* Search data */
	lev     int32   /* Current search tree level */
	anc     int32   /* Level of greatest common ancestor with zeta */
	anctar  []int32 /* Copy of target cell at anc */
	kanctar int32   /* Location within anctar to iterate from */
	start   []int32 /* Location of target at each level */
	indmin  int32   /* Used for group size computation */
	match   int32   /* Have we not diverged from previous left? */

	/* Search: orbit partition */
	theta   []int32 /* Running approximation of orbit partition */
	thsize  []int32 /* Size of cells in theta, defined for mcrs */
	thnext  []int32 /* Next rep in list (circular list) */
	thprev  []int32 /* Previous rep in list */
	threp   []int32 /* First rep for a given cell front */
	thfront []int32 /* The cell front associated with this rep */

	/* Search: split record */
	splitvar  []int32 /* The actual value of the splits on the left-most branch */
	splitwho  []int32 /* List of where splits occurred */
	splitfrom []int32 /* List of cells which were split */
	splitlev  []int32 /* Where splitwho/from begins for each level */
	nsplits   int32   /* Number of splits at this point */

	/* Search: differences from leftmost */
	diffmark   []byte  /* Marked for diff labels */
	diffs      []int32 /* List of diff labels */
	difflev    []int32 /* How many labels diffed at each level */
	ndiffs     int32   /* Current number of diffs */
	undifflev  []int32 /* How many diff labels fixed at each level */
	nundiffs   int32   /* Current number of diffs in singletons (fixed) */
	unsupp     []int32 /* Inverted diff array */
	specmin    []int32 /* Speculated mappings */
	pairs      []int32 /* Not-undiffed diffs that can make two-cycles */
	unpairs    []int32 /* Indices into pairs */
	npairs     int32   /* Number of such pairs */
	diffnons   []int32 /* Diffs that haven't been undiffed */
	undiffnons []int32 /* Inverse of that */
	ndiffnons  int32   /* Number of such diffs */

	/* Polymorphic functions */
	split                func(*saucy, *coloring, int32, int32) int32
	is_automorphism      func(*saucy) int32
	ref_singleton        func(*saucy, *coloring, int32) int32
	ref_nonsingle        func(*saucy, *coloring, int32) int32
	select_decomposition func(*saucy, *int32, *int32, *int32)

	/* Statistics structure */
	stats *saucy_stats

	/* New data structures for Boolean formulas */
	pNtk                   *Abc_Ntk_t
	pNtk_permuted          *Abc_Ntk_t
	depAdj                 []int32
	depEdg                 []int32
	iDep, oDep             []*Vec_Int_t
	obs, ctrl              []*Vec_Int_t
	topOrder               []*Vec_Ptr_t
	randomVectorArray_sim1 *Vec_Ptr_t
	randomVectorSplit_sim1 []int32
	randomVectorArray_sim2 *Vec_Ptr_t
	randomVectorSplit_sim2 []int32
	marks                  []byte
	pModel                 []int32
	satCounterExamples     *Vec_Ptr_t
	activityInc            float64

	fBooleanMatching int32
	fPrintTree       int32
	fLookForSwaps    int32
	gFile            *os.File

	refineBySim1       func(*saucy, *coloring) int32
	refineBySim2       func(*saucy, *coloring) int32
	print_automorphism func(*os.File, int32, []int32, int32, []int32, []byte, *Abc_Ntk_t) int32
}

// inlining of macro ABC_ALLOC
func ints(n int32) []int32 { return make([]int32, n) }
func print_automorphism_ntk(f *os.File, n int32, gamma []int32, nsupp int32, support []int32, marks []byte, pNtk *Abc_Ntk_t) int32 {
	/* We presume support is already sorted */
	for i := int32(0); i < nsupp; i++ {
		k := support[i]

		/* Skip elements already seen */
		if marks[k] != 0 {
			continue
		}

		/* Start an orbit */
		marks[k] = 1
		fmt.Fprintf(f, "(%s", getVertexName(pNtk, k))

		/* Mark and notify elements in this orbit */
		for j := gamma[k]; j != k; j = gamma[j] {
			marks[j] = 1
			fmt.Fprintf(f, " %s", getVertexName(pNtk, j))
		}

		/* Finish off the orbit */
		fmt.Fprintf(f, ")")
	}
	fmt.Fprintf(f, "\n")

	/* Clean up after ourselves */
	for i := int32(0); i < nsupp; i++ {
		marks[support[i]] = 0
	}

	return 1
}

func print_automorphism_ntk2(f *os.File, n int32, gamma []int32, nsupp int32, support []int32, marks []byte, pNtk *Abc_Ntk_t) int32 {
	/* We presume support is already sorted */
	for i := int32(0); i < nsupp; i++ {
		k := support[i]

		/* Skip elements already seen */
		if marks[k] != 0 {
			continue
		}

		/* Start an orbit */
		marks[k] = 1
		fmt.Fprintf(f, "%d", k-1)

		/* Mark and notify elements in this orbit */
		for j := gamma[k]; j != k; j = gamma[j] {
			marks[j] = 1
			fmt.Fprintf(f, " %d ", j-1)
		}

		/* Finish off the orbit */
	}
	fmt.Fprintf(f, "-1\n")

	/* Clean up after ourselves */
	for i := int32(0); i < nsupp; i++ {
		marks[support[i]] = 0
	}

	return 1
}

func print_automorphism_quiet(f *os.File, n int32, gamma []int32, nsupp int32, support []int32, marks []byte, pNtk *Abc_Ntk_t) int32 {
	return 1
}

func array_find_min(a []int32) int32 {
	min := 0
	for i := 1; i < len(a); i++ {
		if a[i] < a[min] {
			min = i
		}
	}
	return int32(min)
}

func swap(a []int32, x, y int32) {
	a[x], a[y] = a[y], a[x]
}

func sift_up(a []int32, k int32) {
	for k > 0 {
		p := (k - 1) / 2
		if a[k] <= a[p] {
			return
		}
		swap(a, k, p)
		k = p
	}
}

func sift_down(a []int32, n int32) {
	for p := int32(0); ; {
		k := 2*p + 1
		if k >= n {
			return
		}
		if k+1 < n && a[k] < a[k+1] {
			k++
		}
		if a[p] >= a[k] {
			return
		}
		swap(a, p, k)
		p = k
	}
}

func heap_sort(a []int32) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		sift_down(a, int32(len(a)))
	}
	for end := len(a) - 1; end > 0; end-- {
		swap(a, 0, int32(end))
		sift_down(a[:end], int32(end))
	}
}

func insertion_sort(a []int32) {
	for i := 1; i < len(a); i++ {
		k := a[i]
		j := i
		for j > 0 && a[j-1] > k {
			a[j] = a[j-1]
			j--
		}
		a[j] = k
	}
}

func partition(a []int32, m int32) int32 {
	f, b := 0, len(a)
	for {
		for f < len(a) && a[f] <= m {
			f++
		}
		b--
		for b >= 0 && m <= a[b] {
			b--
		}
		if f < b {
			swap(a, int32(f), int32(b))
			f++
		} else {
			return int32(f)
		}
	}
}

func log_base2(n int32) int32 {
	var k int32
	for n > 1 {
		k++
		n >>= 1
	}
	return k
}

func median(a int32, b int32, c int32) int32 {
	if a <= b {
		if b <= c {
			return b
		}
		if a <= c {
			return c
		}
		return a

	} else {
		if a <= c {
			return a
		}
		if b <= c {
			return c
		}
		return b
	}
}

func introsort_loop(a []int32, lim int32) {
	for len(a) > 16 {
		if lim == 0 {
			heap_sort(a)
			return
		}
		lim--
		pivot := median(a[0], a[len(a)/2], a[len(a)-1])
		p := partition(a, pivot)
		introsort_loop(a[p:], lim)
		a = a[:p]
	}
}

func introsort(a []int32) {
	introsort_loop(a, 2*log_base2(int32(len(a))))
	insertion_sort(a)
}

func do_find_min(c *coloring, t int32) int32 {
	return array_find_min(c.lab[t:t+c.clen[t]+1]) + t
}

func find_min(s *saucy, t int32) int32 {
	return do_find_min(&s.right, t)
}

func set_label(c *coloring, index, value int32) {
	c.lab[index] = value
	c.unlab[value] = index
}

func swap_labels(c *coloring, a, b int32) {
	tmp := c.lab[a]
	set_label(c, a, c.lab[b])
	set_label(c, b, tmp)
}

func move_to_back(s *saucy, c *coloring, k int32) {
	cf := c.cfront[k]
	cb := cf + c.clen[cf]
	offset := s.conncnts[cf]
	s.conncnts[cf]++

	/* Move this connected label to the back of its cell */
	swap_labels(c, cb-offset, c.unlab[k])

	/* Add it to the cell list if it's the first one swapped */
	if offset == 0 {
		s.clist[s.csize] = cf
		s.csize++
	}
}

func data_mark(s *saucy, c *coloring, k int32) {
	cf := c.cfront[k]

	/* Move connects to the back of nonsingletons */
	if c.clen[cf] != 0 {
		move_to_back(s, c, k)
	}
}

func data_count(s *saucy, c *coloring, k int32) {
	cf := c.cfront[k]

	/* Move to back and count the number of connections */
	if c.clen[cf] != 0 && s.ccount[k] == 0 {
		s.ccount[k]++
		move_to_back(s, c, k)
	} else if c.clen[cf] != 0 {
		s.ccount[k]++
	}
}

func check_mapping(s *saucy, adj, edg []int32, k int32) int32 {

	/* Mark gamma of neighbors */
	for i := adj[k]; i != adj[k+1]; i++ {
		s.stuff[s.gamma[edg[i]]] = 1
	}

	/* Check neighbors of gamma */
	gk := s.gamma[k]
	for i := adj[gk]; i != adj[gk+1]; i++ {
		if s.stuff[edg[i]] == 0 {
			for j := adj[k]; j != adj[k+1]; j++ {
				s.stuff[s.gamma[edg[j]]] = 0
			}
			return 0
		}
	}

	/* Clear out bit vector before we leave */
	for i := adj[k]; i != adj[k+1]; i++ {
		s.stuff[s.gamma[edg[i]]] = 0
	}

	return 1
}

func add_conterexample(s *saucy, cex *sim_result) int32 {
	nins := baseabc.Abc_NtkPiNum(s.pNtk)
	cex.inVecSignature = 0
	for i := int32(0); i < nins; i++ {
		if cex.inVec[i] != 0 {
			cex.inVecSignature += cex.inVec[i] * i * i
			cex.inVecSignature ^= 0xABCD
		}
	}
	for _, value := range vec.Vec_PtrArray(s.satCounterExamples) {
		saved := value.(*sim_result)
		if saved.inVecSignature == cex.inVecSignature {
			return 0
		}
	}
	Vec_PtrPush(s.satCounterExamples, cex)
	bumpActivity(s, cex)
	return 1
}

func is_undirected_automorphism(s *saucy) int32 {
	for i := int32(0); i < s.ndiffs; i++ {
		if check_mapping(s, s.adj, s.edg, s.unsupp[i]) == 0 {
			return 0
		}
	}
	ret := Abc_NtkCecSat_saucy(s.pNtk, s.pNtk_permuted, s.pModel)
	if BACKTRACK_BY_SAT != 0 && ret == 0 {
		add_conterexample(s, analyzeConflict(s.pNtk, s.pModel, s.fPrintTree))
		add_conterexample(s, analyzeConflict(s.pNtk_permuted, s.pModel, s.fPrintTree))
		s.activityInc *= 1 / CLAUSE_DECAY
		if Vec_PtrSize(s.satCounterExamples) >= MAX_LEARNTS {
			reduceDB(s)
		}
	}
	return ret
}

func is_directed_automorphism(s *saucy) int32 {
	for i := int32(0); i < s.ndiffs; i++ {
		k := s.unsupp[i]
		if check_mapping(s, s.adj, s.edg, k) == 0 || check_mapping(s, s.dadj, s.dedg, k) == 0 {
			return 0
		}
	}
	return 1
}

func add_induce(s *saucy, c *coloring, who int32) {
	if c.clen[who] == 0 {
		s.sinduce[s.nsinduce] = who
		s.nsinduce++
	} else {
		s.ninduce[s.nninduce] = who
		s.nninduce++
	}
	s.indmark[who] = 1
}

func fix_fronts(c *coloring, cf, ff int32) {
	end := cf + c.clen[cf]
	for i := ff; i <= end; i++ {
		c.cfront[c.lab[i]] = cf
	}
}

func array_indirect_sort(a, b []int32) {
	n := int32(len(a))

	/* Shell sort, as implemented in nauty, (C) Brendan McKay */
	j := n / 3
	h := int32(1)
	for h < j {
		h = 3*h + 1
	}

	for h > 0 {
		for i := h; i < n; i++ {
			k := a[i]
			j = i
			for j >= h && b[a[j-h]] > b[k] {
				a[j] = a[j-h]
				j -= h
			}
			a[j] = k
		}
		h /= 3
	}
}

func at_terminal(s *saucy) bool {
	return s.nsplits == s.n
}

func add_diffnon(s *saucy, k int32) {
	/* Only add if we're in a consistent state */
	if s.ndiffnons == -1 {
		return
	}

	s.undiffnons[k] = s.ndiffnons
	s.diffnons[s.ndiffnons] = k
	s.ndiffnons++
}

func remove_diffnon(s *saucy, k int32) {
	if s.undiffnons[k] == -1 {
		return
	}

	s.ndiffnons--
	j := s.diffnons[s.ndiffnons]
	s.diffnons[s.undiffnons[k]] = j
	s.undiffnons[j] = s.undiffnons[k]

	s.undiffnons[k] = -1
}

func add_diff(s *saucy, k int32) {
	if s.diffmark[k] == 0 {
		s.diffmark[k] = 1
		s.diffs[s.ndiffs] = k
		s.ndiffs++
		add_diffnon(s, k)
	}
}

func is_a_pair(s *saucy, k int32) bool {
	return s.unpairs[k] != -1
}

func in_cell_range(c *coloring, ff, cf int32) bool {
	cb := cf + c.clen[cf]
	return cf <= ff && ff <= cb
}

func add_pair(s *saucy, k int32) {
	if s.npairs != -1 {
		s.unpairs[k] = s.npairs
		s.pairs[s.npairs] = k
		s.npairs++
	}
}

func eat_pair(s *saucy, k int32) {
	s.npairs--
	j := s.pairs[s.npairs]
	s.pairs[s.unpairs[k]] = j
	s.unpairs[j] = s.unpairs[k]
	s.unpairs[k] = -1
}

func pick_all_the_pairs(s *saucy) {
	for i := int32(0); i < s.npairs; i++ {
		s.unpairs[s.pairs[i]] = -1
	}
	s.npairs = 0
}

func clear_undiffnons(s *saucy) {
	for i := int32(0); i < s.ndiffnons; i++ {
		s.undiffnons[s.diffnons[i]] = -1
	}
}

func fix_diff_singleton(s *saucy, cf int32) {
	r := s.right.lab[cf]
	l := s.left.lab[cf]

	if s.right.clen[cf] == 0 && r != l {

		/* Make sure diff is marked */
		add_diff(s, r)

		/* It is now undiffed since it is singleton */
		s.nundiffs++
		remove_diffnon(s, r)

		/* Mark the other if not singleton already */
		rcfl := s.right.cfront[l]
		if s.right.clen[rcfl] != 0 {
			add_diff(s, l)

			/* Check for pairs */
			if in_cell_range(&s.right, s.left.unlab[r], rcfl) {
				add_pair(s, l)
			}

			/* Otherwise we might be eating a pair */
		} else if is_a_pair(s, r) {
			eat_pair(s, r)
		}
	}
}

func fix_diff_subtract(s *saucy, cf int32, a, b []int32) {
	cb := cf + s.right.clen[cf]

	/* Mark the contents of the first set */
	for i := cf; i <= cb; i++ {
		s.stuff[a[i]] = 1
	}

	/* Add elements from second set not present in the first */
	for i := cf; i <= cb; i++ {
		k := b[i]
		if s.stuff[k] == 0 {
			add_diff(s, k)
		}
	}

	/* Clear the marks of the first set */
	for i := cf; i <= cb; i++ {
		s.stuff[a[i]] = 0
	}
}

func fix_diffs(s *saucy, cf, ff int32) {

	/* Check for singleton cases in both cells */
	fix_diff_singleton(s, cf)
	fix_diff_singleton(s, ff)

	/* If they're both nonsingleton, do subtraction on smaller */
	if s.right.clen[cf] != 0 && s.right.clen[ff] != 0 {
		min := ff
		if s.right.clen[cf] < s.right.clen[ff] {
			min = cf
		}
		fix_diff_subtract(s, min, s.left.lab, s.right.lab)
		fix_diff_subtract(s, min, s.right.lab, s.left.lab)
	}
}

func split_color(c *coloring, cf, ff int32) {
	/* Fix lengths */
	fb := ff - 1
	cb := cf + c.clen[cf]
	c.clen[cf] = fb - cf
	c.clen[ff] = cb - ff

	/* Fix cell front pointers */
	fix_fronts(c, ff, ff)
}

func split_common(s *saucy, c *coloring, cf, ff int32) {
	split_color(c, cf, ff)

	/* Add to refinement */
	if s.indmark[cf] != 0 || c.clen[ff] < c.clen[cf] {
		add_induce(s, c, ff)
	} else {
		add_induce(s, c, cf)
	}
}

func split_left(s *saucy, c *coloring, cf, ff int32) int32 {
	/* Record the split */
	s.splitwho[s.nsplits] = ff
	s.splitfrom[s.nsplits] = cf
	s.nsplits++

	/* Do common splitting tasks */
	split_common(s, c, cf, ff)

	/* Always succeeds */
	return 1
}

func split_init(s *saucy, c *coloring, cf, ff int32) int32 {
	split_left(s, c, cf, ff)

	/* Maintain nonsingleton list for finding new targets */
	if c.clen[ff] != 0 {
		s.prevnon[s.nextnon[cf+1]] = ff
		s.nextnon[ff+1] = s.nextnon[cf+1]
		s.prevnon[ff] = cf
		s.nextnon[cf+1] = ff
	}
	if c.clen[cf] == 0 {
		s.nextnon[s.prevnon[cf]+1] = s.nextnon[cf+1]
		s.prevnon[s.nextnon[cf+1]] = s.prevnon[cf]
	}

	/* Always succeeds */
	return 1
}

func split_other(s *saucy, c *coloring, cf, ff int32) int32 {
	k := s.nsplits

	/* Verify the split with init */
	if s.splitwho[k] != ff || s.splitfrom[k] != cf || k >= s.splitlev[s.lev] {
		return 0
	}
	s.nsplits++

	/* Do common splitting tasks */
	split_common(s, c, cf, ff)

	/* Fix differences with init */
	fix_diffs(s, cf, ff)

	/* If we got this far we succeeded */
	return 1
}

func print_partition(left, right *coloring, n int32, pNtk *Abc_Ntk_t, fNames bool) int32 {
	fmt.Printf("top = |")
	for i := int32(0); i < n; i += left.clen[i] + 1 {
		for j := int32(0); j < left.clen[i]+1; j++ {
			if fNames {
				fmt.Printf("%s ", getVertexName(pNtk, left.lab[i+j]))
			} else {
				fmt.Printf("%d ", left.lab[i+j])
			}
		}
		if i+left.clen[i]+1 < n {
			fmt.Printf("|")
		}
	}
	fmt.Printf("|\n")

	/*printf("(cfront = {");
	  for (i = 0; i < n; i++)
	      printf("%d ", left->cfront[i]);
	  printf("})\n");*/

	if right == nil {
		return 1
	}

	fmt.Printf("bot = |")
	for i := int32(0); i < n; i += right.clen[i] + 1 {
		for j := int32(0); j < right.clen[i]+1; j++ {
			if fNames {
				fmt.Printf("%s ", getVertexName(pNtk, right.lab[i+j]))
			} else {
				fmt.Printf("%d ", right.lab[i+j])
			}
		}
		if i+right.clen[i]+1 < n {
			fmt.Printf("|")
		}
	}
	fmt.Printf("|\n")

	/*printf("(cfront = {");
	  for (i = 0; i < n; i++)
	      printf("%d ", right->cfront[i]);
	  printf("})\n");*/

	return 1
}

func refine_cell(s *saucy, c *coloring, refine func(*saucy, *coloring, int32) int32) int32 {
	if s.lev > 1 {
		introsort(s.clist[:s.csize])
	}
	ret := int32(1)
	for i := int32(0); ret != 0 && i < s.csize; i++ {
		cf := s.clist[i]
		ret = refine(s, c, cf)
	}
	for i := int32(0); i < s.csize; i++ {
		s.conncnts[s.clist[i]] = 0
	}
	s.csize = 0
	return ret
}
func maybe_split(s *saucy, c *coloring, cf, ff int32) int32 {
	if cf == ff {
		return 1
	}
	return s.split(s, c, cf, ff)
}
func ref_single_cell(s *saucy, c *coloring, cf int32) int32 {
	zcnt := c.clen[cf] + 1 - s.conncnts[cf]
	return maybe_split(s, c, cf, cf+zcnt)
}
func ref_singleton(s *saucy, c *coloring, adj, edg []int32, cf int32) int32 {
	k := c.lab[cf]
	for i := adj[k]; i < adj[k+1]; i++ {
		data_mark(s, c, edg[i])
	}
	return refine_cell(s, c, ref_single_cell)
}
func ref_singleton_directed(s *saucy, c *coloring, cf int32) int32 {
	if ref_singleton(s, c, s.adj, s.edg, cf) == 0 {
		return 0
	}
	return ref_singleton(s, c, s.dadj, s.dedg, cf)
}
func ref_singleton_undirected(s *saucy, c *coloring, cf int32) int32 {
	return ref_singleton(s, c, s.adj, s.edg, cf)
}
func ref_nonsingle_cell(s *saucy, c *coloring, cf int32) int32 {
	cb := cf + c.clen[cf]
	nzf := cb - s.conncnts[cf] + 1
	ff := nzf
	cnt := s.ccount[c.lab[ff]]
	s.count[ff] = cnt
	bmin, bmax := cnt, cnt
	s.bucket[cnt] = 1
	for ff++; ff <= cb; ff++ {
		cnt = s.ccount[c.lab[ff]]
		for bmin > cnt {
			bmin--
			s.bucket[bmin] = 0
		}
		for bmax < cnt {
			bmax++
			s.bucket[bmax] = 0
		}
		s.bucket[cnt]++
		s.count[ff] = cnt
	}
	if bmin == bmax && cf == nzf {
		return 1
	}
	ff = nzf
	fb := nzf
	for i := bmin; i <= bmax; i++ {
		ff = fb
		if s.bucket[i] == 0 {
			continue
		}
		fb = ff + s.bucket[i]
		s.bucket[i] = fb
	}
	for i := nzf; i <= cb; i++ {
		b := s.count[i]
		s.bucket[b]--
		s.junk[s.bucket[b]] = c.lab[i]
	}
	for i := nzf; i <= cb; i++ {
		set_label(c, i, s.junk[i])
	}
	for i := bmax; i > bmin; i-- {
		ff = s.bucket[i]
		if ff != 0 && s.split(s, c, cf, ff) == 0 {
			return 0
		}
	}
	return maybe_split(s, c, cf, s.bucket[bmin])
}
func ref_nonsingle(s *saucy, c *coloring, adj, edg []int32, cf int32) int32 {
	cb := cf + c.clen[cf]
	size := cb - cf + 1
	if cf == cb {
		return ref_singleton(s, c, adj, edg, cf)
	}
	copy(s.junk[:size], c.lab[cf:cb+1])
	for i := int32(0); i < size; i++ {
		k := s.junk[i]
		for j := adj[k]; j < adj[k+1]; j++ {
			data_count(s, c, edg[j])
		}
	}
	ret := refine_cell(s, c, ref_nonsingle_cell)
	for i := cf; i <= cb; i++ {
		k := c.lab[i]
		for j := adj[k]; j < adj[k+1]; j++ {
			s.ccount[edg[j]] = 0
		}
	}
	return ret
}
func ref_nonsingle_directed(s *saucy, c *coloring, cf int32) int32 {
	if ref_nonsingle(s, c, s.adj, s.edg, cf) == 0 {
		return 0
	}
	return ref_nonsingle(s, c, s.dadj, s.dedg, cf)
}
func ref_nonsingle_undirected(s *saucy, c *coloring, cf int32) int32 {
	return ref_nonsingle(s, c, s.adj, s.edg, cf)
}
func clear_refine(s *saucy) {
	for i := int32(0); i < s.nninduce; i++ {
		s.indmark[s.ninduce[i]] = 0
	}
	for i := int32(0); i < s.nsinduce; i++ {
		s.indmark[s.sinduce[i]] = 0
	}
	s.nninduce = 0
	s.nsinduce = 0
}
func refine(s *saucy, c *coloring) int32 {
	for {
		if at_terminal(s) {
			clear_refine(s)
			return 1
		}
		if s.nsinduce > 0 {
			front := s.sinduce[s.nsinduce-1]
			s.nsinduce--
			s.indmark[front] = 0
			if s.ref_singleton(s, c, front) == 0 {
				break
			}
		} else if s.nninduce > 0 {
			front := s.ninduce[s.nninduce-1]
			s.nninduce--
			s.indmark[front] = 0
			if s.ref_nonsingle(s, c, front) == 0 {
				break
			}
		} else {
			return 1
		}
	}
	clear_refine(s)
	return 0
}

func refineByDepGraph(s *saucy, c *coloring) int32 {
	s.adj = s.depAdj
	s.edg = s.depEdg
	return refine(s, c)
}
func backtrackBysatCounterExamples(s *saucy, c *coloring) int32 {
	if c == &s.left || Vec_PtrSize(s.satCounterExamples) == 0 {
		return 1
	}
	flag := make([]int32, Vec_PtrSize(s.satCounterExamples))
	for i := int32(0); i < Vec_PtrSize(s.satCounterExamples); i++ {
		cex1 := Vec_PtrEntry(s.satCounterExamples, i).(*sim_result)
		for j := int32(0); j < Vec_PtrSize(s.satCounterExamples); j++ {
			if flag[j] != 0 {
				continue
			}
			cex2 := Vec_PtrEntry(s.satCounterExamples, j).(*sim_result)
			res := ifInputVectorsAreConsistent(s, cex1.inVec, cex2.inVec)
			if res == -2 {
				flag[j] = 1
				continue
			}
			if res == -1 {
				break
			}
			if res == 0 {
				continue
			}
			if cex1.outVecOnes != cex2.outVecOnes {
				bumpActivity(s, cex1)
				bumpActivity(s, cex2)
				return 0
			}
			if baseabc.Abc_NtkPoNum(s.pNtk) == 1 {
				continue
			}
			if ifOutputVectorsAreConsistent(s, cex1.outVec, cex2.outVec) == 0 {
				bumpActivity(s, cex1)
				bumpActivity(s, cex2)
				return 0
			}
		}
	}
	return 1
}

func allOutputsDistinguished(c *coloring, n int32) bool {
	for j := int32(0); j < n; j++ {
		if c.clen[j] != 0 {
			return false
		}
	}
	return true
}
func induceAll(s *saucy, c *coloring) {
	for j := int32(0); j < s.n; j += c.clen[j] + 1 {
		add_induce(s, c, j)
	}
}
func refineBySim1_init(s *saucy, c *coloring) int32 {
	if baseabc.Abc_NtkPoNum(s.pNtk) == 1 {
		return 1
	}
	for i := int32(0); i < NUM_SIM1_ITERATION; i++ {
		if allOutputsDistinguished(c, baseabc.Abc_NtkPoNum(s.pNtk)) {
			break
		}
		randVec := assignRandomBitsToCells(s.pNtk, c)
		g := buildSim1Graph(s.pNtk, c, randVec, s.iDep, s.oDep)
		assert(g != nil, "g != nil")
		s.adj = g.adj
		s.edg = g.edg
		nsplits := s.nsplits
		induceAll(s, c)
		refine(s, c)
		if s.nsplits > nsplits {
			i = 0
			induceAll(s, c)
			refineByDepGraph(s, c)
		}
		Vec_IntFree(randVec)
		g.adj = nil
		g.edg = nil
	}
	return 1
}
func refineBySim1_left(s *saucy, c *coloring) int32 {
	if baseabc.Abc_NtkPoNum(s.pNtk) == 1 {
		return 1
	}
	for i := int32(0); i < NUM_SIM1_ITERATION; i++ {
		if allOutputsDistinguished(c, baseabc.Abc_NtkPoNum(s.pNtk)) {
			break
		}
		randVec := assignRandomBitsToCells(s.pNtk, c)
		g := buildSim1Graph(s.pNtk, c, randVec, s.iDep, s.oDep)
		assert(g != nil, "g != nil")
		s.adj = g.adj
		s.edg = g.edg
		nsplits := s.nsplits
		induceAll(s, c)
		refine(s, c)
		if s.nsplits > nsplits {
			Vec_PtrPush(s.randomVectorArray_sim1, randVec)
			i = 0
			induceAll(s, c)
			refineByDepGraph(s, c)
		} else {
			Vec_IntFree(randVec)
		}
		g.adj = nil
		g.edg = nil
	}
	s.randomVectorSplit_sim1[s.lev] = Vec_PtrSize(s.randomVectorArray_sim1)
	return 1
}
func refineBySim1_other(s *saucy, c *coloring) int32 {
	for i := s.randomVectorSplit_sim1[s.lev-1]; i < s.randomVectorSplit_sim1[s.lev]; i++ {
		randVec := Vec_PtrEntry(s.randomVectorArray_sim1, i).(*Vec_Int_t)
		g := buildSim1Graph(s.pNtk, c, randVec, s.iDep, s.oDep)
		if g == nil {
			assert(c == &s.right, "c == &s.right")
			return 0
		}
		s.adj = g.adj
		s.edg = g.edg
		nsplits := s.nsplits
		induceAll(s, c)
		ret := refine(s, c)
		if s.nsplits == nsplits {
			assert(c == &s.right, "c == &s.right")
			ret = 0
		}
		if ret != 0 {
			induceAll(s, c)
			ret = refineByDepGraph(s, c)
		}
		g.adj = nil
		g.edg = nil
		if ret == 0 {
			return 0
		}
	}
	return 1
}
func refineBySim2_init(s *saucy, c *coloring) int32 {
	for i := int32(0); i < NUM_SIM2_ITERATION; i++ {
		randVec := assignRandomBitsToCells(s.pNtk, c)
		g := buildSim2Graph(s.pNtk, c, randVec, s.iDep, s.oDep, s.topOrder, s.obs, s.ctrl)
		assert(g != nil, "g != nil")
		s.adj = g.adj
		s.edg = g.edg
		nsplits := s.nsplits
		induceAll(s, c)
		refine(s, c)
		if s.nsplits > nsplits {
			i = 0
			induceAll(s, c)
			refineByDepGraph(s, c)
		}
		Vec_IntFree(randVec)
		g.adj = nil
		g.edg = nil
	}
	return 1
}
func refineBySim2_left(s *saucy, c *coloring) int32 {
	for i := int32(0); i < NUM_SIM2_ITERATION; i++ {
		randVec := assignRandomBitsToCells(s.pNtk, c)
		g := buildSim2Graph(s.pNtk, c, randVec, s.iDep, s.oDep, s.topOrder, s.obs, s.ctrl)
		assert(g != nil, "g != nil")
		s.adj = g.adj
		s.edg = g.edg
		nsplits := s.nsplits
		induceAll(s, c)
		refine(s, c)
		if s.nsplits > nsplits {
			Vec_PtrPush(s.randomVectorArray_sim2, randVec)
			i = 0
			induceAll(s, c)
			refineByDepGraph(s, c)
		} else {
			Vec_IntFree(randVec)
		}
		g.adj = nil
		g.edg = nil
	}
	s.randomVectorSplit_sim2[s.lev] = Vec_PtrSize(s.randomVectorArray_sim2)
	return 1
}
func refineBySim2_other(s *saucy, c *coloring) int32 {
	for i := s.randomVectorSplit_sim2[s.lev-1]; i < s.randomVectorSplit_sim2[s.lev]; i++ {
		randVec := Vec_PtrEntry(s.randomVectorArray_sim2, i).(*Vec_Int_t)
		g := buildSim2Graph(s.pNtk, c, randVec, s.iDep, s.oDep, s.topOrder, s.obs, s.ctrl)
		if g == nil {
			assert(c == &s.right, "c == &s.right")
			return 0
		}
		s.adj = g.adj
		s.edg = g.edg
		nsplits := s.nsplits
		induceAll(s, c)
		ret := refine(s, c)
		if s.nsplits == nsplits {
			assert(c == &s.right, "c == &s.right")
			ret = 0
		}
		if ret != 0 {
			induceAll(s, c)
			ret = refineByDepGraph(s, c)
		}
		g.adj = nil
		g.edg = nil
		if ret == 0 {
			return 0
		}
	}
	return 1
}

func check_OPP_only_has_swaps(s *saucy, c *coloring) int32 {
	if c == &s.left {
		return 1
	}
	left, right := Vec_IntAlloc(1), Vec_IntAlloc(1)
	defer Vec_IntFree(left)
	defer Vec_IntFree(right)
	for cell := int32(0); cell < s.n; cell += s.left.clen[cell] + 1 {
		for j := cell; j <= cell+s.left.clen[cell]; j++ {
			Vec_IntPush(left, s.left.cfront[s.right.lab[j]])
			Vec_IntPush(right, s.right.cfront[s.left.lab[j]])
		}
		vec.Vec_IntSortUnsigned(left)
		vec.Vec_IntSortUnsigned(right)
		for j := int32(0); j < Vec_IntSize(left); j++ {
			if Vec_IntEntry(left, j) != Vec_IntEntry(right, j) {
				return 0
			}
		}
		vec.Vec_IntClear(left)
		vec.Vec_IntClear(right)
	}
	return 1
}
func check_OPP_for_Boolean_matching(s *saucy, c *coloring) int32 {
	if c == &s.left {
		return 1
	}
	for cell := int32(0); cell < s.n; cell += s.right.clen[cell] + 1 {
		var n1l, n2l, n1r, n2r int32
		for j := cell; j <= cell+s.right.clen[cell]; j++ {
			a, b := getVertexName(s.pNtk, s.left.lab[j]), getVertexName(s.pNtk, s.right.lab[j])
			assert(len(a) > 2 && a[0] == 'N' && a[2] == ':', "left node name format")
			assert(len(b) > 2 && b[0] == 'N' && b[2] == ':', "right node name format")
			if a[1] == '1' {
				n1l++
			} else {
				assert(a[1] == '2', "left network prefix")
				n2l++
			}
			if b[1] == '1' {
				n1r++
			} else {
				assert(b[1] == '2', "right network prefix")
				n2r++
			}
		}
		if n1l != n2r || n2l != n1r {
			return 0
		}
	}
	return 1
}

func double_check_OPP_isomorphism(s *saucy, c *coloring) int32 {
	if c == &s.left {
		return 1
	}
	for i := s.nsplits - 1; i > s.splitlev[s.lev-1]; i-- {
		v := c.lab[s.splitwho[i]]
		var sum1, xor1 int32
		for j := s.adj[v]; j < s.adj[v+1]; j++ {
			sum1 += c.cfront[s.edg[j]]
			xor1 ^= c.cfront[s.edg[j]]
		}
		v = s.left.lab[s.splitwho[i]]
		var sum2, xor2 int32
		for j := s.adj[v]; j < s.adj[v+1]; j++ {
			sum2 += s.left.cfront[s.edg[j]]
			xor2 ^= s.left.cfront[s.edg[j]]
		}
		if sum1 != sum2 || xor1 != xor2 {
			return 0
		}
		v = c.lab[s.splitfrom[i]]
		sum1, xor1 = 0, 0
		for j := s.adj[v]; j < s.adj[v+1]; j++ {
			sum1 += c.cfront[s.edg[j]]
			xor1 ^= c.cfront[s.edg[j]]
		}
		v = s.left.lab[s.splitfrom[i]]
		sum2, xor2 = 0, 0
		for j := s.adj[v]; j < s.adj[v+1]; j++ {
			sum2 += s.left.cfront[s.edg[j]]
			xor2 ^= s.left.cfront[s.edg[j]]
		}
		if sum1 != sum2 || xor1 != xor2 {
			return 0
		}
	}
	return 1
}
func descend(s *saucy, c *coloring, target, min int32) int32 {
	back := target + c.clen[target]
	s.stats.nodes++
	swap_labels(c, min, back)
	s.difflev[s.lev] = s.ndiffs
	s.undifflev[s.lev] = s.nundiffs
	s.lev++
	s.split(s, c, target, back)
	if refineByDepGraph(s, c) == 0 {
		return 0
	}
	if s.fBooleanMatching != 0 && check_OPP_for_Boolean_matching(s, c) == 0 {
		return 0
	}
	if REFINE_BY_SIM_1 != 0 && s.refineBySim1(s, c) == 0 {
		return 0
	}
	if REFINE_BY_SIM_2 != 0 && s.refineBySim2(s, c) == 0 {
		return 0
	}
	if s.fBooleanMatching != 0 && check_OPP_for_Boolean_matching(s, c) == 0 {
		return 0
	}
	if s.fLookForSwaps != 0 && check_OPP_only_has_swaps(s, c) == 0 {
		return 0
	}
	if double_check_OPP_isomorphism(s, c) == 0 {
		return 0
	}
	return 1
}

func select_smallest_max_connected_cell(s *saucy, start, end int32) int32 {
	smallest, maxconn := int32(-1), int32(-1)
	smallSize := s.n
	connections := make([]byte, s.n)
	cell := start
	for cell < end && s.left.clen[cell] == 0 {
		cell++
	}
	for cell < end {
		if s.left.clen[cell] <= smallSize {
			var count int32
			v := s.left.lab[cell]
			for i := s.depAdj[v]; i < s.depAdj[v+1]; i++ {
				x := s.depEdg[i]
				if connections[x] == 0 {
					count++
					connections[x] = 1
				}
			}
			if s.left.clen[cell] < smallSize || count > maxconn {
				smallSize = s.left.clen[cell]
				maxconn = count
				smallest = cell
			}
			for i := s.depAdj[v]; i < s.depAdj[v+1]; i++ {
				connections[s.depEdg[i]] = 0
			}
		}
		cell = s.nextnon[cell+1]
	}
	return smallest
}
func descend_leftmost(s *saucy) int32 {
	for !at_terminal(s) {
		sentinel := s.nextnon[0]
		var target, min int32
		if sentinel < baseabc.Abc_NtkPoNum(s.pNtk) {
			target = select_smallest_max_connected_cell(s, sentinel, baseabc.Abc_NtkPoNum(s.pNtk))
			min = target
		} else {
			target = select_smallest_max_connected_cell(s, baseabc.Abc_NtkPoNum(s.pNtk), s.n)
			min = target
		}
		if s.fPrintTree != 0 {
			fmt.Printf("%s->%s\n", getVertexName(s.pNtk, s.left.lab[min]), getVertexName(s.pNtk, s.left.lab[min]))
		}
		s.splitvar[s.lev] = s.left.lab[min]
		s.start[s.lev] = target
		s.splitlev[s.lev] = s.nsplits
		if descend(s, &s.left, target, min) == 0 {
			return 0
		}
	}
	s.splitlev[s.lev] = s.n
	return 1
}
func zeta_fixed(s *saucy) int32 {
	if s.ndiffs == s.nundiffs {
		return 1
	}
	return 0
}

func select_dynamically(s *saucy, target, lmin, rmin *int32) {
	for i := int32(0); i < s.npairs; i++ {
		k := s.pairs[i]
		*target = s.right.cfront[k]
		*lmin = s.left.unlab[s.right.lab[s.left.unlab[k]]]
		*rmin = s.right.unlab[k]
		if s.left.clen[*target] != 0 && in_cell_range(&s.left, *lmin, *target) && in_cell_range(&s.right, *rmin, *target) {
			return
		}
	}
	for i := int32(0); i < s.n; i += s.left.clen[i] + 1 {
		if s.left.clen[i] == 0 {
			continue
		}
		*target = i
		*lmin = i
		*rmin = i
		if s.right.cfront[s.left.lab[*lmin]] == *target {
			*rmin = s.right.unlab[s.left.lab[*lmin]]
		}
		return
	}
	panic("saucy: no target cell")
}
func select_statically(s *saucy, target, lmin, rmin *int32) {
	*target = s.left.cfront[s.splitvar[s.lev]]
	*rmin = *target
	*lmin = s.left.unlab[s.splitvar[s.lev]]
	for i := *rmin; i <= *rmin+s.right.clen[*target]; i++ {
		if s.right.lab[i] == s.left.lab[*lmin] {
			*rmin = i
			break
		}
	}
}
func descend_left(s *saucy) int32 {
	if s.nsplits != s.splitlev[s.lev] {
		return 0
	}
	for !at_terminal(s) {
		var target, lmin, rmin int32
		s.select_decomposition(s, &target, &lmin, &rmin)
		if s.fPrintTree != 0 {
			fmt.Printf("in level %d: %s->%s\n", s.lev, getVertexName(s.pNtk, s.left.lab[lmin]), getVertexName(s.pNtk, s.right.lab[rmin]))
		}
		s.match = 0
		s.start[s.lev] = target
		s.split = split_left
		if SELECT_DYNAMICALLY != 0 {
			s.refineBySim1 = refineBySim1_left
			s.refineBySim2 = refineBySim2_left
		}
		descend(s, &s.left, target, lmin)
		s.splitlev[s.lev] = s.nsplits
		s.split = split_other
		if SELECT_DYNAMICALLY != 0 {
			s.refineBySim1 = refineBySim1_other
			s.refineBySim2 = refineBySim2_other
		}
		s.lev--
		s.nsplits = s.splitlev[s.lev]
		s.specmin[s.lev] = s.right.lab[rmin]
		if descend(s, &s.right, target, rmin) == 0 {
			return 0
		}
		if s.nsplits != s.splitlev[s.lev] {
			return 0
		}
	}
	return 1
}

func find_representative(k int32, theta []int32) int32 {
	rep := k
	for rep != theta[rep] {
		rep = theta[rep]
	}
	for theta[k] != rep {
		tmp := theta[k]
		theta[k] = rep
		k = tmp
	}
	return rep
}
func update_theta(s *saucy) {
	for i := int32(0); i < s.ndiffs; i++ {
		k := s.unsupp[i]
		x := find_representative(k, s.theta)
		y := find_representative(s.gamma[k], s.theta)
		if x == y {
			continue
		}
		if x > y {
			x, y = y, x
		}
		s.theta[y] = x
		s.thsize[x] += s.thsize[y]
		s.thnext[s.thprev[y]] = s.thnext[y]
		s.thprev[s.thnext[y]] = s.thprev[y]
		s.threp[s.thfront[y]] = s.thnext[y]
	}
}
func theta_prune(s *saucy) int32 {
	start := s.start[s.lev]
	irep := find_representative(s.indmin, s.theta)
	for s.kanctar > 0 {
		s.kanctar--
		label := s.anctar[s.kanctar]
		rep := find_representative(label, s.theta)
		if rep == label && rep != irep {
			return s.right.unlab[label] - start
		}
	}
	return -1
}
func orbit_prune(s *saucy) int32 {
	k := s.start[s.lev]
	size := s.right.clen[k] + 1
	cell := s.right.lab[k : k+size]
	fixed := cell[size-1]
	min := int32(-1)
	for i := int32(0); i < size-1; i++ {
		label := cell[i]
		if label <= fixed {
			continue
		}
		if min != -1 && label > cell[min] {
			continue
		}
		min = i
	}
	return min
}

func note_anctar_reps(s *saucy) {
	for i := s.splitlev[s.anc+1] - 1; i >= s.splitlev[s.anc]; i-- {
		f := s.splitfrom[i]
		j := s.threp[f]
		k := s.threp[s.splitwho[i]]
		s.thnext[s.thprev[j]] = k
		s.thnext[s.thprev[k]] = j
		tmp := s.thprev[j]
		s.thprev[j] = s.thprev[k]
		s.thprev[k] = tmp
		for m := k; m != j; m = s.thnext[m] {
			s.thfront[m] = f
		}
	}
	s.kanctar = 0
	rep := s.threp[s.start[s.lev]]
	s.anctar[s.kanctar] = rep
	s.kanctar++
	for k := s.thnext[rep]; k != rep; k = s.thnext[k] {
		s.anctar[s.kanctar] = k
		s.kanctar++
	}
	array_indirect_sort(s.anctar[:s.kanctar], s.thsize)
}
func multiply_index(s *saucy, k int32) {
	s.stats.grpsize_base *= float64(k)
	if s.stats.grpsize_base > 1e10 {
		s.stats.grpsize_base /= 1e10
		s.stats.grpsize_exp += 10
	}
}
func backtrack_leftmost(s *saucy) int32 {
	rep := find_representative(s.indmin, s.theta)
	repsize := s.thsize[rep]
	min := int32(-1)
	pick_all_the_pairs(s)
	clear_undiffnons(s)
	s.ndiffs = 0
	s.nundiffs = 0
	s.npairs = 0
	s.ndiffnons = 0
	if repsize != s.right.clen[s.start[s.lev]]+1 {
		min = theta_prune(s)
	}
	if min == -1 {
		multiply_index(s, repsize)
	}
	return min
}
func backtrack_other(s *saucy) int32 {
	cf := s.start[s.lev]
	cb := cf + s.right.clen[cf]
	spec := s.specmin[s.lev]
	pick_all_the_pairs(s)
	clear_undiffnons(s)
	s.npairs = -1
	s.ndiffnons = -1
	var min int32
	if s.right.lab[cb] == spec {
		min = find_min(s, cf)
		if min == cb {
			min = orbit_prune(s)
		} else {
			min -= cf
		}
	} else {
		min = orbit_prune(s)
		if min != -1 && s.right.lab[min+cf] == spec {
			swap_labels(&s.right, min+cf, cb)
			min = orbit_prune(s)
		}
	}
	return min
}
func rewind_coloring(s *saucy, c *coloring, lev int32) {
	splits := s.splitlev[lev]
	for i := s.nsplits - 1; i >= splits; i-- {
		cf := s.splitfrom[i]
		ff := s.splitwho[i]
		c.clen[cf] += c.clen[ff] + 1
		fix_fronts(c, cf, ff)
	}
}

func rewind_simulation_vectors(s *saucy, lev int32) {
	for i := s.randomVectorSplit_sim1[lev]; i < Vec_PtrSize(s.randomVectorArray_sim1); i++ {
		Vec_IntFree(Vec_PtrEntry(s.randomVectorArray_sim1, i).(*Vec_Int_t))
	}
	vec.Vec_PtrShrink(s.randomVectorArray_sim1, int(s.randomVectorSplit_sim1[lev]))
	for i := s.randomVectorSplit_sim2[lev]; i < Vec_PtrSize(s.randomVectorArray_sim2); i++ {
		Vec_IntFree(Vec_PtrEntry(s.randomVectorArray_sim2, i).(*Vec_Int_t))
	}
	vec.Vec_PtrShrink(s.randomVectorArray_sim2, int(s.randomVectorSplit_sim2[lev]))
}
func do_backtrack(s *saucy) int32 {
	rewind_coloring(s, &s.right, s.lev)
	s.nsplits = s.splitlev[s.lev]
	for i := s.ndiffs - 1; i >= s.difflev[s.lev]; i-- {
		s.diffmark[s.diffs[i]] = 0
	}
	s.ndiffs = s.difflev[s.lev]
	s.nundiffs = s.undifflev[s.lev]
	cf := s.start[s.lev]
	cb := cf + s.right.clen[cf]
	if s.anc > s.lev {
		s.anc = s.lev
		s.indmin = s.left.lab[cb]
		s.match = 1
		note_anctar_reps(s)
	}
	if s.lev == s.anc {
		return backtrack_leftmost(s)
	}
	return backtrack_other(s)
}
func backtrack_loop(s *saucy) int32 {
	for s.lev--; s.lev != 0; s.lev-- {
		min := do_backtrack(s)
		if min != -1 {
			return min + s.start[s.lev]
		}
	}
	return -1
}
func backtrack(s *saucy) int32 {
	old := s.nsplits
	min := backtrack_loop(s)
	tmp := s.nsplits
	s.nsplits = old
	rewind_coloring(s, &s.left, s.lev+1)
	s.nsplits = tmp
	if SELECT_DYNAMICALLY != 0 {
		rewind_simulation_vectors(s, s.lev+1)
	}
	return min
}
func backtrack_bad(s *saucy) int32 {
	old := s.lev
	min := backtrack_loop(s)
	if BACKTRACK_BY_SAT != 0 {
		oldLev := s.lev
		for backtrackBysatCounterExamples(s, &s.right) == 0 {
			min = backtrack_loop(s)
			if s.lev == 0 {
				if s.fPrintTree != 0 {
					fmt.Printf("Backtrack by SAT from level %d to %d\n", oldLev, 0)
				}
				return -1
			}
		}
		if s.fPrintTree != 0 && s.lev < oldLev {
			fmt.Printf("Backtrack by SAT from level %d to %d\n", oldLev, s.lev)
		}
	}
	tmp := s.nsplits
	s.nsplits = s.splitlev[old]
	rewind_coloring(s, &s.left, s.lev+1)
	s.nsplits = tmp
	if SELECT_DYNAMICALLY != 0 {
		rewind_simulation_vectors(s, s.lev+1)
	}
	return min
}

func prepare_permutation_ntk(s *saucy) {
	numouts := baseabc.Abc_NtkPoNum(s.pNtk)
	baseabc.Abc_NtkResetNames(s.pNtk_permuted)
	for i := int32(0); i < s.n; i++ {
		var src, dst *Abc_Obj_t
		if i < numouts {
			src = baseabc.Abc_NtkPo(s.pNtk, i)
			dst = baseabc.Abc_NtkPo(s.pNtk_permuted, s.gamma[i])
		} else {
			src = baseabc.Abc_NtkPi(s.pNtk, i-numouts)
			dst = baseabc.Abc_NtkPi(s.pNtk_permuted, s.gamma[i]-numouts)
		}
		baseabc.Abc_ObjAssignName(dst, src.Name(), "")
	}
	baseabc.Abc_NtkOrderObjsByName(s.pNtk_permuted, 1)
}
func prepare_permutation(s *saucy) {
	for i := int32(0); i < s.ndiffs; i++ {
		k := s.right.unlab[s.diffs[i]]
		s.unsupp[i] = s.left.lab[k]
		s.gamma[s.left.lab[k]] = s.right.lab[k]
	}
	prepare_permutation_ntk(s)
}
func unprepare_permutation_ntk(s *saucy) {
	numouts := baseabc.Abc_NtkPoNum(s.pNtk)
	baseabc.Abc_NtkResetNames(s.pNtk_permuted)
	for i := int32(0); i < s.n; i++ {
		var src, dst *Abc_Obj_t
		if i < numouts {
			src = baseabc.Abc_NtkPo(s.pNtk, s.gamma[i])
			dst = baseabc.Abc_NtkPo(s.pNtk_permuted, i)
		} else {
			src = baseabc.Abc_NtkPi(s.pNtk, s.gamma[i]-numouts)
			dst = baseabc.Abc_NtkPi(s.pNtk_permuted, i-numouts)
		}
		baseabc.Abc_ObjAssignName(dst, src.Name(), "")
	}
	baseabc.Abc_NtkOrderObjsByName(s.pNtk_permuted, 1)
}
func unprepare_permutation(s *saucy) {
	unprepare_permutation_ntk(s)
	for i := int32(0); i < s.ndiffs; i++ {
		s.gamma[s.unsupp[i]] = s.unsupp[i]
	}
}
func do_search(s *saucy) int32 {
	unprepare_permutation(s)
	if s.lev > s.anc {
		s.lev = s.anc + 1
	}
	min := backtrack(s)
	if s.fBooleanMatching != 0 && (s.stats.grpsize_base > 1 || s.stats.grpsize_exp > 0) {
		return 0
	}
	if s.fPrintTree != 0 && s.lev > 0 {
		fmt.Printf("in level %d: %s->%s\n", s.lev, getVertexName(s.pNtk, s.left.lab[s.splitwho[s.nsplits]]), getVertexName(s.pNtk, s.right.lab[min]))
	}
	for s.lev > 0 {
		if descend(s, &s.right, s.start[s.lev], min) != 0 && descend_left(s) != 0 {
			prepare_permutation(s)
			if s.is_automorphism(s) != 0 {
				s.stats.gens++
				s.stats.support += s.ndiffs
				s.stats.nodes += 0
				update_theta(s)
				s.print_automorphism(s.gFile, s.n, s.gamma, s.ndiffs, s.unsupp, s.marks, s.pNtk)
				unprepare_permutation(s)
				return 1
			}
			unprepare_permutation(s)
		}
		s.stats.bads++
		min = backtrack_bad(s)
		if s.fPrintTree != 0 {
			fmt.Println("BAD NODE")
			if s.lev > 0 {
				fmt.Printf("in level %d: %s->%s\n", s.lev, getVertexName(s.pNtk, s.left.lab[s.splitwho[s.nsplits]]), getVertexName(s.pNtk, s.right.lab[min]))
			}
		}
	}
	for s.stats.grpsize_base >= 10 {
		s.stats.grpsize_base /= 10
		s.stats.grpsize_exp++
	}
	return 0
}

func saucy_search(pNtk *Abc_Ntk_t, s *saucy, directed int32, colors []int32, stats *saucy_stats) {
	s.pNtk = pNtk
	s.pNtk_permuted = baseabc.Abc_NtkDup(pNtk)
	g := buildDepGraph(pNtk, s.iDep, s.oDep)
	s.n = g.n
	s.depAdj = g.adj
	s.depEdg = g.edg
	s.stats = stats
	if directed != 0 {
		s.is_automorphism = is_directed_automorphism
		s.ref_singleton = ref_singleton_directed
		s.ref_nonsingle = ref_nonsingle_directed
	} else {
		s.is_automorphism = is_undirected_automorphism
		s.ref_singleton = ref_singleton_undirected
		s.ref_nonsingle = ref_nonsingle_undirected
	}
	s.indmin = 0
	s.lev = 1
	s.anc = 1
	s.ndiffs = 0
	s.nundiffs = 0
	s.ndiffnons = 0
	s.activityInc = 1
	for i := int32(0); i < s.n; i++ {
		s.theta[i] = i
		s.gamma[i] = i
		s.thsize[i] = 1
		s.thprev[i] = i
		s.thnext[i] = i
		s.unpairs[i] = -1
		s.undiffnons[i] = -1
	}
	s.npairs = 0
	stats.grpsize_base = 1
	stats.grpsize_exp = 0
	stats.nodes = 1
	stats.bads = 0
	stats.gens = 0
	stats.support = 0
	s.nninduce = 0
	s.nsinduce = 0
	s.csize = 0
	var max int32
	for i := int32(0); i < s.n; i++ {
		s.ccount[colors[i]]++
		if max < colors[i] {
			max = colors[i]
		}
	}
	s.nsplits = max + 1
	s.left.clen[0] = s.ccount[0] - 1
	for i := int32(0); i < max; i++ {
		s.left.clen[s.ccount[i]] = s.ccount[i+1] - 1
		s.ccount[i+1] += s.ccount[i]
	}
	for i := int32(0); i < s.n; i++ {
		s.ccount[colors[i]]--
		set_label(&s.left, s.ccount[colors[i]], i)
	}
	for i := int32(0); i <= max; i++ {
		s.ccount[i] = 0
	}
	for i := int32(0); i < s.n; i += s.left.clen[i] + 1 {
		add_induce(s, &s.left, i)
		fix_fronts(&s.left, i, i)
	}
	for i, j := int32(0), int32(-1); i < s.n; i += s.left.clen[i] + 1 {
		if s.left.clen[i] == 0 {
			continue
		}
		s.prevnon[i] = j
		s.nextnon[j+1] = i
		j = i
	}
	s.prevnon[s.n] = int32(-1)
	s.nextnon[s.n+1] = int32(-1)
	s.split = split_init
	s.refineBySim1 = refineBySim1_init
	s.refineBySim2 = refineBySim2_init
	fmt.Print("Initial Refine by Dependency graph ... ")
	refineByDepGraph(s, &s.left)
	fmt.Println("done!")
	fmt.Print("Initial Refine by Simulation ... ")
	if REFINE_BY_SIM_1 != 0 {
		s.refineBySim1(s, &s.left)
	}
	if REFINE_BY_SIM_2 != 0 {
		s.refineBySim2(s, &s.left)
	}
	fmt.Println("done!\n\t--------------------")
	s.refineBySim1 = refineBySim1_left
	s.refineBySim2 = refineBySim2_left
	descend_leftmost(s)
	s.split = split_other
	s.refineBySim1 = refineBySim1_other
	s.refineBySim2 = refineBySim2_other
	stats.levels = s.lev
	s.anc = s.lev
	copy(s.right.lab, s.left.lab)
	copy(s.right.unlab, s.left.unlab)
	copy(s.right.clen, s.left.clen)
	copy(s.right.cfront, s.left.cfront)
	copy(s.threp, s.left.lab)
	copy(s.thfront, s.left.unlab)
	if SELECT_DYNAMICALLY != 0 {
		s.select_decomposition = select_dynamically
	} else {
		s.select_decomposition = select_statically
	}
	for do_search(s) != 0 {
	}
}

func saucy_free(s *saucy) {
	if s == nil {
		return
	}
	for _, list := range [][]*Vec_Int_t{s.iDep, s.obs, s.oDep, s.ctrl} {
		for _, v := range list {
			Vec_IntFree(v)
		}
	}
	for _, v := range s.topOrder {
		Vec_PtrFree(v)
	}
	for i := int32(0); i < Vec_PtrSize(s.randomVectorArray_sim1); i++ {
		Vec_IntFree(Vec_PtrEntry(s.randomVectorArray_sim1, i).(*Vec_Int_t))
	}
	for i := int32(0); i < Vec_PtrSize(s.randomVectorArray_sim2); i++ {
		Vec_IntFree(Vec_PtrEntry(s.randomVectorArray_sim2, i).(*Vec_Int_t))
	}
	Vec_PtrFree(s.randomVectorArray_sim1)
	Vec_PtrFree(s.randomVectorArray_sim2)
	Vec_PtrFree(s.satCounterExamples)
	if s.pNtk_permuted != nil {
		baseabc.Abc_NtkDelete(s.pNtk_permuted)
	}
	*s = saucy{}
}
func saucy_alloc(pNtk *Abc_Ntk_t) *saucy {
	n := baseabc.Abc_NtkPiNum(pNtk) + baseabc.Abc_NtkPoNum(pNtk)
	ni, no := baseabc.Abc_NtkPiNum(pNtk), baseabc.Abc_NtkPoNum(pNtk)
	s := &saucy{ninduce: ints(n), sinduce: ints(n), indmark: make([]byte, n), left: coloring{cfront: ints(n), clen: ints(n), lab: ints(n), unlab: ints(n)}, right: coloring{cfront: ints(n), clen: ints(n), lab: ints(n), unlab: ints(n)}, stuff: make([]byte, n+1), bucket: ints(n + 2), count: ints(n + 1), ccount: ints(n), clist: ints(n), nextnon: ints(n + 2), prevnon: ints(n + 1), anctar: ints(n), start: ints(n), gamma: ints(n), junk: ints(n), theta: ints(n), thsize: ints(n), splitvar: ints(n), splitwho: ints(n), splitfrom: ints(n), splitlev: ints(n + 1), unsupp: ints(n), conncnts: ints(n), diffmark: make([]byte, n), diffs: ints(n), difflev: ints(n), undifflev: ints(n), specmin: ints(n), thnext: ints(n), thprev: ints(n), threp: ints(n), thfront: ints(n), pairs: ints(n), unpairs: ints(n), diffnons: ints(n), undiffnons: ints(n), marks: make([]byte, n), iDep: make([]*Vec_Int_t, ni), oDep: make([]*Vec_Int_t, no), obs: make([]*Vec_Int_t, ni), ctrl: make([]*Vec_Int_t, no), randomVectorArray_sim1: Vec_PtrAlloc(n), randomVectorSplit_sim1: ints(n), randomVectorArray_sim2: Vec_PtrAlloc(n), randomVectorSplit_sim2: ints(n), satCounterExamples: Vec_PtrAlloc(1), pModel: ints(ni)}
	for i := range s.iDep {
		s.iDep[i] = Vec_IntAlloc(1)
		s.obs[i] = Vec_IntAlloc(1)
	}
	for i := range s.oDep {
		s.oDep[i] = Vec_IntAlloc(1)
		s.ctrl[i] = Vec_IntAlloc(1)
	}
	return s
}

func print_stats(f *os.File, stats saucy_stats) {
	fmt.Fprintf(f, "group size = %fe%d\n", stats.grpsize_base, stats.grpsize_exp)
	fmt.Fprintf(f, "levels = %d\n", stats.levels)
	fmt.Fprintf(f, "nodes = %d\n", stats.nodes)
	fmt.Fprintf(f, "generators = %d\n", stats.gens)
	fmt.Fprintf(f, "total support = %d\n", stats.support)
	fmt.Fprintf(f, "average support = %.2f\n", float64(stats.support)/float64(stats.gens))
	fmt.Fprintf(f, "nodes per generator = %.2f\n", float64(stats.nodes)/float64(stats.gens))
	fmt.Fprintf(f, "bad nodes = %d\n", stats.bads)
}

/* From this point up are SAUCY functions*/
/////////////////////////////////////////////////////////////////////////////////////////////////////////
/* From this point down are new functions */

func getVertexName(pNtk *Abc_Ntk_t, v int32) string {
	numouts := baseabc.Abc_NtkPoNum(pNtk)
	if v < numouts {
		return baseabc.Abc_NtkPo(pNtk, v).Name()
	}
	return baseabc.Abc_NtkPi(pNtk, v-numouts).Name()
}

func saucyFindTopologicalOrder(pNtk *Abc_Ntk_t) []*Vec_Ptr_t {
	n := baseabc.Abc_NtkPiNum(pNtk)
	nodes := make([]*Vec_Ptr_t, n)
	for i := range nodes {
		nodes[i] = Vec_PtrAlloc(50)
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		pi := baseabc.Abc_NtkCi(pNtk, i)
		baseabc.Abc_NtkIncrementTravId(pNtk)
		baseabc.Abc_NodeSetTravIdCurrent(pi)
		root := pi
		if baseabc.Abc_NtkIsNetlist(pNtk) != 0 {
			root = baseabc.Abc_ObjFanout0(pi)
		}
		var ordered []*baseabc.Object
		for k := int32(0); k < baseabc.Abc_ObjFanoutNum(root); k++ {
			baseabc.Abc_NtkDfsReverse_rec(baseabc.Abc_ObjFanout(root, k), &ordered)
		}
		for _, obj := range ordered {
			Vec_PtrPush(nodes[i], obj)
		}
	}
	return nodes
}

func saucyGetDependencies(pNtk *Abc_Ntk_t, iDep, oDep []*Vec_Int_t) {
	supports := sim.Sim_ComputeFunSupp(baseabc.Abc_NtkStrashToGia(pNtk), false)
	for output := range supports {
		for input := 0; input < int(baseabc.Abc_NtkPiNum(pNtk)); input++ {
			if sim.Sim_HasBit(supports[output], input) {
				Vec_IntPushOrder(oDep[output], int32(input))
			}
		}
	}
	for output := range supports {
		for j := int32(0); j < Vec_IntSize(oDep[output]); j++ {
			Vec_IntPush(iDep[Vec_IntEntry(oDep[output], j)], int32(output))
		}
	}
}

func getDependenciesDummy(pNtk *Abc_Ntk_t, iDep, oDep []*Vec_Int_t) {
	for i := int32(0); i < baseabc.Abc_NtkPoNum(pNtk); i++ {
		for j := int32(0); j < baseabc.Abc_NtkPiNum(pNtk); j++ {
			Vec_IntPush(oDep[i], j)
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkPiNum(pNtk); i++ {
		for j := int32(0); j < baseabc.Abc_NtkPoNum(pNtk); j++ {
			Vec_IntPush(iDep[i], j)
		}
	}
}

func buildDepGraph(pNtk *Abc_Ntk_t, iDep, oDep []*Vec_Int_t) *saucy_graph {
	no := baseabc.Abc_NtkPoNum(pNtk)
	n := no + baseabc.Abc_NtkPiNum(pNtk)
	var e int32
	for i := int32(0); i < no; i++ {
		e += Vec_IntSize(oDep[i])
	}
	g := &saucy_graph{n: n, e: e, adj: ints(n + 1), edg: ints(2 * e)}
	for i := int32(0); i < n; i++ {
		if i < no {
			g.adj[i+1] = g.adj[i] + Vec_IntSize(oDep[i])
			for k := int32(0); k < Vec_IntSize(oDep[i]); k++ {
				g.edg[g.adj[i]+k] = Vec_IntEntry(oDep[i], k) + no
			}
		} else {
			pi := i - no
			g.adj[i+1] = g.adj[i] + Vec_IntSize(iDep[pi])
			for k := int32(0); k < Vec_IntSize(iDep[pi]); k++ {
				g.edg[g.adj[i]+k] = Vec_IntEntry(iDep[pi], k)
			}
		}
	}
	return g
}

func assignRandomBitsToCells(pNtk *Abc_Ntk_t, c *coloring) *Vec_Int_t {
	randVec := Vec_IntAlloc(1)
	for i := int32(0); i < baseabc.Abc_NtkPiNum(pNtk); i += c.clen[i+baseabc.Abc_NtkPoNum(pNtk)] + 1 {
		Vec_IntPush(randVec, int32(cascadeRandN(2)))
	}
	return randVec
}

func generateProperInputVector(pNtk *Abc_Ntk_t, c *coloring, randomVector *Vec_Int_t) []int32 {
	no, ni := baseabc.Abc_NtkPoNum(pNtk), baseabc.Abc_NtkPiNum(pNtk)
	values := make([]int32, ni)
	k := int32(0)
	i := no
	for ; i < no+ni; i += c.clen[i] + 1 {
		if k == Vec_IntSize(randomVector) {
			break
		}
		bit := Vec_IntEntry(randomVector, k)
		for j := i; j <= i+c.clen[i]; j++ {
			input := c.lab[j] - no
			values[input] = bit
		}
		k++
	}
	if i < no+ni {
		return nil
	}
	return values
}

func ifInputVectorsAreConsistent(s *saucy, leftVec, rightVec []int32) int32 {
	no, n := baseabc.Abc_NtkPoNum(s.pNtk), s.n
	for i := no; i < n; i += s.right.clen[i] + 1 {
		lab := s.left.lab[i] - no
		leftBit := leftVec[lab]
		for j := i + 1; j <= i+s.right.clen[i]; j++ {
			if leftBit != leftVec[s.left.lab[j]-no] {
				return -1
			}
		}
		rightBit := rightVec[s.right.lab[i]-no]
		for j := i + 1; j <= i+s.right.clen[i]; j++ {
			if rightBit != rightVec[s.right.lab[j]-no] {
				return 0
			}
		}
		if leftBit != rightBit {
			return 0
		}
	}
	return 1
}

func ifOutputVectorsAreConsistent(s *saucy, leftVec, rightVec []int32) int32 {
	for i := int32(0); i < baseabc.Abc_NtkPoNum(s.pNtk); i += s.right.clen[i] + 1 {
		var leftCount, rightCount int32
		for j := i; j <= i+s.right.clen[i]; j++ {
			if leftVec[s.left.lab[j]] != 0 {
				leftCount++
			}
			if rightVec[s.right.lab[j]] != 0 {
				rightCount++
			}
		}
		if leftCount != rightCount {
			return 0
		}
	}
	return 1
}

func buildSim1Graph(pNtk *Abc_Ntk_t, c *coloring, randVec *Vec_Int_t, iDep, oDep []*Vec_Int_t) *saucy_graph {
	piValues := generateProperInputVector(pNtk, c, randVec)
	if piValues == nil {
		return nil
	}
	output := Abc_NtkVerifySimulatePatternSlice(pNtk, piValues)
	no, ni := baseabc.Abc_NtkPoNum(pNtk), baseabc.Abc_NtkPiNum(pNtk)
	var ones int32
	for _, value := range output {
		if value != 0 {
			ones++
		}
	}
	g := &saucy_graph{n: no + ni, e: ni * ones, adj: ints(no + ni + 1), edg: ints(2 * ni * ones)}
	for i := int32(0); i < no; i++ {
		g.adj[i+1] = g.adj[i]
		if output[i] != 0 {
			g.adj[i+1] += Vec_IntSize(oDep[i])
		}
	}
	for i := int32(0); i < ni; i++ {
		degree := int32(0)
		for j := int32(0); j < Vec_IntSize(iDep[i]); j++ {
			if output[Vec_IntEntry(iDep[i], j)] != 0 {
				degree++
			}
		}
		g.adj[no+i+1] = g.adj[no+i] + degree
	}
	for i := int32(0); i < no; i++ {
		k := g.adj[i]
		if output[i] != 0 {
			for j := int32(0); j < Vec_IntSize(oDep[i]); j++ {
				g.edg[k] = Vec_IntEntry(oDep[i], j) + no
				k++
			}
		}
	}
	for i := int32(0); i < ni; i++ {
		k := g.adj[no+i]
		for j := int32(0); j < Vec_IntSize(iDep[i]); j++ {
			out := Vec_IntEntry(iDep[i], j)
			if output[out] != 0 {
				g.edg[k] = out
				k++
			}
		}
	}
	return g
}

func buildSim2Graph(pNtk *Abc_Ntk_t, c *coloring, randVec *Vec_Int_t, iDep, oDep []*Vec_Int_t, topOrder []*Vec_Ptr_t, obs, ctrl []*Vec_Int_t) *saucy_graph {
	values := generateProperInputVector(pNtk, c, randVec)
	if values == nil {
		return nil
	}
	output := Abc_NtkVerifySimulatePatternSlice(pNtk, values)
	no, ni := baseabc.Abc_NtkPoNum(pNtk), baseabc.Abc_NtkPiNum(pNtk)
	e := int32(0)
	for i := int32(0); i < ni; i++ {
		if c.clen[c.cfront[i+no]] == 0 {
			continue
		}
		values[i] ^= 1
		changed := Abc_NtkSimulateOneNode(pNtk, values, i, topOrder)
		for j := int32(0); j < Vec_IntSize(iDep[i]); j++ {
			out := Vec_IntEntry(iDep[i], j)
			if output[out] != changed[out] {
				Vec_IntPush(obs[i], out)
				Vec_IntPush(ctrl[out], i)
				e++
			}
		}
		values[i] ^= 1
	}
	g := &saucy_graph{n: no + ni, e: e, adj: ints(no + ni + 1), edg: ints(2 * e)}
	for i := int32(0); i < no; i++ {
		g.adj[i+1] = g.adj[i] + Vec_IntSize(ctrl[i])
		for j := int32(0); j < Vec_IntSize(ctrl[i]); j++ {
			g.edg[g.adj[i]+j] = Vec_IntEntry(ctrl[i], j) + no
		}
	}
	for i := int32(0); i < ni; i++ {
		g.adj[no+i+1] = g.adj[no+i] + Vec_IntSize(obs[i])
		for j := int32(0); j < Vec_IntSize(obs[i]); j++ {
			g.edg[g.adj[no+i]+j] = Vec_IntEntry(obs[i], j)
		}
	}
	for i := range obs {
		Vec_IntClear(obs[i])
	}
	for i := range ctrl {
		Vec_IntClear(ctrl[i])
	}
	return g
}

func bumpActivity(s *saucy, cex *sim_result) {
	cex.activity += s.activityInc
	if cex.activity > 1e20 {
		for i := int32(0); i < Vec_PtrSize(s.satCounterExamples); i++ {
			Vec_PtrEntry(s.satCounterExamples, i).(*sim_result).activity *= 1e-20
		}
		s.activityInc *= 1e-20
	}
}

func reduceDB(s *saucy) {
	extra := s.activityInc / float64(Vec_PtrSize(s.satCounterExamples))
	for float64(Vec_PtrSize(s.satCounterExamples)) > 0.7*MAX_LEARNTS {
		kept := int32(0)
		for i := int32(0); i < Vec_PtrSize(s.satCounterExamples); i++ {
			cex := Vec_PtrEntry(s.satCounterExamples, i).(*sim_result)
			if cex.activity < extra {
				continue
			}
			if kept < i {
				Vec_PtrWriteEntry(s.satCounterExamples, int(kept), cex)
			}
			kept++
		}
		Vec_PtrShrink(s.satCounterExamples, int(kept))
		extra *= 2
	}
}

func analyzeConflict(pNtk *Abc_Ntk_t, model []int32, verbose int32) *sim_result {
	ni, no := baseabc.Abc_NtkPiNum(pNtk), baseabc.Abc_NtkCoNum(pNtk)
	cex := &sim_result{inVec: ints(ni), outVec: ints(no)}
	values := Abc_NtkVerifySimulatePatternSlice(pNtk, model)
	for i := int32(0); i < baseabc.Abc_NtkCiNum(pNtk); i++ {
		obj := baseabc.Abc_NtkCi(pNtk, i)
		cex.inVec[baseabc.Abc_ObjId(obj)-1] = model[i]
		if verbose != 0 {
			fmt.Printf(" %s=%d", obj.Name(), model[i])
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(pNtk); i++ {
		obj := baseabc.Abc_NtkCo(pNtk, i)
		v := values[i]
		cex.outVec[baseabc.Abc_ObjId(obj)-ni-1] = v
		if v != 0 {
			cex.outVecOnes++
		}
	}
	if verbose != 0 {
		fmt.Println()
	}
	return cex
}

func Abc_NtkCecSat_saucy(pNtk1, pNtk2 *Abc_Ntk_t, model []int32) int32 {
	miter := Abc_NtkMiter(pNtk1, pNtk2, 1, 0, 0, 0)
	if miter == nil {
		fmt.Println("Miter computation has failed.")
		os.Exit(1)
	}
	ret := Abc_NtkMiterIsConstant(miter)
	if ret == 0 {
		clean := Abc_NtkVerifyGetCleanModel(miter, 1)
		for i := int32(0); i < baseabc.Abc_NtkPiNum(pNtk1); i++ {
			model[i] = clean[i]
		}
		Abc_NtkDelete(miter)
		return 0
	}
	if ret == 1 {
		Abc_NtkDelete(miter)
		return 1
	}
	cnf := Abc_NtkMulti(miter, 0, 100, 1, 0, 0, 0)
	Abc_NtkDelete(miter)
	if cnf == nil {
		fmt.Println("Renoding for CNF has failed.")
		os.Exit(1)
	}
	ret = Abc_NtkMiterSat(cnf, 10000, 0, 0, nil, nil)
	if ret == -1 {
		fmt.Println("Networks are undecided (SAT solver timed out).")
		os.Exit(1)
	}
	if cnfModel := baseabc.Abc_NtkReadModel(cnf); cnfModel != nil {
		for i := int32(0); i < baseabc.Abc_NtkPiNum(pNtk1); i++ {
			model[i] = cnfModel[i]
		}
	}
	Abc_NtkDelete(cnf)
	return ret
}

func saucyGateWay(orig *Abc_Ntk_t, po *Abc_Obj_t, gFile *os.File, booleanMatching, lookForSwaps, fixOutputs, fixInputs, quiet, printTree int32) {
	started := time.Now()
	var ntk *Abc_Ntk_t
	if po == nil {
		ntk = baseabc.Abc_NtkDup(orig)
	} else {
		ntk = baseabc.Abc_NtkCreateCone(orig, baseabc.Abc_ObjFanin0(po), po.Name(), 0)
	}
	if baseabc.Abc_NtkPiNum(ntk) == 0 {
		fmt.Println("This output is not dependent on any input")
		baseabc.Abc_NtkDelete(ntk)
		return
	}
	s := saucy_alloc(ntk)
	fmt.Print("Build functional dependency graph (dependency stats are below) ... ")
	saucyGetDependencies(ntk, s.iDep, s.oDep)
	fmt.Println("\t--------------------")
	s.topOrder = saucyFindTopologicalOrder(ntk)
	colors := ints(s.n)
	no := baseabc.Abc_NtkPoNum(ntk)
	for i := int32(0); i < no; i++ {
		if fixOutputs != 0 {
			colors[i] = i
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkPiNum(ntk); i++ {
		if fixInputs != 0 {
			if fixOutputs != 0 {
				colors[no+i] = no + i
			} else {
				colors[no+i] = 1 + i
			}
		} else if fixOutputs != 0 {
			colors[no+i] = no
		} else {
			colors[no+i] = 1
		}
	}
	s.fBooleanMatching = booleanMatching
	if booleanMatching != 0 {
		NUM_SIM1_ITERATION = 50
		NUM_SIM2_ITERATION = 50
	} else {
		NUM_SIM1_ITERATION = 200
		NUM_SIM2_ITERATION = 200
	}
	if quiet == 0 {
		s.print_automorphism = print_automorphism_ntk
	} else {
		s.print_automorphism = print_automorphism_quiet
	}
	if gFile == nil {
		s.gFile = os.Stdout
	} else {
		s.gFile = gFile
	}
	s.fPrintTree = printTree
	s.fLookForSwaps = lookForSwaps
	var stats saucy_stats
	saucy_search(ntk, s, 0, colors, &stats)
	print_stats(os.Stdout, stats)
	if booleanMatching != 0 {
		if stats.grpsize_base > 1 || stats.grpsize_exp > 0 {
			fmt.Println("*** Networks are equivalent ***")
		} else {
			fmt.Println("*** Networks are NOT equivalent ***")
		}
	}
	saucy_free(s)
	baseabc.Abc_NtkDelete(ntk)
	if f, err := os.OpenFile("hadi.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err == nil {
		fmt.Fprintf(f, "group size = %fe%d\n", stats.grpsize_base, stats.grpsize_exp)
		_ = f.Close()
	}
	fmt.Printf("Runtime = %9.2f sec\n", time.Since(started).Seconds())
}

//ABC_NAMESPACE_IMPL_END

// Source review: the tail keeps the C cell traversal, output/input vertex order,
// activity decay threshold, and conflict database reduction order. Pointer-owned
// C buffers are represented by Go slices and become collectible after references
// are removed from the vector. Simulation graph e retains the C allocation count.

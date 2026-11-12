package abci

import (
	"fmt"
	"time"
)

import "os"

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
const REFINE_BY_SIM_1 =     0
const REFINE_BY_SIM_2 =     0
const BACKTRACK_BY_SAT =    1
const SELECT_DYNAMICALLY =  0

/* number of iterations for sim1 and sim2 refinements */
var NUM_SIM1_ITERATION int32;
var NUM_SIM2_ITERATION int32;

/* conflict analysis */
const CLAUSE_DECAY = 0.9
const MAX_LEARNTS =  50

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
    grpsize_base float64;
    grpsize_exp int32;
    levels int32;
    nodes int32;
    bads int32;
    gens int32;
    support int32;
}

type saucy_graph struct {
    n int32;
    e int32;
    adj *int32;
    edg *int32;
}

type coloring struct {
    lab *int32;        /* Labelling of objects */
    unlab *int32;      /* Inverse of lab */
    cfront *int32;     /* Pointer to front of cells */
    clen *int32;       /* Length of cells (defined for cfront's) */
}

type sim_result struct {
    inVec *int32;
    outVec *int32;
    inVecSignature int32;
    outVecOnes int32; 
    activity float64;
}

type saucy struct {
    /* Graph data */
    n int32;           /* Size of domain */
    adj *int32;        /* Neighbors of k: edg[adj[k]]..edg[adj[k+1]] */
    edg *int32;        /* Actual neighbor data */
    dadj *int32;       /* Fanin neighbor indices, for digraphs */
    dedg *int32;       /* Fanin neighbor data, for digraphs */    

    /* Coloring data */
    struct coloring left, right;
    nextnon *int32;    /* Forward next-nonsingleton pointers */ 
    prevnon *int32;    /* Backward next-nonsingleton pointers */

    /* Refinement: inducers */
    indmark string;   /* Induce marks */
    ninduce *int32;    /* Nonsingletons that might induce refinement */
    sinduce *int32;    /* Singletons that might induce refinement */
    nninduce int32;    /* Size of ninduce stack */
    nsinduce int32;    /* Size of sinduce stack */

    /* Refinement: marked cells */
    clist *int32;      /* List of cells marked for refining */
    csize int32;       /* Number of cells in clist */

    /* Refinement: workspace */
    stuff string;     /* Bit vector, but one char per bit */
    ccount *int32;     /* Number of connections to refining cell */
    bucket *int32;     /* Workspace */
    count *int32;      /* Num vertices with same adj count to ref cell */
    junk *int32;       /* More workspace */
    gamma *int32;      /* Working permutation */
    conncnts *int32;   /* Connection counts for cell fronts */

    /* Search data */
    lev int32;         /* Current search tree level */
    anc int32;         /* Level of greatest common ancestor with zeta */
    anctar *int32;     /* Copy of target cell at anc */
    kanctar int32;     /* Location within anctar to iterate from */
    start *int32;      /* Location of target at each level */
    indmin int32;      /* Used for group size computation */
    match int32;       /* Have we not diverged from previous left? */

    /* Search: orbit partition */
    theta *int32;      /* Running approximation of orbit partition */
    thsize *int32;     /* Size of cells in theta, defined for mcrs */
    thnext *int32;     /* Next rep in list (circular list) */
    thprev *int32;     /* Previous rep in list */
    threp *int32;      /* First rep for a given cell front */
    thfront *int32;    /* The cell front associated with this rep */

    /* Search: split record */
    splitvar *int32;   /* The actual value of the splits on the left-most branch */
    splitwho *int32;   /* List of where splits occurred */
    splitfrom *int32;  /* List of cells which were split */
    splitlev *int32;   /* Where splitwho/from begins for each level */
    nsplits int32;     /* Number of splits at this point */

    /* Search: differences from leftmost */
    diffmark string;  /* Marked for diff labels */
    diffs *int32;      /* List of diff labels */
    difflev *int32;    /* How many labels diffed at each level */
    ndiffs int32;      /* Current number of diffs */
    undifflev *int32;  /* How many diff labels fixed at each level */
    nundiffs int32;    /* Current number of diffs in singletons (fixed) */
    unsupp *int32;     /* Inverted diff array */
    specmin *int32;    /* Speculated mappings */
    pairs *int32;      /* Not-undiffed diffs that can make two-cycles */
    unpairs *int32;    /* Indices into pairs */
    npairs int32;      /* Number of such pairs */
    diffnons *int32;   /* Diffs that haven't been undiffed */
    undiffnons *int32; /* Inverse of that */
    ndiffnons int32;   /* Number of such diffs */

    /* Polymorphic functions */
    int (*split)(struct saucy *, struct coloring *, int, int);
    int (*is_automorphism)(struct saucy *);
    int (*ref_singleton)(struct saucy *, struct coloring *, int);
    int (*ref_nonsingle)(struct saucy *, struct coloring *, int);
    void (*select_decomposition)(struct saucy *, int *, int *, int *);

     /* Statistics structure */
    struct saucy_stats *stats;

    /* New data structures for Boolean formulas */
    pNtk  *Abc_Ntk_t;
    pNtk_permuted  *Abc_Ntk_t;
    depAdj  *int32;
    depEdg  *int32;
    Vec_Int_t ** iDep, ** oDep;
    Vec_Int_t ** obs, ** ctrl;  
    topOrder  **Vec_Ptr_t;
    randomVectorArray_sim1  *Vec_Ptr_t;
    randomVectorSplit_sim1  *int32;
    randomVectorArray_sim2  *Vec_Ptr_t;
    randomVectorSplit_sim2  *int32;
    marks  string;
    pModel  *int32;
    satCounterExamples  *Vec_Ptr_t;
    activityInc float64;

    fBooleanMatching int32;
    fPrintTree int32;
    fLookForSwaps int32;
    gFile  *os.File;
    
    int (*refineBySim1)(struct saucy *, struct coloring *);
    int (*refineBySim2)(struct saucy *, struct coloring *);
    int (*print_automorphism)(FILE *f, int n, int *gamma, int nsupp, int *support, char * marks, Abc_Ntk_t * pNtk);
}

// inlining of macro ABC_ALLOC
func ints(n int32) *int32 { return ((int *) malloc(sizeof(int) * (size_t)(n))); }
// inlining of macro ABC_CALLOC
func zeros(n int32) *int32 { return ((int *) calloc((size_t)(n), sizeof(int))); }
// inlining of macro ABC_CALLOC
func bits(n int32) string { return ((char *) calloc((size_t)(n), sizeof(char))); }

// static char *               getVertexName(Abc_Ntk_t *pNtk, int v);

// static int *                generateProperInputVector(Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randomVector);

// static int                  ifInputVectorsAreConsistent(struct saucy * s, int * leftVec, int * rightVec);

// static int                  ifOutputVectorsAreConsistent(struct saucy * s, int * leftVec, int * rightVec);

// static Vec_Ptr_t **         findTopologicalOrder(Abc_Ntk_t * pNtk);

// static void                 getDependencies(Abc_Ntk_t *pNtk, Vec_Int_t** iDep, Vec_Int_t** oDep);

static struct saucy_graph * buildDepGraph (Abc_Ntk_t *pNtk, Vec_Int_t ** iDep, Vec_Int_t ** oDep);
static struct saucy_graph * buildSim1Graph(Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randVec, Vec_Int_t ** iDep, Vec_Int_t ** oDep);
static struct saucy_graph * buildSim2Graph(Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randVec, Vec_Int_t ** iDep, Vec_Int_t ** oDep, Vec_Ptr_t ** topOrder, Vec_Int_t ** obs,  Vec_Int_t ** ctrl);
// static Vec_Int_t *          assignRandomBitsToCells(Abc_Ntk_t * pNtk, struct coloring *c);

// static int                  Abc_NtkCecSat_saucy(Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel);

static struct sim_result *  analyzeConflict(Abc_Ntk_t * pNtk, int * pModel, int fVerbose);
// static void                 bumpActivity (struct saucy * s, struct sim_result * cex);

// static void                 reduceDB(struct saucy * s);



func print_automorphism_ntk(f *os.File, n int32, gamma *int32, nsupp int32, support *int32, marks string, pNtk *Abc_Ntk_t) int32 {
    var i, j, k int32;    

    /* We presume support is already sorted */
    for i = 0; i < nsupp; ++i {
        k = support[i];

        /* Skip elements already seen */
        if marks[k] { continue; }

        /* Start an orbit */
        marks[k] = 1;
        fmt.Fprintf(f, "(%s", getVertexName(pNtk, k));

        /* Mark and notify elements in this orbit */
        for j = gamma[k]; j != k; j = gamma[j] {
            marks[j] = 1;
            fmt.Fprintf(f, " %s", getVertexName(pNtk, j));
        }

        /* Finish off the orbit */
        fmt.Fprintf(f, ")");
    }
    fmt.Fprintf(f, "\n");

    /* Clean up after ourselves */
    for i = 0; i < nsupp; ++i {
        marks[support[i]] = 0;
    }

    return 1;
}

func print_automorphism_ntk2(f *os.File, n int32, gamma *int32, nsupp int32, support *int32, marks string, pNtk *Abc_Ntk_t) int32 {
    var i, j, k int32;    

    /* We presume support is already sorted */
    for i = 0; i < nsupp; ++i {
        k = support[i];

        /* Skip elements already seen */
        if marks[k] { continue; }

        /* Start an orbit */
        marks[k] = 1;
        fmt.Fprintf(f, "%d", k-1);

        /* Mark and notify elements in this orbit */
        for j = gamma[k]; j != k; j = gamma[j] {
            marks[j] = 1;
            fmt.Fprintf(f, " %d ", j-1);
        }

        /* Finish off the orbit */      
    }
    fmt.Fprintf(f, "-1\n");

    /* Clean up after ourselves */
    for i = 0; i < nsupp; ++i {
        marks[support[i]] = 0;
    }

    return 1;
}

func print_automorphism_quiet(f *os.File, n int32, gamma *int32, nsupp int32, support *int32, marks string, pNtk *Abc_Ntk_t) int32 {   
    return 1;
}

func array_find_min(a *int32, n int32) int32 {
    int *start = a, *end = a + n, *min = a;
    for ++a != end {
        if *a < *min { min = a; }
    }
    return min - start;
}

func swap(a *int32, x int32, y int32) {
    var tmp int32  = a[x];
    a[x] = a[y];
    a[y] = tmp;
}

func sift_up(a *int32, k int32) {
    var p int32;
    for {
        p = k / 2;
        if a[k] <= a[p] {
            return;
        
        } else {
            swap(a, k, p);
            k = p;
        }
    
if !(k > 1) { break }
} 
}

func sift_down(a *int32, n int32) {
    var p int32 = 1;
    var k int32 = 2;
    for k <= n {
        if k < n && a[k] < a[k+1] { ++k; }
        if a[p] < a[k] {
            swap(a, p, k);
            p = k;
            k = 2 * p;
        
        } else {
            return;
        }
    }
}

func heap_sort(a *int32, n int32) {
    var i int32;
    for i = 1; i < n; ++i {
        sift_up(a-1, i+1);
    }
    --i;
    for i > 0 {
        swap(a, 0, i);
        sift_down(a-1, i--);
    }
}

func insertion_sort(a *int32, n int32) {
    var i, j, k int32;
    for i = 1; i < n; ++i {
        k = a[i];
        for j = i; j > 0 && a[j-1] > k; --j {
            a[j] = a[j-1];
        }
        a[j] = k;
    }
}

func partition(a *int32, n int32, m int32) int32 {
    var f int32;
    var b int32 = n;
    for {
        for a[f] <= m { ++f; }
        do  --b; while (m <= a[b]);
        if f < b {
            swap(a, f, b);
            ++f;
        
        } else { break; }
    }
    return f;
}

func log_base2(n int32) int32 {
    var k int32;
    for n > 1 {
        ++k;
        n >>= 1;
    }
    return k;
}

func median(a int32, b int32, c int32) int32 {
    if a <= b {
        if b <= c { return b; }
        if a <= c { return c; }
        return a;
    
    } else {
        if a <= c { return a; }
        if b <= c { return c; }
        return b;
    }
}

func introsort_loop(a *int32, n int32, lim int32) {
    var p int32;
    for n > 16 {
        if lim == 0 {
            heap_sort(a, n);
            return;
        }
        --lim;
        p = partition(a, n, median(a[0], a[n/2], a[n-1]));
        introsort_loop(a + p, n - p, lim);
        n = p;
    }
}

func introsort(a *int32, n int32) {
    introsort_loop(a, n, 2 * log_base2(n));
    insertion_sort(a, n);
}

static int
do_find_min(struct coloring *c, int t)
{
    return array_find_min(c.lab + t, c.clen[t] + 1) + t;
}

static int
find_min(struct saucy *s, int t)
{
    return do_find_min(&s.right, t);
}

static void
set_label(struct coloring *c, int index, int value)
{
    c.lab[index] = value;
    c.unlab[value] = index;
}

static void
swap_labels(struct coloring *c, int a, int b)
{
    var tmp int32  = c.lab[a];
    set_label(c, a, c.lab[b]);
    set_label(c, b, tmp);
}

static void
move_to_back(struct saucy *s, struct coloring *c, int k)
{
    var cf int32  = c.cfront[k];
    var cb int32  = cf + c.clen[cf];
    var offset int32  = s.conncnts[cf]++;

    /* Move this connected label to the back of its cell */
    swap_labels(c, cb - offset, c.unlab[k]);

    /* Add it to the cell list if it's the first one swapped */
    if !(offset != 0) { s.clist[s.csize++] = cf; }
}

static void
data_mark(struct saucy *s, struct coloring *c, int k)
{
    var cf int32  = c.cfront[k];

    /* Move connects to the back of nonsingletons */
    if c.clen[cf] { move_to_back(s, c, k); }
}

static void
data_count(struct saucy *s, struct coloring *c, int k)
{
    var cf int32  = c.cfront[k];

    /* Move to back and count the number of connections */
    if c.clen[cf] && !s.ccount[k]++ { move_to_back(s, c, k); }
}

static int
check_mapping(struct saucy *s, int *adj, int *edg, int k)
{
    var i int32;
    var gk int32;
    var ret int32 = 1;

    /* Mark gamma of neighbors */
    for i = adj[k]; i != adj[k+1]; ++i {
        s.stuff[s.gamma[edg[i]]] = 1;
    }

    /* Check neighbors of gamma */
    gk = s.gamma[k];
    for i = adj[gk]; (ret != 0) && i != adj[gk+1]; ++i {
        ret = s.stuff[edg[i]];
    }

    /* Clear out bit vector before we leave */
    for i = adj[k]; i != adj[k+1]; ++i {
        s.stuff[s.gamma[edg[i]]] = 0;
    }

    return ret;
}

static int
add_conterexample(struct saucy *s, struct sim_result * cex)
{
    var i int32;
    var nins int32  = Abc_NtkPiNum(s.pNtk);
    struct sim_result * savedcex;
    
    cex.inVecSignature = 0;
    for i = 0; i < nins; i++ {
        if cex.inVec[i] {
            cex.inVecSignature += (cex.inVec[i] * i * i);
            cex.inVecSignature ^= 0xABCD;
        }
    }

    for i = 0; i < Vec_PtrSize(s.satCounterExamples); i++ {
        savedcex = (struct sim_result *)Vec_PtrEntry(s.satCounterExamples, i);
        if savedcex.inVecSignature == cex.inVecSignature {
            //bumpActivity(s, savedcex);
            return 0;
        }
    }
    
    Vec_PtrPush(s.satCounterExamples, cex);
    bumpActivity(s, cex);
    return 1;
}

static int
is_undirected_automorphism(struct saucy *s)
{
    var i, j, ret int32;  

    for i = 0; i < s.ndiffs; ++i {
        j = s.unsupp[i];
        if !check_mapping(s, s.adj, s.edg, j) { return 0; }
    }

    ret = Abc_NtkCecSat_saucy(s.pNtk, s.pNtk_permuted, s.pModel);
    
    if BACKTRACK_BY_SAT && !(ret != 0) {
        struct sim_result * cex;

        cex = analyzeConflict( s.pNtk, s.pModel, s.fPrintTree );
        add_conterexample(s, cex);

        cex = analyzeConflict( s.pNtk_permuted, s.pModel, s.fPrintTree );
        add_conterexample(s, cex);      
        
        s.activityInc *= (1 / CLAUSE_DECAY);
        if Vec_PtrSize(s.satCounterExamples) >= MAX_LEARNTS {
            reduceDB(s); }
    }

    return ret;
}

static int
is_directed_automorphism(struct saucy *s)
{
    var i, j int32;

    for i = 0; i < s.ndiffs; ++i {
        j = s.unsupp[i];
        if !check_mapping(s, s.adj, s.edg, j) { return 0; }
        if !check_mapping(s, s.dadj, s.dedg, j) { return 0; }
    }
    return 1;
}

static void
add_induce(struct saucy *s, struct coloring *c, int who)
{
    if !c.clen[who] {
        s.sinduce[s.nsinduce++] = who;
    
    } else {
        s.ninduce[s.nninduce++] = who;
    }
    s.indmark[who] = 1;
}

static void
fix_fronts(struct coloring *c, int cf, int ff)
{
    var i int32;
    var end int32 = cf + c.clen[cf];
    for i = ff; i <= end; ++i {
        c.cfront[c.lab[i]] = cf;
    }
}

func array_indirect_sort(a *int32, b *int32, n int32) {
    var h, i, j, k int32;

    /* Shell sort, as implemented in nauty, (C) Brendan McKay */
    j = n / 3;
    h = 1;
    for { h = 3 * h + 1; 
if !(h < j) { break }
} 

    for {
        for i = h; i < n; ++i {
            k = a[i];
            for j = i; b[a[j-h]] > b[k];  {
                a[j] = a[j-h];
                if (j -= h) < h { break; }
            }
            a[j] = k;
        }
        h /= 3;
    
if !(h > 0) { break }
} 
}

static int
at_terminal(struct saucy *s)
{
    return s.nsplits == s.n;
}

static void
add_diffnon(struct saucy *s, int k)
{
    /* Only add if we're in a consistent state */
    if s.ndiffnons == -1 { return; }

    s.undiffnons[k] = s.ndiffnons;
    s.diffnons[s.ndiffnons++] = k;
}

static void
remove_diffnon(struct saucy *s, int k)
{
    var j int32;

    if s.undiffnons[k] == -1 { return; }

    j = s.diffnons[--s.ndiffnons];
    s.diffnons[s.undiffnons[k]] = j;
    s.undiffnons[j] = s.undiffnons[k];

    s.undiffnons[k] = -1;
}

static void
add_diff(struct saucy *s, int k)
{
    if !s.diffmark[k] {
        s.diffmark[k] = 1;
        s.diffs[s.ndiffs++] = k;
        add_diffnon(s, k);
    }
}

static int
is_a_pair(struct saucy *s, int k)
{
    return s.unpairs[k] != -1;
}

static int
in_cell_range(struct coloring *c, int ff, int cf)
{
    var cb int32  = cf + c.clen[cf];
    return cf <= ff && ff <= cb;
}

static void
add_pair(struct saucy *s, int k)
{
    if s.npairs != -1 {
        s.unpairs[k] = s.npairs;
        s.pairs[s.npairs++] = k;
    }
}

static void
eat_pair(struct saucy *s, int k)
{
    var j int32;
    j = s.pairs[--s.npairs];
    s.pairs[s.unpairs[k]] = j;
    s.unpairs[j] = s.unpairs[k];
    s.unpairs[k] = -1;
}

static void
pick_all_the_pairs(struct saucy *s)
{
    var i int32;
    for i = 0; i < s.npairs; ++i {
        s.unpairs[s.pairs[i]] = -1;
    }
    s.npairs = 0;
}

static void
clear_undiffnons(struct saucy *s)
{
    var i int32;
    for i = 0 ; i < s.ndiffnons ; ++i {
        s.undiffnons[s.diffnons[i]] = -1;
    }
}

static void
fix_diff_singleton(struct saucy *s, int cf)
{
    var r int32  = s.right.lab[cf];
    var l int32  = s.left.lab[cf];
    var rcfl int32;

    if !s.right.clen[cf] && r != l {

        /* Make sure diff is marked */
        add_diff(s, r);

        /* It is now undiffed since it is singleton */
        ++s.nundiffs;
        remove_diffnon(s, r);

        /* Mark the other if not singleton already */
        rcfl = s.right.cfront[l];
        if s.right.clen[rcfl] {
            add_diff(s, l);

            /* Check for pairs */
            if in_cell_range(&s.right, s.left.unlab[r], rcfl) {
                add_pair(s, l);
            }
        
        /* Otherwise we might be eating a pair */
        } else if is_a_pair(s, r) {
            eat_pair(s, r);
        }
    }
}

static void
fix_diff_subtract(struct saucy *s, int cf, int *a, int *b)
{
    var i, k int32;
    var cb int32  = cf + s.right.clen[cf];

    /* Mark the contents of the first set */
    for i = cf; i <= cb; ++i {
        s.stuff[a[i]] = 1;
    }

    /* Add elements from second set not present in the first */
    for i = cf; i <= cb; ++i {
        k = b[i];
        if !s.stuff[k] { add_diff(s, k); }
    }

    /* Clear the marks of the first set */
    for i = cf; i <= cb; ++i {
        s.stuff[a[i]] = 0;
    }
}

static void
fix_diffs(struct saucy *s, int cf, int ff)
{
    var min int32;

    /* Check for singleton cases in both cells */
    fix_diff_singleton(s, cf);
    fix_diff_singleton(s, ff);

    /* If they're both nonsingleton, do subtraction on smaller */
    if s.right.clen[cf] && s.right.clen[ff] {
        if s.right.clen[cf] < s.right.clen[ff] {
            min = cf;
        } else {
            min = ff;
        }
        fix_diff_subtract(s, min, s.left.lab, s.right.lab);
        fix_diff_subtract(s, min, s.right.lab, s.left.lab);
    }
}

static void
split_color(struct coloring *c, int cf, int ff)
{
    var cb, fb int32;

    /* Fix lengths */
    fb = ff - 1;
    cb = cf + c.clen[cf];
    c.clen[cf] = fb - cf;
    c.clen[ff] = cb - ff;

    /* Fix cell front pointers */
    fix_fronts(c, ff, ff);
}

static void
split_common(struct saucy *s, struct coloring *c, int cf, int ff)
{
    split_color(c, cf, ff);

    /* Add to refinement */
    if s.indmark[cf] || c.clen[ff] < c.clen[cf] {
        add_induce(s, c, ff);
    
    } else {
        add_induce(s, c, cf);
    }
}

static int
split_left(struct saucy *s, struct coloring *c, int cf, int ff)
{
    /* Record the split */
    s.splitwho[s.nsplits] = ff;
    s.splitfrom[s.nsplits] = cf;
    ++s.nsplits;

    /* Do common splitting tasks */
    split_common(s, c, cf, ff);

    /* Always succeeds */
    return 1;
}

static int
split_init(struct saucy *s, struct coloring *c, int cf, int ff)
{
    split_left(s, c, cf, ff);

    /* Maintain nonsingleton list for finding new targets */
    if c.clen[ff] {
        s.prevnon[s.nextnon[cf]] = ff;
        s.nextnon[ff] = s.nextnon[cf];
        s.prevnon[ff] = cf;
        s.nextnon[cf] = ff;
    }
    if !c.clen[cf] {
        s.nextnon[s.prevnon[cf]] = s.nextnon[cf];
        s.prevnon[s.nextnon[cf]] = s.prevnon[cf];
    }

    /* Always succeeds */
    return 1;
}

static int
split_other(struct saucy *s, struct coloring *c, int cf, int ff)
{
    var k int32  = s.nsplits;

    /* Verify the split with init */
    if s.splitwho[k] != ff || s.splitfrom[k] != cf
            || k >= s.splitlev[s.lev] {
        return 0;
    }
    ++s.nsplits;

    /* Do common splitting tasks */
    split_common(s, c, cf, ff);

    /* Fix differences with init */
    fix_diffs(s, cf, ff);

    /* If we got this far we succeeded */
    return 1;
}

static int
print_partition(struct coloring *left, struct coloring *right, int n, Abc_Ntk_t * pNtk, int fNames)
{
        var i, j int32;        

        fmt.Printf("top = |");
        for i = 0; i < n; i += (left.clen[i]+1) {
            for j = 0; j < (left.clen[i]+1); j++ {
                if fNames { fmt.Printf("%s ", getVertexName(pNtk, left.lab[i+j]));
                } else {        fmt.Printf("%d ", left.lab[i+j]); }
            }
            if (i+left.clen[i]+1) < n { fmt.Printf("|"); }
        }
        fmt.Printf("|\n");

        /*printf("(cfront = {");
        for (i = 0; i < n; i++)
            printf("%d ", left->cfront[i]);
        printf("})\n");*/

        if right == nil { return 1; }

        fmt.Printf("bot = |");
        for i = 0; i < n; i += (right.clen[i]+1) {
            for j = 0; j < (right.clen[i]+1); j++ {
                if fNames { fmt.Printf("%s ", getVertexName(pNtk, right.lab[i+j]));
                } else {        fmt.Printf("%d ", right.lab[i+j]); }
            }
            if (i+right.clen[i]+1) < n { fmt.Printf("|"); }
        }
        fmt.Printf("|\n");

        /*printf("(cfront = {");
        for (i = 0; i < n; i++)
            printf("%d ", right->cfront[i]);
        printf("})\n");*/

        return 1;
}

static int
refine_cell(struct saucy *s, struct coloring *c,
    int (*refine)(struct saucy *, struct coloring *, int))
{
    var i int32;
    var cf int32;
    var ret int32 = 1;

    /*
     * The connected list must be consistent.  This is for
     * detecting mappings across nodes at a given level.  However,
     * at the root of the tree, we never have to map with another
     * node, so we lack this consistency constraint in that case.
     */
    if s.lev > 1 { introsort(s.clist, s.csize); }

    /* Now iterate over the marked cells */
    for i = 0; (ret != 0) && i < s.csize; ++i {
        cf = s.clist[i];
        ret = refine(s, c, cf);
    }

    /* Clear the connected marks */
    for i = 0; i < s.csize; ++i {
        cf = s.clist[i];
        s.conncnts[cf] = 0;
    }
    s.csize = 0;
    return ret;
}

static int
maybe_split(struct saucy *s, struct coloring *c, int cf, int ff)
{
    if cf == ff {
        return 1;
    } else {
        return s.split(s, c, cf, ff);
    }
}

static int
ref_single_cell(struct saucy *s, struct coloring *c, int cf)
{
    var zcnt int32  = c.clen[cf] + 1 - s.conncnts[cf];
    return maybe_split(s, c, cf, cf + zcnt);
}

static int
ref_singleton(struct saucy *s, struct coloring *c,
    int *adj, int *edg, int cf)
{
    var i int32;
    var k int32 = c.lab[cf];

    /* Find the cells we're connected to, and mark our neighbors */
    for i = adj[k]; i != adj[k+1]; ++i {
        data_mark(s, c, edg[i]);
    }

    /* Refine the cells we're connected to */
    return refine_cell(s, c, ref_single_cell);
}

static int
ref_singleton_directed(struct saucy *s, struct coloring *c, int cf)
{
    return ref_singleton(s, c, s.adj, s.edg, cf)
        && ref_singleton(s, c, s.dadj, s.dedg, cf);
}

static int
ref_singleton_undirected(struct saucy *s, struct coloring *c, int cf)
{
    return ref_singleton(s, c, s.adj, s.edg, cf);
}

static int
ref_nonsingle_cell(struct saucy *s, struct coloring *c, int cf)
{
    var cnt, i, cb, nzf, ff, fb, bmin, bmax int32;

    /* Find the front and back */
    cb = cf + c.clen[cf];
    nzf = cb - s.conncnts[cf] + 1;

    /* Prepare the buckets */
    ff = nzf;
    cnt = s.ccount[c.lab[ff]];
    s.count[ff] = bmin = bmax = cnt;
    s.bucket[cnt] = 1;

    /* Iterate through the rest of the vertices */
    for ++ff <= cb {
        cnt = s.ccount[c.lab[ff]];

        /* Initialize intermediate buckets */
        for bmin > cnt { s.bucket[--bmin] = 0; }
        for bmax < cnt { s.bucket[++bmax] = 0; }

        /* Mark this count */
        ++s.bucket[cnt];
        s.count[ff] = cnt;
    }

    /* If they all had the same count, bail */
    if bmin == bmax && cf == nzf { return 1; }
    ff = fb = nzf;

    /* Calculate bucket locations, sizes */
    for i = bmin; i <= bmax; ++i, ff = fb {
        if !s.bucket[i] { continue; }
        fb = ff + s.bucket[i];
        s.bucket[i] = fb;
    }

    /* Repair the partition nest */
    for i = nzf; i <= cb; ++i {
        s.junk[--s.bucket[s.count[i]]] = c.lab[i];
    }
    for i = nzf; i <= cb; ++i {
        set_label(c, i, s.junk[i]);
    }

    /* Split; induce */
    for i = bmax; i > bmin; --i {
        ff = s.bucket[i];
        if (ff != 0) && !s.split(s, c, cf, ff) { return 0; }
    }

    /* If there was a zero area, then there's one more cell */
    return maybe_split(s, c, cf, s.bucket[bmin]);
}

static int
ref_nonsingle(struct saucy *s, struct coloring *c,
    int *adj, int *edg, int cf)
{
    var i, j, k, ret int32;
    int cb = cf + c.clen[cf];
    int size = cb - cf + 1;

    /* Double check for nonsingles which became singles later */
    if cf == cb {
        return ref_singleton(s, c, adj, edg, cf);
    }

    /* Establish connected list */
    memcpy(s.junk, c.lab + cf, size * sizeof(int));
    for i = 0; i < size; ++i {
        k = s.junk[i];
        for j = adj[k]; j != adj[k+1]; ++j {
            data_count(s, c, edg[j]);
        }
    }

    /* Refine the cells we're connected to */
    ret = refine_cell(s, c, ref_nonsingle_cell);

    /* Clear the counts; use lab because junk was overwritten */
    for i = cf; i <= cb; ++i {
        k = c.lab[i];
        for j = adj[k]; j != adj[k+1]; ++j {
            s.ccount[edg[j]] = 0;
        }
    }

    return ret;
}

static int
ref_nonsingle_directed(struct saucy *s, struct coloring *c, int cf)
{
    return ref_nonsingle(s, c, s.adj, s.edg, cf)
        && ref_nonsingle(s, c, s.dadj, s.dedg, cf);
}

static int
ref_nonsingle_undirected(struct saucy *s, struct coloring *c, int cf)
{
    return ref_nonsingle(s, c, s.adj, s.edg, cf);
}

static void
clear_refine(struct saucy *s)
{
    var i int32;
    for i = 0; i < s.nninduce; ++i {
        s.indmark[s.ninduce[i]] = 0;
    }
    for i = 0; i < s.nsinduce; ++i {
        s.indmark[s.sinduce[i]] = 0;
    }
    s.nninduce = s.nsinduce = 0;
}

static int
refine(struct saucy *s, struct coloring *c)
{
    var front int32;

    /* Keep going until refinement stops */
    for {

        /* If discrete, bail */
        if at_terminal(s) {
            clear_refine(s);
            return 1;
        };

        /* Look for something else to refine on */
        if s.nsinduce {
            front = s.sinduce[--s.nsinduce];
            s.indmark[front] = 0;
            if !s.ref_singleton(s, c, front) { break; }
        
        } else if s.nninduce {
            front = s.ninduce[--s.nninduce];
            s.indmark[front] = 0;
            if !s.ref_nonsingle(s, c, front) { break; }
        
        } else {
            return 1;
        };
    }

    clear_refine(s);
    return 0;
}

static int
refineByDepGraph(struct saucy *s, struct coloring *c) 
{
    s.adj = s.depAdj;
    s.edg = s.depEdg; 

    return refine(s, c);
}

static int
backtrackBysatCounterExamples(struct saucy *s, struct coloring *c)
{
    var i, j, res int32;  
    struct sim_result * cex1, * cex2;
    var flag *int32  = zeros(Vec_PtrSize(s.satCounterExamples));

    if c == &s.left { return 1; }
    if Vec_PtrSize(s.satCounterExamples) == 0 { return 1; }      

    for i = 0; i < Vec_PtrSize(s.satCounterExamples); i++ {  
        cex1 = (struct sim_result *)Vec_PtrEntry(s.satCounterExamples, i);      

        for j = 0; j < Vec_PtrSize(s.satCounterExamples); j++ {
            if flag[j] { continue; }
        
            cex2 = (struct sim_result *)Vec_PtrEntry(s.satCounterExamples, j);
            res = ifInputVectorsAreConsistent(s, cex1.inVec, cex2.inVec);

            if res == -2 {
                flag[j] = 1;
                continue;
            }
            if res == -1 { break; }
            if res == 0 { continue; }

            if cex1.outVecOnes != cex2.outVecOnes {
                bumpActivity(s, cex1);
                bumpActivity(s, cex2);
                // inlining of macro ABC_FREE
                ((flag) ? (free((char *) (flag)), (flag) = 0) : 0);
                return 0;
            }

            /* if two input vectors produce the same number of ones (look above), and
             * pNtk's number of outputs is 1, then output vectors are definitely consistent. */
            if Abc_NtkPoNum(s.pNtk) == 1 { continue; }

            if !ifOutputVectorsAreConsistent(s, cex1.outVec, cex2.outVec) {
                bumpActivity(s, cex1);
                bumpActivity(s, cex2);
                // inlining of macro ABC_FREE
                ((flag) ? (free((char *) (flag)), (flag) = 0) : 0);
                return 0;
            }
        }
    }

    // inlining of macro ABC_FREE
    ((flag) ? (free((char *) (flag)), (flag) = 0) : 0);
    return 1;
}

static int
refineBySim1_init(struct saucy *s, struct coloring *c) 
{
    struct saucy_graph *g;
    var randVec *Vec_Int_t;
    var i, j int32;
    var allOutputsAreDistinguished int32;
    var nsplits int32;

    if Abc_NtkPoNum(s.pNtk) == 1 { return 1; }
    
    for i = 0; i < NUM_SIM1_ITERATION; i++ {

        /* if all outputs are distinguished, quit */
        allOutputsAreDistinguished = 1;
        for j = 0; j < Abc_NtkPoNum(s.pNtk); j++ {
            if c.clen[j] {
                allOutputsAreDistinguished = 0;
                break;
            }
        }
        if allOutputsAreDistinguished != 0 { break; }

        randVec = assignRandomBitsToCells(s.pNtk, c);      
        g = buildSim1Graph(s.pNtk, c, randVec, s.iDep, s.oDep);
        assert(g != nil, "g != nil");

        s.adj = g.adj;
        s.edg = g.edg;

        nsplits = s.nsplits;

        for j = 0; j < s.n; j += c.clen[j]+1 {            
                add_induce(s, c, j);
        }
        refine(s, c);

        if s.nsplits > nsplits {         
            i = 0; /* reset i */
            /* do more refinement by dependency graph */
            for j = 0; j < s.n; j += c.clen[j]+1 {                
                    add_induce(s, c, j);
            }
            refineByDepGraph(s, c);
        }
        
        Vec_IntFree(randVec);
        // inlining of macro ABC_FREE
        ((g.adj) ? (free((char *) (g.adj)), (g.adj) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g.edg) ? (free((char *) (g.edg)), (g.edg) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g) ? (free((char *) (g)), (g) = 0) : 0);
    }   

    return 1;
}


static int
refineBySim1_left(struct saucy *s, struct coloring *c) 
{
    struct saucy_graph *g;
    var randVec *Vec_Int_t;
    var i, j int32;
    var allOutputsAreDistinguished int32;
    var nsplits int32;

    if Abc_NtkPoNum(s.pNtk) == 1 { return 1; }
    
    for i = 0; i < NUM_SIM1_ITERATION; i++ {

        /* if all outputs are distinguished, quit */
        allOutputsAreDistinguished = 1;
        for j = 0; j < Abc_NtkPoNum(s.pNtk); j++ {
            if c.clen[j] {
                allOutputsAreDistinguished = 0;
                break;
            }
        }
        if allOutputsAreDistinguished != 0 { break; }

        randVec = assignRandomBitsToCells(s.pNtk, c);      
        g = buildSim1Graph(s.pNtk, c, randVec, s.iDep, s.oDep);
        assert(g != nil, "g != nil");

        s.adj = g.adj;
        s.edg = g.edg;

        nsplits = s.nsplits;

        for j = 0; j < s.n; j += c.clen[j]+1 {            
                add_induce(s, c, j);
        }
        refine(s, c);

        if s.nsplits > nsplits {
            /* save the random vector */
            Vec_PtrPush(s.randomVectorArray_sim1, randVec);            
            i = 0;  /* reset i */
            /* do more refinement by dependency graph */
            for j = 0; j < s.n; j += c.clen[j]+1 {                
                    add_induce(s, c, j);
            }
            refineByDepGraph(s, c);
        
        } else {
            Vec_IntFree(randVec); }

        // inlining of macro ABC_FREE
        ((g.adj) ? (free((char *) (g.adj)), (g.adj) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g.edg) ? (free((char *) (g.edg)), (g.edg) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g) ? (free((char *) (g)), (g) = 0) : 0);
    }

    s.randomVectorSplit_sim1[s.lev] = Vec_PtrSize(s.randomVectorArray_sim1); 

    return 1;
}

static int
refineBySim1_other(struct saucy *s, struct coloring *c) 
{
    struct saucy_graph *g;
    var randVec *Vec_Int_t;
    var i, j int32;
    var ret, nsplits int32;

    for i = s.randomVectorSplit_sim1[s.lev-1]; i < s.randomVectorSplit_sim1[s.lev]; i++ {
        randVec = (Vec_Int_t *)Vec_PtrEntry(s.randomVectorArray_sim1, i);
        g = buildSim1Graph(s.pNtk, c, randVec, s.iDep, s.oDep);

        if g == nil {
            assert(c == &s.right, "c == &s.right");
            return 0;
        }

        s.adj = g.adj;
        s.edg = g.edg;

        nsplits = s.nsplits;

        for j = 0; j < s.n; j += c.clen[j]+1 {            
                add_induce(s, c, j);
        }
        ret = refine(s, c);

        if s.nsplits == nsplits {
            assert(c == &s.right, "c == &s.right");
            ret = 0;
        }

        if ret != 0 {      
            /* do more refinement now by dependency graph */
            for j = 0; j < s.n; j += c.clen[j]+1 {                
                    add_induce(s, c, j);
            }
            ret = refineByDepGraph(s, c);
        }
        
        // inlining of macro ABC_FREE
        ((g.adj) ? (free((char *) (g.adj)), (g.adj) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g.edg) ? (free((char *) (g.edg)), (g.edg) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g) ? (free((char *) (g)), (g) = 0) : 0);

        if !(ret != 0) { return 0; }
    }   

    return 1;
}

static int
refineBySim2_init(struct saucy *s, struct coloring *c) 
{
    struct saucy_graph *g;
    var randVec *Vec_Int_t;    
    var i, j int32;
    var nsplits int32;    
    
    for i = 0; i < NUM_SIM2_ITERATION; i++ {
        randVec = assignRandomBitsToCells(s.pNtk, c);      
        g = buildSim2Graph(s.pNtk, c, randVec, s.iDep, s.oDep, s.topOrder, s.obs,  s.ctrl);
        assert(g != nil, "g != nil");

        s.adj = g.adj;
        s.edg = g.edg;

        nsplits = s.nsplits;
        
        for j = 0; j < s.n; j += c.clen[j]+1 {            
                add_induce(s, c, j);
        }
        refine(s, c);

        if s.nsplits > nsplits {                     
            i = 0; /* reset i */
            /* do more refinement by dependency graph */
            for j = 0; j < s.n; j += c.clen[j]+1 {                
                    add_induce(s, c, j);
            }
            refineByDepGraph(s, c);
        }
        
        Vec_IntFree(randVec);
        
        // inlining of macro ABC_FREE
        ((g.adj) ? (free((char *) (g.adj)), (g.adj) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g.edg) ? (free((char *) (g.edg)), (g.edg) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g) ? (free((char *) (g)), (g) = 0) : 0);
    }

    return 1;
}

static int
refineBySim2_left(struct saucy *s, struct coloring *c) 
{
    struct saucy_graph *g;
    var randVec *Vec_Int_t;    
    var i, j int32;
    var nsplits int32;    
    
    for i = 0; i < NUM_SIM2_ITERATION; i++ {
        randVec = assignRandomBitsToCells(s.pNtk, c);      
        g = buildSim2Graph(s.pNtk, c, randVec, s.iDep, s.oDep, s.topOrder, s.obs,  s.ctrl);
        assert(g != nil, "g != nil");

        s.adj = g.adj;
        s.edg = g.edg;

        nsplits = s.nsplits;
        
        for j = 0; j < s.n; j += c.clen[j]+1 {            
                add_induce(s, c, j);
        }
        refine(s, c);

        if s.nsplits > nsplits {
            /* save the random vector */
            Vec_PtrPush(s.randomVectorArray_sim2, randVec);            
            i = 0; /* reset i */
            /* do more refinement by dependency graph */
            for j = 0; j < s.n; j += c.clen[j]+1 {                
                    add_induce(s, c, j);
            }
            refineByDepGraph(s, c);
        
        } else {
            Vec_IntFree(randVec); }
        
        // inlining of macro ABC_FREE
        ((g.adj) ? (free((char *) (g.adj)), (g.adj) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g.edg) ? (free((char *) (g.edg)), (g.edg) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g) ? (free((char *) (g)), (g) = 0) : 0);
    }

    s.randomVectorSplit_sim2[s.lev] = Vec_PtrSize(s.randomVectorArray_sim2); 

    return 1;
}

static int
refineBySim2_other(struct saucy *s, struct coloring *c) 
{
    struct saucy_graph *g;
    var randVec *Vec_Int_t;
    var i, j int32;
    var ret, nsplits int32;

    for i = s.randomVectorSplit_sim2[s.lev-1]; i < s.randomVectorSplit_sim2[s.lev]; i++ {
        randVec = (Vec_Int_t *)Vec_PtrEntry(s.randomVectorArray_sim2, i);       
        g = buildSim2Graph(s.pNtk, c, randVec, s.iDep, s.oDep, s.topOrder, s.obs,  s.ctrl);

        if g == nil {
            assert(c == &s.right, "c == &s.right");
            return 0;
        }
        
        s.adj = g.adj;
        s.edg = g.edg;    

        nsplits = s.nsplits;       

        for j = 0; j < s.n; j += c.clen[j]+1 {            
                add_induce(s, c, j);
        }
        ret = refine(s, c);

        if s.nsplits == nsplits {
            assert(c == &s.right, "c == &s.right");
            ret = 0;
        }

        if ret != 0 {
            /* do more refinement by dependency graph */
            for j = 0; j < s.n; j += c.clen[j]+1 {                
                    add_induce(s, c, j);
            }
            ret = refineByDepGraph(s, c);
        }

        // inlining of macro ABC_FREE
        ((g.adj) ? (free((char *) (g.adj)), (g.adj) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g.edg) ? (free((char *) (g.edg)), (g.edg) = 0) : 0);
        // inlining of macro ABC_FREE
        ((g) ? (free((char *) (g)), (g) = 0) : 0);

        if !(ret != 0) {
            assert(c == &s.right, "c == &s.right");
            return 0;
        }
    }

    return 1;
}

static int
check_OPP_only_has_swaps(struct saucy *s, struct coloring *c)
{
    var j, cell int32;
    var left_cfront, right_cfront *Vec_Int_t;

    if c == &s.left {
        return 1; }

    left_cfront = Vec_IntAlloc (1);
    right_cfront = Vec_IntAlloc (1);

    for cell = 0; cell < s.n; cell += (s.left.clen[cell]+1) {               
        for j = cell; j <= (cell+s.left.clen[cell]); j++ {
            Vec_IntPush(left_cfront, s.left.cfront[s.right.lab[j]]);
            Vec_IntPush(right_cfront, s.right.cfront[s.left.lab[j]]);             
        }
        Vec_IntSortUnsigned(left_cfront);
        Vec_IntSortUnsigned(right_cfront);
        for j = 0; j < Vec_IntSize(left_cfront); j++ {
            if Vec_IntEntry(left_cfront, j) != Vec_IntEntry(right_cfront, j) {
                Vec_IntFree(left_cfront);
                Vec_IntFree(right_cfront);
                return 0;
            }
        }
        Vec_IntClear(left_cfront);
        Vec_IntClear(right_cfront);
    }

    Vec_IntFree(left_cfront);
    Vec_IntFree(right_cfront);

    return 1;
}

static int
check_OPP_for_Boolean_matching(struct saucy *s, struct coloring *c)
{
    var j, cell int32;
    var countN1Left, countN2Left int32;
    var countN1Right, countN2Right int32;
    var name string; 

    if c == &s.left {
        return 1; }       

    for cell = 0; cell < s.n; cell += (s.right.clen[cell]+1) {  
        countN1Left = countN2Left = countN1Right = countN2Right = 0;

        for j = cell; j <= (cell+s.right.clen[cell]); j++ {

            name = getVertexName(s.pNtk, s.left.lab[j]);
            assert(name[0] == 'N' && name[2] == ':', "name[0] == 'N' && name[2] == ':'");
            if name[1] == '1' {
                countN1Left++;
            } else {
                assert(name[1] == '2', "name[1] == '2'");
                countN2Left++;
            }

            name = getVertexName(s.pNtk, s.right.lab[j]);
            assert(name[0] == 'N' && name[2] == ':', "name[0] == 'N' && name[2] == ':'");
            if name[1] == '1' { 
                countN1Right++;
            } else {
                assert(name[1] == '2', "name[1] == '2'");
                countN2Right++;
            }

        }

        if countN1Left != countN2Right || countN2Left != countN1Right {
            return 0; }
    }   
    
    return 1;
}

static int
double_check_OPP_isomorphism(struct saucy *s, struct coloring * c)
{
    /* This is the new enhancement in saucy 3.0 */
    var i, j, v, sum1, sum2, xor1, xor2 int32;

    if c == &s.left {
        return 1; }
    
    for i = s.nsplits - 1; i > s.splitlev[s.lev-1]; --i {
        v = c.lab[s.splitwho[i]];
        sum1 = xor1 = 0;
        for j = s.adj[v]; j < s.adj[v+1]; j++ {
            sum1 += c.cfront[s.edg[j]];
            xor1 ^= c.cfront[s.edg[j]];
        }
        v = s.left.lab[s.splitwho[i]];
        sum2 = xor2 = 0;
        for j = s.adj[v]; j < s.adj[v+1]; j++ {
            sum2 += s.left.cfront[s.edg[j]];
            xor2 ^= s.left.cfront[s.edg[j]];
        }
        if (sum1 != sum2) || (xor1 != xor2) {
            return 0; }
        v = c.lab[s.splitfrom[i]];
        sum1 = xor1 = 0;
        for j = s.adj[v]; j < s.adj[v+1]; j++ {
            sum1 += c.cfront[s.edg[j]];
            xor1 ^= c.cfront[s.edg[j]];
        }
        v = s.left.lab[s.splitfrom[i]];
        sum2 = xor2 = 0;
        for j = s.adj[v]; j < s.adj[v+1]; j++ {
            sum2 += s.left.cfront[s.edg[j]];
            xor2 ^= s.left.cfront[s.edg[j]];
        }
        if (sum1 != sum2) || (xor1 != xor2) {
            return 0; }
    }   

    return 1;
}

static int
descend(struct saucy *s, struct coloring *c, int target, int min)
{
    var back int32  = target + c.clen[target];

    /* Count this node */
    ++s.stats.nodes;

    /* Move the minimum label to the back */
    swap_labels(c, min, back);

    /* Split the cell */
    s.difflev[s.lev] = s.ndiffs;
    s.undifflev[s.lev] = s.nundiffs;
    ++s.lev;
    s.split(s, c, target, back);   

    /* Now go and do some work */
    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);    
    if !refineByDepGraph(s, c) { return 0; }

    /* if we are looking for a Boolean matching, check the OPP and 
     * backtrack if the OPP maps part of one network to itself */
    if s.fBooleanMatching && !check_OPP_for_Boolean_matching(s, c) { return 0; }

    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);
    if REFINE_BY_SIM_1 && !s.refineBySim1(s, c) { return 0; }

    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);
    if REFINE_BY_SIM_2 && !s.refineBySim2(s, c) { return 0; }

    /* do the check once more, maybe the check fails, now that refinement is complete */
    if s.fBooleanMatching && !check_OPP_for_Boolean_matching(s, c) { return 0; }

    if s.fLookForSwaps && !check_OPP_only_has_swaps(s, c) { return 0; }

    if !double_check_OPP_isomorphism(s, c) { return 0; }
    
    return 1;
}

static int
select_smallest_max_connected_cell(struct saucy *s, int start, int end)
{
    var smallest_cell int32 = -1;
    var cell int32;
    var smallest_cell_size int32  = s.n;
    var max_connections int32  = -1;
    var connection_list *int32  = zeros(s.n);
    
    cell = start;
    for  !s.left.clen[cell]  { cell++; }
    for  cell < end  {       
        if s.left.clen[cell] <= smallest_cell_size {
            var i, connections int32;;
            for i = s.depAdj[s.left.lab[cell]]; i < s.depAdj[s.left.lab[cell]+1]; i++ {
                if !connection_list[s.depEdg[i]] {
                    connections++;
                    connection_list[s.depEdg[i]] = 1;
                }
            }
            if (s.left.clen[cell] < smallest_cell_size) || (connections > max_connections) {
                smallest_cell_size = s.left.clen[cell];
                max_connections = connections;
                smallest_cell = cell;
            }
            for i = s.depAdj[s.left.lab[cell]]; i < s.depAdj[s.left.lab[cell]+1]; i++ {
                connection_list[s.depEdg[i]] = 0;
            }
        }
        cell = s.nextnon[cell];
    }
    
    // inlining of macro ABC_FREE
    ((connection_list) ? (free((char *) (connection_list)), (connection_list) = 0) : 0);
    return smallest_cell;
}

static int
descend_leftmost(struct saucy *s)
{
    var target, min int32;

    /* Keep going until we're discrete */
    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);
    for !at_terminal(s) {
        //target = min = s->nextnon[-1];
        if s.nextnon[-1] < Abc_NtkPoNum(s.pNtk) {
            target = min = select_smallest_max_connected_cell(s, s.nextnon[-1], Abc_NtkPoNum(s.pNtk));            
        } else {
            target = min = select_smallest_max_connected_cell(s, Abc_NtkPoNum(s.pNtk), s.n); }
        if s.fPrintTree { 
            fmt.Printf("%s->%s\n", getVertexName(s.pNtk, s.left.lab[min]), getVertexName(s.pNtk, s.left.lab[min])); }     
        s.splitvar[s.lev] = s.left.lab[min];
        s.start[s.lev] = target;
        s.splitlev[s.lev] = s.nsplits;
        if !descend(s, &s.left, target, min) { return 0; }       
    }
    s.splitlev[s.lev] = s.n;
    return 1;
}

/*
 * If the remaining nonsingletons in this partition match exactly
 * those nonsingletons from the leftmost branch of the search tree,
 * then there is no point in continuing descent.
 */

static int
zeta_fixed(struct saucy *s)
{
    return s.ndiffs == s.nundiffs;
}

static void
select_dynamically(struct saucy *s, int *target, int *lmin, int *rmin)
{
    /* Both clens are equal; this clarifies the code a bit */
    int *clen = s.left.clen;
    var i, k int32;
    //int cf;

    /*
     * If there's a pair, use it.  pairs[0] should always work,
     * but we use a checked loop instead because I'm not 100% sure
     * I'm "unpairing" at every point I should be.
     */
    for i = 0; i < s.npairs; ++i {
        k = s.pairs[i];
        *target = s.right.cfront[k];
        *lmin = s.left.unlab[s.right.lab[s.left.unlab[k]]];
        *rmin = s.right.unlab[k];

        if clen[*target]
                && in_cell_range(&s.left, *lmin, *target)
                && in_cell_range(&s.right, *rmin, *target) {
            return; }
    }

    /* Diffnons is only consistent when there are no baddies */
    /*if (s->ndiffnons != -1) {
        *target = *lmin = *rmin = s->right.cfront[s->diffnons[0]];
        return;
    }*/

    /* Pick any old target cell and element */
    /*for (i = 0; i < s->ndiffs; ++i) {
        cf = s->right.cfront[s->diffs[i]];
        if (clen[cf]) {
            *lmin = *rmin = *target = cf;
            return;
        }
    }*/

    for i = 0; i < s.n; i += (clen[i]+1) {
        if !clen[i] { continue; }
        *rmin = *lmin = *target = i;
        if s.right.cfront[s.left.lab[*lmin]] == *target {
            *rmin = s.right.unlab[s.left.lab[*lmin]]; }
        return;
    }   
    
    /* we should never get here */
    abort();
}

static void
select_statically(struct saucy *s, int *target, int *lmin, int *rmin)
{
    var i int32;

    *target = *rmin = s.left.cfront[s.splitvar[s.lev]];
    *lmin = s.left.unlab[s.splitvar[s.lev]];
    /* try to map identically! */
    for i = *rmin; i <= (*rmin + s.right.clen[*target]); i++ {
        if s.right.lab[*rmin] == s.left.lab[*lmin] {
            *rmin = i;
            break;
        }
    }
}

static int
descend_left(struct saucy *s)
{
    var target, lmin, rmin int32;

    /* Check that we ended at the right spot */
    if s.nsplits != s.splitlev[s.lev] { return 0; }

    /* Keep going until we're discrete */
    for !at_terminal(s) /*&& !zeta_fixed(s)*/ {

        /* We can pick any old target cell and element */
        s.select_decomposition(s, &target, &lmin, &rmin);

        if s.fPrintTree {
            //printf("in level %d: %d->%d\n", s->lev, s->left.lab[lmin], s->right.lab[rmin]);
            fmt.Printf("in level %d: %s->%s\n", s.lev, getVertexName(s.pNtk, s.left.lab[lmin]), getVertexName(s.pNtk, s.right.lab[rmin]));
        }

        /* Check if we need to refine on the left */
        s.match = 0;
        s.start[s.lev] = target;
        s.split = split_left;
        if SELECT_DYNAMICALLY {
            s.refineBySim1 = refineBySim1_left;
            s.refineBySim2 = refineBySim2_left;
        }
        descend(s, &s.left, target, lmin);
        s.splitlev[s.lev] = s.nsplits;
        s.split = split_other;     
        if SELECT_DYNAMICALLY {
            s.refineBySim1 = refineBySim1_other;
            s.refineBySim2 = refineBySim2_other;
        }
        --s.lev;
        s.nsplits = s.splitlev[s.lev];

        /* Now refine on the right and ensure matching */
        s.specmin[s.lev] = s.right.lab[rmin];
        if !descend(s, &s.right, target, rmin) { return 0; }
        if s.nsplits != s.splitlev[s.lev] { return 0; }
    }
    return 1;
}

func find_representative(k int32, theta *int32) int32 {
    var rep, tmp int32;

    /* Find the minimum cell representative */
    for rep = k; rep != theta[rep]; rep = theta[rep] {}

    /* Update all intermediaries */
    for theta[k] != rep {
        tmp = theta[k]; theta[k] = rep; k = tmp;
    }
    return rep;
}

static void
update_theta(struct saucy *s)
{
    var i, k, x, y, tmp int32;

    for i = 0; i < s.ndiffs; ++i {
        k = s.unsupp[i];
        x = find_representative(k, s.theta);
        y = find_representative(s.gamma[k], s.theta);

        if x != y {
            if x > y {
                tmp = x;
                x = y;
                y = tmp;
            }
            s.theta[y] = x;
            s.thsize[x] += s.thsize[y];

            s.thnext[s.thprev[y]] = s.thnext[y];
            s.thprev[s.thnext[y]] = s.thprev[y];
            s.threp[s.thfront[y]] = s.thnext[y];
        }
    }
}

static int
theta_prune(struct saucy *s)
{
    var start int32  = s.start[s.lev];
    var label, rep, irep int32;

    irep = find_representative(s.indmin, s.theta);
    for s.kanctar {
        label = s.anctar[--s.kanctar];
        rep = find_representative(label, s.theta);
        if rep == label && rep != irep {
            return s.right.unlab[label] - start;
        }
    }
    return -1;
}

static int
orbit_prune(struct saucy *s)
{
    var i int32;
    var label int32;
    var fixed int32;
    var min int32 = -1;
    var k int32  = s.start[s.lev];
    var size int32  = s.right.clen[k] + 1;
    var cell *int32  = s.right.lab + k;

    /* The previously fixed value */
    fixed = cell[size-1];

    /* Look for the next minimum cell representative */
    for i = 0; i < size-1; ++i {
        label = cell[i];

        /* Skip things we've already considered */
        if label <= fixed { continue; }

        /* Skip things that we'll consider later */
        if min != -1 && label > cell[min] { continue; }

        /* New candidate minimum */
        min = i;
    }

    return min;
}

static void
note_anctar_reps(struct saucy *s)
{
    var i, j, k, m, f, rep, tmp int32;

    /*
     * Undo the previous level's splits along leftmost so that we
     * join the appropriate lists of theta reps.
     */
    for i = s.splitlev[s.anc+1]-1; i >= s.splitlev[s.anc]; --i {
        f = s.splitfrom[i];
        j = s.threp[f];
        k = s.threp[s.splitwho[i]];

        s.thnext[s.thprev[j]] = k;
        s.thnext[s.thprev[k]] = j;

        tmp = s.thprev[j];
        s.thprev[j] = s.thprev[k];
        s.thprev[k] = tmp;

        for m = k; m != j; m = s.thnext[m] {
            s.thfront[m] = f;
        }
    }

    /*
     * Just copy over the target's reps and sort by cell size, in
     * the hopes of trimming some otherwise redundant generators.
     */
    s.kanctar = 0;
    s.anctar[s.kanctar++] = rep = s.threp[s.start[s.lev]];
    for k = s.thnext[rep]; k != rep; k = s.thnext[k] {
        s.anctar[s.kanctar++] = k;
    }
    array_indirect_sort(s.anctar, s.thsize, s.kanctar);
}

static void
multiply_index(struct saucy *s, int k)
{
    if (s.stats.grpsize_base *= k) > 1e10 {
        s.stats.grpsize_base /= 1e10;
        s.stats.grpsize_exp += 10;
    }
}

static int
backtrack_leftmost(struct saucy *s)
{
    var rep int32  = find_representative(s.indmin, s.theta);
    var repsize int32  = s.thsize[rep];
    var min int32  = -1;

    pick_all_the_pairs(s);
    clear_undiffnons(s);
    s.ndiffs = s.nundiffs = s.npairs = s.ndiffnons = 0;

    if repsize != s.right.clen[s.start[s.lev]]+1 {
        min = theta_prune(s);
    }

    if min == -1 {
        multiply_index(s, repsize);
    }

    return min;
}

static int
backtrack_other(struct saucy *s)
{
    var cf int32  = s.start[s.lev];
    var cb int32  = cf + s.right.clen[cf];
    var spec int32  = s.specmin[s.lev];
    var min int32;

    /* Avoid using pairs until we get back to leftmost. */
    pick_all_the_pairs(s);

    clear_undiffnons(s);

    s.npairs = s.ndiffnons = -1;

    if s.right.lab[cb] == spec {
        min = find_min(s, cf);
        if min == cb {
            min = orbit_prune(s);
        
        } else {
            min -= cf;
        }
    
    } else {
        min = orbit_prune(s);
        if min != -1 && s.right.lab[min + cf] == spec {
            swap_labels(&s.right, min + cf, cb);
            min = orbit_prune(s);
        }
    }
    return min;
}

static void
rewind_coloring(struct saucy *s, struct coloring *c, int lev)
{
    var i int32;
    var cf int32;
    var ff int32;
    var splits int32 = s.splitlev[lev];
    for i = s.nsplits - 1; i >= splits; --i {
        cf = s.splitfrom[i];
        ff = s.splitwho[i];
        c.clen[cf] += c.clen[ff] + 1;
        fix_fronts(c, cf, ff);
    }
}

static void
rewind_simulation_vectors(struct saucy *s, int lev)
{   
    var i int32;
    for i = s.randomVectorSplit_sim1[lev]; i < Vec_PtrSize(s.randomVectorArray_sim1); i++ {
        Vec_IntFree((Vec_Int_t *)Vec_PtrEntry(s.randomVectorArray_sim1, i));
    }
    Vec_PtrShrink(s.randomVectorArray_sim1, s.randomVectorSplit_sim1[lev]);

    for i = s.randomVectorSplit_sim2[lev]; i < Vec_PtrSize(s.randomVectorArray_sim2); i++ {                      
        Vec_IntFree((Vec_Int_t *)Vec_PtrEntry(s.randomVectorArray_sim2, i));
    }    
    Vec_PtrShrink(s.randomVectorArray_sim2, s.randomVectorSplit_sim2[lev]);
}

static int
do_backtrack(struct saucy *s)
{
    var i, cf, cb int32;

    /* Undo the splits up to this level */
    rewind_coloring(s, &s.right, s.lev);
    s.nsplits = s.splitlev[s.lev];

    /* Rewind diff information */
    for i = s.ndiffs - 1; i >= s.difflev[s.lev]; --i {
        s.diffmark[s.diffs[i]] = 0;
    }
    s.ndiffs = s.difflev[s.lev];
    s.nundiffs = s.undifflev[s.lev];

    /* Point to the target cell */
    cf = s.start[s.lev];
    cb = cf + s.right.clen[cf];

    /* Update ancestor with zeta if we've rewound more */
    if s.anc > s.lev {
        s.anc = s.lev;
        s.indmin = s.left.lab[cb];
        s.match = 1;
        note_anctar_reps(s);
    }

    /* Perform backtracking appropriate to our location */
    if s.lev == s.anc {
        return backtrack_leftmost(s);
    } else {
        return backtrack_other(s);
    }
}

static int
backtrack_loop(struct saucy *s)
{
    var min int32;

    /* Backtrack as long as we're exhausting target cells */
    for --s.lev; s.lev; --s.lev {
        min = do_backtrack(s);
        if min != -1 { return min + s.start[s.lev]; }
    }
    return -1;
}

static int
backtrack(struct saucy *s)
{
    var min, old, tmp int32;
    old = s.nsplits;
    min = backtrack_loop(s);
    tmp = s.nsplits;
    s.nsplits = old;
    rewind_coloring(s, &s.left, s.lev+1); 
    s.nsplits = tmp;
    if SELECT_DYNAMICALLY {
        rewind_simulation_vectors(s, s.lev+1); }

    return min;
}

static int
backtrack_bad(struct saucy *s)
{
    var min, old, tmp int32;
    old = s.lev;
    min = backtrack_loop(s);
    if BACKTRACK_BY_SAT {
        var oldLev int32  = s.lev;
        for !backtrackBysatCounterExamples(s, &s.right) {                          
            min = backtrack_loop(s);            
            if !s.lev {
                if s.fPrintTree {
                    fmt.Printf("Backtrack by SAT from level %d to %d\n", oldLev, 0); }
                return -1;
            }
        }
        if s.fPrintTree {          
            if s.lev < oldLev { 
                fmt.Printf("Backtrack by SAT from level %d to %d\n", oldLev, s.lev); } }
    }
    tmp = s.nsplits;
    s.nsplits = s.splitlev[old];
    rewind_coloring(s, &s.left, s.lev+1); 
    s.nsplits = tmp;
    if SELECT_DYNAMICALLY {
        rewind_simulation_vectors(s, s.lev+1); }

    return min;
}

void
prepare_permutation_ntk(struct saucy *s) 
{       
    var i int32;
    var pObj, pObjPerm *Abc_Obj_t;
    var numouts int32  = Abc_NtkPoNum(s.pNtk);

    Nm_ManFree( s.pNtk_permuted.pManName );
    s.pNtk_permuted.pManName = Nm_ManCreate( Abc_NtkCiNum(s.pNtk) + Abc_NtkCoNum(s.pNtk) + Abc_NtkBoxNum(s.pNtk) );    

    for i = 0; i < s.n; ++i {        
        if i < numouts {
            pObj     = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk.vPos, i);
            pObjPerm = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk_permuted.vPos, s.gamma[i]);           
        
        } else {          
            pObj     = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk.vPis, i - numouts);
            pObjPerm = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk_permuted.vPis, s.gamma[i] - numouts);
            
        }
        Abc_ObjAssignName( pObjPerm, Abc_ObjName(pObj), nil );         
    }

    Abc_NtkOrderObjsByName( s.pNtk_permuted, 1 );

    /* print the permutation */
    /*for (i = 0; i < s->ndiffs; ++i)
        printf(" %d->%d", s->unsupp[i], s->diffs[i]);
    printf("\n");
    Abc_NtkForEachCo( s->pNtk, pObj, i )
        printf (" %d", Abc_ObjId(pObj)-1-Abc_NtkPiNum(s->pNtk));
    Abc_NtkForEachCi( s->pNtk, pObj, i )
        printf (" %d", Abc_ObjId(pObj)-1+Abc_NtkPoNum(s->pNtk));
    printf("\n");
    Abc_NtkForEachCo( s->pNtk_permuted, pObj, i )
        printf (" %d", Abc_ObjId(pObj)-1-Abc_NtkPiNum(s->pNtk_permuted));
    Abc_NtkForEachCi( s->pNtk_permuted, pObj, i )
        printf (" %d", Abc_ObjId(pObj)-1+Abc_NtkPoNum(s->pNtk_permuted));   
    printf("\n");*/
}


static void
prepare_permutation(struct saucy *s)
{
    var i, k int32;
    for i = 0; i < s.ndiffs; ++i {
        k = s.right.unlab[s.diffs[i]];
        s.unsupp[i] = s.left.lab[k];
        s.gamma[s.left.lab[k]] = s.right.lab[k];
    }
    prepare_permutation_ntk(s);
}

void
unprepare_permutation_ntk(struct saucy *s) 
{       
    var i int32;
    var pObj, pObjPerm *Abc_Obj_t;
    var numouts int32  = Abc_NtkPoNum(s.pNtk);

    Nm_ManFree( s.pNtk_permuted.pManName );
    s.pNtk_permuted.pManName = Nm_ManCreate( Abc_NtkCiNum(s.pNtk) + Abc_NtkCoNum(s.pNtk) + Abc_NtkBoxNum(s.pNtk) );    

    for i = 0; i < s.n; ++i {        
        if i < numouts {
            pObj     = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk.vPos, s.gamma[i]);
            pObjPerm = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk_permuted.vPos, i);         
        
        } else {          
            pObj     = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk.vPis, s.gamma[i] - numouts);
            pObjPerm = (Abc_Obj_t *)Vec_PtrEntry(s.pNtk_permuted.vPis, i - numouts);
            
        }
        Abc_ObjAssignName( pObjPerm, Abc_ObjName(pObj), nil );         
    }

    Abc_NtkOrderObjsByName( s.pNtk_permuted, 1 );
}


static void
unprepare_permutation(struct saucy *s)  
{
    var i int32;
    unprepare_permutation_ntk(s);
    for i = 0; i < s.ndiffs; ++i {
        s.gamma[s.unsupp[i]] = s.unsupp[i];
    }   
}

static int
do_search(struct saucy *s)
{
    var min int32;

    unprepare_permutation(s);

    /* Backtrack to the ancestor with zeta */
    if s.lev > s.anc { s.lev = s.anc + 1; }

    /* Perform additional backtracking */
    min = backtrack(s);

    if s.fBooleanMatching && (s.stats.grpsize_base > 1 || s.stats.grpsize_exp > 0) {
        return 0; }

    if s.fPrintTree && s.lev > 0 {
        //printf("in level %d: %d->%d\n", s->lev, s->left.lab[s->splitwho[s->nsplits]], s->right.lab[min]);
        fmt.Printf("in level %d: %s->%s\n", s.lev, getVertexName(s.pNtk, s.left.lab[s.splitwho[s.nsplits]]), getVertexName(s.pNtk, s.right.lab[min]));
    }   

    /* Keep going while there are tree nodes to expand */
    for s.lev {

        /* Descend to a new leaf node */    
        if descend(s, &s.right, s.start[s.lev], min)
                && descend_left(s) {

            /* Prepare permutation */
            prepare_permutation(s);

            /* If we found an automorphism, return it */
            if s.is_automorphism(s) {
                ++s.stats.gens;
                s.stats.support += s.ndiffs;
                update_theta(s);                
                s.print_automorphism(s.gFile, s.n, s.gamma, s.ndiffs, s.unsupp, s.marks, s.pNtk);
                unprepare_permutation(s);               
                return 1;
            
            } else {
                unprepare_permutation(s);
            }
        }

        /* If we get here, something went wrong; backtrack */
        ++s.stats.bads;
        min = backtrack_bad(s);
        if s.fPrintTree {
            fmt.Printf("BAD NODE\n");
            if s.lev > 0 {
                //printf("in level %d: %d->%d\n", s->lev, s->left.lab[s->splitwho[s->nsplits]], s->right.lab[min]);     
                fmt.Printf("in level %d: %s->%s\n", s.lev, getVertexName(s.pNtk, s.left.lab[s.splitwho[s.nsplits]]), getVertexName(s.pNtk, s.right.lab[min]));                               
            }
        }
    }

    /* Normalize group size */
    for s.stats.grpsize_base >= 10.0 {
        s.stats.grpsize_base /= 10;
        ++s.stats.grpsize_exp;
    }
    return 0;
}

void
saucy_search(
    Abc_Ntk_t * pNtk,
    struct saucy *s,
    int directed,
    int *colors,
    struct saucy_stats *stats)
{
    var i, j, max int32;  
    struct saucy_graph *g;

    extern Abc_Ntk_t * Abc_NtkDup( Abc_Ntk_t * pNtk );

    /* Save network information */
    s.pNtk = pNtk;
    s.pNtk_permuted = Abc_NtkDup( pNtk );  

    /* Builde dependency graph */
    g = buildDepGraph(pNtk, s.iDep, s.oDep);  

    /* Save graph information */
    s.n = g.n;
    s.depAdj = g.adj;
    s.depEdg = g.edg;
    /*s->dadj = g->adj + g->n + 1;
    s->dedg = g->edg + g->e;*/

    /* Save client information */
    s.stats = stats;       

    /* Polymorphism */
    if directed {
        s.is_automorphism = is_directed_automorphism;
        s.ref_singleton = ref_singleton_directed;
        s.ref_nonsingle = ref_nonsingle_directed;
    
    } else {
        s.is_automorphism = is_undirected_automorphism;
        s.ref_singleton = ref_singleton_undirected;
        s.ref_nonsingle = ref_nonsingle_undirected;
    }

    /* Initialize scalars */
    s.indmin = 0;  
    s.lev = s.anc = 1;
    s.ndiffs = s.nundiffs = s.ndiffnons = 0;
    s.activityInc = 1;

    /* The initial orbit partition is discrete */
    for i = 0; i < s.n; ++i {
        s.theta[i] = i;
    }

    /* The initial permutation is the identity */
    for i = 0; i < s.n; ++i {
        s.gamma[i] = i;
    }

    /* Initially every cell of theta has one element */
    for i = 0; i < s.n; ++i {
        s.thsize[i] = 1;
    }

    /* Every theta rep list is singleton */
    for i = 0; i < s.n; ++i {
        s.thprev[i] = s.thnext[i] = i;
    }

    /* We have no pairs yet */
    s.npairs = 0;
    for i = 0; i < s.n; ++i {
        s.unpairs[i] = -1;
    }

    /* Ensure no stray pointers in undiffnons, which is checked by removed_diffnon() */
    for i = 0; i < s.n; ++i {
        s.undiffnons[i] = -1;
    }

    /* Initialize stats */
    s.stats.grpsize_base = 1.0;
    s.stats.grpsize_exp = 0;
    s.stats.nodes = 1;
    s.stats.bads = s.stats.gens = s.stats.support = 0;

    /* Prepare for refinement */
    s.nninduce = s.nsinduce = 0;
    s.csize = 0;

    /* Count cell sizes */
    for i = 0; i < s.n; ++i {
        s.ccount[colors[i]]++;
        if max < colors[i] { max = colors[i]; }
    }
    s.nsplits = max + 1;

    /* Build cell lengths */
    s.left.clen[0] = s.ccount[0] - 1;
    for i = 0; i < max; ++i {
        s.left.clen[s.ccount[i]] = s.ccount[i+1] - 1;
        s.ccount[i+1] += s.ccount[i];
    }

    /* Build the label array */
    for i = 0; i < s.n; ++i {
        set_label(&s.left, --s.ccount[colors[i]], i);
    }

    /* Clear out ccount */
    for i = 0; i <= max; ++i {
        s.ccount[i] = 0;
    }

    /* Update refinement stuff based on initial partition */
    for i = 0; i < s.n; i += s.left.clen[i]+1 {
        add_induce(s, &s.left, i);
        fix_fronts(&s.left, i, i);
    }

    /* Prepare lists based on cell lengths */
    for i, j = 0, -1; i < s.n; i += s.left.clen[i] + 1 {
        if !s.left.clen[i] { continue; }
        s.prevnon[i] = j;
        s.nextnon[j] = i;
        j = i;
    }

    /* Fix the end */
    s.prevnon[s.n] = j;
    s.nextnon[j] = s.n;

    /* Preprocessing after initial coloring */
    s.split = split_init;  
    s.refineBySim1 = refineBySim1_init;
    s.refineBySim2 = refineBySim2_init;    

    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);
    fmt.Printf("Initial Refine by Dependency graph ... ");
    refineByDepGraph(s, &s.left);
    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);
    fmt.Printf("done!\n");
    
    fmt.Printf("Initial Refine by Simulation ... ");
    if REFINE_BY_SIM_1 { s.refineBySim1(s, &s.left); }
    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);
    if REFINE_BY_SIM_2 { s.refineBySim2(s, &s.left); }
    //print_partition(&s->left, NULL, s->n, s->pNtk, 1);
    fmt.Printf("done!\n\t--------------------\n");

    /* Descend along the leftmost branch and compute zeta */
    s.refineBySim1 = refineBySim1_left;
    s.refineBySim2 = refineBySim2_left;
    descend_leftmost(s);    
    s.split = split_other; 
    s.refineBySim1 = refineBySim1_other;
    s.refineBySim2 = refineBySim2_other;   

    /* Our common ancestor with zeta is the current level */
    s.stats.levels = s.anc = s.lev;

    /* Copy over this data to our non-leftmost coloring */
    memcpy(s.right.lab, s.left.lab, s.n * sizeof(int));
    memcpy(s.right.unlab, s.left.unlab, s.n * sizeof(int));
    memcpy(s.right.clen, s.left.clen, s.n * sizeof(int));
    memcpy(s.right.cfront, s.left.cfront, s.n * sizeof(int));

    /* The reps are just the labels at this point */
    memcpy(s.threp, s.left.lab, s.n * sizeof(int));
    memcpy(s.thfront, s.left.unlab, s.n * sizeof(int));

    /* choose cell selection method */
    if SELECT_DYNAMICALLY { s.select_decomposition = select_dynamically;
    } else {                    s.select_decomposition = select_statically; }

    /* Keep running till we're out of automorphisms */
    for do_search(s) {; }
}

void
saucy_free(struct saucy *s)
{
    var i int32;

    // inlining of macro ABC_FREE
    ((s.undiffnons) ? (free((char *) (s.undiffnons)), (s.undiffnons) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.diffnons) ? (free((char *) (s.diffnons)), (s.diffnons) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.unpairs) ? (free((char *) (s.unpairs)), (s.unpairs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.pairs) ? (free((char *) (s.pairs)), (s.pairs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.thfront) ? (free((char *) (s.thfront)), (s.thfront) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.threp) ? (free((char *) (s.threp)), (s.threp) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.thnext) ? (free((char *) (s.thnext)), (s.thnext) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.thprev) ? (free((char *) (s.thprev)), (s.thprev) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.specmin) ? (free((char *) (s.specmin)), (s.specmin) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.anctar) ? (free((char *) (s.anctar)), (s.anctar) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.thsize) ? (free((char *) (s.thsize)), (s.thsize) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.undifflev) ? (free((char *) (s.undifflev)), (s.undifflev) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.difflev) ? (free((char *) (s.difflev)), (s.difflev) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.diffs) ? (free((char *) (s.diffs)), (s.diffs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.diffmark) ? (free((char *) (s.diffmark)), (s.diffmark) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.conncnts) ? (free((char *) (s.conncnts)), (s.conncnts) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.unsupp) ? (free((char *) (s.unsupp)), (s.unsupp) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.splitlev) ? (free((char *) (s.splitlev)), (s.splitlev) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.splitfrom) ? (free((char *) (s.splitfrom)), (s.splitfrom) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.splitwho) ? (free((char *) (s.splitwho)), (s.splitwho) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.splitvar) ? (free((char *) (s.splitvar)), (s.splitvar) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.right.unlab) ? (free((char *) (s.right.unlab)), (s.right.unlab) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.right.lab) ? (free((char *) (s.right.lab)), (s.right.lab) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.left.unlab) ? (free((char *) (s.left.unlab)), (s.left.unlab) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.left.lab) ? (free((char *) (s.left.lab)), (s.left.lab) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.theta) ? (free((char *) (s.theta)), (s.theta) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.junk) ? (free((char *) (s.junk)), (s.junk) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.gamma) ? (free((char *) (s.gamma)), (s.gamma) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.start) ? (free((char *) (s.start)), (s.start) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.prevnon) ? (free((char *) (s.prevnon)), (s.prevnon) = 0) : 0);
    free(s.nextnon-1);
    // inlining of macro ABC_FREE
    ((s.clist) ? (free((char *) (s.clist)), (s.clist) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.ccount) ? (free((char *) (s.ccount)), (s.ccount) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.count) ? (free((char *) (s.count)), (s.count) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.bucket) ? (free((char *) (s.bucket)), (s.bucket) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.stuff) ? (free((char *) (s.stuff)), (s.stuff) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.right.clen) ? (free((char *) (s.right.clen)), (s.right.clen) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.right.cfront) ? (free((char *) (s.right.cfront)), (s.right.cfront) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.left.clen) ? (free((char *) (s.left.clen)), (s.left.clen) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.left.cfront) ? (free((char *) (s.left.cfront)), (s.left.cfront) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.indmark) ? (free((char *) (s.indmark)), (s.indmark) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.sinduce) ? (free((char *) (s.sinduce)), (s.sinduce) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.ninduce) ? (free((char *) (s.ninduce)), (s.ninduce) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.depAdj) ? (free((char *) (s.depAdj)), (s.depAdj) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.depEdg) ? (free((char *) (s.depEdg)), (s.depEdg) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.marks) ? (free((char *) (s.marks)), (s.marks) = 0) : 0);
    for i = 0; i < Abc_NtkPiNum(s.pNtk); i++ {
        Vec_IntFree( s.iDep[i] );
        Vec_IntFree( s.obs[i] );
        Vec_PtrFree( s.topOrder[i] );
    }
    for i = 0; i < Abc_NtkPoNum(s.pNtk); i++ {
        Vec_IntFree( s.oDep[i] );
        Vec_IntFree( s.ctrl[i] );
    }
    for i = 0; i < Vec_PtrSize(s.randomVectorArray_sim1); i++ {
        Vec_IntFree((Vec_Int_t *)Vec_PtrEntry(s.randomVectorArray_sim1, i));
    }
    for i = 0; i < Vec_PtrSize(s.randomVectorArray_sim2); i++ {
        Vec_IntFree((Vec_Int_t *)Vec_PtrEntry(s.randomVectorArray_sim2, i));
    }
    Vec_PtrFree( s.randomVectorArray_sim1 );
    Vec_PtrFree( s.randomVectorArray_sim2 );
    // inlining of macro ABC_FREE
    ((s.randomVectorSplit_sim1) ? (free((char *) (s.randomVectorSplit_sim1)), (s.randomVectorSplit_sim1) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.randomVectorSplit_sim2) ? (free((char *) (s.randomVectorSplit_sim2)), (s.randomVectorSplit_sim2) = 0) : 0);
    Abc_NtkDelete( s.pNtk_permuted );
    for i = 0; i < Vec_PtrSize(s.satCounterExamples); i++ {
        struct sim_result * cex = (struct sim_result *)Vec_PtrEntry(s.satCounterExamples, i);
        // inlining of macro ABC_FREE
        ((cex.inVec) ? (free((char *) (cex.inVec)), (cex.inVec) = 0) : 0);
        // inlining of macro ABC_FREE
        ((cex.outVec) ? (free((char *) (cex.outVec)), (cex.outVec) = 0) : 0);
        // inlining of macro ABC_FREE
        ((cex) ? (free((char *) (cex)), (cex) = 0) : 0);
    }
    Vec_PtrFree(s.satCounterExamples);
    // inlining of macro ABC_FREE
    ((s.pModel) ? (free((char *) (s.pModel)), (s.pModel) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.iDep) ? (free((char *) (s.iDep)), (s.iDep) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.oDep) ? (free((char *) (s.oDep)), (s.oDep) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.obs) ? (free((char *) (s.obs)), (s.obs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.ctrl) ? (free((char *) (s.ctrl)), (s.ctrl) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s.topOrder) ? (free((char *) (s.topOrder)), (s.topOrder) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s) ? (free((char *) (s)), (s) = 0) : 0);
}

func saucy_alloc(pNtk *Abc_Ntk_t) *struct saucy {
    var i int32;
    var numouts int32  = Abc_NtkPoNum(pNtk);
    var numins int32  =  Abc_NtkPiNum(pNtk);
    var n int32  = numins + numouts;
    // inlining of macro ABC_ALLOC
    struct saucy *s = ((struct saucy *) malloc(sizeof(struct saucy) * (size_t)(1)));
    if s == nil { return nil; }

    s.ninduce = ints(n);
    s.sinduce = ints(n);
    s.indmark = bits(n);
    s.left.cfront = zeros(n);
    s.left.clen = ints(n);
    s.right.cfront = zeros(n);
    s.right.clen = ints(n);
    s.stuff = bits(n+1);
    s.bucket = ints(n+2);
    s.count = ints(n+1);
    s.ccount = zeros(n);
    s.clist = ints(n);
    s.nextnon = ints(n+1) + 1;
    s.prevnon = ints(n+1);
    s.anctar = ints(n);
    s.start = ints(n);
    s.gamma = ints(n);
    s.junk = ints(n);
    s.theta = ints(n);
    s.thsize = ints(n);
    s.left.lab = ints(n);
    s.left.unlab = ints(n);
    s.right.lab = ints(n);
    s.right.unlab = ints(n);
    s.splitvar = ints(n);
    s.splitwho = ints(n);
    s.splitfrom = ints(n);
    s.splitlev = ints(n+1);
    s.unsupp = ints(n);
    s.conncnts = zeros(n);
    s.diffmark = bits(n);
    s.diffs = ints(n);
    s.difflev = ints(n);
    s.undifflev = ints(n);
    s.specmin = ints(n);
    s.thnext = ints(n);
    s.thprev = ints(n);
    s.threp = ints(n);
    s.thfront = ints(n);
    s.pairs = ints(n);
    s.unpairs = ints(n);
    s.diffnons = ints(n);
    s.undiffnons = ints(n);
    s.marks = bits(n);

    // inlining of macro ABC_ALLOC
    s.iDep = ((Vec_Int_t* *) malloc(sizeof(Vec_Int_t*) * (size_t)(numins)));
    // inlining of macro ABC_ALLOC
    s.oDep = ((Vec_Int_t* *) malloc(sizeof(Vec_Int_t*) * (size_t)(numouts)));
    // inlining of macro ABC_ALLOC
    s.obs  = ((Vec_Int_t* *) malloc(sizeof(Vec_Int_t*) * (size_t)(numins)));
    // inlining of macro ABC_ALLOC
    s.ctrl = ((Vec_Int_t* *) malloc(sizeof(Vec_Int_t*) * (size_t)(numouts)));

    for i = 0; i < numins; i++ {       
        s.iDep[i] = Vec_IntAlloc( 1 );
        s.obs[i] = Vec_IntAlloc( 1 );
    }
    for i = 0; i < numouts; i++ {      
        s.oDep[i] = Vec_IntAlloc( 1 );
        s.ctrl[i] = Vec_IntAlloc( 1 );
    }   

    s.randomVectorArray_sim1 = Vec_PtrAlloc( n );  
    s.randomVectorSplit_sim1 = zeros( n );
    s.randomVectorArray_sim2 = Vec_PtrAlloc( n );  
    s.randomVectorSplit_sim2= zeros( n );

    s.satCounterExamples = Vec_PtrAlloc( 1 );
    s.pModel = ints( numins );

    if s.ninduce && s.sinduce && s.left.cfront && s.left.clen
        && s.right.cfront && s.right.clen
        && s.stuff && s.bucket && s.count && s.ccount
        && s.clist && s.nextnon-1 && s.prevnon
        && s.start && s.gamma && s.theta && s.left.unlab
        && s.right.lab && s.right.unlab
        && s.left.lab &&  s.splitvar && s.splitwho && s.junk
        && s.splitfrom && s.splitlev && s.thsize
        && s.unsupp && s.conncnts && s.anctar
        && s.diffmark && s.diffs && s.indmark
        && s.thnext && s.thprev && s.threp && s.thfront
        && s.pairs && s.unpairs && s.diffnons && s.undiffnons
        && s.difflev && s.undifflev && s.specmin {
        return s;
    
    } else {
        saucy_free(s);
        return nil;
    }
}

static void
print_stats(FILE *f, struct saucy_stats stats )
{
    fmt.Fprintf(f, "group size = %fe%d\n", stats.grpsize_base, stats.grpsize_exp);
    fmt.Fprintf(f, "levels = %d\n", stats.levels);
    fmt.Fprintf(f, "nodes = %d\n", stats.nodes);
    fmt.Fprintf(f, "generators = %d\n", stats.gens);
    fmt.Fprintf(f, "total support = %d\n", stats.support);
    fmt.Fprintf(f, "average support = %.2f\n", float64((stats.support))/float64((stats.gens)));
    fmt.Fprintf(f, "nodes per generator = %.2f\n", float64((stats.nodes))/float64((stats.gens)));
    fmt.Fprintf(f, "bad nodes = %d\n", stats.bads);
}


/* From this point up are SAUCY functions*/
/////////////////////////////////////////////////////////////////////////////////////////////////////////
/* From this point down are new functions */

func getVertexName(pNtk *Abc_Ntk_t, v int32) string {   
    var pObj *Abc_Obj_t;
    var numouts int32  =  Abc_NtkPoNum(pNtk);

    if v < numouts {    
        pObj = (Abc_Obj_t *)Vec_PtrEntry(pNtk.vPos, v);
    } else {
        pObj = (Abc_Obj_t *)Vec_PtrEntry(pNtk.vPis, v - numouts); }       
    
    return Abc_ObjName(pObj);
}

func findTopologicalOrder(pNtk *Abc_Ntk_t) **Vec_Ptr_t {
    var vNodes **Vec_Ptr_t;
    var pObj, pFanout *Abc_Obj_t;
    var i, k int32;    
    
    extern void Abc_NtkDfsReverse_rec( Abc_Obj_t * pNode, Vec_Ptr_t * vNodes );
    
    /* start the array of nodes */
    // inlining of macro ABC_ALLOC
    vNodes = ((Vec_Ptr_t * *) malloc(sizeof(Vec_Ptr_t *) * (size_t)(Abc_NtkPiNum(pNtk))));
    for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
        vNodes[i] = Vec_PtrAlloc(50);
    }   
    
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        /* set the traversal ID */
        Abc_NtkIncrementTravId( pNtk );
        Abc_NodeSetTravIdCurrent( pObj );
        pObj = Abc_ObjFanout0Ntk(pObj);
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_NtkDfsReverse_rec( pFanout, vNodes[i] );
        }
    }
   
    return vNodes;
}

func getDependencies(pNtk *Abc_Ntk_t, iDep **Vec_Int_t, oDep **Vec_Int_t) {   
    var vSuppFun *Vec_Ptr_t;
    var i, j int32;   
    
    vSuppFun = Sim_ComputeFunSupp(pNtk, 0);
    for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
        var seg string  = (char *)vSuppFun.pArray[i];
        
        for j = 0; j < Abc_NtkPiNum(pNtk); j+=8 {
            if ((*seg) & 0x01) == 0x01 {
                Vec_IntPushOrder(oDep[i], j); }
            if ((*seg) & 0x02) == 0x02 {
                Vec_IntPushOrder(oDep[i], j+1); }
            if ((*seg) & 0x04) == 0x04 {
                Vec_IntPushOrder(oDep[i], j+2); }
            if ((*seg) & 0x08) == 0x08 {
                Vec_IntPushOrder(oDep[i], j+3); }
            if ((*seg) & 0x10) == 0x10 {
                Vec_IntPushOrder(oDep[i], j+4); }
            if ((*seg) & 0x20) == 0x20 {
                Vec_IntPushOrder(oDep[i], j+5); }
            if ((*seg) & 0x40) == 0x40 {
                Vec_IntPushOrder(oDep[i], j+6); }
            if ((*seg) & 0x80) == 0x80 {
                Vec_IntPushOrder(oDep[i], j+7); }

            seg++;
        }
    }

    for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
        for j = 0; j < Vec_IntSize(oDep[i]); j++ {
            Vec_IntPush(iDep[Vec_IntEntry(oDep[i], j)], i);
        }
    } 
    

    /*for(i = 0; i < Abc_NtkPoNum(pNtk); i++)
    {
        printf("Output %d: ", i);
        for(j = 0; j < Vec_IntSize(oDep[i]); j++)
            printf("%d ", Vec_IntEntry(oDep[i], j));
        printf("\n");
    }

    printf("\n");

    for(i = 0; i < Abc_NtkPiNum(pNtk); i++)
    {
        printf("Input %d: ", i);
        for(j = 0; j < Vec_IntSize(iDep[i]); j++)
            printf("%d ", Vec_IntEntry(iDep[i], j));
        printf("\n");
    }

    printf("\n");   */  
}

func getDependenciesDummy(pNtk *Abc_Ntk_t, iDep **Vec_Int_t, oDep **Vec_Int_t) {       
    var i, j int32;   
    
    /* let's assume that every output is dependent on every input */
    for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
        for j = 0; j < Abc_NtkPiNum(pNtk); j++ {
            Vec_IntPush(oDep[i], j);
        }
    }

    for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
        for j = 0; j < Abc_NtkPoNum(pNtk); j++ {
            Vec_IntPush(iDep[i], j);
        }
    }
}

func buildDepGraph(pNtk *Abc_Ntk_t, iDep **Vec_Int_t, oDep **Vec_Int_t) *struct saucy_graph {       
    var i, j, k int32;    
    struct saucy_graph *g = nil;
    var n int32;
    var e int32;
    var adj *int32;
    var edg *int32;   

    n = Abc_NtkPoNum(pNtk) + Abc_NtkPiNum(pNtk);
    for e, i = 0, 0; i < Abc_NtkPoNum(pNtk); i++ {
        e += Vec_IntSize(oDep[i]);
    }  

    // inlining of macro ABC_ALLOC
    g = ((struct saucy_graph *) malloc(sizeof(struct saucy_graph) * (size_t)(1)));
    adj = zeros(n+1);
    edg = ints(2*e);    

    g.n = n;
    g.e = e;
    g.adj = adj;
    g.edg = edg;   

    adj[0] = 0;
    for i = 0; i < n; i++ {
        /* first add outputs and then inputs */
        if i < Abc_NtkPoNum(pNtk) {          
            adj[i+1] = adj[i] + Vec_IntSize(oDep[i]);
            for k = 0, j = adj[i]; j < adj[i+1]; j, k = j+1, k+1 {
                edg[j] = Vec_IntEntry(oDep[i], k) + Abc_NtkPoNum(pNtk);
            }
        
        } else {          
            adj[i+1] = adj[i] + Vec_IntSize(iDep[i-Abc_NtkPoNum(pNtk)]);
            for k = 0, j = adj[i]; j < adj[i+1]; j, k = j+1, k+1 {
                edg[j] = Vec_IntEntry(iDep[i-Abc_NtkPoNum(pNtk)], k);
            }
        }
    }

    /* print graph for testing */
    /*for (i = 0; i < n; i++) {
        printf("%d: ", i);
        for (j = adj[i]; j < adj[i+1]; j++)
            printf("%d ", edg[j]);
        printf("\n");
    }*/

    return g;
}

static Vec_Int_t * 
assignRandomBitsToCells(Abc_Ntk_t * pNtk, struct coloring *c)
{
    var randVec *Vec_Int_t  = Vec_IntAlloc( 1 );
    var i, bit int32;

    for i = 0; i < Abc_NtkPiNum(pNtk); i += (c.clen[i+Abc_NtkPoNum(pNtk)]+1) {
        bit = int32((SIM_RANDOM_UNSIGNED % 2));
        Vec_IntPush(randVec, bit);
    }

    return randVec;
}

static int * 
generateProperInputVector( Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randomVector )
{
    var vPiValues *int32;    
    var i, j, k, bit, input int32;
    var numouts int32  =  Abc_NtkPoNum(pNtk);
    var numins int32  =  Abc_NtkPiNum(pNtk);
    var n int32  = numouts + numins;

    // inlining of macro ABC_ALLOC
    vPiValues = ((int *) malloc(sizeof(int) * (size_t)(numins)));   

    for i, k = numouts, 0; i < n; i += (c.clen[i]+1), k++ {
        if k == Vec_IntSize(randomVector) { break; }

        bit = Vec_IntEntry(randomVector, k);
        for j = i; j <= (i + c.clen[i]); j++ {
            input = c.lab[j] - numouts;
            vPiValues[input] = bit;
        }
    }

    //if (k != Vec_IntSize(randomVector)) {
    if i < n {
        // inlining of macro ABC_FREE
        ((vPiValues) ? (free((char *) (vPiValues)), (vPiValues) = 0) : 0);
        return nil;
    }

    return vPiValues;
}

static int
ifInputVectorsAreConsistent( struct saucy * s, int * leftVec, int * rightVec )
{
    /* This function assumes that left and right partitions are isomorphic */
    var i, j int32;
    var lab int32;
    var left_bit, right_bit int32;
    var numouts int32  =  Abc_NtkPoNum(s.pNtk);
    var n int32  = numouts + Abc_NtkPiNum(s.pNtk);

    for i = numouts; i < n; i += (s.right.clen[i]+1) {       
        lab = s.left.lab[i] - numouts;
        left_bit = leftVec[lab];
        for j = i+1; j <= (i + s.right.clen[i]); j++ {
            lab = s.left.lab[j] - numouts;         
            if left_bit != leftVec[lab] { return -1; }
        }       
        
        lab = s.right.lab[i] - numouts;
        right_bit = rightVec[lab];
        for j = i+1; j <= (i + s.right.clen[i]); j++ {
            lab = s.right.lab[j] - numouts;            
            if right_bit != rightVec[lab] { return 0; }           
        }

        if left_bit != right_bit { 
             return 0; }
    }

    return 1;
}

static int
ifOutputVectorsAreConsistent( struct saucy * s, int * leftVec, int * rightVec )
{
    /* This function assumes that left and right partitions are isomorphic */
    var i, j int32;
    var count1, count2 int32;

    for i = 0; i < Abc_NtkPoNum(s.pNtk); i += (s.right.clen[i]+1) {     
        count1 = count2 = 0;
        for j = i; j <= (i + s.right.clen[i]); j++ {         
            if leftVec[s.left.lab[j]] { count1++; }
            if rightVec[s.right.lab[j]] { count2++; }
        }

        if count1 != count2 { return 0; }
    }

    return 1;
}

static struct saucy_graph *
buildSim1Graph( Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randVec, Vec_Int_t ** iDep, Vec_Int_t ** oDep )
{
    var i, j, k int32;
    struct saucy_graph *g;
    var n int32;
    var e int32;
    var adj *int32;
    var edg *int32;
    var vPiValues, output *int32;
    var numOneOutputs int32;
    var numouts int32  =  Abc_NtkPoNum(pNtk);
    var numins int32  = Abc_NtkPiNum(pNtk);

    vPiValues = generateProperInputVector(pNtk, c, randVec);
    if vPiValues == nil { 
        return nil; }

    output = Abc_NtkVerifySimulatePattern(pNtk, vPiValues);

    for i = 0; i < numouts; i++ {
        if output[i] {
            numOneOutputs++; }
    }

    // inlining of macro ABC_ALLOC
    g = ((struct saucy_graph *) malloc(sizeof(struct saucy_graph) * (size_t)(1)));
    n = numouts + numins;
    e = numins * numOneOutputs;
    adj = ints(n+1);
    edg = ints(2*e);        
    g.n = n;
    g.e = e;
    g.adj = adj;
    g.edg = edg;   

    adj[0] = 0;
    for i = 0; i < numouts; i++ {
        if output[i] {
            adj[i+1] = adj[i] + Vec_IntSize(oDep[i]);
            for j = adj[i], k = 0; j < adj[i+1]; j, k = j+1, k+1 {
                edg[j] = Vec_IntEntry(oDep[i], k) + numouts;
            }
        } else {
            adj[i+1] = adj[i];
        }
    }

    for i = 0; i < numins; i++ {
        adj[i+numouts+1] = adj[i+numouts];
        for k = 0, j = adj[i+numouts]; k < Vec_IntSize(iDep[i]); k++ {
            if output[Vec_IntEntry(iDep[i], k)] {
                edg[j++] = Vec_IntEntry(iDep[i], k);
                adj[i+numouts+1]++;
            }
        }
    }

    /* print graph */
    /*for (i = 0; i < n; i++) {
        printf("%d: ", i);
        for (j = adj[i]; j < adj[i+1]; j++)
            printf("%d ", edg[j]);
        printf("\n");
    }*/

    // inlining of macro ABC_FREE
    ((vPiValues) ? (free((char *) (vPiValues)), (vPiValues) = 0) : 0);  
    // inlining of macro ABC_FREE
    ((output) ? (free((char *) (output)), (output) = 0) : 0);
    
    return g;   
}

static struct saucy_graph *
buildSim2Graph( Abc_Ntk_t * pNtk, struct coloring *c, Vec_Int_t * randVec, Vec_Int_t ** iDep, Vec_Int_t ** oDep, Vec_Ptr_t ** topOrder, Vec_Int_t ** obs,  Vec_Int_t ** ctrl )
{
    var i, j, k int32;
    struct saucy_graph *g = nil;
    var n int32;
    var e int32;
    var adj *int32;
    var edg *int32;
    var vPiValues *int32;
    var output, output2 *int32;
    var numouts int32  =  Abc_NtkPoNum(pNtk);
    var numins int32  =  Abc_NtkPiNum(pNtk);

    extern int * Abc_NtkSimulateOneNode( Abc_Ntk_t * , int * , int , Vec_Ptr_t ** );    
    
    vPiValues = generateProperInputVector(pNtk, c, randVec);
    if vPiValues == nil { 
        return nil; }

    output = Abc_NtkVerifySimulatePattern( pNtk, vPiValues );   
    
    for i = 0; i < numins; i++ {
        if !c.clen[c.cfront[i+numouts]] { continue; }
        if vPiValues[i] == 0 {  vPiValues[i] = 1;
        } else {                    vPiValues[i] = 0; }

        output2 = Abc_NtkSimulateOneNode( pNtk, vPiValues, i, topOrder );

        for j = 0; j < Vec_IntSize(iDep[i]); j++ {
            if output[Vec_IntEntry(iDep[i], j)] != output2[Vec_IntEntry(iDep[i], j)] {
                Vec_IntPush(obs[i], Vec_IntEntry(iDep[i], j));
                Vec_IntPush(ctrl[Vec_IntEntry(iDep[i], j)], i);
                e++;
            }
        }

        if vPiValues[i] == 0 {  vPiValues[i] = 1;
        } else {                    vPiValues[i] = 0; }

        // inlining of macro ABC_FREE
        ((output2) ? (free((char *) (output2)), (output2) = 0) : 0);
    }       

    /* build the graph */
    // inlining of macro ABC_ALLOC
    g = ((struct saucy_graph *) malloc(sizeof(struct saucy_graph) * (size_t)(1)));
    n = numouts + numins;
    adj = ints(n+1);
    edg = ints(2*e);        
    g.n = n;
    g.e = e;
    g.adj = adj;
    g.edg = edg;       

    adj[0] = 0;
    for i = 0; i < numouts; i++ {
        adj[i+1] = adj[i] + Vec_IntSize(ctrl[i]);
        for k = 0, j = adj[i]; j < adj[i+1]; j, k = j+1, k+1 {
            edg[j] = Vec_IntEntry(ctrl[i], k) + numouts;
        }
    }
    for i = 0; i < numins; i++ {
        adj[i+numouts+1] = adj[i+numouts] + Vec_IntSize(obs[i]);
        for k = 0, j = adj[i+numouts]; j < adj[i+numouts+1]; j, k = j+1, k+1 {
            edg[j] = Vec_IntEntry(obs[i], k);
        }
    }

    /* print graph */
    /*for (i = 0; i < n; i++) {
        printf("%d: ", i);
        for (j = adj[i]; j < adj[i+1]; j++)
            printf("%d ", edg[j]);
        printf("\n");
    }*/

    // inlining of macro ABC_FREE
    ((output) ? (free((char *) (output)), (output) = 0) : 0);
    // inlining of macro ABC_FREE
    ((vPiValues) ? (free((char *) (vPiValues)), (vPiValues) = 0) : 0);  
    for j = 0; j < numins; j++ {
        Vec_IntClear(obs[j]);
    }
    for j = 0; j < numouts; j++ {
        Vec_IntClear(ctrl[j]);
    }

    return g;
}

static void 
bumpActivity( struct saucy * s, struct sim_result * cex )
{
    var i int32;
    struct sim_result * cex2;

    if (cex.activity += s.activityInc) > 1e20 {
        /* Rescale: */
        for i = 0; i < Vec_PtrSize(s.satCounterExamples); i++ {
            cex2 = (struct sim_result *)Vec_PtrEntry(s.satCounterExamples, i);
            cex2.activity *= 1e-20;
        }
        s.activityInc *= 1e-20; 
    } 
}

static void
reduceDB( struct saucy * s )
{   
    var i, j int32;
    var extra_lim float64  = s.activityInc / Vec_PtrSize(s.satCounterExamples);    /* Remove any clause below this activity */
    struct sim_result * cex;

    for Vec_PtrSize(s.satCounterExamples) > (0.7 * MAX_LEARNTS) {
        for i = j = 0; i < Vec_PtrSize(s.satCounterExamples); i++ {
            cex = (struct sim_result *)Vec_PtrEntry(s.satCounterExamples, i);
            if cex.activity < extra_lim {
                // inlining of macro ABC_FREE
                ((cex.inVec) ? (free((char *) (cex.inVec)), (cex.inVec) = 0) : 0);
                // inlining of macro ABC_FREE
                ((cex.outVec) ? (free((char *) (cex.outVec)), (cex.outVec) = 0) : 0);
                // inlining of macro ABC_FREE
                ((cex) ? (free((char *) (cex)), (cex) = 0) : 0);                      
            
            } else if j < i {
                Vec_PtrWriteEntry(s.satCounterExamples, j, cex);
                j++;
            }
        }
        //printf("Database size reduced from %d to %d\n", Vec_PtrSize(s->satCounterExamples), j);
        Vec_PtrShrink(s.satCounterExamples, j);        
        extra_lim *= 2;
    }
    
    assert(Vec_PtrSize(s.satCounterExamples) <= (0.7 * MAX_LEARNTS), "Vec_PtrSize(s.satCounterExamples) <= (0.7 * MAX_LEARNTS)");
}

func analyzeConflict(pNtk *Abc_Ntk_t, pModel *int32, fVerbose int32) *struct sim_result {   
    var pNode *Abc_Obj_t;
    var i, count int32;
    var pValues *int32;
    struct sim_result * cex;
    var numouts int32  = Abc_NtkPoNum(pNtk);
    var numins int32  = Abc_NtkPiNum(pNtk);
    
    // inlining of macro ABC_ALLOC
    cex = ((struct sim_result *) malloc(sizeof(struct sim_result) * (size_t)(1)));  
    cex.inVec = ints( numins );
    cex.outVec =  ints( numouts ); 

    /* get the CO values under this model */
    pValues = Abc_NtkVerifySimulatePattern( pNtk, pModel );    

    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  { 
        cex.inVec[Abc_ObjId(pNode)-1] = pModel[i];
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        cex.outVec[Abc_ObjId(pNode)-numins-1] = pValues[i];
        if pValues[i] { count++; }
    }   
    
    cex.outVecOnes = count;
    cex.activity = 0;  

    if fVerbose != 0 {
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
            fmt.Printf(" %s=%d", Abc_ObjName(pNode), pModel[i]);
        }
        fmt.Printf("\n");
    }

    // inlining of macro ABC_FREE
    ((pValues) ? (free((char *) (pValues)), (pValues) = 0) : 0);    

    return cex;
}

func Abc_NtkCecSat_saucy(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pModel *int32) int32 {
    extern Abc_Ntk_t * Abc_NtkMulti( Abc_Ntk_t * pNtk, int nThresh, int nFaninMax, int fCnf, int fMulti, int fSimple, int fFactor );
    var pMiter *Abc_Ntk_t;
    var pCnf *Abc_Ntk_t;
    var RetValue int32;
    var nConfLimit int32;
    var nInsLimit int32;
    var i int32;

    nConfLimit = 10000;
    nInsLimit = 0;  

    /* get the miter of the two networks */
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 1, 0, 0, 0 );
    if pMiter == nil {
        fmt.Printf("Miter computation has failed.\n");
        os.Exit(int(1));
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        //printf( "Networks are NOT EQUIVALENT after structural hashing.\n" );
        /* report the error */
        pMiter.pModel = Abc_NtkVerifyGetCleanModel( pMiter, 1 );       
        for i = 0; i < Abc_NtkPiNum(pNtk1); i++ {
            pModel[i] = pMiter.pModel[i];
        }  
        // inlining of macro ABC_FREE
        ((pMiter.pModel) ? (free((char *) (pMiter.pModel)), (pMiter.pModel) = 0) : 0);
        Abc_NtkDelete( pMiter );
        return 0;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pMiter );
        //printf( "Networks are equivalent after structural hashing.\n" );
        return 1;
    }

    /* convert the miter into a CNF */
    pCnf = Abc_NtkMulti( pMiter, 0, 100, 1, 0, 0, 0 );
    Abc_NtkDelete( pMiter );
    if pCnf == nil {
        fmt.Printf("Renoding for CNF has failed.\n");
        os.Exit(int(1));
    }

    /* solve the CNF using the SAT solver */
    RetValue = Abc_NtkMiterSat( pCnf, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)nInsLimit, 0, nil, nil );
    if RetValue == -1 {
        fmt.Printf("Networks are undecided (SAT solver timed out).\n");
        os.Exit(int(1));
    }
    /*else if ( RetValue == 0 )
        printf( "Networks are NOT EQUIVALENT after SAT.\n" );
    else
        printf( "Networks are equivalent after SAT.\n" );*/
    if pCnf.pModel {       
        for i = 0; i < Abc_NtkPiNum(pNtk1); i++ {
            pModel[i] = pCnf.pModel[i];
        }
    }
    // inlining of macro ABC_FREE
    ((pCnf.pModel) ? (free((char *) (pCnf.pModel)), (pCnf.pModel) = 0) : 0);
    Abc_NtkDelete( pCnf );

    return RetValue;
}


func saucyGateWay(pNtkOrig *Abc_Ntk_t, pNodePo *Abc_Obj_t, gFile *os.File, fBooleanMatching int32, fLookForSwaps int32, fFixOutputs int32, fFixInputs int32, fQuiet int32, fPrintTree int32) {
    var pNtk *Abc_Ntk_t;
    struct saucy *s;
    struct saucy_stats stats;
    var colors *int32;
    var i int32;
    var clk int32 = time.Now().UnixNano();   
    
    if pNodePo == nil {
        pNtk = Abc_NtkDup( pNtkOrig );
    } else {
        pNtk = Abc_NtkCreateCone( pNtkOrig, Abc_ObjFanin0(pNodePo), Abc_ObjName(pNodePo), 0 ); }
    
    if Abc_NtkPiNum(pNtk) == 0 {
        Abc_Print( 0, "This output is not dependent on any input\n" );
        Abc_NtkDelete( pNtk );
        return;
    }

    s = saucy_alloc( pNtk );    

    /******* Getting Dependencies *******/  
    fmt.Printf("Build functional dependency graph (dependency stats are below) ... ");      
    getDependencies( pNtk, s.iDep, s.oDep );
    fmt.Printf("\t--------------------\n");
    /************************************/

    /* Finding toplogical orde */
    s.topOrder = findTopologicalOrder( pNtk );                 

    /* Setting graph colors: outputs = 0 and inputs = 1 */
    colors = ints(Abc_NtkPoNum(pNtk) + Abc_NtkPiNum(pNtk));
    if fFixOutputs != 0 {
        for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
            colors[i] = i;
        }
    } else {
        for i = 0; i < Abc_NtkPoNum(pNtk); i++ {
            colors[i] = 0;
        }
    }
    if fFixInputs != 0 {
        var c int32;
        if fFixOutputs != 0 {
            c = Abc_NtkPoNum(pNtk);
        } else {
            c = 1;
        }
        for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
            colors[i+Abc_NtkPoNum(pNtk)] = c+i;
        }     
    } else {
        var c int32;
        if fFixOutputs != 0 {
            c = Abc_NtkPoNum(pNtk);
        } else {
            c = 1;
        }
        for i = 0; i < Abc_NtkPiNum(pNtk); i++ {
            colors[i+Abc_NtkPoNum(pNtk)] = c;
        }   
    }   

    /* Are we looking for Boolean matching? */
    s.fBooleanMatching = fBooleanMatching;
    if fBooleanMatching != 0 {
        NUM_SIM1_ITERATION = 50;
        NUM_SIM2_ITERATION = 50;
    } else {
        NUM_SIM1_ITERATION = 200;
        NUM_SIM2_ITERATION = 200;
    }

    /* Set the print automorphism routine */
    if !(fQuiet != 0) {
        s.print_automorphism = print_automorphism_ntk;
    } else {
        s.print_automorphism = print_automorphism_quiet; }

    /* Set the output file for generators */
    if gFile == nil {
        s.gFile = os.Stdout;
    } else {
        s.gFile = gFile; }

    /* Set print tree option */
    s.fPrintTree = fPrintTree;

    /* Set input permutations option */
    s.fLookForSwaps = fLookForSwaps;

    saucy_search(pNtk, s, 0, colors, &stats);   
    print_stats(os.Stdout, stats);
    if fBooleanMatching != 0 {
        if stats.grpsize_base > 1 || stats.grpsize_exp > 0 {
            fmt.Printf("*** Networks are equivalent ***\n");
        } else {
            fmt.Printf("*** Networks are NOT equivalent ***\n"); }
    }
    saucy_free(s);
    Abc_NtkDelete(pNtk);

    if true {
        var hadi *os.File  = os.OpenFile("hadi.txt", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666);
        fmt.Fprintf(hadi, "group size = %fe%d\n", stats.grpsize_base, stats.grpsize_exp); 
        _ = (hadi).Close();
    }

    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;

}
//ABC_NAMESPACE_IMPL_END



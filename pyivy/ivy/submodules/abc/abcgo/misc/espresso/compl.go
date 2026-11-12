package espresso

import "fmt"

/*
 * Revision Control Information
 *
 * $Source$
 * $Author$
 * $Revision$
 * $Date$
 *
 */
/*
 *  module: compl.c
 *  purpose: compute the complement of a multiple-valued function
 *
 *  The "unate recursive paradigm" is used.  After a set of special
 *  cases are examined, the function is split on the "most active
 *  variable".  These two halves are complemented recursively, and then
 *  the results are merged.
 *
 *  Changes (from Version 2.1 to Version 2.2)
 *      1. Minor bug in compl_lifting -- cubes in the left half were
 *      not marked as active, so that when merging a leaf from the left
 *      hand side, the active flags were essentially random.  This led
 *      to minor impredictability problem, but never affected the
 *      accuracy of the results.
 */

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


const USE_COMPL_LIFT =            0
const USE_COMPL_LIFT_ONSET =        1
const USE_COMPL_LIFT_ONSET_COMPLEX =    2
const NO_LIFTING =            3

// static bool compl_special_cases();

static pcover compl_merge();
// static void compl_d1merge();

static pcover compl_cube();
// static void compl_lift();

// static void compl_lift_onset();

// static void compl_lift_onset_complex();

// static bool simp_comp_special_cases();

// static bool simplify_special_cases();



/* complement -- compute the complement of T */
pcover complement(T)
pcube *T;            /* T will be disposed of */
{
    register pcube cl, cr;
    register int best;
    pcover Tbar, Tl, Tr;
    int lifting;
    static int compl_level = 0;

    if (debug & COMPL)
    debug_print(T, "COMPLEMENT", compl_level++);

    if compl_special_cases(T, &Tbar) == MAYBE {

    /* Allocate space for the partition cubes */
    // inlining of macro new_cube
    // inlining of macro set_new
    cl = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    cr = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    best = binate_split_select(T, cl, cr, COMPL);

    /* Complement the left and right halves */
    Tl = complement(scofactor(T, cl, best));
    Tr = complement(scofactor(T, cr, best));

    // inlining of macro CUBELISTSIZE
    if Tr.count*Tl.count > (Tr.count+Tl.count)*(((pcube *) T[1] - T) - 3) {
        lifting = USE_COMPL_LIFT_ONSET;
    } else {
        lifting = USE_COMPL_LIFT;
    }
    Tbar = compl_merge(T, Tl, Tr, cl, cr, best, lifting);

    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cl);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cr);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    }

    if (debug & COMPL)
    debug1_print(Tbar, "exit COMPLEMENT", --compl_level);
    return Tbar;
}

static bool compl_special_cases(T, Tbar)
pcube *T;            /* will be disposed if answer is determined */
pcover *Tbar;            /* returned only if answer determined */
{
    register pcube *T1, p, ceil, cof=T[0];
    pcover A, ceil_compl;

    /* Check for no cubes in the cover */
    if T[2] == NULL {
    // inlining of macro new_cover
    *Tbar = sf_addset(sf_new(1, cube.size), cube.fullset);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for only a single cube in the cover */
    if T[3] == NULL {
    *Tbar = compl_cube(set_or(cof, cof, T[2]));
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for a row of all 1's (implies complement is null) */
    for T1 = T+2; (p = *T1++) != NULL;  {
    if full_row(p, cof) {
        // inlining of macro new_cover
        *Tbar = sf_new(0, cube.size);
        // inlining of macro free_cubelist
        FREE(T[0]); FREE(T);;
        return TRUE;
    }
    }

    /* Check for a column of all 0's which can be factored out */
    // inlining of macro set_save
    ceil = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cof))), cof);
    for T1 = T+2; (p = *T1++) != NULL;  {
    INLINEset_or(ceil, ceil, p);
    }
    if ! setp_equal(ceil, cube.fullset) {
    ceil_compl = compl_cube(ceil);
    (void) set_or(cof, cof, set_diff(ceil, cube.fullset, ceil));
    // inlining of macro set_free
    FREE(ceil);
    *Tbar = sf_append(complement(T), ceil_compl);
    return TRUE;
    }
    // inlining of macro set_free
    FREE(ceil);

    /* Collect column counts, determine unate variables, etc. */
    massive_count(T);

    /* If single active variable not factored out above, then tautology ! */
    if cdata.vars_active == 1 {
    // inlining of macro new_cover
    *Tbar = sf_new(0, cube.size);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;

    /* Check for unate cover */
    } else if cdata.vars_unate == cdata.vars_active {
    A = map_cover_to_unate(T);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    A = unate_compl(A);
    *Tbar = map_unate_to_cover(A);
    sf_free(A);
    return TRUE;

    /* Not much we can do about it */
    } else {
    return MAYBE;
    }
}

/*
 *  compl_merge -- merge the two cofactors around the splitting
 *  variable
 *
 *  The merge operation involves intersecting each cube of the left
 *  cofactor with cl, and intersecting each cube of the right cofactor
 *  with cr.  The union of these two covers is the merged result.
 *
 *  In order to reduce the number of cubes, a distance-1 merge is
 *  performed (note that two cubes can only combine distance-1 in the
 *  splitting variable).  Also, a simple expand is performed in the
 *  splitting variable (simple implies the covering check for the
 *  expansion is not full containment, but single-cube containment).
 */

static pcover compl_merge(T1, L, R, cl, cr, var, lifting)
pcube *T1;            /* Original ON-set */
pcover L, R;            /* Complement from each recursion branch */
register pcube cl, cr;        /* cubes used for cofactoring */
var var int32;            /* splitting variable */
var lifting int32;            /* whether to perform lifting or not */
{
    register pcube p, last, pt;
    pcover T, Tbar;
    pcube *L1, *R1;

    if debug & COMPL {
    (void) fmt.Printf("compl_merge: left %d, right %d\n", L.count, R.count);
    (void) fmt.Printf("%s (cl)\n%s (cr)\nLeft is\n", pc1(cl), pc2(cr));
    cprint(L);
    (void) fmt.Printf("Right is\n");
    cprint(R);
    }

    /* Intersect each cube with the cofactored cube */
    // inlining of macro foreach_set
    for p=L.data,last=p+L.count*L.wsize;p<last;p+=L.wsize {
    INLINEset_and(p, p, cl);
    // inlining of macro SET
    (p[0] |= (ACTIVE));
    }
    // inlining of macro foreach_set
    for p=R.data,last=p+R.count*R.wsize;p<last;p+=R.wsize {
    INLINEset_and(p, p, cr);
    // inlining of macro SET
    (p[0] |= (ACTIVE));
    }

    /* Sort the arrays for a distance-1 merge */
    (void) set_copy(cube.temp[0], cube.var_mask[var]);
    qsort((char *) (L1 = sf_list(L)), (size_t)L.count, sizeof(pset), (int (*)()) d1_order);
    qsort((char *) (R1 = sf_list(R)), (size_t)R.count, sizeof(pset), (int (*)()) d1_order);

    /* Perform distance-1 merge */
    compl_d1merge(L1, R1);

    /* Perform lifting */
    switch lifting {
    case USE_COMPL_LIFT_ONSET:
        T = cubeunlist(T1);
        compl_lift_onset(L1, T, cr, var);
        compl_lift_onset(R1, T, cl, var);
        // inlining of macro free_cover
        sf_free(T);
        break;
    case USE_COMPL_LIFT_ONSET_COMPLEX:
        T = cubeunlist(T1);
        compl_lift_onset_complex(L1, T, var);
        compl_lift_onset_complex(R1, T, var);
        // inlining of macro free_cover
        sf_free(T);
        break;
    case USE_COMPL_LIFT:
        compl_lift(L1, R1, cr, var);
        compl_lift(R1, L1, cl, var);
        break;
    case NO_LIFTING:
        break;
    default:
        ;
    }
    FREE(L1);
    FREE(R1);

    /* Re-create the merged cover */
    // inlining of macro new_cover
    Tbar = sf_new(L.count + R.count, cube.size);
    pt = Tbar.data;
    // inlining of macro foreach_set
    for p=L.data,last=p+L.count*L.wsize;p<last;p+=L.wsize {
    INLINEset_copy(pt, p);
    Tbar.count++;
    pt += Tbar.wsize;
    }
    // inlining of macro foreach_active_set
    // inlining of macro foreach_set
    // inlining of macro TESTP
    for(p=R.data,last=p+R.count*R.wsize;p<last;p+=R.wsize) if p[0] & (ACTIVE) {
    INLINEset_copy(pt, p);
    Tbar.count++;
    pt += Tbar.wsize;
    }

    if debug & COMPL {
    (void) fmt.Printf("Result %d\n", Tbar.count);
    if (verbose_debug)
        cprint(Tbar);
    }

    // inlining of macro free_cover
    sf_free(L);
    // inlining of macro free_cover
    sf_free(R);
    return Tbar;
}

/*
 *  compl_lift_simple -- expand in the splitting variable using single
 *  cube containment against the other recursion branch to check
 *  validity of the expansion, and expanding all (or none) of the
 *  splitting variable.
 */
static void compl_lift(A1, B1, bcube, var)
pcube *A1, *B1, bcube;
var var int32;
{
    register pcube a, b, *B2, lift=cube.temp[4], liftor=cube.temp[5];
    pcube mask = cube.var_mask[var];

    (void) set_and(liftor, bcube, mask);

    /* for each cube in the first array ... */
    for (a = *A1++) != NULL {
    // inlining of macro TESTP
    if a[0] & (ACTIVE) {

        /* create a lift of this cube in the merging coord */
        (void) set_merge(lift, bcube, a, mask);

        /* for each cube in the second array */
        for B2 = B1; (b = *B2++) != NULL;  {
        INLINEsetp_implies(lift, b, /* when_false => */ continue);
        /* when_true => fall through to next statement */

        /* cube of A1 was contained by some cube of B1, so raise */
        INLINEset_or(a, a, liftor);
        break;
        }
    }
    }
}



/*
 *  compl_lift_onset -- expand in the splitting variable using a
 *  distance-1 check against the original on-set; expand all (or
 *  none) of the splitting variable.  Each cube of A1 is expanded
 *  against the original on-set T.
 */
static void compl_lift_onset(A1, T, bcube, var)
pcube *A1;
pcover T;
pcube bcube;
var var int32;
{
    register pcube a, last, p, lift=cube.temp[4], mask=cube.var_mask[var];

    /* for each active cube from one branch of the complement */
    for (a = *A1++) != NULL {
    // inlining of macro TESTP
    if a[0] & (ACTIVE) {

        /* create a lift of this cube in the merging coord */
        INLINEset_and(lift, bcube, mask);    /* isolate parts to raise */
        INLINEset_or(lift, a, lift);    /* raise these parts in a */

        /* for each cube in the ON-set, check for intersection */
        // inlining of macro foreach_set
        for p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize {
        if cdist0(p, lift) {
            goto nolift;
        }
        }
        INLINEset_copy(a, lift);        /* save the raising */
        // inlining of macro SET
        (a[0] |= (ACTIVE));
nolift : ;
    }
    }
}

/*
 *  compl_lift_complex -- expand in the splitting variable, but expand all
 *  parts which can possibly expand.
 *  T is the original ON-set
 *  A1 is either the left or right cofactor
 */
static void compl_lift_onset_complex(A1, T, var)
pcube *A1;            /* array of pointers to new result */
pcover T;            /* original ON-set */
var var int32;            /* which variable we split on */
{
    register int dist;
    register pcube last, p, a, xlower;

    /* for each cube in the complement */
    // inlining of macro new_cube
    // inlining of macro set_new
    xlower = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    for (a = *A1++) != NULL {

    // inlining of macro TESTP
    if a[0] & (ACTIVE) {

        /* Find which parts of the splitting variable are forced low */
        INLINEset_clear(xlower, cube.size);
        // inlining of macro foreach_set
        for p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize {
        if (dist = cdist01(p, a)) < 2 {
            if dist == 0 {
            fatal("compl: ON-set and OFF-set are not orthogonal");
            } else {
            (void) force_lower(xlower, p, a);
            }
        }
        }

        (void) set_diff(xlower, cube.var_mask[var], xlower);
        (void) set_or(a, a, xlower);
        // inlining of macro free_cube
        // inlining of macro set_free
        FREE(xlower);
    }
    }
}



/*
 *  compl_d1merge -- distance-1 merge in the splitting variable
 */
static void compl_d1merge(L1, R1)
register pcube *L1, *R1;
{
    register pcube pl, pr;

    /* Find equal cubes between the two cofactors */
    for(pl = *L1, pr = *R1; (pl != NULL) && (pr != NULL); )
    switch d1_order(L1, R1) {
        case 1:
        pr = *(++R1); break;            /* advance right pointer */
        case -1:
        pl = *(++L1); break;            /* advance left pointer */
        case 0:
        // inlining of macro RESET
        (pr[0] &= ~ (ACTIVE));
        INLINEset_or(pl, pl, pr);
        pr = *(++R1);
        fallthrough
default:
        ;
    }
}



/* compl_cube -- return the complement of a single cube (De Morgan's law) */
static pcover compl_cube(p)
register pcube p;
{
    register pcube diff=cube.temp[7], pdest, mask, full=cube.fullset;
    int var;
    pcover R;

    /* Allocate worst-case size cover (to avoid checking overflow) */
    // inlining of macro new_cover
    R = sf_new(cube.num_vars, cube.size);

    /* Compute bit-wise complement of the cube */
    INLINEset_diff(diff, full, p);

    for var = 0; var < cube.num_vars; var++ {
    mask = cube.var_mask[var];
    /* If the bit-wise complement is not empty in var ... */
    if ! setp_disjoint(diff, mask) {
        // inlining of macro GETSET
        pdest = ((R).data + (R).wsize * (R.count++));
        INLINEset_merge(pdest, diff, full, mask);
    }
    }
    return R;
}

/* simp_comp -- quick simplification of T */
void simp_comp(T, Tnew, Tbar)
pcube *T;            /* T will be disposed of */
pcover *Tnew;
pcover *Tbar;
{
    register pcube cl, cr;
    register int best;
    pcover Tl, Tr, Tlbar, Trbar;
    int lifting;
    static int simplify_level = 0;

    if (debug & COMPL)
    debug_print(T, "SIMPCOMP", simplify_level++);

    if simp_comp_special_cases(T, Tnew, Tbar) == MAYBE {

    /* Allocate space for the partition cubes */
    // inlining of macro new_cube
    // inlining of macro set_new
    cl = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    cr = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    best = binate_split_select(T, cl, cr, COMPL);

    /* Complement the left and right halves */
    simp_comp(scofactor(T, cl, best), &Tl, &Tlbar);
    simp_comp(scofactor(T, cr, best), &Tr, &Trbar);

    lifting = USE_COMPL_LIFT;
    *Tnew = compl_merge(T, Tl, Tr, cl, cr, best, lifting);

    lifting = USE_COMPL_LIFT;
    *Tbar = compl_merge(T, Tlbar, Trbar, cl, cr, best, lifting);

    /* All of this work for nothing ? Let's hope not ... */
    // inlining of macro CUBELISTSIZE
    if (*Tnew).count > (((pcube *) T[1] - T) - 3) {
        sf_free(*Tnew);
        *Tnew = cubeunlist(T);
    }

    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cl);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cr);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    }

    if debug & COMPL {
    debug1_print(*Tnew, "exit SIMPCOMP (new)", simplify_level);
    debug1_print(*Tbar, "exit SIMPCOMP (compl)", simplify_level);
    simplify_level--;
    }
}

static bool simp_comp_special_cases(T, Tnew, Tbar)
pcube *T;            /* will be disposed if answer is determined */
pcover *Tnew;            /* returned only if answer determined */
pcover *Tbar;            /* returned only if answer determined */
{
    register pcube *T1, p, ceil, cof=T[0];
    pcube last;
    pcover A;

    /* Check for no cubes in the cover (function is empty) */
    if T[2] == NULL {
    // inlining of macro new_cover
    *Tnew = sf_new(1, cube.size);
    // inlining of macro new_cover
    *Tbar = sf_addset(sf_new(1, cube.size), cube.fullset);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for only a single cube in the cover */
    if T[3] == NULL {
    (void) set_or(cof, cof, T[2]);
    // inlining of macro new_cover
    *Tnew = sf_addset(sf_new(1, cube.size), cof);
    *Tbar = compl_cube(cof);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for a row of all 1's (function is a tautology) */
    for T1 = T+2; (p = *T1++) != NULL;  {
    if full_row(p, cof) {
        // inlining of macro new_cover
        *Tnew = sf_addset(sf_new(1, cube.size), cube.fullset);
        // inlining of macro new_cover
        *Tbar = sf_new(1, cube.size);
        // inlining of macro free_cubelist
        FREE(T[0]); FREE(T);;
        return TRUE;
    }
    }

    /* Check for a column of all 0's which can be factored out */
    // inlining of macro set_save
    ceil = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cof))), cof);
    for T1 = T+2; (p = *T1++) != NULL;  {
    INLINEset_or(ceil, ceil, p);
    }
    if ! setp_equal(ceil, cube.fullset) {
    // inlining of macro new_cube
    // inlining of macro set_new
    p = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    (void) set_diff(p, cube.fullset, ceil);
    (void) set_or(cof, cof, p);
    // inlining of macro set_free
    FREE(p);
    simp_comp(T, Tnew, Tbar);

    /* Adjust the ON-set */
    A = *Tnew;
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
        INLINEset_and(p, p, ceil);
    }

    /* Compute the new complement */
    *Tbar = sf_append(*Tbar, compl_cube(ceil));
    // inlining of macro set_free
    FREE(ceil);
    return TRUE;
    }
    // inlining of macro set_free
    FREE(ceil);

    /* Collect column counts, determine unate variables, etc. */
    massive_count(T);

    /* If single active variable not factored out above, then tautology ! */
    if cdata.vars_active == 1 {
    // inlining of macro new_cover
    *Tnew = sf_addset(sf_new(1, cube.size), cube.fullset);
    // inlining of macro new_cover
    *Tbar = sf_new(1, cube.size);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;

    /* Check for unate cover */
    } else if cdata.vars_unate == cdata.vars_active {
    /* Make the cover minimum by single-cube containment */
    A = cubeunlist(T);
    *Tnew = sf_contain(A);

    /* Now form a minimum representation of the complement */
    A = map_cover_to_unate(T);
    A = unate_compl(A);
    *Tbar = map_unate_to_cover(A);
    sf_free(A);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;

    /* Not much we can do about it */
    } else {
    return MAYBE;
    }
}

/* simplify -- quick simplification of T */
pcover simplify(T)
pcube *T;            /* T will be disposed of */
{
    register pcube cl, cr;
    register int best;
    pcover Tbar, Tl, Tr;
    int lifting;
    static int simplify_level = 0;

    if debug & COMPL {
    debug_print(T, "SIMPLIFY", simplify_level++);
    }

    if simplify_special_cases(T, &Tbar) == MAYBE {

    /* Allocate space for the partition cubes */
    // inlining of macro new_cube
    // inlining of macro set_new
    cl = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    cr = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);

    best = binate_split_select(T, cl, cr, COMPL);

    /* Complement the left and right halves */
    Tl = simplify(scofactor(T, cl, best));
    Tr = simplify(scofactor(T, cr, best));

    lifting = USE_COMPL_LIFT;
    Tbar = compl_merge(T, Tl, Tr, cl, cr, best, lifting);

    /* All of this work for nothing ? Let's hope not ... */
    // inlining of macro CUBELISTSIZE
    if Tbar.count > (((pcube *) T[1] - T) - 3) {
        sf_free(Tbar);
        Tbar = cubeunlist(T);
    }

    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cl);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cr);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    }

    if debug & COMPL {
    debug1_print(Tbar, "exit SIMPLIFY", --simplify_level);
    }
    return Tbar;
}

static bool simplify_special_cases(T, Tnew)
pcube *T;            /* will be disposed if answer is determined */
pcover *Tnew;            /* returned only if answer determined */
{
    register pcube *T1, p, ceil, cof=T[0];
    pcube last;
    pcover A;

    /* Check for no cubes in the cover */
    if T[2] == NULL {
    // inlining of macro new_cover
    *Tnew = sf_new(0, cube.size);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for only a single cube in the cover */
    if T[3] == NULL {
    // inlining of macro new_cover
    *Tnew = sf_addset(sf_new(1, cube.size), set_or(cof, cof, T[2]));
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;
    }

    /* Check for a row of all 1's (implies function is a tautology) */
    for T1 = T+2; (p = *T1++) != NULL;  {
    if full_row(p, cof) {
        // inlining of macro new_cover
        *Tnew = sf_addset(sf_new(1, cube.size), cube.fullset);
        // inlining of macro free_cubelist
        FREE(T[0]); FREE(T);;
        return TRUE;
    }
    }

    /* Check for a column of all 0's which can be factored out */
    // inlining of macro set_save
    ceil = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cof))), cof);
    for T1 = T+2; (p = *T1++) != NULL;  {
    INLINEset_or(ceil, ceil, p);
    }
    if ! setp_equal(ceil, cube.fullset) {
    // inlining of macro new_cube
    // inlining of macro set_new
    p = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    (void) set_diff(p, cube.fullset, ceil);
    (void) set_or(cof, cof, p);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(p);

    A = simplify(T);
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
        INLINEset_and(p, p, ceil);
    }
    *Tnew = A;
    // inlining of macro set_free
    FREE(ceil);
    return TRUE;
    }
    // inlining of macro set_free
    FREE(ceil);

    /* Collect column counts, determine unate variables, etc. */
    massive_count(T);

    /* If single active variable not factored out above, then tautology ! */
    if cdata.vars_active == 1 {
    // inlining of macro new_cover
    *Tnew = sf_addset(sf_new(1, cube.size), cube.fullset);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;

    /* Check for unate cover */
    } else if cdata.vars_unate == cdata.vars_active {
    A = cubeunlist(T);
    *Tnew = sf_contain(A);
    // inlining of macro free_cubelist
    FREE(T[0]); FREE(T);;
    return TRUE;

    /* Not much we can do about it */
    } else {
    return MAYBE;
    }
}
//ABC_NAMESPACE_IMPL_END


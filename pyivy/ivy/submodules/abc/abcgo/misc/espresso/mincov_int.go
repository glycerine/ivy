package espresso

/*
 * Revision Control Information
 *
 * $Source$
 * $Author$
 * $Revision$
 * $Date$
 *
 */
//#include "port.h"
//#include "utility.h"
//#include "sparse.h"
//#include "mincov.h"

//#include "util_hack.h" // added


   
type stats_t struct {
    debug int32;            /* 1 if debugging is enabled */
    max_print_depth int32;    /* dump stats for levels up to this level */
    max_depth int32;        /* deepest the recursion has gone */
    nodes int32;            /* total nodes visited */
    component int32;        /* currently solving a component */
    comp_count int32;        /* number of components detected */
    gimpel_count int32;        /* number of times Gimpel reduction applied */
    gimpel int32;            /* currently inside Gimpel reduction */
    start_time long;        /* cpu time when the covering started */
    no_branching int32;
    lower_bound int32;
}



   
type solution_t struct {
    row *sm_row;
    cost int32;
}


// extern solution_t *solution_alloc();

// extern void solution_free();

// extern solution_t *solution_dup();

// extern void solution_accept();

// extern void solution_reject();

// extern void solution_add();

// extern solution_t *solution_choose_best();


// extern solution_t *sm_maximal_independent_set();

// extern solution_t *sm_mincov();

// extern int gimpel_reduce();



#define ABC__misc__espresso__mincov_int_h

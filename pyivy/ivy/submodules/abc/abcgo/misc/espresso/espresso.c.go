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
/*
 *  Module: espresso.c
 *  Purpose: The main espresso algorithm
 *
 *  Returns a minimized version of the ON-set of a function
 *
 *  The following global variables affect the operation of Espresso:
 *
 *  MISCELLANEOUS:
 *      trace
 *          print trace information as the minimization progresses
 *
 *      remove_essential
 *          remove essential primes
 *
 *      single_expand
 *          if true, stop after first expand/irredundant
 *
 *  LAST_GASP or SUPER_GASP strategy:
 *      use_super_gasp
 *          uses the super_gasp strategy rather than last_gasp
 *
 *  SETUP strategy:
 *      recompute_onset
 *          recompute onset using the complement before starting
 *
 *      unwrap_onset
 *          unwrap the function output part before first expand
 *
 *  MAKE_SPARSE strategy:
 *      force_irredundant
 *          iterates make_sparse to force a minimal solution (used
 *          indirectly by make_sparse)
 *
 *      skip_make_sparse
 *          skip the make_sparse step (used by opo only)
 */

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


pcover espresso(F, D1, R)
pcover F, D1, R;
{
    pcover E, D, Fsave;
    pset last, p;
    cost_t cost, best_cost;

begin:
    Fsave = sf_save(F);        /* save original function */
    D = sf_save(D1);        /* make a scratch copy of D */

    /* Setup has always been a problem */
    if recompute_onset {
    // inlining of macro EXEC
    {long t=ptime();E = simplify(cube1list(F));if(trace)print_trace(E,"SIMPLIFY   ",ptime()-t);};
    // inlining of macro free_cover
    sf_free(F);
    F = E;
    }
    cover_cost(F, &cost);
    if unwrap_onset && (cube.part_size[cube.num_vars - 1] > 1)
      && (cost.out != cost.cubes*cube.part_size[cube.num_vars-1])
      && (cost.out < 5000) {
    // inlining of macro EXEC
    long t=ptime();F = sf_contain(unravel(F, cube.num_vars - 1));if(trace)print_trace(F,"SETUP      ",ptime()-t);};

    /* Initial expand and irredundant */
    // inlining of macro foreach_set
    for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
    // inlining of macro RESET
    (p[0] &= ~ (PRIME));
    }
    // inlining of macro EXECUTE
    {long t=ptime();F = expand(F, R, FALSE);totals(t,EXPAND_TIME,F,&(cost));};
    // inlining of macro EXECUTE
    {long t=ptime();F = irredundant(F, D);totals(t,IRRED_TIME,F,&(cost));};

    if ! single_expand {
    if remove_essential {
        // inlining of macro EXECUTE
        {long t=ptime();E = essential(&F, &D);totals(t,ESSEN_TIME,E,&(cost));};
    } else {
        // inlining of macro new_cover
        E = sf_new(0, cube.size);
    }

    cover_cost(F, &cost);
    for {

        /* Repeat inner loop until solution becomes "stable" */
        for {
        copy_cost(&cost, &best_cost);
        // inlining of macro EXECUTE
        {long t=ptime();F = reduce(F, D);totals(t,REDUCE_TIME,F,&(cost));};
        // inlining of macro EXECUTE
        {long t=ptime();F = expand(F, R, FALSE);totals(t,EXPAND_TIME,F,&(cost));};
        // inlining of macro EXECUTE
        {long t=ptime();F = irredundant(F, D);totals(t,IRRED_TIME,F,&(cost));};
        
if !(cost.cubes < best_cost.cubes) { break }
} 

        /* Perturb solution to see if we can continue to iterate */
        copy_cost(&cost, &best_cost);
        if use_super_gasp {
        F = super_gasp(F, D, R, &cost);
        if (cost.cubes >= best_cost.cubes)
            break;
        } else {
        F = last_gasp(F, D, R, &cost);
        }

    
if !(cost.cubes < best_cost.cubes ||
        (cost.cubes == best_cost.cubes && cost.total < best_cost.total)) { break }
} 

    /* Append the essential cubes to F */
    F = sf_append(F, E);                /* disposes of E */
    if (trace) size_stamp(F, "ADJUST     ");
    }

    /* Free the D which we used */
    // inlining of macro free_cover
    sf_free(D);

    /* Attempt to make the PLA matrix sparse */
    if ! skip_make_sparse {
    F = make_sparse(F, D1, R);
    }

    /*
     *  Check to make sure function is actually smaller !!
     *  This can only happen because of the initial unravel.  If we fail,
     *  then run the whole thing again without the unravel.
     */
    if Fsave.count < F.count {
    // inlining of macro free_cover
    sf_free(F);
    F = Fsave;
    unwrap_onset = FALSE;
    goto begin;
    } else {
    // inlining of macro free_cover
    sf_free(Fsave);
    }

    return F;
}
//ABC_NAMESPACE_IMPL_END


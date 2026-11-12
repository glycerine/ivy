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
    module: sparse.c

    make_sparse is a last-step cleanup to reduce the total number
    of literals in the cover.

    This is done by reducing the "sparse" variables (using a modified
    version of irredundant rather than reduce), followed by expanding
    the "dense" variables (using modified version of expand).
*/

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


pcover make_sparse(F, D, R)
pcover F, D, R;
{
    cost_t cost, best_cost;

    cover_cost(F, &best_cost);

    for {
    // inlining of macro EXECUTE
    {long t=ptime();F = mv_reduce(F, D);totals(t,MV_REDUCE_TIME,F,&(cost));};
    if (cost.total == best_cost.total)
        break;
    copy_cost(&cost, &best_cost);

    // inlining of macro EXECUTE
    {long t=ptime();F = expand(F, R, TRUE);totals(t,RAISE_IN_TIME,F,&(cost));};
    if (cost.total == best_cost.total)
        break;
    copy_cost(&cost, &best_cost);
    
if !(force_irredundant) { break }
} 

    return F;
}

/*
    mv_reduce -- perform an "optimal" reduction of the variables which
    we desire to be sparse

    This could be done using "reduce" and then saving just the desired
    part of the reduction.  Instead, this version uses IRRED to find
    which cubes of an output are redundant.  Note that this gets around
    the cube-ordering problem.

    In normal use, it is expected that the cover is irredundant and
    hence no cubes will be reduced to the empty cube (however, this is
    checked for and such cubes will be deleted)
*/

pcover
mv_reduce(F, D)
pcover F, D;
{
    register int i, var;
    register pcube p, p1, last;
    int index;
    pcover F1, D1;
    pcube *F_cube_table;

    /* loop for each multiple-valued variable */
    for var = 0; var < cube.num_vars; var++ {

    if cube.sparse[var] {

        /* loop for each part of the variable */
        for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {

        /* remember mapping of F1 cubes back to F cubes */
        F_cube_table = ALLOC(pcube, F.count);

        /* 'cofactor' against part #i of variable #var */
        // inlining of macro new_cover
        F1 = sf_new(F.count, cube.size);
        // inlining of macro foreach_set
        for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
            // inlining of macro is_in_set
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            if p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
            F_cube_table[F1.count] = p;
            // inlining of macro GETSET
            p1 = ((F1).data + (F1).wsize * (F1.count++));
            (void) set_diff(p1, p, cube.var_mask[var]);
            // inlining of macro set_insert
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            (p1[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
            }
        }

        /* 'cofactor' against part #i of variable #var */
        /* not really necessary -- just more efficient ? */
        // inlining of macro new_cover
        D1 = sf_new(D.count, cube.size);
        // inlining of macro foreach_set
        for p=D.data,last=p+D.count*D.wsize;p<last;p+=D.wsize {
            // inlining of macro is_in_set
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            if p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
            // inlining of macro GETSET
            p1 = ((D1).data + (D1).wsize * (D1.count++));
            (void) set_diff(p1, p, cube.var_mask[var]);
            // inlining of macro set_insert
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            (p1[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
            }
        }

        mark_irredundant(F1, D1);

        /* now remove part i from cubes which are redundant */
        index = 0;
        // inlining of macro foreach_set
        for p1=F1.data,last=p1+F1.count*F1.wsize;p1<last;p1+=F1.wsize {
            // inlining of macro TESTP
            if ! (p1[0] & (ACTIVE)) {
            p = F_cube_table[index];

            /*   don't reduce a variable which is full
             *   (unless it is the output variable)
             */
            if var == cube.num_vars-1 ||
                 ! setp_implies(cube.var_mask[var], p) {
                // inlining of macro set_remove
                // inlining of macro WHICH_WORD
                // inlining of macro WHICH_BIT
                (p[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
            }
            // inlining of macro RESET
            (p[0] &= ~ (PRIME));
            }
            index++;
        }

        // inlining of macro free_cover
        sf_free(F1);
        // inlining of macro free_cover
        sf_free(D1);
        FREE(F_cube_table);
        }
    }
    }

    /* Check if any cubes disappeared */
    (void) sf_active(F);
    for var = 0; var < cube.num_vars; var++ {
    if cube.sparse[var] {
        // inlining of macro foreach_active_set
        // inlining of macro foreach_set
        // inlining of macro TESTP
        for(p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize) if p[0] & (ACTIVE) {
        if setp_disjoint(p, cube.var_mask[var]) {
            // inlining of macro RESET
            (p[0] &= ~ (ACTIVE));
            F.active_count--;
        }
        }
    }
    }

    if F.count != F.active_count {
    F = sf_inactive(F);
    }
    return F;
}
//ABC_NAMESPACE_IMPL_END


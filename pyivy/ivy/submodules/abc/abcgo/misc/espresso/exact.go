package espresso

import (
	"fmt"
	"os"
	"strings"
)

/*
 * Revision Control Information
 *
 * $Source$
 * $Author$
 * $Revision$
 * $Date$
 *
 */
//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START



// static void dump_irredundant();

static pcover do_minimize();


/*
 *  minimize_exact -- main entry point for exact minimization
 *
 *  Global flags which affect this routine are:
 *
 *      debug
 *      skip_make_sparse
 */

pcover
minimize_exact(F, D, R, exact_cover)
pcover F, D, R;
var exact_cover int32;
{
    return do_minimize(F, D, R, exact_cover, /*weighted*/ 0);
}


pcover
minimize_exact_literals(F, D, R, exact_cover)
pcover F, D, R;
var exact_cover int32;
{
    return do_minimize(F, D, R, exact_cover, /*weighted*/ 1);
}



static pcover
do_minimize(F, D, R, exact_cover, weighted)
pcover F, D, R;
var exact_cover int32;
var weighted int32;
{
    pcover newF, E, Rt, Rp;
    pset p, last;
    int heur, level, *weights, i;
    sm_matrix *table;
    sm_row *cover;
    sm_element *pe;
    int debug_save = debug;

    if debug & EXACT {
    debug |= (IRRED | MINCOV);
    }
#if defined(sun) || defined(bsd4_2)            /* hack ... */
    if (debug & MINCOV) {
    setlinebuf(stdout);
    }
#endif
    level = (debug & MINCOV) ? 4 : 0;
    heur = ! exact_cover;

    /* Generate all prime implicants */
    // inlining of macro EXEC
    {long t=ptime();F = primes_consensus(cube2list(F, D));if(trace)print_trace(F,"PRIMES     ",ptime()-t);};

    /* Setup the prime implicant table */
    // inlining of macro EXEC
    {long t=ptime();irred_split_cover(F, D, &E, &Rt, &Rp);if(trace)print_trace(E,"ESSENTIALS ",ptime()-t);};
    // inlining of macro EXEC
    {long t=ptime();table = irred_derive_table(D, E, Rp);if(trace)print_trace(Rp,"PI-TABLE   ",ptime()-t);};

    /* Solve either a weighted or nonweighted covering problem */
    if weighted != 0 {
    /* correct only for all 2-valued variables */
    weights = ALLOC(int, F.count);
    // inlining of macro foreach_set
    for p=Rp.data,last=p+Rp.count*Rp.wsize;p<last;p+=Rp.wsize {
        weights[SIZE(p)] = cube.size - set_ord(p);
        /* We have added the 0's in the output part instead of the 1's.
           This loop corrects the literal count. */
        for i = cube.first_part[cube.output];
         i <= cube.last_part[cube.output]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) ? weights[SIZE(p)]++ : weights[SIZE(p)]--;
        }
    }
    } else {
    weights = NIL(int);
    }
    // inlining of macro EXEC
    {long t=ptime();cover=sm_minimum_cover(table,weights,heur,level);if(trace)print_trace(F,"MINCOV     ",ptime()-t);};
    if weights != 0 {
    FREE(weights);
    }

    if debug & EXACT {
    dump_irredundant(E, Rt, Rp, table);
    }

    /* Form the result cover */
    // inlining of macro new_cover
    newF = sf_new(100, cube.size);
    // inlining of macro foreach_set
    for p=E.data,last=p+E.count*E.wsize;p<last;p+=E.wsize {
    newF = sf_addset(newF, p);
    }
    sm_foreach_row_element(cover, pe) {
    // inlining of macro GETSET
    newF = sf_addset(newF, ((F).data + (F).wsize * (pe.col_num)));
    }

    // inlining of macro free_cover
    sf_free(E);
    // inlining of macro free_cover
    sf_free(Rt);
    // inlining of macro free_cover
    sf_free(Rp);
    sm_free(table);
    sm_row_free(cover);
    // inlining of macro free_cover
    sf_free(F);

    /* Attempt to make the results more sparse */
    debug &= ~ (IRRED | SHARP | MINCOV);
    if ! skip_make_sparse && R != 0 {
    newF = make_sparse(newF, D, R);
    }

    debug = debug_save;
    return newF;
}

static void
dump_irredundant(E, Rt, Rp, table)
pcover E, Rt, Rp;
var table *sm_matrix;
{
    FILE *fp_pi_table, *fp_primes;
    pPLA PLA;
    pset last, p;
    char *file;

    if filename == 0 || int32(strings.Compare(filename, "(stdin)")) == 0 {
    fp_pi_table = fp_primes = os.Stdout;
    } else {
    file = ALLOC(char, len(filename)+20);
    (void) sprintf(file, "%s.primes", filename);
    if (fp_primes = os.OpenFile(file, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)) == NULL {
        (void) fmt.Fprintf(os.Stderr, "espresso: Unable to open %s\n", file);
        fp_primes = os.Stdout;
    }
    (void) sprintf(file, "%s.pi", filename);
    if (fp_pi_table = os.OpenFile(file, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)) == NULL {
        (void) fmt.Fprintf(os.Stderr, "espresso: Unable to open %s\n", file);
        fp_pi_table = os.Stdout;
    }
    FREE(file);
    }

    PLA = new_PLA();
    PLA_labels(PLA);

    fpr_header(fp_primes, PLA, F_type);
    free_PLA(PLA);

    (void) fmt.Fprintf(fp_primes, "# Essential primes are\n");
    // inlining of macro foreach_set
    for p=E.data,last=p+E.count*E.wsize;p<last;p+=E.wsize {
    (void) fmt.Fprintf(fp_primes, "%s\n", pc1(p));
    }
    (void) fmt.Fprintf(fp_primes, "# Totally redundant primes are\n");
    // inlining of macro foreach_set
    for p=Rt.data,last=p+Rt.count*Rt.wsize;p<last;p+=Rt.wsize {
    (void) fmt.Fprintf(fp_primes, "%s\n", pc1(p));
    }
    (void) fmt.Fprintf(fp_primes, "# Partially redundant primes are\n");
    // inlining of macro foreach_set
    for p=Rp.data,last=p+Rp.count*Rp.wsize;p<last;p+=Rp.wsize {
    (void) fmt.Fprintf(fp_primes, "%s\n", pc1(p));
    }
    if fp_primes != os.Stdout {
    (void) fclose(fp_primes);
    }
    
    sm_write(fp_pi_table, table);
    if fp_pi_table != os.Stdout {
    (void) fclose(fp_pi_table);
    }
}
//ABC_NAMESPACE_IMPL_END


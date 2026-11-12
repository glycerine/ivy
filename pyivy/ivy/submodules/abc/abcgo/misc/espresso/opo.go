package espresso

import (
	"fmt"
	"os"
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


/*
 *   Phase assignment technique (T. Sasao):
 *
 *      1. create a function with 2*m outputs which implements the
 *      original function and its complement for each output
 *
 *      2. minimize this function
 *
 *      3. choose the minimum number of prime implicants from the
 *      result of step 2 which are needed to realize either a function
 *      or its complement for each output
 *
 *  Step 3 is performed in a rather crude way -- by simply multiplying
 *  out a large expression of the form:
 *
 *      I = (ab + cdef)(acd + bgh) ...
 *
 *  which is a product of m expressions where each expression has two
 *  product terms -- one representing which primes are needed for the
 *  function, and one representing which primes are needed for the
 *  complement.  The largest product term resulting shows which primes
 *  to keep to implement one function or the other for each output.
 *  For problems with many outputs, this may grind to a
 *  halt.
 *
 *  Untried: form complement of I and use unate_complement ...
 *
 *  I have unsuccessfully tried several modifications to the basic
 *  algorithm.  The first is quite simple: use Sasao's technique, but
 *  only commit to a single output at a time (rather than all
 *  outputs).  The goal would be that the later minimizations can "take
 *  into account" the partial assignment at each step.  This is
 *  expensive (m+1 minimizations rather than 2), and the results are
 *  discouraging.
 *
 *  The second modification is rather complicated.  The result from the
 *  minimization in step 2 is guaranteed to be minimal.  Hence, for
 *  each output, the set of primes with a 1 in that output are both
 *  necessary and sufficient to implement the function.  Espresso
 *  achieves the minimality using the routine MAKE_SPARSE.  The
 *  modification is to prevent MAKE_SPARSE from running.  Hence, there
 *  are potentially many subsets of the set of primes with a 1 in a
 *  column which can be used to implement that output.  We use
 *  IRREDUNDANT to enumerate all possible subsets and then proceed as
 *  before.
 */

var opo_no_make_sparse int32;
var opo_repeated int32;
var opo_exact int32;
// static void minimize();


void phase_assignment(PLA, opo_strategy)
var PLA pPLA;
var opo_strategy int32;
{
    opo_no_make_sparse = opo_strategy % 2;
    skip_make_sparse = opo_no_make_sparse;
    opo_repeated = (opo_strategy / 2) % 2;
    opo_exact = (opo_strategy / 4) % 2;

    /* Determine a phase assignment */
    if PLA.phase != NULL {
    FREE(PLA.phase);
    }

    if opo_repeated != 0 {
    // inlining of macro set_save
    PLA.phase = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cube.fullset))), cube.fullset);
    repeated_phase_assignment(PLA);
    } else {
    PLA.phase = find_phase(PLA, 0, (pcube) NULL);
    }

    /* Now minimize with this assignment */
    skip_make_sparse = FALSE;
    (void) set_phase(PLA);
    minimize(PLA);
}

/*
 *  repeated_phase_assignment -- an alternate strategy which commits
 *  to a single phase assignment a step at a time.  Performs m + 1
 *  minimizations !
 */
void repeated_phase_assignment(PLA)
var PLA pPLA;
{
    int i;
    pcube phase;

    for i = 0; i < cube.part_size[cube.output]; i++ {

    /* Find best assignment for all undecided outputs */
    phase = find_phase(PLA, i, PLA.phase);

    /* Commit for only a single output ... */
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if ! (phase[(((cube.first_part[cube.output] + i) >> LOGBPI) + 1)] & (1 << ((cube.first_part[cube.output] + i) & (BPI-1)))) {
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (PLA.phase[(((cube.first_part[cube.output] + i) >> LOGBPI) + 1)] &= ~ (1 << ((cube.first_part[cube.output] + i) & (BPI-1))));
    }

    if trace || summary {
        fmt.Printf("\nOPO loop for output #%d\n", i);
        fmt.Printf("PLA->phase is %s\n", pc1(PLA.phase));
        fmt.Printf("phase      is %s\n", pc1(phase));
    }
    // inlining of macro set_free
    FREE(phase);
    }
}


/*
 *  find_phase -- find a phase assignment for the PLA for all outputs starting
 *  with output number first_output.
 */
pcube find_phase(PLA, first_output, phase1)
var PLA pPLA;
var first_output int32;
pcube phase1;
{
    pcube phase;
    pPLA PLA1;

    // inlining of macro set_save
    phase = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cube.fullset))), cube.fullset);

    /* setup the double-phase characteristic function, resize the cube */
    PLA1 = new_PLA();
    PLA1.F = sf_save(PLA.F);
    PLA1.R = sf_save(PLA.R);
    PLA1.D = sf_save(PLA.D);
    if phase1 != NULL {
    // inlining of macro set_save
    PLA1.phase = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(phase1))), phase1);
    (void) set_phase(PLA1);
    }
    // inlining of macro EXEC_S
    {long t=ptime();output_phase_setup(PLA1, first_output);if(summary)print_trace(PLA1.F,"OPO-SETUP ",ptime()-t);};

    /* minimize the double-phase function */
    minimize(PLA1);

    /* set the proper phases according to what gives a minimum solution */
    // inlining of macro EXEC_S
    {long t=ptime();PLA1.F = opo(phase, PLA1.F, PLA1.D, PLA1.R, first_output);if(summary)print_trace(PLA1.F,"OPO       ",ptime()-t);};
    free_PLA(PLA1);

    /* set the cube structure to reflect the old size */
    setdown_cube();
    cube.part_size[cube.output] -=
    (cube.part_size[cube.output] - first_output) / 2;
    cube_setup();

    return phase;
}

/*
 *  opo -- multiply the expression out to determine a minimum subset of
 *  primes.
 */

/*ARGSUSED*/
pcover opo(phase, T, D, R, first_output)
pcube phase;
pcover T, D, R;
var first_output int32;
{
    int offset, output, i, last_output, ind;
    pset pdest, select, p, p1, last, last1, not_covered, tmp;
    pset_family temp, T1, T2;

    /* must select all primes for outputs [0 .. first_output-1] */
    // inlining of macro set_full
    select = set_fill(ALLOC(unsigned int, SET_SIZE(T.count)), T.count);
    for output = 0; output < first_output; output++ {
    ind = cube.first_part[cube.output] + output;
    // inlining of macro foreachi_set
    for p, i = T.data, 0;i<T.count;p+=T.wsize,i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((ind) >> LOGBPI) + 1)] & (1 << ((ind) & (BPI-1))) {
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (select[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
        }
    }
    }

    /* Recursively perform the intersections */
    offset = (cube.part_size[cube.output] - first_output) / 2;
    last_output = first_output + offset - 1;
    temp = opo_recur(T, D, select, offset, first_output, last_output);

    /* largest set is on top -- select primes which are inferred from it */
    pdest = temp.data;
    // inlining of macro new_cover
    T1 = sf_new(T.count, cube.size);
    // inlining of macro foreachi_set
    for p, i = T.data, 0;i<T.count;p+=T.wsize,i++ {
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if ! (pdest[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) {
        T1 = sf_addset(T1, p);
    }
    }

    // inlining of macro set_free
    FREE(select);
    sf_free(temp);

    /* finding phases is difficult -- see which functions are not covered */
    T2 = complement(cube1list(T1));
    // inlining of macro new_cube
    // inlining of macro set_new
    not_covered = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    tmp = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro foreach_set
    for p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize {
    // inlining of macro foreach_set
    for p1=T2.data,last1=p1+T2.count*T2.wsize;p1<last1;p1+=T2.wsize {
        if cdist0(p, p1) {
        (void) set_or(not_covered, not_covered, set_and(tmp, p, p1));
        }
    }
    }
    // inlining of macro free_cover
    sf_free(T);
    // inlining of macro free_cover
    sf_free(T2);
    // inlining of macro set_free
    FREE(tmp);

    /* Now reflect the phase choice in a single cube */
    for output = first_output; output <= last_output; output++ {
    ind = cube.first_part[cube.output] + output;
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if not_covered[(((ind) >> LOGBPI) + 1)] & (1 << ((ind) & (BPI-1))) {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if not_covered[(((ind + offset) >> LOGBPI) + 1)] & (1 << ((ind + offset) & (BPI-1))) {
        fatal("error in output phase assignment");
        } else {
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (phase[(((ind) >> LOGBPI) + 1)] &= ~ (1 << ((ind) & (BPI-1))));
        }
    }
    }
    // inlining of macro set_free
    FREE(not_covered);
    return T1;
}

pset_family opo_recur(T, D, select, offset, first, last)
pcover T, D;
pcube select;
var offset, first, last int32;
{
    static int level = 0;
    int middle;
    pset_family sl, sr, temp;

    level++;
    if first == last {
     
                             
                                                                     
            
                                                          
     
     
    temp = opo_leaf(T, select, first, first + offset);
      
    } else {
    middle = (first + last) / 2;
    sl = opo_recur(T, D, select, offset, first, middle);
    sr = opo_recur(T, D, select, offset, middle+1, last);
    temp = unate_intersect(sl, sr, level == 1);
    if trace {
        fmt.Printf("# OPO[%d]: %4d = %4d x %4d, time = %s\n", level - 1, temp.count, sl.count, sr.count, print_time(ptime()));
        (void) fflush(os.Stdout);
    }
    // inlining of macro free_cover
    sf_free(sl);
    // inlining of macro free_cover
    sf_free(sr);
    }
    level--;
    return temp;
}


pset_family opo_leaf(T, select, out1, out2)
register pcover T;
var select pset;
var out1, out2 int32;
{
    register pset_family temp;
    register pset p, pdest;
    register int i;

    out1 += cube.first_part[cube.output];
    out2 += cube.first_part[cube.output];

    /* Allocate space for the result */
    temp = sf_new(2, T.count);

    /* Find which primes are needed for the ON-set of this fct */
    // inlining of macro GETSET
    pdest = ((temp).data + (temp).wsize * (temp.count++));
    (void) set_copy(pdest, select);
    // inlining of macro foreachi_set
    for p, i = T.data, 0;i<T.count;p+=T.wsize,i++ {
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if p[(((out1) >> LOGBPI) + 1)] & (1 << ((out1) & (BPI-1))) {
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pdest[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
    }
    }

    /* Find which primes are needed for the OFF-set of this fct */
    // inlining of macro GETSET
    pdest = ((temp).data + (temp).wsize * (temp.count++));
    (void) set_copy(pdest, select);
    // inlining of macro foreachi_set
    for p, i = T.data, 0;i<T.count;p+=T.wsize,i++ {
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if p[(((out2) >> LOGBPI) + 1)] & (1 << ((out2) & (BPI-1))) {
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pdest[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
    }
    }

    return temp;
}

     
                                                   
            
            
                                                                  
 
                   
                     
                                    

                                             
                       
                                
                           
                  
     
                           
                     
     

                                           
                                                 
                                             

                       
                   
 


                                        
            
                     
      
 
                               
              
                 
                               
          

                                      

                                                                   
                             
                           
                          
                                      
                                                   
                              
                         
     

                                                                   
                             
                                
                              
                                
                                  
                                                         
                     
     
                                
                                 
                             
     

                                                                    
                                  
                                 
                             
                         
                                    
                    
     
     
                                 

                                                             
                                
                                   
                             
     

                         
                
                   
 
      

/*
 *  Take a PLA (ON-set, OFF-set and DC-set) and create the
 *  "double-phase characteristic function" which is merely a new
 *  function which has twice as many outputs and realizes both the
 *  function and the complement.
 *
 *  The cube structure is assumed to represent the PLA upon entering.
 *  It will be modified to represent the double-phase function upon
 *  exit.
 *
 *  Only the outputs numbered starting with "first_output" are
 *  duplicated in the output part
 */

void output_phase_setup(PLA, first_output)
INOUT pPLA PLA;
var first_output int32;
{
    pcover F, R, D;
    pcube mask, mask1, last;
    int first_part, offset;
    bool save;
    register pcube p, pr, pf;
    register int i, last_part;

    if (cube.output == -1)
    fatal("output_phase_setup: must have an output");

    F = PLA.F;
    D = PLA.D;
    R = PLA.R;
    first_part = cube.first_part[cube.output] + first_output;
    last_part = cube.last_part[cube.output];
    offset = cube.part_size[cube.output] - first_output;

    /* Change the output size, setup the cube structure */
    setdown_cube();
    cube.part_size[cube.output] += offset;
    cube_setup();

    /* Create a mask to select that part of the cube which isn't changing */
    // inlining of macro set_save
    mask = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cube.fullset))), cube.fullset);
    for(i = first_part; i < cube.size; i++)
    // inlining of macro set_remove
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (mask[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
    // inlining of macro set_save
    mask1 = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(mask))), mask);
    for i = cube.first_part[cube.output]; i < first_part; i++ {
    // inlining of macro set_remove
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (mask1[(((i) >> LOGBPI) + 1)] &= ~ (1 << ((i) & (BPI-1))));
    }

    // inlining of macro new_cover
    PLA.F = sf_new(F.count + R.count, cube.size);
    // inlining of macro new_cover
    PLA.R = sf_new(F.count + R.count, cube.size);
    // inlining of macro new_cover
    PLA.D = sf_new(D.count, cube.size);

    // inlining of macro foreach_set
    for p=F.data,last=p+F.count*F.wsize;p<last;p+=F.wsize {
    // inlining of macro GETSET
    pf = ((PLA.F).data + (PLA.F).wsize * ((PLA.F).count++));
    // inlining of macro GETSET
    pr = ((PLA.R).data + (PLA.R).wsize * ((PLA.R).count++));
    INLINEset_and(pf, mask, p);
    INLINEset_and(pr, mask1, p);
    for(i = first_part; i <= last_part; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if ((p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))))
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pf[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
    save = FALSE;
    for(i = first_part; i <= last_part; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if ((p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))))
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        save = TRUE, (pr[(((i+offset) >> LOGBPI) + 1)] |= 1 << ((i+offset) & (BPI-1)));
    if (! save) PLA.R.count--;
    }

    // inlining of macro foreach_set
    for p=R.data,last=p+R.count*R.wsize;p<last;p+=R.wsize {
    // inlining of macro GETSET
    pf = ((PLA.F).data + (PLA.F).wsize * ((PLA.F).count++));
    // inlining of macro GETSET
    pr = ((PLA.R).data + (PLA.R).wsize * ((PLA.R).count++));
    INLINEset_and(pf, mask1, p);
    INLINEset_and(pr, mask, p);
    save = FALSE;
    for(i = first_part; i <= last_part; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if ((p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))))
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        save = TRUE, (pf[(((i+offset) >> LOGBPI) + 1)] |= 1 << ((i+offset) & (BPI-1)));
    if (! save) PLA.F.count--;
    for(i = first_part; i <= last_part; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if ((p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))))
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pr[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
    }

    // inlining of macro foreach_set
    for p=D.data,last=p+D.count*D.wsize;p<last;p+=D.wsize {
    // inlining of macro GETSET
    pf = ((PLA.D).data + (PLA.D).wsize * ((PLA.D).count++));
    INLINEset_and(pf, mask, p);
    for(i = first_part; i <= last_part; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pf[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pf[(((i+offset) >> LOGBPI) + 1)] |= 1 << ((i+offset) & (BPI-1)));
        }
    }

    // inlining of macro free_cover
    sf_free(F);
    // inlining of macro free_cover
    sf_free(D);
    // inlining of macro free_cover
    sf_free(R);
    // inlining of macro set_free
    FREE(mask);
    // inlining of macro set_free
    FREE(mask1);
}

/*
 *  set_phase -- given a "cube" which describes which phases of the output
 *  are to be implemented, compute the appropriate on-set and off-set
 */
pPLA set_phase(PLA)
INOUT pPLA PLA;
{
    pcover F1, R1;
    register pcube last, p, outmask;
    register pcube temp=cube.temp[0], phase=PLA.phase, phase1=cube.temp[1];

    outmask = cube.var_mask[cube.num_vars - 1];
    set_diff(phase1, outmask, phase);
    set_or(phase1, set_diff(temp, cube.fullset, outmask), phase1);
    // inlining of macro new_cover
    F1 = sf_new((PLA.F).count + (PLA.R).count, cube.size);
    // inlining of macro new_cover
    R1 = sf_new((PLA.F).count + (PLA.R).count, cube.size);

    // inlining of macro foreach_set
    for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
    if (! setp_disjoint(set_and(temp, p, phase), outmask))
        // inlining of macro GETSET
        set_copy(((F1).data + (F1).wsize * (F1.count++)), temp);
    if (! setp_disjoint(set_and(temp, p, phase1), outmask))
        // inlining of macro GETSET
        set_copy(((R1).data + (R1).wsize * (R1.count++)), temp);
    }
    // inlining of macro foreach_set
    for p=PLA.R.data,last=p+PLA.R.count*PLA.R.wsize;p<last;p+=PLA.R.wsize {
    if (! setp_disjoint(set_and(temp, p, phase), outmask))
        // inlining of macro GETSET
        set_copy(((R1).data + (R1).wsize * (R1.count++)), temp);
    if (! setp_disjoint(set_and(temp, p, phase1), outmask))
        // inlining of macro GETSET
        set_copy(((F1).data + (F1).wsize * (F1.count++)), temp);
    }
    // inlining of macro free_cover
    sf_free(PLA.F);
    // inlining of macro free_cover
    sf_free(PLA.R);
    PLA.F = F1;
    PLA.R = R1;
    return PLA;
}

// Original macro:
// #define POW2(x)        (1 << (x))


void opoall(PLA, first_output, last_output, opo_strategy)
var PLA pPLA;
var first_output, last_output int32;
var opo_strategy int32;
{
    pcover F, D, R, best_F, best_D, best_R;
    int i, j, ind, num;
    pcube bestphase;

    opo_exact = opo_strategy;

    if PLA.phase != NULL {
    // inlining of macro set_free
    FREE(PLA.phase);
    }

    // inlining of macro set_save
    bestphase = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cube.fullset))), cube.fullset);
    best_F = sf_save(PLA.F);
    best_D = sf_save(PLA.D);
    best_R = sf_save(PLA.R);

    // inlining of macro POW2
    for i = 0; i < (1 << (last_output - first_output + 1)); i++ {

    /* save the initial PLA covers */
    F = sf_save(PLA.F);
    D = sf_save(PLA.D);
    R = sf_save(PLA.R);

    /* compute the phase cube for this iteration */
    // inlining of macro set_save
    PLA.phase = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cube.fullset))), cube.fullset);
    num = i;
    for j = last_output; j >= first_output; j-- {
        if num % 2 == 0 {
        ind = cube.first_part[cube.output] + j;
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (PLA.phase[(((ind) >> LOGBPI) + 1)] &= ~ (1 << ((ind) & (BPI-1))));
        }
        num /= 2;
    }

    /* set the phase and minimize */
    (void) set_phase(PLA);
    fmt.Printf("# phase is %s\n", pc1(PLA.phase));
    summary = TRUE;
    minimize(PLA);

    /* see if this is the best so far */
    if PLA.F.count < best_F.count {
        /* save new best solution */
        set_copy(bestphase, PLA.phase);
        sf_free(best_F);
        sf_free(best_D);
        sf_free(best_R);
        best_F = PLA.F;
        best_D = PLA.D;
        best_R = PLA.R;
    } else {
        /* throw away the solution */
        // inlining of macro free_cover
        sf_free(PLA.F);
        // inlining of macro free_cover
        sf_free(PLA.D);
        // inlining of macro free_cover
        sf_free(PLA.R);
    }
    // inlining of macro set_free
    FREE(PLA.phase);

    /* restore the initial PLA covers */
    PLA.F = F;
    PLA.D = D;
    PLA.R = R;
    }

    /* one more minimization to restore the best answer */
    PLA.phase = bestphase;
    sf_free(PLA.F);
    sf_free(PLA.D);
    sf_free(PLA.R);
    PLA.F = best_F;
    PLA.D = best_D;
    PLA.R = best_R;
}

static void minimize(PLA)
var PLA pPLA;
{
    if opo_exact != 0 {
    // inlining of macro EXEC_S
    {long t=ptime();PLA.F = minimize_exact(PLA.F,PLA.D,PLA.R,1);if(summary)print_trace(PLA.F,"EXACT",ptime()-t);};
    } else {
    // inlining of macro EXEC_S
    {long t=ptime();PLA.F = espresso(PLA.F, PLA.D, PLA.R);if(summary)print_trace(PLA.F,"ESPRESSO  ",ptime()-t);};
    }
}
//ABC_NAMESPACE_IMPL_END


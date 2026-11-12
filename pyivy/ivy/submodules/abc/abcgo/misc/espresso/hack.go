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


void map_dcset(PLA)
var PLA pPLA;
{
    int var, i;
    pcover Tplus, Tminus, Tplusbar, Tminusbar;
    pcover newf, term1, term2, dcset, dcsetbar;
    pcube cplus, cminus, last, p;

    if (PLA.label == NIL(char *) || PLA.label[0] == NIL(char))
    return;

    /* try to find a binary variable named "DONT_CARE" */
    var = -1;
    for i = 0; i < cube.num_binary_vars * 2; i++ {
    if strncmp(PLA.label[i], "DONT_CARE", 9) == 0 ||
      strncmp(PLA.label[i], "DONTCARE", 8) == 0 ||
      strncmp(PLA.label[i], "dont_care", 9) == 0 ||
      strncmp(PLA.label[i], "dontcare", 8) == 0 {
        var = i/2;
        break;
    }
    }
    if var == -1 {
    return;
    }

    /* form the cofactor cubes for the don't-care variable */
    // inlining of macro set_save
    cplus = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cube.fullset))), cube.fullset);
    // inlining of macro set_save
    cminus = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(cube.fullset))), cube.fullset);
    // inlining of macro set_remove
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (cplus[(((var*2) >> LOGBPI) + 1)] &= ~ (1 << ((var*2) & (BPI-1))));
    // inlining of macro set_remove
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (cminus[(((var*2 + 1) >> LOGBPI) + 1)] &= ~ (1 << ((var*2 + 1) & (BPI-1))));

    /* form the don't-care set */
    // inlining of macro EXEC
    {long t=ptime();simp_comp(cofactor(cube1list(PLA.F), cplus), &Tplus, &Tplusbar);if(trace)print_trace(Tplus,"simpcomp+",ptime()-t);};
    // inlining of macro EXEC
    {long t=ptime();simp_comp(cofactor(cube1list(PLA.F), cminus), &Tminus, &Tminusbar);if(trace)print_trace(Tminus,"simpcomp-",ptime()-t);};
    // inlining of macro EXEC
    {long t=ptime();term1 = cv_intersect(Tplus, Tminusbar);if(trace)print_trace(term1,"term1    ",ptime()-t);};
    // inlining of macro EXEC
    {long t=ptime();term2 = cv_intersect(Tminus, Tplusbar);if(trace)print_trace(term2,"term2    ",ptime()-t);};
    // inlining of macro EXEC
    {long t=ptime();dcset = sf_union(term1, term2);if(trace)print_trace(dcset,"union     ",ptime()-t);};
    // inlining of macro EXEC
    {long t=ptime();simp_comp(cube1list(dcset), &PLA.D, &dcsetbar);if(trace)print_trace(PLA.D,"simplify",ptime()-t);};
    // inlining of macro EXEC
    {long t=ptime();newf = cv_intersect(PLA.F, dcsetbar);if(trace)print_trace(PLA.F,"separate  ",ptime()-t);};
    // inlining of macro free_cover
    sf_free(PLA.F);
    PLA.F = newf;
    // inlining of macro free_cover
    sf_free(Tplus);
    // inlining of macro free_cover
    sf_free(Tminus);
    // inlining of macro free_cover
    sf_free(Tplusbar);
    // inlining of macro free_cover
    sf_free(Tminusbar);
    // inlining of macro free_cover
    sf_free(dcsetbar);

    /* remove any cubes dependent on the DONT_CARE variable */
    (void) sf_active(PLA.F);
    // inlining of macro foreach_set
    for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if ! (p[(((var*2) >> LOGBPI) + 1)] & (1 << ((var*2) & (BPI-1)))) || ! (p[(((var*2+1) >> LOGBPI) + 1)] & (1 << ((var*2+1) & (BPI-1)))) {
        // inlining of macro RESET
        (p[0] &= ~ (ACTIVE));
    }
    }
    PLA.F = sf_inactive(PLA.F);

    /* resize the cube and delete the don't-care variable */
    setdown_cube();
    for i = 2*var+2; i < cube.size; i++ {
    PLA.label[i-2] = PLA.label[i];
    }
    for i = var+1; i < cube.num_vars; i++ {
    cube.part_size[i-1] = cube.part_size[i];
    }
    cube.num_binary_vars--;
    cube.num_vars--;
    cube_setup();
    PLA.F = sf_delc(PLA.F, 2*var, 2*var+1);
    PLA.D = sf_delc(PLA.D, 2*var, 2*var+1);
}

void map_output_symbolic(PLA)
var PLA pPLA;
{
    pset_family newF, newD;
    pset compress;
    symbolic_t *p1;
    symbolic_list_t *p2;
    int i, bit, tot_size, base, old_size;

    /* Remove the DC-set from the ON-set (is this necessary ??) */
    if PLA.D.count > 0 {
    sf_free(PLA.F);
    PLA.F = complement(cube2list(PLA.D, PLA.R));
    }

    /* tot_size = width added for all symbolic variables */
    tot_size = 0;
    for p1=PLA.symbolic_output; p1!=NIL(symbolic_t); p1=p1.next {
    for p2=p1.symbolic_list; p2!=NIL(symbolic_list_t); p2=p2.next {
        if p2.pos<0 || p2.pos>=cube.part_size[cube.output] {
        fatal("symbolic-output index out of range");
/*        } else if (p2->variable != cube.output) {
        fatal("symbolic-output label must be an output");*/
        }
    }
    tot_size += 1 << p1.symbolic_list_length;
    }

    /* adjust the indices to skip over new outputs */
    for p1=PLA.symbolic_output; p1!=NIL(symbolic_t); p1=p1.next {
    for p2=p1.symbolic_list; p2!=NIL(symbolic_list_t); p2=p2.next {
        p2.pos += tot_size;
    }
    }

    /* resize the cube structure -- add enough for the one-hot outputs */
    old_size = cube.size;
    cube.part_size[cube.output] += tot_size;
    setdown_cube();
    cube_setup();

    /* insert space in the output part for the one-hot output */
    base = cube.first_part[cube.output];
    PLA.F = sf_addcol(PLA.F, base, tot_size);
    PLA.D = sf_addcol(PLA.D, base, tot_size);
    PLA.R = sf_addcol(PLA.R, base, tot_size);

    /* do the real work */
    for p1=PLA.symbolic_output; p1!=NIL(symbolic_t); p1=p1.next {
    // inlining of macro new_cover
    newF = sf_new(100, cube.size);
    // inlining of macro new_cover
    newD = sf_new(100, cube.size);
    find_inputs(NIL(set_family_t), PLA, p1.symbolic_list, base, 0,
                &newF, &newD);
/*
 *  Not sure what this means
    find_dc_inputs(PLA, p1->symbolic_list,
                base, 1 << p1->symbolic_list_length, &newF, &newD);
 */
    // inlining of macro free_cover
    sf_free(PLA.F);
    PLA.F = newF;
/*
 *  retain OLD DC-set -- but we've lost the don't-care arc information
 *  (it defaults to branch to the zero state)
    free_cover(PLA->D);
    PLA->D = newD;
 */
    // inlining of macro free_cover
    sf_free(newD);
    base += 1 << p1.symbolic_list_length;
    }

    /* delete the old outputs, and resize the cube */
    // inlining of macro set_full
    compress = set_fill(ALLOC(unsigned int, SET_SIZE(newF.sf_size)), newF.sf_size);
    for p1=PLA.symbolic_output; p1!=NIL(symbolic_t); p1=p1.next {
    for p2=p1.symbolic_list; p2!=NIL(symbolic_list_t); p2=p2.next {
        bit = cube.first_part[cube.output] + p2.pos;
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (compress[(((bit) >> LOGBPI) + 1)] &= ~ (1 << ((bit) & (BPI-1))));
    }
    }
    cube.part_size[cube.output] -= newF.sf_size - set_ord(compress);
    setdown_cube();
    cube_setup();
    PLA.F = sf_compress(PLA.F, compress);
    PLA.D = sf_compress(PLA.D, compress);
    if (cube.size != PLA.F.sf_size) fatal("error");

    /* Quick minimization */
    PLA.F = sf_contain(PLA.F);
    PLA.D = sf_contain(PLA.D);
    for i = 0; i < cube.num_vars; i++ {
    PLA.F = d1merge(PLA.F, i);
    PLA.D = d1merge(PLA.D, i);
    }
    PLA.F = sf_contain(PLA.F);
    PLA.D = sf_contain(PLA.D);

    // inlining of macro free_cover
    sf_free(PLA.R);
    // inlining of macro new_cover
    PLA.R = sf_new(0, cube.size);

    symbolic_hack_labels(PLA, PLA.symbolic_output,
                compress, cube.size, old_size, tot_size);
    // inlining of macro set_free
    FREE(compress);
}


void find_inputs(A, PLA, list, base, value, newF, newD)
pcover A;
var PLA pPLA;
var list *symbolic_list_t;
var base, value int32;
pcover *newF, *newD;
{
    pcover S, S1;
    register pset last, p;

    /*
     *  A represents th 'input' values for which the outputs assume
     *  the integer value 'value
     */
    if list == NIL(symbolic_list_t) {
    /*
     *  Simulate these inputs against the on-set; then, insert into the
     *  new on-set a 1 in the proper position
     */
    S = cv_intersect(A, PLA.F);
    // inlining of macro foreach_set
    for p=S.data,last=p+S.count*S.wsize;p<last;p+=S.wsize {
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (p[(((base + value) >> LOGBPI) + 1)] |= 1 << ((base + value) & (BPI-1)));
    }
    *newF = sf_append(*newF, S);

    /*
     *  'simulate' these inputs against the don't-care set
    S = cv_intersect(A, PLA->D);
    *newD = sf_append(*newD, S);
     */

    } else {
    /* intersect and recur with the OFF-set */
    S = cof_output(PLA.R, cube.first_part[cube.output] + list.pos);
    if A != NIL(set_family_t) {
        S1 = cv_intersect(A, S);
        // inlining of macro free_cover
        sf_free(S);
        S = S1;
    }
    find_inputs(S, PLA, list.next, base, value*2, newF, newD);
    // inlining of macro free_cover
    sf_free(S);

    /* intersect and recur with the ON-set */
    S = cof_output(PLA.F, cube.first_part[cube.output] + list.pos);
    if A != NIL(set_family_t) {
        S1 = cv_intersect(A, S);
        // inlining of macro free_cover
        sf_free(S);
        S = S1;
    }
    find_inputs(S, PLA, list.next, base, value*2 + 1, newF, newD);
    // inlining of macro free_cover
    sf_free(S);
    }
}


     
                                                   
         
                      
                 
                    
 
                    
                        
                          
                   

                                                                       
                          
                                                         
                                                                   
                                 
              
            
                                
                      
                      
               
     
     

                                
                                

                                
                             
                                           
                         
     
     
                                
                  
 
      

void map_symbolic(PLA)
var PLA pPLA;
{
    symbolic_t *p1;
    symbolic_list_t *p2;
    int var, base, num_vars, num_binary_vars, *new_part_size;
    int new_size, size_added, num_deleted_vars, num_added_vars, newvar;
    pset compress;

    /* Verify legal values are in the symbolic lists */
    for p1 = PLA.symbolic; p1 != NIL(symbolic_t); p1 = p1.next {
    for p2=p1.symbolic_list; p2!=NIL(symbolic_list_t); p2=p2.next {
        if p2.variable  < 0 || p2.variable >= cube.num_binary_vars {
        fatal(".symbolic requires binary variables");
        }
    }
    }

    /*
     *  size_added = width added for all symbolic variables
     *  num_deleted_vars = # binary variables to be deleted
     *  num_added_vars = # new mv variables
     *  compress = a cube which will be used to compress the set families
     */
    size_added = 0;
    num_added_vars = 0;
    for p1 = PLA.symbolic; p1 != NIL(symbolic_t); p1 = p1.next {
    size_added += 1 << p1.symbolic_list_length;
    num_added_vars++;
    }
    // inlining of macro set_full
    compress = set_fill(ALLOC(unsigned int, SET_SIZE(PLA.F.sf_size + size_added)), PLA.F.sf_size + size_added);
    for p1 = PLA.symbolic; p1 != NIL(symbolic_t); p1 = p1.next {
    for p2=p1.symbolic_list; p2!=NIL(symbolic_list_t); p2=p2.next {
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (compress[(((p2.variable*2) >> LOGBPI) + 1)] &= ~ (1 << ((p2.variable*2) & (BPI-1))));
        // inlining of macro set_remove
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (compress[(((p2.variable*2+1) >> LOGBPI) + 1)] &= ~ (1 << ((p2.variable*2+1) & (BPI-1))));
    }
    }
    num_deleted_vars = ((PLA.F.sf_size + size_added) - set_ord(compress))/2;

    /* compute the new cube constants */
    num_vars = cube.num_vars - num_deleted_vars + num_added_vars;
    num_binary_vars = cube.num_binary_vars - num_deleted_vars;
    new_size = cube.size - num_deleted_vars*2 + size_added;
    new_part_size = ALLOC(int, num_vars);
    new_part_size[num_vars-1] = cube.part_size[cube.num_vars-1];
    for var = cube.num_binary_vars; var < cube.num_vars-1; var++ {
    new_part_size[var-num_deleted_vars] = cube.part_size[var];
    }

    /* re-size the covers, opening room for the new mv variables */
    base = cube.first_part[cube.output];
    PLA.F = sf_addcol(PLA.F, base, size_added);
    PLA.D = sf_addcol(PLA.D, base, size_added);
    PLA.R = sf_addcol(PLA.R, base, size_added);

    /* compute the values for the new mv variables */
    newvar = (cube.num_vars - 1) - num_deleted_vars;
    for p1 = PLA.symbolic; p1 != NIL(symbolic_t); p1 = p1.next {
    PLA.F = map_symbolic_cover(PLA.F, p1.symbolic_list, base);
    PLA.D = map_symbolic_cover(PLA.D, p1.symbolic_list, base);
    PLA.R = map_symbolic_cover(PLA.R, p1.symbolic_list, base);
    base += 1 << p1.symbolic_list_length;
    new_part_size[newvar++] = 1 << p1.symbolic_list_length;
    }

    /* delete the binary variables which disappear */
    PLA.F = sf_compress(PLA.F, compress);
    PLA.D = sf_compress(PLA.D, compress);
    PLA.R = sf_compress(PLA.R, compress);

    symbolic_hack_labels(PLA, PLA.symbolic, compress,
        new_size, cube.size, size_added);
    setdown_cube();
    FREE(cube.part_size);
    cube.num_vars = num_vars;
    cube.num_binary_vars = num_binary_vars;
    cube.part_size = new_part_size;
    cube_setup();
    // inlining of macro set_free
    FREE(compress);
}


pcover map_symbolic_cover(T, list, base)
pcover T;
var list *symbolic_list_t;
var base int32;
{
    pset last, p;
    // inlining of macro foreach_set
    for p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize {
    form_bitvector(p, base, 0, list);
    }
    return T;
}


void form_bitvector(p, base, value, list)
var p pset;            /* old cube, looking at binary variables */
var base int32;        /* where in mv cube the new variable starts */
var value int32;        /* current value for this recursion */
var list *symbolic_list_t;    /* current place in the symbolic list */
{
    if list == NIL(symbolic_list_t) {
    // inlining of macro set_insert
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (p[(((base + value) >> LOGBPI) + 1)] |= 1 << ((base + value) & (BPI-1)));
    } else {
    // inlining of macro GETINPUT
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    switch (p[(((2*list.variable) >> LOGBPI) + 1)] >> ((2*list.variable) & (BPI-1))) & 3 {
        case ZERO:
        form_bitvector(p, base, value*2, list.next);
        break;
        case ONE:
        form_bitvector(p, base, value*2+1, list.next);
        break;
        case TWO:
        form_bitvector(p, base, value*2, list.next);
        form_bitvector(p, base, value*2+1, list.next);
        break;
        default:
        fatal("bad cube in form_bitvector");
    }
    }
}


void symbolic_hack_labels(PLA, list, compress, new_size, old_size, size_added)
var PLA pPLA;
var list *symbolic_t;
var compress pset;
var new_size, old_size, size_added int32;
{
    int i, base;
    char **oldlabel;
    symbolic_t *p1;
    symbolic_label_t *p3;

    /* hack with the labels */
    if ((oldlabel = PLA.label) == NIL(char *))
    return;
    PLA.label = ALLOC(char *, new_size);
    for i = 0; i < new_size; i++ {
    PLA.label[i] = NIL(char);
    }

    /* copy the binary variable labels and unchanged mv variable labels */
    base = 0;
    for i = 0; i < cube.first_part[cube.output]; i++ {
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if compress[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
        PLA.label[base++] = oldlabel[i];
    } else {
        if oldlabel[i] != NIL(char) {
        FREE(oldlabel[i]);
        }
    }
    }

    /* add the user-defined labels for the symbolic outputs */
    for p1 = list; p1 != NIL(symbolic_t); p1 = p1.next {
    p3 = p1.symbolic_label;
    for i = 0; i < (1 << p1.symbolic_list_length); i++ {
        if p3 == NIL(symbolic_label_t) {
        PLA.label[base+i] = ALLOC(char, 10);
        (void) sprintf(PLA.label[base+i], "X%d", i);
        } else {
        PLA.label[base+i] = p3.label;
        p3 = p3.next;
        }
    }
    base += 1 << p1.symbolic_list_length;
    }

    /* copy the labels for the binary outputs which remain */
    for i = cube.first_part[cube.output]; i < old_size; i++ {
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if compress[(((i + size_added) >> LOGBPI) + 1)] & (1 << ((i + size_added) & (BPI-1))) {
        PLA.label[base++] = oldlabel[i];
    } else {
        if oldlabel[i] != NIL(char) {
        FREE(oldlabel[i]);
        }
    }
    }
    FREE(oldlabel);
}

static pcover fsm_simplify(F)
pcover F;
{
    pcover D, R;
    // inlining of macro new_cover
    D = sf_new(0, cube.size);
    R = complement(cube1list(F));
    F = espresso(F, D, R);
    // inlining of macro free_cover
    sf_free(D);
    // inlining of macro free_cover
    sf_free(R);
    return F;
}


void disassemble_fsm(PLA, verbose_mode)
var PLA pPLA;
var verbose_mode int32;
{
    int nin, nstates, nout;
    int before, after, present_state, next_state, i, j;
    pcube next_state_mask, present_state_mask, state_mask, p, p1, last;
    pcover go_nowhere, F, tF;

    /* We make the DISGUSTING assumption that the first 'n' outputs have
     *  been created by .symbolic-output, and represent a one-hot encoding
     * of the next state.  'n' is the size of the second-to-last multiple-
     * valued variable (i.e., before the outputs
     */

    if cube.num_vars - cube.num_binary_vars != 2 {
    (void) fmt.Fprintf(os.Stderr, "use .symbolic and .symbolic-output to specify\n");
    (void) fmt.Fprintf(os.Stderr, "the present state and next state field information\n");
    fatal("disassemble_pla: need two multiple-valued variables\n");
    }

    nin = cube.num_binary_vars;
    nstates = cube.part_size[cube.num_binary_vars];
    nout = cube.part_size[cube.num_vars - 1];
    if nout < nstates {
    (void) fmt.Fprintf(os.Stderr, "use .symbolic and .symbolic-output to specify\n");
    (void) fmt.Fprintf(os.Stderr, "the present state and next state field information\n");
    fatal("disassemble_pla: # outputs < # states\n");
    }


    present_state = cube.first_part[cube.num_binary_vars];
    // inlining of macro new_cube
    // inlining of macro set_new
    present_state_mask = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    for i = 0; i < nstates; i++ {
    // inlining of macro set_insert
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (present_state_mask[(((i + present_state) >> LOGBPI) + 1)] |= 1 << ((i + present_state) & (BPI-1)));
    }

    next_state = cube.first_part[cube.num_binary_vars+1];
    // inlining of macro new_cube
    // inlining of macro set_new
    next_state_mask = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    for i = 0; i < nstates; i++ {
    // inlining of macro set_insert
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    (next_state_mask[(((i + next_state) >> LOGBPI) + 1)] |= 1 << ((i + next_state) & (BPI-1)));
    }

    // inlining of macro new_cube
    // inlining of macro set_new
    state_mask = set_or(set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size), next_state_mask, present_state_mask);

    // inlining of macro new_cover
    F = sf_new(10, cube.size);


    /*
     *  check for arcs which go from ANY state to state #i
     */
    for i = 0; i < nstates; i++ {
    // inlining of macro new_cover
    tF = sf_new(10, cube.size);
    // inlining of macro foreach_set
    for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
        if setp_implies(present_state_mask, p) { /* from any state ! */
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((next_state + i) >> LOGBPI) + 1)] & (1 << ((next_state + i) & (BPI-1))) {
            tF = sf_addset(tF, p);
        }
        }
    }
    before = tF.count;
    if before > 0 {
        tF = fsm_simplify(tF);
        /* don't allow the next state to disappear ... */
        // inlining of macro foreach_set
        for p=tF.data,last=p+tF.count*tF.wsize;p<last;p+=tF.wsize {
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (p[(((next_state + i) >> LOGBPI) + 1)] |= 1 << ((next_state + i) & (BPI-1)));
        }
        after = tF.count;
        F = sf_append(F, tF);
        if verbose_mode != 0 {
        fmt.Printf("# state EVERY to %d, before=%d after=%d\n", i, before, after);
        }
    }
    }


    /*
     *  some 'arcs' may NOT have a next state -- handle these
     *  we must unravel the present state part
     */
    // inlining of macro new_cover
    go_nowhere = sf_new(10, cube.size);
    // inlining of macro foreach_set
    for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
    if setp_disjoint(p, next_state_mask) { /* no next state !! */
        go_nowhere = sf_addset(go_nowhere, p);
    }
    }
    before = go_nowhere.count;
    go_nowhere = unravel_range(go_nowhere,
                cube.num_binary_vars, cube.num_binary_vars);
    after = go_nowhere.count;
    F = sf_append(F, go_nowhere);
    if verbose_mode != 0 {
    fmt.Printf("# state ANY to NOWHERE, before=%d after=%d\n", before, after);
    }


    /*
     *  minimize cover for all arcs from state #i to state #j
     */
    for i = 0; i < nstates; i++ {
    for j = 0; j < nstates; j++ {
        // inlining of macro new_cover
        tF = sf_new(10, cube.size);
        // inlining of macro foreach_set
        for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
        /* not EVERY state */
        if ! setp_implies(present_state_mask, p) {
            // inlining of macro is_in_set
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            if p[(((present_state + i) >> LOGBPI) + 1)] & (1 << ((present_state + i) & (BPI-1))) {
            // inlining of macro is_in_set
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            if p[(((next_state + j) >> LOGBPI) + 1)] & (1 << ((next_state + j) & (BPI-1))) {
                // inlining of macro set_save
                p1 = set_copy(ALLOC(unsigned int, SET_SIZE(NELEM(p))), p);
                set_diff(p1, p1, state_mask);
                // inlining of macro set_insert
                // inlining of macro WHICH_WORD
                // inlining of macro WHICH_BIT
                (p1[(((present_state + i) >> LOGBPI) + 1)] |= 1 << ((present_state + i) & (BPI-1)));
                // inlining of macro set_insert
                // inlining of macro WHICH_WORD
                // inlining of macro WHICH_BIT
                (p1[(((next_state + j) >> LOGBPI) + 1)] |= 1 << ((next_state + j) & (BPI-1)));
                tF = sf_addset(tF, p1);
                // inlining of macro set_free
                FREE(p1);
            }
            }
        }
        }
        before = tF.count;
        if before > 0 {
        tF = fsm_simplify(tF);
        /* don't allow the next state to disappear ... */
        // inlining of macro foreach_set
        for p=tF.data,last=p+tF.count*tF.wsize;p<last;p+=tF.wsize {
            // inlining of macro set_insert
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            (p[(((next_state + j) >> LOGBPI) + 1)] |= 1 << ((next_state + j) & (BPI-1)));
        }
        after = tF.count;
        F = sf_append(F, tF);
        if verbose_mode != 0 {
            fmt.Printf("# state %d to %d, before=%d after=%d\n", i, j, before, after);
        }
        }
    }
    }


    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(state_mask);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(present_state_mask);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(next_state_mask);

    // inlining of macro free_cover
    sf_free(PLA.F);
    PLA.F = F;
    // inlining of macro free_cover
    sf_free(PLA.D);
    // inlining of macro new_cover
    PLA.D = sf_new(0, cube.size);

    setdown_cube();
    FREE(cube.part_size);
    cube.num_binary_vars = nin;
    cube.num_vars = nin + 3;
    cube.part_size = ALLOC(int, cube.num_vars);
    cube.part_size[cube.num_binary_vars] = nstates;
    cube.part_size[cube.num_binary_vars+1] = nstates;
    cube.part_size[cube.num_binary_vars+2] = nout - nstates;
    cube_setup();

    // inlining of macro foreach_set
    for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
    kiss_print_cube(os.Stdout, PLA, p, "~1");
    }
}
//ABC_NAMESPACE_IMPL_END


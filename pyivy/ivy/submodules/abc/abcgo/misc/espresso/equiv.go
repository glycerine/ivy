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
//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START



void find_equiv_outputs(PLA)
var PLA pPLA;
{
    int i, j, ipart, jpart, some_equiv;
    pcover *R, *F;

    some_equiv = FALSE;

    makeup_labels(PLA);

    F = ALLOC(pcover, cube.part_size[cube.output]);
    R = ALLOC(pcover, cube.part_size[cube.output]);

    for i = 0; i < cube.part_size[cube.output]; i++ {
    ipart = cube.first_part[cube.output] + i;
    R[i] = cof_output(PLA.R, ipart);
    F[i] = complement(cube1list(R[i]));
    }
    
    for i = 0; i < cube.part_size[cube.output]-1; i++ {
    for j = i+1; j < cube.part_size[cube.output]; j++ {
        ipart = cube.first_part[cube.output] + i;
        jpart = cube.first_part[cube.output] + j;

        if check_equiv(F[i], F[j]) {
        (void) fmt.Printf("# Outputs %d and %d (%s and %s) are equivalent\n", i, j, PLA.label[ipart], PLA.label[jpart]);
        some_equiv = TRUE;
        } else if check_equiv(F[i], R[j]) {
        (void) fmt.Printf("# Outputs %d and NOT %d (%s and %s) are equivalent\n", i, j, PLA.label[ipart], PLA.label[jpart]);
        some_equiv = TRUE;
        } else if check_equiv(R[i], F[j]) {
        (void) fmt.Printf("# Outputs NOT %d and %d (%s and %s) are equivalent\n", i, j, PLA.label[ipart], PLA.label[jpart]);
        some_equiv = TRUE;
        } else if check_equiv(R[i], R[j]) {
        (void) fmt.Printf("# Outputs NOT %d and NOT %d (%s and %s) are equivalent\n", i, j, PLA.label[ipart], PLA.label[jpart]);
        some_equiv = TRUE;
        }
    }
    }

    if ! some_equiv {
    (void) fmt.Printf("# No outputs are equivalent\n");
    }

    for i = 0; i < cube.part_size[cube.output]; i++ {
    // inlining of macro free_cover
    sf_free(F[i]);
    // inlining of macro free_cover
    sf_free(R[i]);
    }
    FREE(F);
    FREE(R);
}



int check_equiv(f1, f2)
pcover f1, f2;
{
    register pcube *f1list, *f2list;
    register pcube p, last;

    f1list = cube1list(f1);
    // inlining of macro foreach_set
    for p=f2.data,last=p+f2.count*f2.wsize;p<last;p+=f2.wsize {
    if ! cube_is_covered(f1list, p) {
        return FALSE;
    }
    }
    // inlining of macro free_cubelist
    FREE(f1list[0]); FREE(f1list);;

    f2list = cube1list(f2);
    // inlining of macro foreach_set
    for p=f1.data,last=p+f1.count*f1.wsize;p<last;p+=f1.wsize {
    if ! cube_is_covered(f2list, p) {
        return FALSE;
    }
    }
    // inlining of macro free_cubelist
    FREE(f2list[0]); FREE(f2list);;

    return TRUE;
}
//ABC_NAMESPACE_IMPL_END


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
    Module: cubestr.c -- routines for managing the global cube structure
*/

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


/*
    cube_setup -- assume that the fields "num_vars", "num_binary_vars", and
    part_size[num_binary_vars .. num_vars-1] are setup, and initialize the
    rest of cube and cdata.

    If a part_size is < 0, then the field size is abs(part_size) and the
    field read from the input is symbolic.
*/
func cube_setup() {
    register int i, var;
    register pcube p;

    if cube.num_binary_vars < 0 || cube.num_vars < cube.num_binary_vars {
    fatal("cube size is silly, error in .i/.o or .mv"); }

    cube.num_mv_vars = cube.num_vars - cube.num_binary_vars;
    if cube.num_mv_vars > 0 {
        cube.output = cube.num_vars - 1;
    } else {
        cube.output = -1;
    }

    cube.size = 0;
    cube.first_part = ALLOC(int, cube.num_vars);
    cube.last_part = ALLOC(int, cube.num_vars);
    cube.first_word = ALLOC(int, cube.num_vars);
    cube.last_word = ALLOC(int, cube.num_vars);
    for var = 0; var < cube.num_vars; var++ {
    if var < cube.num_binary_vars {
        cube.part_size[var] = 2; }
    cube.first_part[var] = cube.size;
    // inlining of macro WHICH_WORD
    cube.first_word[var] = (((cube.size) >> LOGBPI) + 1);
    cube.size += ABS(cube.part_size[var]);
    cube.last_part[var] = cube.size - 1;
    // inlining of macro WHICH_WORD
    cube.last_word[var] = (((cube.size - 1) >> LOGBPI) + 1);
    }

    cube.var_mask = ALLOC(pset, cube.num_vars);
    cube.sparse = ALLOC(int, cube.num_vars);
    // inlining of macro new_cube
    // inlining of macro set_new
    cube.binary_mask = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    cube.mv_mask = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    for var = 0; var < cube.num_vars; var++ {
    // inlining of macro new_cube
    // inlining of macro set_new
    p = cube.var_mask[var] = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (p[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
    }
    if var < cube.num_binary_vars {
        INLINEset_or(cube.binary_mask, cube.binary_mask, p);
        cube.sparse[var] = 0;
    } else {
        INLINEset_or(cube.mv_mask, cube.mv_mask, p);
        cube.sparse[var] = 1;
    }
    }
    if cube.num_binary_vars == 0 {
    cube.inword = -1;
    } else {
    cube.inword = cube.last_word[cube.num_binary_vars - 1];
    cube.inmask = cube.binary_mask[cube.inword] & DISJOINT;
    }

    cube.temp = ALLOC(pset, CUBE_TEMP);
    for i = 0; i < CUBE_TEMP; i++ {
    // inlining of macro new_cube
    // inlining of macro set_new
    cube.temp[i] = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    }
    // inlining of macro new_cube
    // inlining of macro set_new
    cube.fullset = set_fill(set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size), cube.size);
    // inlining of macro new_cube
    // inlining of macro set_new
    cube.emptyset = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);

    cdata.part_zeros = ALLOC(int, cube.size);
    cdata.var_zeros = ALLOC(int, cube.num_vars);
    cdata.parts_active = ALLOC(int, cube.num_vars);
    cdata.is_unate = ALLOC(int, cube.num_vars);
}

/*
    setdown_cube -- free memory allocated for the cube/cdata structs
    (free's all but the part_size array)

    (I wanted to call this cube_setdown, but that violates the 8-character
    external routine limit on the IBM !)
*/
func setdown_cube() {
    register int i, var;

    FREE(cube.first_part);
    FREE(cube.last_part);
    FREE(cube.first_word);
    FREE(cube.last_word);
    FREE(cube.sparse);

    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cube.binary_mask);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cube.mv_mask);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cube.fullset);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cube.emptyset);
    for var = 0; var < cube.num_vars; var++ {
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cube.var_mask[var]);
    }
    FREE(cube.var_mask);

    for i = 0; i < CUBE_TEMP; i++ {
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(cube.temp[i]);
    }
    FREE(cube.temp);

    FREE(cdata.part_zeros);
    FREE(cdata.var_zeros);
    FREE(cdata.parts_active);
    FREE(cdata.is_unate);

    cube.first_part = cube.last_part = (int *) nil;
    cube.first_word = cube.last_word = (int *) nil;
    cube.sparse = (int *) nil;
    cube.binary_mask = cube.mv_mask = (pcube) nil;
    cube.fullset = cube.emptyset = (pcube) nil;
    cube.var_mask = cube.temp = (pcube *) nil;

    cdata.part_zeros = cdata.var_zeros = cdata.parts_active = (int *) nil;
    cdata.is_unate = (bool *) nil;
}


func save_cube_struct() {
    temp_cube_save = cube;              /* structure copy ! */
    temp_cdata_save = cdata;            /*      ""          */

    cube.first_part = cube.last_part = (int *) nil;
    cube.first_word = cube.last_word = (int *) nil;
    cube.part_size = (int *) nil;
    cube.binary_mask = cube.mv_mask = (pcube) nil;
    cube.fullset = cube.emptyset = (pcube) nil;
    cube.var_mask = cube.temp = (pcube *) nil;

    cdata.part_zeros = cdata.var_zeros = cdata.parts_active = (int *) nil;
    cdata.is_unate = (bool *) nil;
}


func restore_cube_struct() {
    cube = temp_cube_save;              /* structure copy ! */
    cdata = temp_cdata_save;            /*      ""          */
}
//ABC_NAMESPACE_IMPL_END


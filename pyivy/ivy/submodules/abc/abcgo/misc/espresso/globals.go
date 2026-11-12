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
//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


/*
 *    Global Variable Declarations
 */

unsigned int debug;              /* debug parameter */
var verbose_debug bool;              /* -v:  whether to print a lot */
var total_name [TIME_COUNT]string;    /* basic function names */
var total_time [TIME_COUNT]long;     /* time spent in basic fcts */
var total_calls [TIME_COUNT]int32;     /* # calls to each fct */

var echo_comments bool;         /* turned off by -eat option */
var echo_unknown_commands bool;     /* always true ?? */
var force_irredundant bool;          /* -nirr command line option */
var skip_make_sparse bool;
var kiss bool;                       /* -kiss command line option */
var pos bool;                        /* -pos command line option */
var print_solution bool;             /* -x command line option */
var recompute_onset bool;            /* -onset command line option */
var remove_essential bool;           /* -ness command line option */
var single_expand bool;              /* -fast command line option */
var summary bool;                    /* -s command line option */
var trace bool;                      /* -t command line option */
var unwrap_onset bool;               /* -nunwrap command line option */
var use_random_order bool;         /* -random command line option */
var use_super_gasp bool;         /* -strong command line option */
var filename string;             /* filename PLA was read from */

struct pla_types_struct pla_types[] = {
    {"-f", F_type},
    {"-r", R_type},
    {"-d", D_type},
    {"-fd", FD_type},
    {"-fr", FR_type},
    {"-dr", DR_type},
    {"-fdr", FDR_type},
    {"-fc", F_type | CONSTRAINTS_type},
    {"-rc", R_type | CONSTRAINTS_type},
    {"-dc", D_type | CONSTRAINTS_type},
    {"-fdc", FD_type | CONSTRAINTS_type},
    {"-frc", FR_type | CONSTRAINTS_type},
    {"-drc", DR_type | CONSTRAINTS_type},
    {"-fdrc", FDR_type | CONSTRAINTS_type},
    {"-pleasure", PLEASURE_type},
    {"-eqn", EQNTOTT_type},
    {"-eqntott", EQNTOTT_type},
    {"-kiss", KISS_type},
    {"-cons", CONSTRAINTS_type},
    {"-scons", SYMBOLIC_CONSTRAINTS_type},
    {0, 0}
};


struct cube_struct cube, temp_cube_save;
struct cdata_struct cdata, temp_cdata_save;

var bit_count [256]int32 = [256]int32{
  0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
};
//ABC_NAMESPACE_IMPL_END


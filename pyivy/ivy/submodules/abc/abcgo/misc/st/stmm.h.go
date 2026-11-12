package st

/*
 * Revision Control Information
 *
 * /projects/hsis/CVS/utilities/st/st.h,v
 * serdar
 * 1.1
 * 1993/07/29 01:00:21
 *
 */
/* LINTLIBRARY */

/* /projects/hsis/CVS/utilities/st/st.h,v 1.1 1993/07/29 01:00:21 serdar Exp */




//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START


/* These are potential duplicates. */
#ifndef EXTERN
#   ifdef __cplusplus
#       ifdef ABC_NAMESPACE
#           define EXTERN extern
#       else
#           define EXTERN extern "C"
#       endif
#   else
#       define EXTERN extern
#   endif
#endif

#ifndef ARGS
#define ARGS(protos) protos
#endif

typedef int (*stmm_compare_func_type)(char*, char*);
typedef int (*stmm_hash_func_type)(char*, int);

   
   
   

type stmm_table_entry struct {
    key string;
    record string;
    next *stmm_table_entry;
}

type stmm_table struct {
    compare stmm_compare_func_type;
    hash stmm_hash_func_type;
    num_bins int32;
    num_entries int32;
    max_density int32;
    reorder_flag int32;
    grow_factor float64;
    bins **stmm_table_entry;
    // memory manager to improve runtime and prevent memory fragmentation
    // added by alanmi - January 16, 2003
    pMemMan  any;
}

type stmm_generator struct {
    table *stmm_table;
    entry *stmm_table_entry;
    index int32;
}

// Original macro:
// #define stmm_is_member(table,key) stmm_lookup(table,key,(char **) 0)

// Original macro:
// #define stmm_count(table) ((table)->num_entries)


type stmm_retval int32
const 
( STMM_CONTINUE stmm_retval = iota
 STMM_STOP
 STMM_DELETE 
)


typedef enum stmm_retval (*STMM_PFSR) (char *, char *, char *);

EXTERN stmm_table *stmm_init_table_with_params
ARGS ((stmm_compare_func_type compare, stmm_hash_func_type hash, int size, int density, double grow_factor, int reorder_flag));
EXTERN stmm_table *stmm_init_table ARGS ((stmm_compare_func_type, stmm_hash_func_type));
EXTERN void stmm_free_table ARGS ((stmm_table *));
EXTERN int stmm_lookup ARGS ((stmm_table *, char *, char **));
EXTERN int stmm_lookup_int ARGS ((stmm_table *, char *, int *));
EXTERN int stmm_insert ARGS ((stmm_table *, char *, char *));
EXTERN int stmm_add_direct ARGS ((stmm_table *, char *, char *));
EXTERN int stmm_find_or_add ARGS ((stmm_table *, char *, char ***));
EXTERN int stmm_find ARGS ((stmm_table *, char *, char ***));
EXTERN stmm_table *stmm_copy ARGS ((stmm_table *));
EXTERN int stmm_delete ARGS ((stmm_table *, char **, char **));
EXTERN int stmm_delete_int ARGS ((stmm_table *, long *, char **));
EXTERN int stmm_foreach ARGS ((stmm_table *, STMM_PFSR, char *));
EXTERN int stmm_strhash ARGS ((char *, int));
EXTERN int stmm_numhash ARGS ((char *, int));
EXTERN int stmm_ptrhash ARGS ((char *, int));
EXTERN int stmm_numcmp ARGS ((char *, char *));
EXTERN int stmm_ptrcmp ARGS ((char *, char *));
EXTERN stmm_generator *stmm_init_gen ARGS ((stmm_table *));
EXTERN int stmm_gen ARGS ((stmm_generator *, char **, char **));
EXTERN int stmm_gen_int ARGS ((stmm_generator *, char **, long *));
EXTERN void stmm_free_gen ARGS ((stmm_generator *));
// additional functions
EXTERN void stmm_clean ARGS ((stmm_table *));



const STMM_DEFAULT_MAX_DENSITY =        5
const STMM_DEFAULT_INIT_TABLE_SIZE =   11
const STMM_DEFAULT_GROW_FACTOR =      2.0
const STMM_DEFAULT_REORDER_FLAG =       0

// added by Zhihong: no need for memory allocation
// Original macro:
// #define stmm_foreach_item2(tb, /* stmm_generator */gen, key, value) \
//     for(gen.table=(tb), gen.entry=NULL, gen.index=0; \
//         stmm_gen(&(gen),key,value);)


// Original macro:
// #define stmm_foreach_item(table, gen, key, value) \
//     for(gen=stmm_init_gen(table); stmm_gen(gen,key,value) || (stmm_free_gen(gen),0);)


// Original macro:
// #define stmm_foreach_item_int(table, gen, key, value) \
//     for(gen=stmm_init_gen(table); stmm_gen_int(gen,key,value) || (stmm_free_gen(gen),0);)


const STMM_OUT_OF_MEM = -10000

/*

// consider adding these other other similar definitions
#define st__table       stmm_table
#define st__insert      stmm_insert
#define st__delete      stmm_delete
#define st__lookup      stmm_lookup
#define st__init_table  stmm_init_table
#define st__free_table  stmm_free_table

*/



//ABC_NAMESPACE_HEADER_END





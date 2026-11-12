package st

/*
 * Revision Control Information
 *
 * /projects/hsis/CVS/utilities/st/st.c,v
 * serdar
 * 1.1
 * 1993/07/29 01:00:13
 *
 */
//#include <stdio.h>
//#include "misc/extra/extra.h"
//#include "stmm.h"

//ABC_NAMESPACE_IMPL_START


// Original macro:
// #define STMM_NUMCMP(x,y) ((x) != (y))

// Original macro:
// #define STMM_NUMHASH(x,size) (Abc_AbsInt((long)x)%(size))

//#define STMM_PTRHASH(x,size) ((int)((ABC_PTRUINT_T)(x)>>2)%size) //  64-bit bug fix 9/17/2007
// Original macro:
// #define STMM_PTRHASH(x,size) ((int)(((ABC_PTRUINT_T)(x)>>2)%size))

// Original macro:
// #define EQUAL(func, x, y) \
//     ((((func) == stmm_numcmp) || ((func) == stmm_ptrcmp)) ?\
//       (STMM_NUMCMP((x),(y)) == 0) : ((*func)((x), (y)) == 0))



// Original macro:
// #define do_hash(key, table)\
//     ((table->hash == stmm_ptrhash) ? STMM_PTRHASH((key),(table)->num_bins) :\
//      (table->hash == stmm_numhash) ? STMM_NUMHASH((key), (table)->num_bins) :\
//      (*table->hash)((key), (table)->num_bins))


// static int rehash (stmm_table *table);

//int stmm_numhash (), stmm_ptrhash (), stmm_numcmp (), stmm_ptrcmp ();

func stmm_init_table_with_params(compare stmm_compare_func_type, hash stmm_hash_func_type, size int32, density int32, grow_factor float64, reorder_flag int32) *stmm_table {
    var i int32;
    var newTable *stmm_table;

    // inlining of macro ABC_ALLOC
    newTable = ((stmm_table *) malloc(sizeof(stmm_table) * (size_t)(1)));
    if newTable == nil {
    return nil;
    }
    newTable.compare = compare;
    newTable.hash = hash;
    newTable.num_entries = 0;
    newTable.max_density = density;
    newTable.grow_factor = grow_factor;
    newTable.reorder_flag = reorder_flag;
    if size <= 0 {
    size = 1;
    }
    newTable.num_bins = size;
    // inlining of macro ABC_ALLOC
    newTable.bins = ((stmm_table_entry * *) malloc(sizeof(stmm_table_entry *) * (size_t)(size)));
    if newTable.bins == nil {
    // inlining of macro ABC_FREE
    ((newTable) ? (free((char *) (newTable)), (newTable) = 0) : 0);
    return nil;
    }
    for i = 0; i < size; i++ {
    newTable.bins[i] = 0;
    }

    // added by alanmi
    newTable.pMemMan = Extra_MmFixedStart(sizeof (stmm_table_entry));
    return newTable;
}

func stmm_init_table(compare stmm_compare_func_type, hash stmm_hash_func_type) *stmm_table {
    return stmm_init_table_with_params (compare, hash,
                    STMM_DEFAULT_INIT_TABLE_SIZE,
                    STMM_DEFAULT_MAX_DENSITY,
                    STMM_DEFAULT_GROW_FACTOR,
                    STMM_DEFAULT_REORDER_FLAG);
}

func stmm_free_table(table *stmm_table) {
/*
    stmm_table_entry *ptr, *next;
    int i;
    for ( i = 0; i < table->num_bins; i++ )
    {
        ptr = table->bins[i];
        while ( ptr != NULL )
        {
            next = ptr->next;
            ABC_FREE( ptr );
            ptr = next;
        }
    }
*/
    // no need to deallocate entries because they are in the memory manager now
    // added by alanmi
    if table.pMemMan {
        Extra_MmFixedStop ((Extra_MmFixed_t *)table.pMemMan); }
    // inlining of macro ABC_FREE
    ((table.bins) ? (free((char *) (table.bins)), (table.bins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((table) ? (free((char *) (table)), (table) = 0) : 0);
}

// this function recycles all the bins
func stmm_clean(table *stmm_table) {
    var i int32;
    // clean the bins
    for i = 0; i < table.num_bins; i++ {
        table.bins[i] = nil;
    }
    // reset the parameters
    table.num_entries = 0;
    // restart the memory manager
    Extra_MmFixedRestart ((Extra_MmFixed_t *)table.pMemMan);
}


// Original macro:
// #define PTR_NOT_EQUAL(table, ptr, user_key)\
// (ptr != NULL && !EQUAL(table->compare, user_key, (ptr)->key))


// Original macro:
// #define FIND_ENTRY(table, hash_val, key, ptr, last) \
//     (last) = &(table)->bins[hash_val];\
//     (ptr) = *(last);\
//     while (PTR_NOT_EQUAL((table), (ptr), (key))) {\
//     (last) = &(ptr)->next; (ptr) = *(last);\
//     }\
//     if ((ptr) != NULL && (table)->reorder_flag) {\
//     *(last) = (ptr)->next;\
//     (ptr)->next = (table)->bins[hash_val];\
//     (table)->bins[hash_val] = (ptr);\
//     }


func stmm_lookup(table *stmm_table, key string, value *string) int32 {
    var hash_val int32;
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro STMM_NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((((table).compare) == stmm_numcmp) || (((table).compare) == stmm_ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    return 0;
    
    } else {
    if value != nil {
        *value = ptr.record;
    }
    return 1;
    }
}

func stmm_lookup_int(table *stmm_table, key string, value *int32) int32 {
    var hash_val int32;
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro STMM_NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((((table).compare) == stmm_numcmp) || (((table).compare) == stmm_ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    return 0;
    
    } else {
    if value != 0 {
        *value = int64( ptr.record);
    }
    return 1;
    }
}

// This macro contained a line
//    new = ABC_ALLOC(stmm_table_entry, 1);
// which was modified by alanmi


/* This macro does not check if memory allocation fails. Use at you own risk */
// Original macro:
// #define ADD_DIRECT(table, key, value, hash_val, new)\
// {\
//     if (table->num_entries/table->num_bins >= table->max_density) {\
//     rehash(table);\
//     hash_val = do_hash(key,table);\
//     }\
//     \
//     new = (stmm_table_entry *)Extra_MmFixedEntryFetch( (Extra_MmFixed_t *)table->pMemMan );\
//     \
//     new->key = key;\
//     new->record = value;\
//     new->next = table->bins[hash_val];\
//     table->bins[hash_val] = new;\
//     table->num_entries++;\
// }


func stmm_insert(table *stmm_table, key string, value string) int32 {
    var hash_val int32;
    var newEntry *stmm_table_entry;
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro STMM_NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((((table).compare) == stmm_numcmp) || (((table).compare) == stmm_ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    if table.num_entries / table.num_bins >= table.max_density {
        if rehash (table) == STMM_OUT_OF_MEM {
        return STMM_OUT_OF_MEM;
        }
        // inlining of macro do_hash
        
        if table.hash == stmm_ptrhash {
            // inlining of macro STMM_PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == stmm_numhash {
                // inlining of macro STMM_NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                hash_val = (*table.hash)((key), (table).num_bins);
            }
        }
    }

//              newEntry = ABC_ALLOC( stmm_table_entry, 1 );
    newEntry = (stmm_table_entry *) Extra_MmFixedEntryFetch ((Extra_MmFixed_t *)table.pMemMan);
    if newEntry == nil {
        return STMM_OUT_OF_MEM;
    }

    newEntry.key = key;
    newEntry.record = value;
    newEntry.next = table.bins[hash_val];
    table.bins[hash_val] = newEntry;
    table.num_entries++;
    return 0;
    
    } else {
    ptr.record = value;
    return 1;
    }
}

func stmm_add_direct(table *stmm_table, key string, value string) int32 {
    var hash_val int32;
    var newEntry *stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }
    if table.num_entries / table.num_bins >= table.max_density {
    if rehash (table) == STMM_OUT_OF_MEM {
        return STMM_OUT_OF_MEM;
    }
    }
    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

//      newEntry = ABC_ALLOC( stmm_table_entry, 1 );
    newEntry = (stmm_table_entry *) Extra_MmFixedEntryFetch ((Extra_MmFixed_t *)table.pMemMan);
    if newEntry == nil {
    return STMM_OUT_OF_MEM;
    }

    newEntry.key = key;
    newEntry.record = value;
    newEntry.next = table.bins[hash_val];
    table.bins[hash_val] = newEntry;
    table.num_entries++;
    return 1;
}

func stmm_find_or_add(table *stmm_table, key string, slot **string) int32 {
    var hash_val int32;
    var newEntry *stmm_table_entry;
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro STMM_NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((((table).compare) == stmm_numcmp) || (((table).compare) == stmm_ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    if table.num_entries / table.num_bins >= table.max_density {
        if rehash (table) == STMM_OUT_OF_MEM {
        return STMM_OUT_OF_MEM;
        }
        // inlining of macro do_hash
        
        if table.hash == stmm_ptrhash {
            // inlining of macro STMM_PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == stmm_numhash {
                // inlining of macro STMM_NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                hash_val = (*table.hash)((key), (table).num_bins);
            }
        }
    }

    // newEntry = ABC_ALLOC( stmm_table_entry, 1 );
    newEntry = (stmm_table_entry *) Extra_MmFixedEntryFetch ((Extra_MmFixed_t *)table.pMemMan);
    if newEntry == nil {
        return STMM_OUT_OF_MEM;
    }

    newEntry.key = key;
    newEntry.record = (char *) 0;
    newEntry.next = table.bins[hash_val];
    table.bins[hash_val] = newEntry;
    table.num_entries++;
    if slot != nil {
         *slot = &newEntry.record; }
    return 0;
    
    } else {
    if slot != nil {
         *slot = &ptr.record; }
    return 1;
    }
}

func stmm_find(table *stmm_table, key string, slot **string) int32 {
    var hash_val int32;
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro STMM_NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((((table).compare) == stmm_numcmp) || (((table).compare) == stmm_ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    return 0;
    
    } else {
    if slot != nil {
        *slot = &ptr.record;
    }
    return 1;
    }
}

func rehash(table *stmm_table) int32 {
    var ptr *stmm_table_entry;
    var next *stmm_table_entry;
    var old_bins **stmm_table_entry;
    var i, old_num_bins, hash_val, old_num_entries int32;

    /* save old values */
    old_bins = table.bins;
    old_num_bins = table.num_bins;
    old_num_entries = table.num_entries;

    /* rehash */
    table.num_bins = int32( (table.grow_factor * old_num_bins));
    if table.num_bins % 2 == 0 {
    table.num_bins += 1;
    }
    table.num_entries = 0;
    // inlining of macro ABC_ALLOC
    table.bins = ((stmm_table_entry * *) malloc(sizeof(stmm_table_entry *) * (size_t)(table.num_bins)));
    if table.bins == nil {
    table.bins = old_bins;
    table.num_bins = old_num_bins;
    table.num_entries = old_num_entries;
    return STMM_OUT_OF_MEM;
    }
    /* initialize */
    for i = 0; i < table.num_bins; i++ {
    table.bins[i] = 0;
    }

    /* copy data over */
    for i = 0; i < old_num_bins; i++ {
    ptr = old_bins[i];
    for ptr != nil {
        next = ptr.next;
        // inlining of macro do_hash
        
        if table.hash == stmm_ptrhash {
            // inlining of macro STMM_PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((ptr.key))>>2)%(table).num_bins)));
        } else {
            if table.hash == stmm_numhash {
                // inlining of macro STMM_NUMHASH
                hash_val = (Abc_AbsInt(int64((ptr.key)))%((table).num_bins));
            } else {
                hash_val = (*table.hash)((ptr.key), (table).num_bins);
            }
        }
        ptr.next = table.bins[hash_val];
        table.bins[hash_val] = ptr;
        table.num_entries++;
        ptr = next;
    }
    }
    // inlining of macro ABC_FREE
    ((old_bins) ? (free((char *) (old_bins)), (old_bins) = 0) : 0);

    return 1;
}

func stmm_copy(old_table *stmm_table) *stmm_table {
    var newEntry_table *stmm_table;
    stmm_table_entry *ptr, /* *newEntryptr, *next, */ *newEntry;
    int i, /*j, */ num_bins = old_table.num_bins;

    // inlining of macro ABC_ALLOC
    newEntry_table = ((stmm_table *) malloc(sizeof(stmm_table) * (size_t)(1)));
    if newEntry_table == nil {
    return nil;
    }

    *newEntry_table = *old_table;
    // inlining of macro ABC_ALLOC
    newEntry_table.bins = ((stmm_table_entry * *) malloc(sizeof(stmm_table_entry *) * (size_t)(num_bins)));
    if newEntry_table.bins == nil {
    // inlining of macro ABC_FREE
    ((newEntry_table) ? (free((char *) (newEntry_table)), (newEntry_table) = 0) : 0);
    return nil;
    }

    // allocate the memory manager for the newEntry table
    newEntry_table.pMemMan = Extra_MmFixedStart (sizeof (stmm_table_entry));

    for i = 0; i < num_bins; i++ {
    newEntry_table.bins[i] = nil;
    ptr = old_table.bins[i];
    for ptr != nil {
//                      newEntry = ABC_ALLOC( stmm_table_entry, 1 );
        newEntry = (stmm_table_entry *)Extra_MmFixedEntryFetch ((Extra_MmFixed_t *)newEntry_table.pMemMan);
        if newEntry == nil {
/*
                for ( j = 0; j <= i; j++ )
                {
                    newEntryptr = newEntry_table->bins[j];
                    while ( newEntryptr != NULL )
                    {
                        next = newEntryptr->next;
                        ABC_FREE( newEntryptr );
                        newEntryptr = next;
                    }
                }
*/
        Extra_MmFixedStop ((Extra_MmFixed_t *)newEntry_table.pMemMan);

        // inlining of macro ABC_FREE
        ((newEntry_table.bins) ? (free((char *) (newEntry_table.bins)), (newEntry_table.bins) = 0) : 0);
        // inlining of macro ABC_FREE
        ((newEntry_table) ? (free((char *) (newEntry_table)), (newEntry_table) = 0) : 0);
        return nil;
        }
        *newEntry = *ptr;
        newEntry.next = newEntry_table.bins[i];
        newEntry_table.bins[i] = newEntry;
        ptr = ptr.next;
    }
    }
    return newEntry_table;
}

func stmm_delete(table *stmm_table, keyp *string, value *string) int32 {
    var hash_val int32;
    var key string  = *keyp;
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro STMM_NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((((table).compare) == stmm_numcmp) || (((table).compare) == stmm_ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    return 0;
    }

    *last = ptr.next;
    if value != nil {
     *value = ptr.record; }
    *keyp = ptr.key;
//      ABC_FREE( ptr );
    Extra_MmFixedEntryRecycle ((Extra_MmFixed_t *)table.pMemMan, (char *) ptr);

    table.num_entries--;
    return 1;
}

func stmm_delete_int(table *stmm_table, keyp *long, value *string) int32 {
    var hash_val int32;
    var key string  = (char *) *keyp;
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;

    // inlining of macro do_hash
    
    if table.hash == stmm_ptrhash {
        // inlining of macro STMM_PTRHASH
        hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
    } else {
        if table.hash == stmm_numhash {
            // inlining of macro STMM_NUMHASH
            hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
        } else {
            hash_val = (*table.hash)((key), (table).num_bins);
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro STMM_NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((((table).compare) == stmm_numcmp) || (((table).compare) == stmm_ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    return 0;
    }

    *last = ptr.next;
    if value != nil {
     *value = ptr.record; }
    *keyp = int64( ptr.key);
//      ABC_FREE( ptr );
    Extra_MmFixedEntryRecycle ((Extra_MmFixed_t *)table.pMemMan, (char *) ptr);

    table.num_entries--;
    return 1;
}

func stmm_foreach(table *stmm_table, func func(string, string, string) stmm_retval, arg string) int32 {
    var ptr *stmm_table_entry;
    var last **stmm_table_entry;
         var retval stmm_retval;
    var i int32;

    for i = 0; i < table.num_bins; i++ {
    last = &table.bins[i];
    ptr = *last;
    for ptr != nil {
        retval = (*func) (ptr.key, ptr.record, arg);
        switch retval {
        case STMM_CONTINUE:
        last = &ptr.next;
        ptr = *last;
        break;
        case STMM_STOP:
        return 0;
        case STMM_DELETE:
        *last = ptr.next;
        table.num_entries--;   /* cstevens@ic */
//                              ABC_FREE( ptr );
        Extra_MmFixedEntryRecycle ((Extra_MmFixed_t *)table.pMemMan, (char *) ptr);

        ptr = *last;
        }
    }
    }
    return 1;
}

func stmm_strhash(string string, modulus int32) int32 {
    var val int32;
    var c int32;

    for (c = *string++) != '\0' {
    val = val * 997 + c;
    }

    return ((val < 0) ? -val : val) % modulus;
}

func stmm_numhash(x string, size int32) int32 {
    // inlining of macro STMM_NUMHASH
    return (Abc_AbsInt(int64(x))%(size));
}

func stmm_ptrhash(x string, size int32) int32 {
    // inlining of macro STMM_PTRHASH
    return (int32((((ABC_PTRUINT_T)(x)>>2)%size)));
}

func stmm_numcmp(x string, y string) int32 {
    // inlining of macro STMM_NUMCMP
    return ((x) != (y));
}

func stmm_ptrcmp(x string, y string) int32 {
    // inlining of macro STMM_NUMCMP
    return ((x) != (y));
}

func stmm_init_gen(table *stmm_table) *stmm_generator {
    var gen *stmm_generator;

    // inlining of macro ABC_ALLOC
    gen = ((stmm_generator *) malloc(sizeof(stmm_generator) * (size_t)(1)));
    if gen == nil {
    return nil;
    }
    gen.table = table;
    gen.entry = nil;
    gen.index = 0;
    return gen;
}


func stmm_gen(gen *stmm_generator, key_p *string, value_p *string) int32 {
    var i int32;

    if gen.entry == nil {
    /* try to find next entry */
    for i = gen.index; i < gen.table.num_bins; i++ {
        if gen.table.bins[i] != nil {
        gen.index = i + 1;
        gen.entry = gen.table.bins[i];
        break;
        }
    }
    if gen.entry == nil {
        return 0;       /* that's all folks ! */
    }
    }
    *key_p = gen.entry.key;
    if value_p != 0 {
    *value_p = gen.entry.record;
    }
    gen.entry = gen.entry.next;
    return 1;
}


func stmm_gen_int(gen *stmm_generator, key_p *string, value_p *long) int32 {
    var i int32;

    if gen.entry == nil {
    /* try to find next entry */
    for i = gen.index; i < gen.table.num_bins; i++ {
        if gen.table.bins[i] != nil {
        gen.index = i + 1;
        gen.entry = gen.table.bins[i];
        break;
        }
    }
    if gen.entry == nil {
        return 0;       /* that's all folks ! */
    }
    }
    *key_p = gen.entry.key;
    if value_p != 0 {
    *value_p = int64( gen.entry.record);
    }
    gen.entry = gen.entry.next;
    return 1;
}


func stmm_free_gen(gen *stmm_generator) {
    // inlining of macro ABC_FREE
    ((gen) ? (free((char *) (gen)), (gen) = 0) : 0);
}
//ABC_NAMESPACE_IMPL_END


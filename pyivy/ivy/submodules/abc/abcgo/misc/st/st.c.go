package st

import "strings"

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
//#include <stdlib.h>
//#include <string.h>

//#include "st.h"

//ABC_NAMESPACE_IMPL_START


// Original macro:
// #define st__NUMCMP(x,y) ((x) != (y))

// Original macro:
// #define st__NUMHASH(x,size) (Abc_AbsInt((long)x)%(size))

//#define st__PTRHASH(x,size) ((int)((ABC_PTRUINT_T)(x)>>2)%size)  // 64-bit bug fix 9/17/2007
// Original macro:
// #define st__PTRHASH(x,size) ((int)(((ABC_PTRUINT_T)(x)>>2)%size))

// Original macro:
// #define EQUAL(func, x, y) \
//     ((func) == NULL ?\
//       (strcmp((x), (y)) == 0) :\
//      (((func) == st__numcmp) || ((func) == st__ptrcmp)) ?\
//       (st__NUMCMP((x),(y)) == 0) :\
//      ((func) == st__strcmp) ?\
//       (strcmp((x), (y)) == 0) : ((*func)((x), (y)) == 0))



// Original macro:
// #define do_hash(key, table)\
//     ((table->hash == NULL) ? st__strhash((key), (table)->num_bins) :\
//      (table->hash == st__ptrhash) ? st__PTRHASH((key),(table)->num_bins) :\
//      (table->hash == st__numhash) ? st__NUMHASH((key), (table)->num_bins) :\
//      (table->hash == st__strhash) ? st__strhash((key), (table)->num_bins) :\
//      (*table->hash)((key), (table)->num_bins))


// static int rehash( st__table *table);


// int st__numhash(const char*, int);

// int st__ptrhash(const char*, int);

// int st__numcmp(const char*, const char*);

// int st__ptrcmp(const char*, const char*);


 func st__init_table_with_params(compare st__compare_func_type, hash st__hash_func_type, size int32, density int32, grow_factor float64, reorder_flag int32) *st__table {
    var i int32;
    var newTable *st__table;

    // inlining of macro ABC_ALLOC
    newTable = ((st__table *) malloc(sizeof(st__table) * (size_t)(1)));
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
    newTable.bins = ((st__table_entry * *) malloc(sizeof(st__table_entry *) * (size_t)(size)));
    if newTable.bins == nil {
    // inlining of macro ABC_FREE
    ((newTable) ? (free((char *) (newTable)), (newTable) = 0) : 0);
    return nil;
    }
    for i = 0; i < size; i++ {
    newTable.bins[i] = 0;
    }
    return newTable;
}

 func st__init_table(compare st__compare_func_type, hash st__hash_func_type) *st__table {
    return st__init_table_with_params(compare, hash, st__DEFAULT_INIT_TABLE_SIZE,
                     st__DEFAULT_MAX_DENSITY,
                     st__DEFAULT_GROW_FACTOR,
                     st__DEFAULT_REORDER_FLAG);
}
                
func st__free_table(table *st__table) {
    var ptr, next *st__table_entry;
    var i int32;

    for i = 0; i < table.num_bins ; i++ {
    ptr = table.bins[i];
    for ptr != nil {
        next = ptr.next;
        // inlining of macro ABC_FREE
        ((ptr) ? (free((char *) (ptr)), (ptr) = 0) : 0);
        ptr = next;
    }
    }
    // inlining of macro ABC_FREE
    ((table.bins) ? (free((char *) (table.bins)), (table.bins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((table) ? (free((char *) (table)), (table) = 0) : 0);
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


func st__lookup(table *st__table, key string, value *string) int32 {
    var hash_val int32;
    var ptr *st__table_entry;
    var last **st__table_entry;

    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro st__NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((table).compare) == nil ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) :     ((((table).compare) == st__numcmp) || (((table).compare) == st__ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) :     (((table).compare) == st__strcmp) ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
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

func st__lookup_int(table *st__table, key string, value *int32) int32 {
    var hash_val int32;
    var ptr *st__table_entry;
    var last **st__table_entry;

    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro st__NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((table).compare) == nil ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) :     ((((table).compare) == st__numcmp) || (((table).compare) == st__ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) :     (((table).compare) == st__strcmp) ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
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

/* This macro does not check if memory allocation fails. Use at you own risk */
// Original macro:
// #define ADD_DIRECT(table, key, value, hash_val, new)\
// {\
//     if (table->num_entries/table->num_bins >= table->max_density) {\
//     rehash(table);\
//     hash_val = do_hash(key,table);\
//     }\
//     \
//     new = ABC_ALLOC( st__table_entry, 1);\
//     \
//     new->key = key;\
//     new->record = value;\
//     new->next = table->bins[hash_val];\
//     table->bins[hash_val] = new;\
//     table->num_entries++;\
// }


func st__insert(table *st__table, key string, value string) int32 {
    var hash_val int32;
    var newEntry *st__table_entry;
    var ptr *st__table_entry;
    var last **st__table_entry;

    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro st__NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((table).compare) == nil ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) :     ((((table).compare) == st__numcmp) || (((table).compare) == st__ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) :     (((table).compare) == st__strcmp) ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    if table.num_entries/table.num_bins >= table.max_density {
        if rehash(table) == st__OUT_OF_MEM {
        return st__OUT_OF_MEM;
        }
        // inlining of macro do_hash
        
        if table.hash == nil {
            hash_val = st__strhash((key), (table).num_bins);
        } else {
            if table.hash == st__ptrhash {
                // inlining of macro st__PTRHASH
                hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
            } else {
                if table.hash == st__numhash {
                    // inlining of macro st__NUMHASH
                    hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
                } else {
                    if table.hash == st__strhash {
                        hash_val = st__strhash((key), (table).num_bins);
                    } else {
                        hash_val = (*table.hash)((key), (table).num_bins);
                    }
                }
            }
        }
    }
    // inlining of macro ABC_ALLOC
    newEntry = ((st__table_entry *) malloc(sizeof(st__table_entry) * (size_t)(1)));
    if newEntry == nil {
        return st__OUT_OF_MEM;
    }
    newEntry.key = (char *)key;
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

func st__add_direct(table *st__table, key string, value string) int32 {
    var hash_val int32;
    var newEntry *st__table_entry;
    
    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }
    if table.num_entries / table.num_bins >= table.max_density {
    if rehash(table) == st__OUT_OF_MEM {
        return st__OUT_OF_MEM;
    }
    }
    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }
    // inlining of macro ABC_ALLOC
    newEntry = ((st__table_entry *) malloc(sizeof(st__table_entry) * (size_t)(1)));
    if newEntry == nil {
    return st__OUT_OF_MEM;
    }
    newEntry.key = key;
    newEntry.record = value;
    newEntry.next = table.bins[hash_val];
    table.bins[hash_val] = newEntry;
    table.num_entries++;
    return 1;
}

func st__find_or_add(table *st__table, key string, slot **string) int32 {
    var hash_val int32;
    var newEntry *st__table_entry;
    var ptr *st__table_entry;
    var last **st__table_entry;

    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro st__NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((table).compare) == nil ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) :     ((((table).compare) == st__numcmp) || (((table).compare) == st__ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) :     (((table).compare) == st__strcmp) ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
    if table.num_entries / table.num_bins >= table.max_density {
        if rehash(table) == st__OUT_OF_MEM {
        return st__OUT_OF_MEM;
        }
        // inlining of macro do_hash
        
        if table.hash == nil {
            hash_val = st__strhash((key), (table).num_bins);
        } else {
            if table.hash == st__ptrhash {
                // inlining of macro st__PTRHASH
                hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
            } else {
                if table.hash == st__numhash {
                    // inlining of macro st__NUMHASH
                    hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
                } else {
                    if table.hash == st__strhash {
                        hash_val = st__strhash((key), (table).num_bins);
                    } else {
                        hash_val = (*table.hash)((key), (table).num_bins);
                    }
                }
            }
        }
    }
    // inlining of macro ABC_ALLOC
    newEntry = ((st__table_entry *) malloc(sizeof(st__table_entry) * (size_t)(1)));
    if newEntry == nil {
        return st__OUT_OF_MEM;
    }
    newEntry.key = key;
    newEntry.record = (char *) 0;
    newEntry.next = table.bins[hash_val];
    table.bins[hash_val] = newEntry;
    table.num_entries++;
    if slot != nil { *slot = &newEntry.record; }
    return 0;
    } else {
    if slot != nil { *slot = &ptr.record; }
    return 1;
    }
}

func st__find(table *st__table, key string, slot **string) int32 {
    var hash_val int32;
    var ptr *st__table_entry;
    var last **st__table_entry;

    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro st__NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((table).compare) == nil ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) :     ((((table).compare) == st__numcmp) || (((table).compare) == st__ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) :     (((table).compare) == st__strcmp) ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
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

func rehash(table *st__table) int32 {
    var ptr *st__table_entry;
    var next *st__table_entry;
    var old_bins **st__table_entry;
    var i, old_num_bins, hash_val, old_num_entries int32;

    /* save old values */
    old_bins = table.bins;
    old_num_bins = table.num_bins;
    old_num_entries = table.num_entries;

    /* rehash */
    table.num_bins = int32((table.grow_factor * old_num_bins));
    if table.num_bins % 2 == 0 {
    table.num_bins += 1;
    }
    table.num_entries = 0;
    // inlining of macro ABC_ALLOC
    table.bins = ((st__table_entry * *) malloc(sizeof(st__table_entry *) * (size_t)(table.num_bins)));
    if table.bins == nil {
    table.bins = old_bins;
    table.num_bins = old_num_bins;
    table.num_entries = old_num_entries;
    return st__OUT_OF_MEM;
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
        
        if table.hash == nil {
            hash_val = st__strhash((ptr.key), (table).num_bins);
        } else {
            if table.hash == st__ptrhash {
                // inlining of macro st__PTRHASH
                hash_val = (int32((((ABC_PTRUINT_T)((ptr.key))>>2)%(table).num_bins)));
            } else {
                if table.hash == st__numhash {
                    // inlining of macro st__NUMHASH
                    hash_val = (Abc_AbsInt(int64((ptr.key)))%((table).num_bins));
                } else {
                    if table.hash == st__strhash {
                        hash_val = st__strhash((ptr.key), (table).num_bins);
                    } else {
                        hash_val = (*table.hash)((ptr.key), (table).num_bins);
                    }
                }
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

 func st__copy(old_table *st__table) *st__table {
    var newEntry_table *st__table;
    var ptr, newEntryptr, next, newEntry *st__table_entry;
    var i int32;
    var j int32;
    var num_bins int32 = old_table.num_bins;

    // inlining of macro ABC_ALLOC
    newEntry_table = ((st__table *) malloc(sizeof(st__table) * (size_t)(1)));
    if newEntry_table == nil {
    return nil;
    }
    
    *newEntry_table = *old_table;
    // inlining of macro ABC_ALLOC
    newEntry_table.bins = ((st__table_entry * *) malloc(sizeof(st__table_entry *) * (size_t)(num_bins)));
    if newEntry_table.bins == nil {
    // inlining of macro ABC_FREE
    ((newEntry_table) ? (free((char *) (newEntry_table)), (newEntry_table) = 0) : 0);
    return nil;
    }
    for i = 0; i < num_bins ; i++ {
    newEntry_table.bins[i] = nil;
    ptr = old_table.bins[i];
    for ptr != nil {
        // inlining of macro ABC_ALLOC
        newEntry = ((st__table_entry *) malloc(sizeof(st__table_entry) * (size_t)(1)));
        if newEntry == nil {
        for j = 0; j <= i; j++ {
            newEntryptr = newEntry_table.bins[j];
            for newEntryptr != nil {
            next = newEntryptr.next;
            // inlining of macro ABC_FREE
            ((newEntryptr) ? (free((char *) (newEntryptr)), (newEntryptr) = 0) : 0);
            newEntryptr = next;
            }
        }
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

func st__delete(table *st__table, keyp *string, value *string) int32 {
    var hash_val int32;
    char *key = *keyp;
    var ptr *st__table_entry;
    var last **st__table_entry;

    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro st__NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((table).compare) == nil ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) :     ((((table).compare) == st__numcmp) || (((table).compare) == st__ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) :     (((table).compare) == st__strcmp) ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };
    
    if ptr == nil {
    return 0;
    }

    *last = ptr.next;
    if value != nil { *value = ptr.record; }
    *keyp = ptr.key;
    // inlining of macro ABC_FREE
    ((ptr) ? (free((char *) (ptr)), (ptr) = 0) : 0);
    table.num_entries--;
    return 1;
}

func st__delete_int(table *st__table, keyp *long, value *string) int32 {
    var hash_val int32;
    var key string  = (char *) *keyp;
    var ptr *st__table_entry;
    var last **st__table_entry;

    // inlining of macro do_hash
    
    if table.hash == nil {
        hash_val = st__strhash((key), (table).num_bins);
    } else {
        if table.hash == st__ptrhash {
            // inlining of macro st__PTRHASH
            hash_val = (int32((((ABC_PTRUINT_T)((key))>>2)%(table).num_bins)));
        } else {
            if table.hash == st__numhash {
                // inlining of macro st__NUMHASH
                hash_val = (Abc_AbsInt(int64((key)))%((table).num_bins));
            } else {
                if table.hash == st__strhash {
                    hash_val = st__strhash((key), (table).num_bins);
                } else {
                    hash_val = (*table.hash)((key), (table).num_bins);
                }
            }
        }
    }

    // inlining of macro FIND_ENTRY
    // inlining of macro PTR_NOT_EQUAL
    // inlining of macro EQUAL
    // inlining of macro st__NUMCMP
    (last) = &(table).bins[hash_val];    (ptr) = *(last);    for (ptr) != nil && !(((table).compare) == nil ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) :     ((((table).compare) == st__numcmp) || (((table).compare) == st__ptrcmp)) ?      (((((key))) != ((((ptr)).key))) == 0) :     (((table).compare) == st__strcmp) ?      (int32(strings.Compare(((key)), (((ptr)).key))) == 0) : ((*(table).compare)(((key)), (((ptr)).key)) == 0)) {    (last) = &(ptr).next; (ptr) = *(last);    }
    if (ptr) != nil && (table).reorder_flag {    *(last) = (ptr).next;    (ptr).next = (table).bins[hash_val];    (table).bins[hash_val] = (ptr);    };

    if ptr == nil {
        return 0;
    }

    *last = ptr.next;
    if value != nil { *value = ptr.record; }
    *keyp = int64( ptr.key);
    // inlining of macro ABC_FREE
    ((ptr) ? (free((char *) (ptr)), (ptr) = 0) : 0);
    table.num_entries--;
    return 1;
}

func st__foreach(table *st__table, func func(string, string, string) st__retval, arg string) int32 {
    var ptr *st__table_entry;
    var last **st__table_entry;
         var retval st__retval;
    var i int32;

    for i = 0; i < table.num_bins; i++ {
    last = &table.bins[i]; ptr = *last;
    for ptr != nil {
        retval = (*func)(ptr.key, ptr.record, arg);
        switch retval {
        case st__CONTINUE:
        last = &ptr.next; ptr = *last;
        break;
        case st__STOP:
        return 0;
        case st__DELETE:
        *last = ptr.next;
        table.num_entries--;   /* cstevens@ic */
        // inlining of macro ABC_FREE
        ((ptr) ? (free((char *) (ptr)), (ptr) = 0) : 0);
        ptr = *last;
        }
    }
    }
    return 1;
}

func st__strcmp(x string, y string) int32 {
    return int32(strings.Compare(x, y));
}

func st__strhash(string string, modulus int32) int32 {
    unsigned char * ustring = (unsigned char *)string;
    var c, val uint32;
    assert(modulus > 0, "modulus > 0");    
    for (c = *ustring++) != '\0' {
        val = val*997 + c;
    }
    return int32((val%modulus));
}

func st__numhash(x string, size int32) int32 {
    // inlining of macro st__NUMHASH
    return (Abc_AbsInt(int64(x))%(size));
}

func st__ptrhash(x string, size int32) int32 {
    // inlining of macro st__PTRHASH
    return (int32((((ABC_PTRUINT_T)(x)>>2)%size)));
}

func st__numcmp(x string, y string) int32 {
    // inlining of macro st__NUMCMP
    return ((x) != (y));
}

func st__ptrcmp(x string, y string) int32 {
    // inlining of macro st__NUMCMP
    return ((x) != (y));
}

 func st__init_gen(table *st__table) *st__generator {
    var gen *st__generator;

    // inlining of macro ABC_ALLOC
    gen = ((st__generator *) malloc(sizeof(st__generator) * (size_t)(1)));
    if gen == nil {
    return nil;
    }
    gen.table = table;
    gen.entry = nil;
    gen.index = 0;
    return gen;
}


func st__gen(gen *st__generator, key_p *string, value_p *string) int32 {
    var i int32;

    if gen.entry == nil {
    /* try to find next entry */
    for i = gen.index; i < gen.table.num_bins; i++ {
        if gen.table.bins[i] != nil {
        gen.index = i+1;
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


func st__gen_int(gen *st__generator, key_p *string, value_p *long) int32 {
    var i int32;

    if gen.entry == nil {
    /* try to find next entry */
    for i = gen.index; i < gen.table.num_bins; i++ {
        if gen.table.bins[i] != nil {
        gen.index = i+1;
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


func st__free_gen(gen *st__generator) {
    // inlining of macro ABC_FREE
    ((gen) ? (free((char *) (gen)), (gen) = 0) : 0);
}

//ABC_NAMESPACE_IMPL_END

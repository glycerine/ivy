package espresso

import "fmt"

import "os"

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
 *   set.c -- routines for maniuplating sets and set families
 */

/* LINTLIBRARY */

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START

var set_family_garbage pset_family = nil;

static void intcpy(d, s, n)
register unsigned int *d, *s;
register long n;
{
    register int i;
    for i = 0; i < n; i++ {
    *d++ = *s++;
    }
}


/* bit_index -- find first bit (from LSB) in a word (MSB=bit n, LSB=bit 0) */
int bit_index(a)
register unsigned int a;
{
    register int i;
    if (a == 0)
    return -1;
    for(i = 0; (a & 1) == 0; a >>= 1, i++)
    ;
    return i;
}


/* set_ord -- count number of elements in a set */
int set_ord(a)
register pset a;
{
    register int i, sum = 0;
    register unsigned int val;
    // inlining of macro LOOP
    for(i = (a[0] & 0x03ff); i > 0; i--)
    if ((val = a[i]) != 0)
        sum += count_ones(val);
    return sum;
}

/* set_dist -- distance between two sets (# elements in common) */
int set_dist(a, b)
register pset a, b;
{
    register int i, sum = 0;
    register unsigned int val;
    // inlining of macro LOOP
    for(i = (a[0] & 0x03ff); i > 0; i--)
    if ((val = a[i] & b[i]) != 0)
        sum += count_ones(val);
    return sum;
}

/* set_clear -- make "r" the empty set of "size" elements */
pset set_clear(r, size)
register pset r;
var size int32;
{
    // inlining of macro LOOPINIT
    // inlining of macro WHICH_WORD
    register int i = ((size <= BPI) ? 1 : ((((size)-1) >> LOGBPI) + 1));
    *r = i; do r[i] = 0; while (--i > 0);
    return r;
}

/* set_fill -- make "r" the universal set of "size" elements */
pset set_fill(r, size)
register pset r;
register int size;
{
    // inlining of macro LOOPINIT
    // inlining of macro WHICH_WORD
    register int i = ((size <= BPI) ? 1 : ((((size)-1) >> LOGBPI) + 1));
    *r = i;
    r[i] = ~ uint32( 0);
    r[i] >>= i * BPI - size;
    while (--i > 0)
    r[i] = ~ uint32( 0);
    return r;
}

/* set_copy -- copy set a into set r */
pset set_copy(r, a)
register pset r, a;
{
    register int i = LOOPCOPY(a);
    do r[i] = a[i]; while (--i >= 0);
    return r;
}

/* set_and -- compute intersection of sets "a" and "b" */
pset set_and(r, a, b)
register pset r, a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    // inlining of macro PUTLOOP
    r[0] &= ~0x03ff;
r[0] |= (i);; do r[i] = a[i] & b[i]; while (--i > 0);
    return r;
}

/* set_or -- compute union of sets "a" and "b" */
pset set_or(r, a, b)
register pset r, a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    // inlining of macro PUTLOOP
    r[0] &= ~0x03ff;
r[0] |= (i);; do r[i] = a[i] | b[i]; while (--i > 0);
    return r;
}

/* set_diff -- compute difference of sets "a" and "b" */
pset set_diff(r, a, b)
register pset r, a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    // inlining of macro PUTLOOP
    r[0] &= ~0x03ff;
r[0] |= (i);; do r[i] = a[i] & ~b[i]; while (--i > 0);
    return r;
}

/* set_xor -- compute exclusive-or of sets "a" and "b" */
pset set_xor(r, a, b)
register pset r, a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
#ifdef IBM_WATC
    PUTLOOP(r,i); do r[i] = (a[i]&~b[i]) | (~a[i]&b[i]); while (--i > 0);
#else
    PUTLOOP(r,i); do r[i] = a[i] ^ b[i]; while (--i > 0);
#endif
    return r;
}

/* set_merge -- compute "a" & "mask" | "b" & ~ "mask" */
pset set_merge(r, a, b, mask)
register pset r, a, b, mask;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    // inlining of macro PUTLOOP
    r[0] &= ~0x03ff;
r[0] |= (i);; do r[i] = (a[i]&mask[i]) | (b[i]&~mask[i]); while (--i > 0);
    return r;
}

/* set_andp -- compute intersection of sets "a" and "b" , TRUE if nonempty */
bool set_andp(r, a, b)
register pset r, a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    register unsigned int x = 0;
    // inlining of macro PUTLOOP
    r[0] &= ~0x03ff;
r[0] |= (i);; for {r[i] = a[i] & b[i]; x |= r[i];
if !(--i > 0) { break }
} 
    return x != 0;
}

/* set_orp -- compute union of sets "a" and "b" , TRUE if nonempty */
bool set_orp(r, a, b)
register pset r, a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    register unsigned int x = 0;
    // inlining of macro PUTLOOP
    r[0] &= ~0x03ff;
r[0] |= (i);; for {r[i] = a[i] | b[i]; x |= r[i];
if !(--i > 0) { break }
} 
    return x != 0;
}

/* setp_empty -- check if the set "a" is empty */
bool setp_empty(a)
register pset a;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    do if (a[i]) return FALSE; while (--i > 0);
    return TRUE;
}

/* setp_full -- check if the set "a" is the full set of "size" elements */
bool setp_full(a, size)
register pset a;
register int size;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    register unsigned int test;
    test = ~ uint32( 0);
    test >>= i * BPI - size;
    if (a[i] != test)
    return FALSE;
    while (--i > 0)
    if (a[i] != (~uint32( 0)))
        return FALSE;
    return TRUE;
}

/* setp_equal -- check if the set "a" equals set "b" */
bool setp_equal(a, b)
register pset a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    do if (a[i] != b[i]) return FALSE; while (--i > 0);
    return TRUE;
}

/* setp_disjoint -- check if intersection of "a" and "b" is empty */
bool setp_disjoint(a, b)
register pset a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    do if (a[i] & b[i]) return FALSE; while (--i > 0);
    return TRUE;
}

/* setp_implies -- check if "a" implies "b" ("b" contains "a") */
bool setp_implies(a, b)
register pset a, b;
{
    // inlining of macro LOOP
    register int i = (a[0] & 0x03ff);
    do if (a[i] & ~b[i]) return FALSE; while (--i > 0);
    return TRUE;
}

/* sf_or -- form the "or" of all sets in a set family */
pset sf_or(A)
var A pset_family;
{
    register pset or, last, p;

    // inlining of macro set_new
    or = set_clear(ALLOC(unsigned int, SET_SIZE(A.sf_size)), A.sf_size);
    // inlining of macro foreach_set
    for(p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize)
    INLINEset_or(or, or, p);
    return or;
}

/* sf_and -- form the "and" of all sets in a set family */
pset sf_and(A)
var A pset_family;
{
    register pset and, last, p;

    // inlining of macro set_new
    and = set_fill(set_clear(ALLOC(unsigned int, SET_SIZE(A.sf_size)), A.sf_size), A.sf_size);
    // inlining of macro foreach_set
    for(p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize)
    INLINEset_and(and, and, p);
    return and;
}

/* sf_active -- make all members of the set family active */
pset_family sf_active(A)
var A pset_family;
{
    register pset p, last;
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    // inlining of macro SET
    (p[0] |= (ACTIVE));
    }
    A.active_count = A.count;
    return A;
}


/* sf_inactive -- remove all inactive cubes in a set family */
pset_family sf_inactive(A)
var A pset_family;
{
    register pset p, last, pdest;

    pdest = A.data;
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    // inlining of macro TESTP
    if p[0] & (ACTIVE) {
        if pdest != p {
        INLINEset_copy(pdest, p);
        }
        pdest += A.wsize;
    } else {
        A.count--;
    }
    }
    return A;
}


/* sf_copy -- copy a set family */
pset_family sf_copy(R, A)
var R, A pset_family;
{
    R.sf_size = A.sf_size;
    R.wsize = A.wsize;
/*R->capacity = A->count;*/
/*R->data = REALLOC(unsigned int, R->data, (long) R->capacity * R->wsize);*/
    R.count = A.count;
    R.active_count = A.active_count;
    intcpy(R.data, A.data, int64( A.wsize) * A.count);
    return R;
}


/* sf_join -- join A and B into a single set_family */
pset_family sf_join(A, B)
var A, B pset_family;
{
    pset_family R;
    long asize = A.count * A.wsize;
    long bsize = B.count * B.wsize;

    if (A.sf_size != B.sf_size) fatal("sf_join: sf_size mismatch");
    R = sf_new(A.count + B.count, A.sf_size);
    R.count = A.count + B.count;
    R.active_count = A.active_count + B.active_count;
    intcpy(R.data, A.data, asize);
    intcpy(R.data + asize, B.data, bsize);
    return R;
}


/* sf_append -- append the sets of B to the end of A, and dispose of B */
pset_family sf_append(A, B)
var A, B pset_family;
{
    long asize = A.count * A.wsize;
    long bsize = B.count * B.wsize;

    if (A.sf_size != B.sf_size) fatal("sf_append: sf_size mismatch");
    A.capacity = A.count + B.count;
    A.data = REALLOC(unsigned int, A.data, int64( A.capacity) * A.wsize);
    intcpy(A.data + asize, B.data, bsize);
    A.count += B.count;
    A.active_count += B.active_count;
    sf_free(B);
    return A;
}


/* sf_new -- allocate "num" sets of "size" elements each */
pset_family sf_new(num, size)
var num, size int32;
{
    pset_family A;
    if set_family_garbage == NULL {
    A = ALLOC(set_family_t, 1);
    } else {
    A = set_family_garbage;
    set_family_garbage = A.next;
    }
    A.sf_size = size;
    A.wsize = SET_SIZE(size);
    A.capacity = num;
    A.data = ALLOC(unsigned int, int64( A.capacity) * A.wsize);
    A.count = 0;
    A.active_count = 0;
    return A;
}


/* sf_save -- create a duplicate copy of a set family */
pset_family sf_save(A)
register pset_family A;
{
    return sf_copy(sf_new(A.count, A.sf_size), A);
}


/* sf_free -- free the storage allocated for a set family */
void sf_free(A)
var A pset_family;
{
    FREE(A.data);
    A.next = set_family_garbage;
    set_family_garbage = A;
}


/* sf_cleanup -- free all of the set families from the garbage list */
func sf_cleanup() {
    register pset_family p, pnext;
    for p = set_family_garbage; p != (pset_family) nil; p = pnext {
    pnext = p.next;
    FREE(p);
    }
    set_family_garbage = (pset_family) nil;
}


/* sf_addset -- add a set to the end of a set family */
pset_family sf_addset(A, s)
var A pset_family;
var s pset;
{
    register pset p;

    if A.count >= A.capacity {
    A.capacity = A.capacity + A.capacity/2 + 1;
    A.data = REALLOC(unsigned int, A.data, int64( A.capacity) * A.wsize);
    }
    // inlining of macro GETSET
    p = ((A).data + (A).wsize * (A.count++));
    INLINEset_copy(p, s);
    return A;
}

/* sf_delset -- delete a set from a set family */
void sf_delset(A, i)
var A pset_family;
var i int32;
// inlining of macro GETSET
{   (void) set_copy(((A).data + (A).wsize * (i)), ((A).data + (A).wsize * (--A.count)));}

/* sf_print -- print a set_family as a set (list the element numbers) */
void sf_print(A)
var A pset_family;
{
    char *ps1();
    register pset p;
    register int i;
    // inlining of macro foreachi_set
    for(p=A.data,i=0;i<A.count;p+=A.wsize,i++)
    fmt.Printf("A[%d] = %s\n", i, ps1(p));
}

/* sf_bm_print -- print a set_family as a bit-matrix */
void sf_bm_print(A)
var A pset_family;
{
    char *pbv1();
    register pset p;
    register int i;
    // inlining of macro foreachi_set
    for(p=A.data,i=0;i<A.count;p+=A.wsize,i++)
    fmt.Printf("[%4d] %s\n", i, pbv1(p, A.sf_size));
}


/* sf_write -- output a set family in an unintelligable manner */
void sf_write(fp, A)
var fp *os.File;
var A pset_family;
{
    register pset p, last;
    (void) fmt.Fprintf(fp, "%d %d\n", A.count, A.sf_size);
    // inlining of macro foreach_set
    for(p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize)
    set_write(fp, p);
    (void) fflush(fp);
}


/* sf_read -- read a set family written by sf_write */
pset_family sf_read(fp)
var fp *os.File;
{
    int i, j;
    register pset p, last;
    pset_family A;

    (void) fscanf(fp, "%d %d\n", &i, &j);
    A = sf_new(i, j);
    A.count = i;
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    (void) fscanf(fp, "%x", p);
    // inlining of macro LOOP
    for(j = 1; j <= (p[0] & 0x03ff); j++)
        (void) fscanf(fp, "%x", p+j);
    }
    return A;
}


/* set_write -- output a set in an unintelligable manner */
void set_write(fp, a)
register FILE *fp;
register pset a;
{
    // inlining of macro LOOP
    register int n = (a[0] & 0x03ff), j;

    for j = 0; j <= n; j++ {
    (void) fmt.Fprintf(fp, "%x ", a[j]);
    if ((j+1) % 8 == 0 && j != n)
        (void) fmt.Fprintf(fp, "\n\t");
    }
    (void) fmt.Fprintf(fp, "\n");
}


/* sf_bm_read -- read a set family written by sf_bm_print (almost) */
pset_family sf_bm_read(fp)
var fp *os.File;
{
    int i, j, rows, cols;
    register pset pdest;
    pset_family A;

    (void) fscanf(fp, "%d %d\n", &rows, &cols);
    A = sf_new(rows, cols);
    for i = 0; i < rows; i++ {
    // inlining of macro GETSET
    pdest = ((A).data + (A).wsize * (A.count++));
    (void) set_clear(pdest, A.sf_size);
    for j = 0; j < cols; j++ {
        switch getc(fp) {
        case '0':
            break;
        case '1':
            // inlining of macro set_insert
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            (pdest[(((j) >> LOGBPI) + 1)] |= 1 << ((j) & (BPI-1)));
            break;
        default:
            fatal("Error reading set family");
        }
    }
    if getc(fp) != '\n' {
        fatal("Error reading set family (at end of line)");
    }
    }
    return A;
}



/* ps1 -- convert a set into a printable string */
const largest_string = 120
var s1 [largest_string]byte;
char *ps1(a)
register pset a;
{
    // inlining of macro NELEM
    // inlining of macro LOOP
    register int i, num, l, len = 0, n = (BPI * (a[0] & 0x03ff));
    char temp[20];
    bool first = TRUE;

    s1[len++] = '[';
    for(i = 0; i < n; i++)
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if a[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
        if (! first)
        s1[len++] = ',';
        first = FALSE; num = i;
        /* Generate digits (reverse order) */
        l = 0; do temp[l++] = num % 10 + '0'; while ((num /= 10) > 0);
        /* Copy them back in correct order */
        do s1[len++] = temp[--l]; while (l > 0);
        if len > largest_string-15 {
        s1[len++] = '.'; s1[len++] = '.'; s1[len++] = '.';
        break;
        }
    }

    s1[len++] = ']';
    s1[len++] = '\0';
    return s1;
}

/* pbv1 -- print bit-vector */
char *pbv1(s, n)
var s pset;
var n int32;
{
    register int i;
    for(i = 0; i < n; i++)
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    s1[i] = (s[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) ? '1' : '0';
    s1[n] = '\0';
    return s1;
}


/* set_adjcnt -- adjust the counts for a set by "weight" */
void
set_adjcnt(a, count, weight)
register pset a;
register int *count, weight;
{
    register int i, base;
    register unsigned int val;

    // inlining of macro LOOP
    for i = (a[0] & 0x03ff); i > 0;  {
    for val = a[i], base = --i << LOGBPI; val != 0; base++, val >>= 1 {
        if val & 1 {
        count[base] += weight;
        }
    }
    }
}



/* sf_count -- perform a column sum over a set family */
int *sf_count(A)
var A pset_family;
{
    register pset p, last;
    register int i, base, *count;
    register unsigned int val;

    count = ALLOC(int, A.sf_size);
    for i = A.sf_size - 1; i >= 0; i-- {
    count[i] = 0;
    }

    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    // inlining of macro LOOP
    for i = (p[0] & 0x03ff); i > 0;  {
        for val = p[i], base = --i << LOGBPI; val != 0; base++, val >>= 1 {
        if val & 1 {
            count[base]++;
        }
        }
    }
    }
    return count;
}


/* sf_count_restricted -- perform a column sum over a set family, restricting
 * to only the columns which are in r; also, the columns are weighted by the
 * number of elements which are in each row
 */
int *sf_count_restricted(A, r)
var A pset_family;
register pset r;
{
    register pset p;
    register int i, base, *count;
    register unsigned int val;
    int weight;
    pset last;

    count = ALLOC(int, A.sf_size);
    for i = A.sf_size - 1; i >= 0; i-- {
    count[i] = 0;
    }

    /* Loop for each set */
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    weight = 1024 / (set_ord(p) - 1);
    // inlining of macro LOOP
    for i = (p[0] & 0x03ff); i > 0;  {
        for val=p[i]&r[i], base= --i<<LOGBPI; val!=0; base++, val >>= 1 {
        if val & 1 {
            count[base] += weight;
        }
        }
    }
    }
    return count;
}


/*
 *  sf_delc -- delete columns first ... last of A
 */
pset_family sf_delc(A, first, last)
var A pset_family;
var first, last int32;
{
    return sf_delcol(A, first, last-first + 1);
}


/*
 *  sf_addcol -- add columns to a set family; includes a quick check to see
 *  if there is already enough room (and hence, can avoid copying)
 */
pset_family sf_addcol(A, firstcol, n)
var A pset_family;
var firstcol, n int32;
{
    int maxsize;

    /* Check if adding columns at the end ... */
    if firstcol == A.sf_size {
    /* If so, check if there is already enough room */
    // inlining of macro LOOPINIT
    // inlining of macro WHICH_WORD
    maxsize = BPI * ((A.sf_size <= BPI) ? 1 : ((((A.sf_size)-1) >> LOGBPI) + 1));
    if (A.sf_size + n) <= maxsize {
        A.sf_size += n;
        return A;
    }
    }
    return sf_delcol(A, firstcol, -n);
}

/*
 * sf_delcol -- add/delete columns to/from a set family
 *
 *  if n > 0 then n columns starting from firstcol are deleted if n < 0
 *  then n blank columns are inserted starting at firstcol
 *      (i.e., the first new column number is firstcol)
 *
 *  This is done by copying columns in the array which is a relatively
 *  slow operation.
 */
pset_family sf_delcol(A, firstcol, n)
var A pset_family;
register int firstcol, n;
{
    register pset p, last, pdest;
    register int i;
    pset_family B;

    B = sf_new(A.count, A.sf_size - n);
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    // inlining of macro GETSET
    pdest = ((B).data + (B).wsize * (B.count++));
    INLINEset_clear(pdest, B.sf_size);
    for(i = 0; i < firstcol; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if ((p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))))
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pdest[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
    for(i = n > 0 ? firstcol + n : firstcol; i < A.sf_size; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if ((p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))))
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pdest[(((i - n) >> LOGBPI) + 1)] |= 1 << ((i - n) & (BPI-1)));
    }
    sf_free(A);
    return B;
}


/*
 *  sf_copy_col -- copy column "srccol" from "src" to column "dstcol" of "dst"
 */
pset_family sf_copy_col(dst, dstcol, src, srccol)
var dst, src pset_family;
var dstcol, srccol int32;
{
    register pset last, p, pdest;
    register int word_test, word_set;
    unsigned int bit_set, bit_test;

    /* CHEAT! form these constants outside the loop */
    // inlining of macro WHICH_WORD
    word_test = (((srccol) >> LOGBPI) + 1);
    // inlining of macro WHICH_BIT
    bit_test = 1 << ((srccol) & (BPI-1));
    // inlining of macro WHICH_WORD
    word_set = (((dstcol) >> LOGBPI) + 1);
    // inlining of macro WHICH_BIT
    bit_set = 1 << ((dstcol) & (BPI-1));

    pdest = dst.data;
    // inlining of macro foreach_set
    for p=src.data,last=p+src.count*src.wsize;p<last;p+=src.wsize {
    if ((p[word_test] & bit_test) != 0)
        pdest[word_set] |= bit_set;
/*
 *  equivalent code for this is ...
 *    if (is_in_set(p, srccol)) set_insert(pdest, destcol);
 */
    pdest += dst.wsize;
    }
    return dst;
}



/*
 *  sf_compress -- delete columns from a matrix
 */
pset_family sf_compress(A, c)
var A pset_family;            /* will be freed */
register pset c;
{
    register pset p;
    register int i, bcol;
    pset_family B;

    /* create a clean set family for the result */
    B = sf_new(A.count, set_ord(c));
    for i = 0; i < A.count; i++ {
    // inlining of macro GETSET
    p = ((B).data + (B).wsize * (B.count++));
    INLINEset_clear(p, B.sf_size);
    }

    /* copy each column of A which has a 1 in c */
    bcol = 0;
    for i = 0; i < A.sf_size; i++ {
    // inlining of macro is_in_set
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    if c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
        (void) sf_copy_col(B, bcol++, A, i);
    }
    }
    sf_free(A);
    return B;
}



/*
 *  sf_transpose -- transpose a bit matrix
 *
 *  There are trickier ways of doing this, but this works.
 */
pset_family sf_transpose(A)
var A pset_family;
{
    pset_family B;
    register pset p;
    register int i, j;

    B = sf_new(A.sf_size, A.count);
    B.count = A.sf_size;
    // inlining of macro foreachi_set
    for p, i = B.data, 0;i<B.count;p+=B.wsize,i++ {
    INLINEset_clear(p, B.sf_size);
    }
    // inlining of macro foreachi_set
    for p, i = A.data, 0;i<A.count;p+=A.wsize,i++ {
    for j = 0; j < A.sf_size; j++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((j) >> LOGBPI) + 1)] & (1 << ((j) & (BPI-1))) {
        // inlining of macro set_insert
        // inlining of macro GETSET
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (((B).data + (B).wsize * (j))[(((i) >> LOGBPI) + 1)] |= 1 << ((i) & (BPI-1)));
        }
    }
    }
    sf_free(A);
    return B;
}


/*
 *   sf_permute -- permute the columns of a set_family
 *
 *   permute is an array of integers containing column numbers of A which
 *   are to be retained.
 */
pset_family sf_permute(A, permute, npermute)
var A pset_family;
register int *permute, npermute;
{
    pset_family B;
    register pset p, last, pdest;
    register int j;

    B = sf_new(A.count, npermute);
    B.count = A.count;
    // inlining of macro foreach_set
    for(p=B.data,last=p+B.count*B.wsize;p<last;p+=B.wsize)
    INLINEset_clear(p, npermute);

    pdest = B.data;
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    for(j = 0; j < npermute; j++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if ((p[(((permute[j]) >> LOGBPI) + 1)] & (1 << ((permute[j]) & (BPI-1)))))
        // inlining of macro set_insert
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        (pdest[(((j) >> LOGBPI) + 1)] |= 1 << ((j) & (BPI-1)));
    pdest += B.wsize;
    }
    sf_free(A);
    return B;
}
//ABC_NAMESPACE_IMPL_END


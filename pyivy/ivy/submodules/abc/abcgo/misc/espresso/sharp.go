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
/*
    sharp.c -- perform sharp, disjoint sharp, and intersection
*/

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


var start_time long;


/* cv_sharp -- form the sharp product between two covers */
pcover cv_sharp(A, B)
pcover A, B;
{
    pcube last, p;
    pcover T;

    // inlining of macro new_cover
    T = sf_new(0, cube.size);
    // inlining of macro foreach_set
    for(p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize)
    T = sf_union(T, cb_sharp(p, B));
    return T;
}


/* cb_sharp -- form the sharp product between a cube and a cover */
pcover cb_sharp(c, T)
pcube c;
pcover T;
{
    if T.count == 0 {
    // inlining of macro new_cover
    T = sf_addset(sf_new(1, cube.size), c);
    } else {
    start_time = ptime();
    T = cb_recur_sharp(c, T, 0, T.count-1, 0);
    }
    return T;
}


/* recursive formulation to provide balanced merging */
pcover cb_recur_sharp(c, T, first, last, level)
pcube c;
pcover T;
var first, last, level int32;
{
    pcover temp, left, right;
    int middle;

    if first == last {
    // inlining of macro GETSET
    temp = sharp(c, ((T).data + (T).wsize * (first)));
    } else {
    middle = (first + last) / 2;
    left = cb_recur_sharp(c, T, first, middle, level+1);
    right = cb_recur_sharp(c, T, middle+1, last, level+1);
    temp = cv_intersect(left, right);
    if (debug & SHARP) && level < 4 {
        fmt.Printf("# SHARP[%d]: %4d = %4d x %4d, time = %s\n", level, temp.count, left.count, right.count, print_time(ptime() - start_time));
        (void) fflush(os.Stdout);
    }
    // inlining of macro free_cover
    sf_free(left);
    // inlining of macro free_cover
    sf_free(right);
    }
    return temp;
}


/* sharp -- form the sharp product between two cubes */
pcover sharp(a, b)
pcube a, b;
{
    register int var;
    register pcube d=cube.temp[0], temp=cube.temp[1], temp1=cube.temp[2];
    // inlining of macro new_cover
    pcover r = sf_new(cube.num_vars, cube.size);

    if cdist0(a, b) {
    set_diff(d, a, b);
    for var = 0; var < cube.num_vars; var++ {
        if ! setp_empty(set_and(temp, d, cube.var_mask[var])) {
        set_diff(temp1, a, cube.var_mask[var]);
        // inlining of macro GETSET
        set_or(((r).data + (r).wsize * (r.count++)), temp, temp1);
        }
    }
    } else {
    r = sf_addset(r, a);
    }
    return r;
}

pcover make_disjoint(A)
pcover A;
{
    pcover R, new;
    register pset last, p;

    // inlining of macro new_cover
    R = sf_new(0, cube.size);
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    new = cb_dsharp(p, R);
    R = sf_append(R, new);
    }
    return R;
}


/* cv_dsharp -- disjoint-sharp product between two covers */
pcover cv_dsharp(A, B)
pcover A, B;
{
    register pcube last, p;
    pcover T;

    // inlining of macro new_cover
    T = sf_new(0, cube.size);
    // inlining of macro foreach_set
    for p=A.data,last=p+A.count*A.wsize;p<last;p+=A.wsize {
    T = sf_union(T, cb_dsharp(p, B));
    }
    return T;
}


/* cb1_dsharp -- disjoint-sharp product between a cover and a cube */
pcover cb1_dsharp(T, c)
pcover T;
pcube c;
{
    pcube last, p;
    pcover R;

    // inlining of macro new_cover
    R = sf_new(T.count, cube.size);
    // inlining of macro foreach_set
    for p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize {
    R = sf_union(R, dsharp(p, c));
    }
    return R;
}


/* cb_dsharp -- disjoint-sharp product between a cube and a cover */
pcover cb_dsharp(c, T)
pcube c;
pcover T;
{
    pcube last, p;
    pcover Y, Y1;

    if T.count == 0 {
    // inlining of macro new_cover
    Y = sf_addset(sf_new(1, cube.size), c);
    } else {
    // inlining of macro new_cover
    Y = sf_new(T.count, cube.size);
    // inlining of macro GETSET
    set_copy(((Y).data + (Y).wsize * (Y.count++)), c);
    // inlining of macro foreach_set
    for p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize {
        Y1 = cb1_dsharp(Y, p);
        // inlining of macro free_cover
        sf_free(Y);
        Y = Y1;
    }
    }
    return Y;
}


/* dsharp -- form the disjoint-sharp product between two cubes */
pcover dsharp(a, b)
pcube a, b;
{
    register pcube mask, diff, and, temp, temp1 = cube.temp[0];
    int var;
    pcover r;

    // inlining of macro new_cover
    r = sf_new(cube.num_vars, cube.size);

    if cdist0(a, b) {
    // inlining of macro new_cube
    // inlining of macro set_new
    diff = set_diff(set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size), a, b);
    // inlining of macro new_cube
    // inlining of macro set_new
    and = set_and(set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size), a, b);
    // inlining of macro new_cube
    // inlining of macro set_new
    mask = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    for var = 0; var < cube.num_vars; var++ {
        /* check if position var of "a and not b" is not empty */
        if ! setp_disjoint(diff, cube.var_mask[var]) {

        /* coordinate var equals the difference between a and b */
        // inlining of macro GETSET
        temp = ((r).data + (r).wsize * (r.count++));
        (void) set_and(temp, diff, cube.var_mask[var]);

        /* coordinates 0 ... var-1 equal the intersection */
        INLINEset_and(temp1, and, mask);
        INLINEset_or(temp, temp, temp1);

        /* coordinates var+1 .. cube.num_vars equal a */
        set_or(mask, mask, cube.var_mask[var]);
        INLINEset_diff(temp1, a, mask);
        INLINEset_or(temp, temp, temp1);
        }
    }
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(diff);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(and);
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(mask);
    } else {
    r = sf_addset(r, a);
    }
    return r;
}

/* cv_intersect -- form the intersection of two covers */

const MAGIC = 500               /* save 500 cubes before containment */

pcover cv_intersect(A, B)
pcover A, B;
{
    register pcube pi, pj, lasti, lastj, pt;
    pcover T, Tsave = NULL;

    /* How large should each temporary result cover be ? */
    // inlining of macro new_cover
    T = sf_new(MAGIC, cube.size);
    pt = T.data;

    /* Form pairwise intersection of each cube of A with each cube of B */
    // inlining of macro foreach_set
    for pi=A.data,lasti=pi+A.count*A.wsize;pi<lasti;pi+=A.wsize {
    // inlining of macro foreach_set
    for pj=B.data,lastj=pj+B.count*B.wsize;pj<lastj;pj+=B.wsize {
        if cdist0(pi, pj) {
        (void) set_and(pt, pi, pj);
        if ++T.count >= T.capacity {
            if (Tsave == NULL)
            Tsave = sf_contain(T);
            else
            Tsave = sf_union(Tsave, sf_contain(T));
            // inlining of macro new_cover
            T = sf_new(MAGIC, cube.size);
            pt = T.data;
        } else
            pt += T.wsize;
        }
    }
    }


    if (Tsave == NULL)
    Tsave = sf_contain(T);
    else
    Tsave = sf_union(Tsave, sf_contain(T));
    return Tsave;
}
//ABC_NAMESPACE_IMPL_END


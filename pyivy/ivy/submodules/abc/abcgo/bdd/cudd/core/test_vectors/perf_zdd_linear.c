/* Exercise CUDD's file-local adjacent transform and its coordinate mapping. */
#include "misc/util/abc_global.h"
#define cuddZddLinearSifting fixtureZddLinearSifting
#include "bdd/cudd/cuddZddLin.c"
#undef cuddZddLinearSifting
#include <time.h>

static int contains(DdManager *dd, DdNode *f, unsigned set)
{
    while (!Cudd_IsConstant(f)) {
        unsigned bit = 1u << f->index;
        if (set & bit) {
            set ^= bit;
            f = cuddT(f);
        } else {
            f = cuddE(f);
        }
    }
    return f == dd->one && set == 0;
}

int main(void)
{
    DdManager *dd = Cudd_Init(0,4,32,128,0);
    empty = dd->zero;
    DdNode *tail = cuddZddGetNode(dd,3,dd->one,dd->zero);
    Cudd_Ref(tail);
    DdNode *roots[16];
    for (int f = 0; f < 16; ++f) {
        DdNode *yt = cuddZddGetNode(dd,2,f & 8 ? tail : empty,f & 2 ? tail : empty);
        Cudd_Ref(yt);
        DdNode *ye = cuddZddGetNode(dd,2,f & 4 ? tail : empty,f & 1 ? tail : empty);
        Cudd_Ref(ye);
        roots[f] = cuddZddGetNode(dd,1,yt,ye);
        Cudd_Ref(roots[f]);
        Cudd_RecursiveDerefZdd(dd,yt);
        Cudd_RecursiveDerefZdd(dd,ye);
    }
    Cudd_RecursiveDerefZdd(dd,tail);
    cuddGarbageCollect(dd,1);
    for (int pass = 1; pass <= 2; ++pass) {
        if (!cuddZddLinearInPlace(dd,1,2)) return 1;
        for (int f = 0; f < 16; ++f)
            for (int set = 0; set < 16; ++set) {
                int bits = (set >> 1) & 3;
                if (pass == 1) bits ^= bits >> 1;
                if (contains(dd,roots[f],set) != ((set & 9) == 8 && (f & (1 << bits)) != 0)) return 2;
            }
    }
    const int runs = 1000000;
    struct timespec a,b;
    clock_gettime(CLOCK_MONOTONIC,&a);
    for (int n = 0; n < runs; ++n)
        if (!cuddZddLinearInPlace(dd,1,2) || !cuddZddLinearInPlace(dd,1,2)) return 3;
    clock_gettime(CLOCK_MONOTONIC,&b);
    printf("16 families, %.2f ns/transform pair\n",((b.tv_sec-a.tv_sec)*1e9+b.tv_nsec-a.tv_nsec)/runs);
    for (int f = 0; f < 16; ++f) Cudd_RecursiveDerefZdd(dd,roots[f]);
    cuddGarbageCollect(dd,1);
    if (dd->keysZ != 4) return 4;
    Cudd_Quit(dd);
    return 0;
}

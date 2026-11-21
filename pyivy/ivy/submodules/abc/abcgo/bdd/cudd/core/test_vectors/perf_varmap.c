#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char **argv)
{
    int repeat = argc == 2 ? atoi(argv[1]) : 0;
    if (repeat <= 0) return 1;
    DdManager *dd = Cudd_Init(16, 0, 64, 4096, 0);
    DdNode *src[256], *dst[256];
    for (int mask = 0; mask < 256; ++mask) {
        DdNode *x = dd->one, *y = dd->one;
        Cudd_Ref(x);
        Cudd_Ref(y);
        for (int i = 7; i >= 0; --i) {
            DdNode *xx = Cudd_bddAnd(dd, Cudd_NotCond(dd->vars[i+8],!(mask & (1<<i))),x);
            Cudd_Ref(xx);
            Cudd_RecursiveDeref(dd,x);
            x = xx;
            DdNode *yy = Cudd_bddAnd(dd, Cudd_NotCond(dd->vars[i],!(mask & (1<<i))),y);
            Cudd_Ref(yy);
            Cudd_RecursiveDeref(dd,y);
            y = yy;
        }
        src[mask] = x;
        dst[mask] = y;
    }
    if (!Cudd_SetVarMap(dd, dd->vars, dd->vars+8, 8)) return 2;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int j = 0; j < repeat; ++j)
        for (int i = 0; i < 256; ++i)
            if (Cudd_bddVarMap(dd,src[i]) != dst[i]) return 2;
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("256 state cubes mapped, %.1f ns/batch\n",ns/repeat);
    Cudd_Quit(dd);
    return 0;
}

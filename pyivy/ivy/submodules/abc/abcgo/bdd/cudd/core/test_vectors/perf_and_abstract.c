#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char **argv)
{
    int repeat = argc == 2 ? atoi(argv[1]) : 0;
    if (repeat <= 0) return 1;
    DdManager *dd = Cudd_Init(5, 0, 256, 4096, 0);
    DdNode *f = Cudd_bddXor(dd, dd->vars[0], dd->vars[3]);
    Cudd_Ref(f);
    Cudd_Ref(f);
    DdNode *g = Cudd_bddXor(dd, dd->vars[1], dd->vars[2]);
    Cudd_Ref(g);
    Cudd_Ref(g);
    DdNode *cube = dd->vars[2];
    if (Cudd_bddAndAbstract(dd, f, g, cube) != f) return 2;
    double hits = Cudd_ReadCacheHits(dd), lookups = Cudd_ReadCacheLookUps(dd);
    if (Cudd_bddAndAbstract(dd, g, f, cube) != f ||
        Cudd_ReadCacheHits(dd) != hits + 1 || Cudd_ReadCacheLookUps(dd) != lookups + 1)
        return 3;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < repeat; ++i)
        if (Cudd_bddAndAbstract(dd, g, f, cube) != f) return 4;
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("commuted image hit, %.1f ns/op\n", ns/repeat);
    Cudd_Quit(dd);
    return 0;
}

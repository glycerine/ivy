#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static int bdd_pressure(void)
{
    DdManager *dd = Cudd_Init(24, 0, 8, 256, 0);
    DdNode *temporary = cuddUniqueInter(dd, 14, dd->vars[18], dd->vars[21]);
    Cudd_Ref(temporary);
    for (int level = 0; level < 14; ++level) {
        for (int i = 0; i < 4; ++i) {
            for (int j = 0; j < 8; ++j) {
                DdNode *n = cuddUniqueInter(dd, level, dd->vars[14+i], Cudd_Not(dd->vars[14+j]));
                Cudd_Ref(n);
                Cudd_RecursiveDeref(dd, n);
            }
        }
    }
    if (dd->dead != 448 || dd->minDead != 800) return 0;
    DdNode *arg = cuddUniqueInter(dd, 13, dd->vars[14], Cudd_Not(dd->vars[15]));
    Cudd_Ref(arg);
    DdNode *root = cuddUniqueInter(dd, 0, arg, temporary);
    int ok = root && dd->garbageCollections == 1 && dd->subtables[0].slots == 8 &&
             dd->dead == 0 && dd->keys == 31;
    Cudd_Quit(dd);
    return ok;
}

static int other_families(void)
{
    DdManager *dd = Cudd_Init(0, 9, 8, 64, 0);
    for (int i = 1; i <= 4; ++i) {
        for (int j = 1; j <= 8; ++j) {
            DdNode *n = cuddUniqueInterZdd(dd, 0, dd->univ[i], dd->univ[j]);
            if (n == dd->univ[0])
                n = cuddUniqueInterZdd(dd, 0, dd->univ[5], dd->zero);
            Cudd_Ref(n);
            Cudd_RecursiveDerefZdd(dd, n);
        }
    }
    if (cuddUniqueInterZdd(dd, 0, dd->univ[1], dd->univ[1]) != dd->univ[0] ||
        dd->garbageCollections != 1 || dd->subtableZ[0].slots != 8 || dd->keysZ != 9)
        return 0;
    Cudd_Quit(dd);
    for (int disabled = 0; disabled < 2; ++disabled) {
        dd = Cudd_Init(0, 0, 8, 64, 0);
        if (disabled) Cudd_DisableGarbageCollection(dd);
        for (int i = 2; i <= 62; ++i) {
            DdNode *n = Cudd_addConst(dd, (double)i);
            Cudd_Ref(n);
            Cudd_RecursiveDeref(dd, n);
        }
        if (Cudd_addConst(dd, 0) != dd->zero ||
            dd->garbageCollections != (disabled ? 0 : 1) ||
            dd->constants.slots != (disabled ? 32 : 16) ||
            dd->constants.keys != (disabled ? 65 : 4))
            return 0;
        Cudd_Quit(dd);
    }
    return 1;
}

int main(int argc, char **argv)
{
    int repeat = argc == 2 ? atoi(argv[1]) : 0;
    if (repeat <= 0 || !other_families()) return 1;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < repeat; ++i)
        if (!bdd_pressure()) return 2;
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("448 released keys, 31 final BDD/ADD keys, %.1f ns/iteration\n", ns/repeat);
    return 0;
}

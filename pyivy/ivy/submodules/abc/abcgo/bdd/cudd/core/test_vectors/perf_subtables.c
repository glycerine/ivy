#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static int check_removal(void)
{
    DdManager *dd = Cudd_Init(4, 0, 8, 64, 0);
    DdNode *live = Cudd_addIthVar(dd, 2);
    Cudd_Ref(live);
    if (cuddDestroySubtables(dd, 2) || dd->size != 4)
        return 0;
    Cudd_RecursiveDeref(dd, live);
    unsigned before = Cudd_ReadGarbageCollections(dd);
    if (!cuddDestroySubtables(dd, 2) || dd->size != 2 ||
        Cudd_ReadGarbageCollections(dd) != before+1)
        return 0;
    DdNode *parent = Cudd_bddAnd(dd, dd->vars[0], dd->vars[1]);
    Cudd_Ref(parent);
    if (cuddDestroySubtables(dd, 1))
        return 0;
    Cudd_RecursiveDeref(dd, parent);
    if (!cuddDestroySubtables(dd, 1))
        return 0;
    Cudd_Quit(dd);
    return 1;
}

int main(int argc, char **argv)
{
    int repeat = argc == 2 ? atoi(argv[1]) : 0;
    if (repeat < 1 || !check_removal()) return 1;
    DdManager *dd = Cudd_Init(256, 0, 8, 4096, 0);
    unsigned slots = dd->slots;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < repeat; ++i) {
        if (!cuddInsertSubtables(dd, 128, 64) || dd->invperm[64] != 256 ||
            !cuddDestroySubtables(dd, 128) || dd->size != 256 ||
            dd->invperm[64] != 64 || dd->slots != slots)
            return 2;
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("%d variables, %u slots, %.1f ns/insert+destroy\n", dd->size, dd->slots, ns/repeat);
    Cudd_Quit(dd);
    return 0;
}

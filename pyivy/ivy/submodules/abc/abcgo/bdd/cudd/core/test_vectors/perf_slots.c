#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static void show(DdManager *dd, const char *state)
{
    unsigned slots = Cudd_ReadSlots(dd);
    printf("%s: slots %u, occupied %.0f, expected %.12f\n", state, slots,
           slots*Cudd_ReadUsedSlots(dd), Cudd_ExpectedUsedSlots(dd));
}

int main(int argc, char **argv)
{
    int iterations = argc == 2 ? atoi(argv[1]) : 0;
    if (iterations <= 0)
        return 1;
    DdManager *dd = Cudd_Init(9, 9, 8, 256, 0);
    show(dd, "initial");
    for (int i = 1; i < 9; ++i)
        for (int j = 1; j < 9; ++j) {
            if (i == j)
                continue;
            DdNode *node = cuddUniqueInter(dd, 0, dd->vars[i], dd->vars[j]);
            Cudd_Ref(node);
        }
    show(dd, "BDD");
    for (int i = 1; i < 9; ++i)
        for (int j = 1; j < 9; ++j) {
            if (i == j)
                continue;
            DdNode *node = cuddUniqueInterZdd(dd, 0, dd->univ[i], dd->univ[j]);
            Cudd_Ref(node);
        }
    show(dd, "ZDD");
    for (int i = 0; i < 80; ++i) {
        DdNode *node = Cudd_addConst(dd, i + 2.0);
        Cudd_Ref(node);
    }
    show(dd, "constants");
    volatile double value = 0;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < iterations; ++i)
        value += Cudd_ReadUsedSlots(dd) + Cudd_ExpectedUsedSlots(dd);
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9 * (end.tv_sec - start.tv_sec) + end.tv_nsec - start.tv_nsec;
    printf("%.2f ns/pair, sum %.3f\n", ns/iterations, value);
    Cudd_Quit(dd);
    return 0;
}

/* Print native numeric terminals and time the same 128-row cover as Go. */
#define _GNU_SOURCE
#include "misc/util/abc_global.h"
#include "bdd/cudd/cudd.h"
#include <math.h>
#include <time.h>

int main(void)
{
    DdManager *dd = Cudd_Init(1, 0, 256, 256, 0);
    double numbers[] = {12345.6789, -0.000123456789, 1e7, INFINITY, -INFINITY};
    for (unsigned i = 0; i < sizeof(numbers)/sizeof(numbers[0]); ++i)
        if (!Cudd_PrintMinterm(dd, Cudd_addConst(dd, numbers[i]))) return 1;
    Cudd_Quit(dd);

    dd = Cudd_Init(8, 0, 256, 256, 0);
    DdNode *f = Cudd_ReadLogicZero(dd);
    Cudd_Ref(f);
    for (int i = 0; i < 8; ++i) {
        DdNode *g = Cudd_bddXor(dd, f, Cudd_bddIthVar(dd, i));
        Cudd_Ref(g);
        Cudd_RecursiveDeref(dd, f);
        f = g;
    }
    char *data = NULL;
    size_t size = 0;
    FILE *out = open_memstream(&data, &size);
    if (!out) return 2;
    Cudd_SetStdout(dd, out);
    struct timespec start, end;
    const int count = 32000;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; ++i) {
        rewind(out);
        if (!Cudd_PrintMinterm(dd, f)) return 3;
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    fclose(out);
    if (size != 128 * 12) return 4;
    printf("%.2f us/parity cover\n", (1e6*(end.tv_sec-start.tv_sec) + 1e-3*(end.tv_nsec-start.tv_nsec))/count);
    free(data);
    Cudd_RecursiveDeref(dd, f);
    Cudd_Quit(dd);
    return 0;
}

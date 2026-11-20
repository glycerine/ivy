#include "sat/cnf/cnf.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

extern Cnf_Cut_t *Cnf_CutAlloc(Cnf_Man_t *, int);

/* An AND over all leaves. Store repeated truth bits as native cuts do. */
static Cnf_Cut_t *make_and(Cnf_Man_t *m, int n, const int *leaves)
{
    Cnf_Cut_t *cut = Cnf_CutAlloc(m, n);
    cut->Cost = n + 1;
    for (int i = 0; i < n; ++i)
        cut->pFanins[i] = leaves[i];
    unsigned word = 0;
    for (int row = 0; row < 32; ++row)
        if ((row & ((1 << n) - 1)) == (1 << n) - 1)
            word |= 1u << row;
    Cnf_CutTruth(cut)[0] = word;
    return cut;
}

int main(int argc, char **argv)
{
    int count = argc == 3 ? atoi(argv[1]) : 0;
    int n = argc == 3 ? atoi(argv[2]) : 0;
    if (count <= 0 || (n != 2 && n != 4))
        return 1;
    Cnf_Man_t *cuts = Cnf_ManStart(), *m = Cnf_ManStart();
    int low[] = {1, 2, 3, 4}, high[] = {5, 6, 7, 8};
    Cnf_Cut_t *top = make_and(cuts, n, high), *fan = make_and(cuts, n, low);
    struct timespec start, end;
    unsigned cost = 0, truth = 0;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; ++i) {
        if ((i & 1023) == 0)
            Aig_MmFlexRestart(m->pMemCuts);
        Cnf_Cut_t *result = Cnf_CutCompose(m, top, fan, high[0]);
        if (!result || result->nFanins != 2*n-1)
            return 2;
        cost = result->Cost;
        truth = Cnf_CutTruth(result)[result->nWords-1];
        Cnf_CutFree(result);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9 * (end.tv_sec - start.tv_sec) + end.tv_nsec - start.tv_nsec;
    printf("%d leaves, cost %u, final word %08x, %.2f ns/compose\n", 2*n-1, cost, truth, ns/count);
    Cnf_Cut_t *verify = Cnf_CutCompose(m, top, fan, high[0]);
    int mismatches = 0, rows = 1 << (2*n-1);
    for (int row = 0; row < rows; ++row)
        mismatches += ((Cnf_CutTruth(verify)[row >> 5] >> (row & 31)) & 1u) != (row == rows-1);
    printf("%d of %d rows differ from the expected AND\n", mismatches, rows);
    Cnf_CutFree(verify);
    Cnf_ManStop(m);
    Cnf_ManStop(cuts);
    return 0;
}

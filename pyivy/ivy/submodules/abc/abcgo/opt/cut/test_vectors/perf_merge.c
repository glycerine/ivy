/* Matches BenchmarkMergeOwnedFourLeafCut. */
#include "opt/cut/cutInt.h"
#include <stdint.h>
#include <stdio.h>
#include <time.h>

int main(void)
{
    const unsigned count = 50000000;
    Cut_Params_t params = {0};
    params.nVarsMax = 4;
    params.fTruth = 1;
    params.nKeepMax = 1000;
    Cut_Man_t *manager = Cut_ManStart(&params);
    Cut_Cut_t *a = Cut_CutAlloc(manager), *b = Cut_CutAlloc(manager);
    a->nLeaves = b->nLeaves = 2;
    a->pLeaves[0] = 0;
    a->pLeaves[1] = 3;
    b->pLeaves[0] = 1;
    b->pLeaves[1] = 2;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (unsigned i = 0; i < count; ++i) {
        Cut_Cut_t *c = Cut_CutMergeTwo(manager, a, b);
        if (c == NULL || c->nLeaves != 4 || c->pLeaves[3] != 3) return 1;
        Cut_CutRecycle(manager, c);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9 * (end.tv_sec - start.tv_sec) + end.tv_nsec - start.tv_nsec;
    printf("four-leaf cut, %.2f ns/merge and recycle\n", ns / count);
    Cut_ManStop(manager);
    return 0;
}

/* Copy and release sixteen temporary cuts as in BenchmarkCutSnapshotList. */
#include "opt/cut/cutInt.h"
#include <stdio.h>
#include <time.h>

int main(void)
{
    Cut_Params_t params = {0};
    params.nVarsMax = 4;
    params.nKeepMax = 1000;
    params.fTruth = 1;
    Cut_Man_t *man = Cut_ManStart(&params);
    Cut_Cut_t *source = NULL;
    for (int i = 15; i >= 0; i--) {
        Cut_Cut_t *cut = Cut_CutCreateTriv(man, i);
        cut->pNext = source;
        source = cut;
    }
    struct timespec start, stop;
    int repeat = 1000000;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < repeat; i++) {
        Cut_Cut_t *copy = Cut_CutDupList(man, source);
        if (copy->pNext->pLeaves[0] != 1 || man->nCutsAlloc != 16 ||
            man->nCutsPeak != 16 || man->nCutsCur != 16 || man->nCutsTriv != 16)
            return 1;
        Cut_CutRecycleList(man, copy);
        if (man->nCutsDealloc != 0) return 2;
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    printf("%.2f process ns/list, allocated %d peak %d recycled %d\n",
           ((stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec)/repeat,
           man->nCutsAlloc, man->nCutsPeak, man->nCutsDealloc);
    Cut_ManStop(man);
    return 0;
}

/* Timed warm-frame startup and teardown of ABC's standard rewrite library. */
#include "base/main/main.h"
#include "opt/rwr/rwr.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char **argv)
{
    int repeat = argc > 1 ? atoi(argv[1]) : 1000;
    if (repeat < 1) return 1;
    Abc_Start();
    Rwr_Man_t *a = Rwr_ManStart(0), *b = Rwr_ManStart(0);
    if (a->puCanons != b->puCanons || a->pPhases != b->pPhases ||
        a->pPerms != b->pPerms || a->pMap != b->pMap ||
        a->pPractical == b->pPractical || a->vForest == b->vForest)
        return 2;
    Rwr_ManStop(a);
    Rwr_ManStop(b);
    struct timespec start, end;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < repeat; i++) {
        Rwr_Man_t *p = Rwr_ManStart(0);
        if (Vec_PtrSize(p->vForest) != 1785 || Vec_VecSize(p->vClasses) != 222)
            return 3;
        Rwr_ManStop(p);
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &end);
    printf("%.2f process us/manager\n",
           ((end.tv_sec-start.tv_sec)*1e6+(end.tv_nsec-start.tv_nsec)*1e-3)/repeat);
    Abc_Stop();
    return 0;
}

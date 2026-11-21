/* Match BenchmarkIvyFetchSlots, including one recycle and complete teardown. */
#include "aig/ivy/ivy.h"
#include <time.h>

int main(void)
{
    const int counts[] = {16, 5000};
    for (int k = 0; k < 2; k++) {
        const int count = counts[k], runs = k ? 12000 : 40000;
        struct timespec a, b;
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &a);
        for (int r = 0; r < runs; r++) {
            Ivy_Man_t p = {0};
            Ivy_ManStartMemory(&p);
            Ivy_Obj_t *last = NULL;
            for (int i = 0; i < count; i++) {
                last = Ivy_ManFetchMemory(&p);
                last->Id = i;
            }
            Ivy_ManRecycleMemory(&p, last);
            Ivy_Obj_t *reused = Ivy_ManFetchMemory(&p);
            if (reused != last || reused->Id != 0) return 1;
            Ivy_ManStopMemory(&p);
        }
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &b);
        printf("%d slots %.2f us/pool\n", count,
               ((b.tv_sec-a.tv_sec)*1e6 + (b.tv_nsec-a.tv_nsec)*1e-3)/runs);
    }
    return 0;
}

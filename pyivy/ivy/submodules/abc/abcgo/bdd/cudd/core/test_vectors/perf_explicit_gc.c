/* Matches BenchmarkExplicitGCMixedKeys. Time only the collection of
   3,072 released keys among 4,096 ANDs sharing 64 projection nodes. */
#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <time.h>

int main(void)
{
    const int runs = 4000;
    double elapsed = 0;
    for (int n = 0; n < runs; ++n) {
        DdManager *dd = Cudd_Init(65, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0);
        Cudd_DisableGarbageCollection(dd);
        for (int i = 0; i < 64; ++i) {
            for (int j = 0; j < 64; ++j) {
                DdNode *f = cuddUniqueInter(dd, 0, dd->vars[i+1], Cudd_Not(dd->vars[j+1]));
                Cudd_Ref(f);
                if ((i+j)%4) Cudd_RecursiveDeref(dd, f);
            }
        }
        struct timespec a, b;
        clock_gettime(CLOCK_MONOTONIC, &a);
        int removed = cuddGarbageCollect(dd, 1);
        clock_gettime(CLOCK_MONOTONIC, &b);
        if (removed != 3072) {
            fprintf(stderr, "collected %d keys\n", removed);
            return 1;
        }
        elapsed += (b.tv_sec-a.tv_sec)*1e6 + (b.tv_nsec-a.tv_nsec)*1e-3;
        Cudd_Quit(dd);
    }
    printf("3072 released, 1024 retained, %.3f us/collection\n", elapsed/runs);
    return 0;
}

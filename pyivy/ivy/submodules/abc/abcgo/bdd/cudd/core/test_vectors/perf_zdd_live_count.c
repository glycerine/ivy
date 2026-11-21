/* Matches BenchmarkZDDReadMixedLiveCount. */
#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <time.h>

int main(void)
{
    DdManager *dd = Cudd_Init(0,32,64,256,0);
    Cudd_DisableGarbageCollection(dd);
    for (int i = 0; i < 31; ++i)
        for (int j = i+1; j < 32; ++j) {
            DdNode *child = cuddUniqueInterZdd(dd,j,dd->one,dd->zero);
            Cudd_Ref(child);
            DdNode *n = cuddUniqueInterZdd(dd,i,child,dd->zero);
            Cudd_Ref(n);
            Cudd_RecursiveDerefZdd(dd,child);
            if (j % 2 == 0) Cudd_RecursiveDerefZdd(dd,n);
        }
    const int runs = 300000000;
    long want = Cudd_zddReadNodeCount(dd);
    if (want < 200) return 1;
    struct timespec a,b;
    clock_gettime(CLOCK_MONOTONIC,&a);
    for (int i = 0; i < runs; ++i)
        if (Cudd_zddReadNodeCount(dd) != want) return 2;
    clock_gettime(CLOCK_MONOTONIC,&b);
    printf("%ld live, %.2f ns/count\n",want,((b.tv_sec-a.tv_sec)*1e9+b.tv_nsec-a.tv_nsec)/runs);
    Cudd_Quit(dd);
    return 0;
}

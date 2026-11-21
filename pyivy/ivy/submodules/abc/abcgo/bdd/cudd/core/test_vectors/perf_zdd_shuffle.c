/* Matches BenchmarkShuffleZDDTripleSets. A separate universe check validates
   collection after swapping variables with an externally pinned root. */
#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <time.h>

static int check_universe(void)
{
    DdManager *dd = Cudd_Init(0, 6, 64, 256, 0);
    int order[] = {4, 3, 1, 5, 2, 0};
    DdNode *root = dd->univ[0];
    Cudd_Ref(root);
    if (!Cudd_zddShuffleHeap(dd, order) || dd->univ[0] != root || dd->keysZ != 6) return 0;
    for (int i = 0; i < 6; ++i)
        if ((int)dd->univ[i]->index != order[i] || dd->univ[i]->ref != 2) return 0;
    Cudd_RecursiveDerefZdd(dd, root);
    cuddZddFreeUniv(dd);
    cuddGarbageCollect(dd, 1);
    int valid = dd->keysZ == 0;
    Cudd_Quit(dd);
    return valid;
}

int main(void)
{
    if (!check_universe()) return 1;
    const int runs = 12000, width = 10;
    int reverse[10];
    for (int i = 0; i < width; ++i) reverse[i] = width-1-i;
    struct timespec a, b;
    clock_gettime(CLOCK_MONOTONIC, &a);
    for (int n = 0; n < runs; ++n) {
        DdManager *dd = Cudd_Init(0, width, 64, 256, 0);
        Cudd_DisableGarbageCollection(dd);
        for (int i = 0; i < width; ++i)
            for (int j = i+1; j < width; ++j)
                for (int k = j+1; k < width; ++k) {
                    DdNode *leaf = cuddUniqueInterZdd(dd, k, dd->one, dd->zero);
                    DdNode *mid = cuddUniqueInterZdd(dd, j, leaf, dd->zero);
                    DdNode *root = cuddUniqueInterZdd(dd, i, mid, dd->zero);
                    Cudd_Ref(root);
                }
        if (!Cudd_zddShuffleHeap(dd, reverse) || dd->invpermZ[0] != 9 || dd->invpermZ[9] != 0) return 2;
        Cudd_Quit(dd);
    }
    clock_gettime(CLOCK_MONOTONIC, &b);
    printf("120 triple roots, %.3f us/build and shuffle\n", ((b.tv_sec-a.tv_sec)*1e6 + (b.tv_nsec-a.tv_nsec)*1e-3)/runs);
    return 0;
}

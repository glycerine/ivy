/* Matches BenchmarkAlignExistingZDDVars and checks growing a pinned cover. */
#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <time.h>

static int contains(DdManager *dd, DdNode *f, unsigned set)
{
    while (!Cudd_IsConstant(f)) {
        unsigned bit = 1u << f->index;
        if (set & bit) {
            set ^= bit;
            f = cuddT(f);
        } else {
            f = cuddE(f);
        }
    }
    return f == dd->one && set == 0;
}

static int check_growth(void)
{
    DdManager *dd = Cudd_Init(4, 4, 32, 64, 0);
    DdNode *old = dd->univ[0];
    Cudd_Ref(old);
    DdNode *a = cuddUniqueInterZdd(dd, 2, dd->one, dd->zero);
    Cudd_Ref(a);
    DdNode *b = cuddUniqueInterZdd(dd, 1, dd->one, dd->zero);
    Cudd_Ref(b);
    DdNode *f = cuddUniqueInterZdd(dd, 0, a, b);
    Cudd_Ref(f);
    Cudd_RecursiveDerefZdd(dd, a);
    Cudd_RecursiveDerefZdd(dd, b);
    int order[] = {1, 0, 3, 2};
    if (!Cudd_ShuffleHeap(dd, order) || !Cudd_zddVarsFromBddVars(dd, 2)) return 0;
    for (unsigned set = 0; set < 256; ++set)
        if (contains(dd, f, set) != (set == 2 || set == 5) ||
            contains(dd, old, set) != (set < 16) || !contains(dd, dd->univ[0], set)) return 0;
    Cudd_RecursiveDerefZdd(dd, f);
    Cudd_RecursiveDerefZdd(dd, old);
    cuddGarbageCollect(dd, 1);
    int valid = dd->keysZ == 8;
    Cudd_Quit(dd);
    return valid;
}

int main(void)
{
    if (!check_growth()) return 1;
    const int width = 12, runs = 2000000;
    DdManager *dd = Cudd_Init(width, width, 32, 128, 0);
    for (int i = 0; i < width; ++i)
        for (int j = i+1; j < width; ++j) {
            DdNode *leaf = cuddUniqueInterZdd(dd, j, dd->one, dd->zero);
            Cudd_Ref(leaf);
            DdNode *pair = cuddUniqueInterZdd(dd, i, leaf, dd->zero);
            Cudd_Ref(pair);
            Cudd_RecursiveDerefZdd(dd, leaf);
        }
    DdNode *root = dd->univ[0];
    struct timespec a,b;
    clock_gettime(CLOCK_MONOTONIC,&a);
    for (int i = 0; i < runs; ++i)
        if (!Cudd_zddVarsFromBddVars(dd,1) || dd->univ[0] != root) return 2;
    clock_gettime(CLOCK_MONOTONIC,&b);
    printf("12 variables, 66 pair roots, %.1f ns/alignment\n",
           ((b.tv_sec-a.tv_sec)*1e9 + b.tv_nsec-a.tv_nsec)/runs);
    Cudd_Quit(dd);
    return 0;
}

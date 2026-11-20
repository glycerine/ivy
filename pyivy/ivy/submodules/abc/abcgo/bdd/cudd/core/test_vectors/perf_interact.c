#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

/* Build disjoint conjunctions in the same node and variable order as Go. */
int main(int argc, char **argv)
{
    if (argc != 4) return 1;
    int groups = atoi(argv[1]), width = atoi(argv[2]), repeat = atoi(argv[3]);
    if (groups < 1 || width < 2 || width > 1024 || groups > 256 || repeat < 1)
        return 1;
    DdManager *dd = Cudd_Init(groups*width, 0, 256, 262144, 0);
    for (int group = 0; group < groups; ++group) {
        int first = group*width;
        DdNode *root = dd->vars[first+width-1];
        Cudd_Ref(root);
        for (int i = first+width-2; i >= first; --i) {
            DdNode *next = cuddUniqueInter(dd, i, root, Cudd_Not(dd->one));
            Cudd_Ref(next);
            Cudd_RecursiveDeref(dd, root);
            root = next;
        }
    }
    if (!cuddInitInteract(dd))
        return 2;
    for (int i = 0; i < groups*width; ++i)
        for (int j = i+1; j < groups*width; ++j)
            if (!!cuddTestInteract(dd, i, j) != (i/width == j/width))
                return 2;
    ABC_FREE(dd->interact);
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < repeat; ++i) {
        if (!cuddInitInteract(dd) || !cuddTestInteract(dd, 0, width-1) ||
            (groups > 1 && cuddTestInteract(dd, 0, width)))
            return 2;
        ABC_FREE(dd->interact);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("%d groups of %d variables, %d pairs, %.1f ns/iteration\n",
           groups, width, groups*width*(width-1)/2, ns/repeat);
    Cudd_Quit(dd);
    return 0;
}

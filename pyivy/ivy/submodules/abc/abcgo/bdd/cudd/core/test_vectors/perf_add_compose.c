#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char **argv)
{
    int repeat = argc == 2 ? atoi(argv[1]) : 0;
    if (repeat <= 0) return 1;
    DdManager *dd = Cudd_Init(5, 0, 256, 4096, 0);
    if (Cudd_bddAdjPermuteX(dd,dd->vars[0],dd->vars,2) != dd->vars[0] ||
        Cudd_bddAdjPermuteX(dd,dd->vars[3],dd->vars,5) != dd->vars[3] ||
        Cudd_bddAdjPermuteX(dd,dd->vars[0],dd->vars,5) != dd->vars[1])
        return 2;
    DdNode *on[5], *off[5], *vector[5];
    for (int i = 0; i < 5; ++i) {
        vector[i] = on[i] = Cudd_addIthVar(dd,i);
        Cudd_Ref(on[i]);
        off[i] = Cudd_addCmpl(dd,on[i]);
        Cudd_Ref(off[i]);
    }
    DdNode *x = on[0], *y = on[1];
    vector[0] = y;
    vector[1] = x;
    if (Cudd_addNonSimCompose(dd,x,vector) != y) return 3;
    int reverse[5] = {4,3,2,1,0};
    if (!Cudd_ShuffleHeap(dd,reverse) || Cudd_addNonSimCompose(dd,y,vector) != x) return 4;
    DdNode *f = Cudd_addIte(dd,x,Cudd_addConst(dd,5),Cudd_addConst(dd,7));
    Cudd_Ref(f);
    on[0] = Cudd_addIte(dd,y,Cudd_addConst(dd,3),dd->zero);
    Cudd_Ref(on[0]);
    off[0] = Cudd_addIte(dd,y,Cudd_addConst(dd,-2),dd->one);
    Cudd_Ref(off[0]);
    DdNode *want = Cudd_addIte(dd,y,dd->one,Cudd_addConst(dd,7));
    Cudd_Ref(want);
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC,&start);
    for (int i = 0; i < repeat; ++i)
        if (Cudd_addGeneralVectorCompose(dd,f,on,off) != want) return 5;
    clock_gettime(CLOCK_MONOTONIC,&end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("weighted leaves 7 and 1, %.1f ns/composition\n",ns/repeat);
    Cudd_Quit(dd);
    return 0;
}

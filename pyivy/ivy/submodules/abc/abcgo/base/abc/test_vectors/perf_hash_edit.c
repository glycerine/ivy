/* Include the implementation to exercise the same internal create/delete paths
   as the port. Link this driver in place of src/base/abc/abcAig.o. */
#include "base/abc/abcAig.c"
#include <time.h>

int main(int argc, char **argv)
{
    int repeat = argc == 2 ? atoi(argv[1]) : 0;
    if (repeat <= 0) return 1;
    Abc_Ntk_t *tiny = Abc_NtkAlloc(ABC_NTK_STRASH,ABC_FUNC_AIG,1);
    Abc_Aig_t *small = tiny->pManFunc;
    small->nBins = 3;
    Abc_Obj_t *x = Abc_NtkCreatePi(tiny), *y = Abc_NtkCreatePi(tiny);
    Abc_Obj_t *oldest = Abc_AigAndCreate(small,x,y);
    for (int i = 1; i < 7; ++i) Abc_AigAndCreate(small,x,y);
    if (small->nBins != 3) return 7;
    Abc_Obj_t *last = Abc_AigAndCreate(small,x,y);
    if (small->nBins != 11) return 8;
    Abc_AigAndDelete(small,last);
    if (Abc_AigAndLookup(small,x,y) != oldest) return 9;
    Abc_NtkDelete(tiny);

    Abc_Ntk_t *n = Abc_NtkAlloc(ABC_NTK_STRASH,ABC_FUNC_AIG,1);
    Abc_Aig_t *m = n->pManFunc;
    Abc_Obj_t *a = Abc_NtkCreatePi(n), *b = Abc_NtkCreatePi(n);
    for (int i = 0; i < 512; ++i) {
        Abc_Obj_t *pi = Abc_NtkCreatePi(n);
        Abc_AigAnd(m,a,pi);
        Abc_AigAnd(m,b,pi);
    }
    Abc_Obj_t *first = Abc_AigAnd(m,a,b);
    Abc_Obj_t *second = Abc_AigAndCreate(m,b,a);
    Abc_Obj_t *third = Abc_AigAndCreate(m,a,b);
    Abc_Obj_t *po = Abc_NtkCreatePo(n);
    Abc_ObjAddFanin(po,first);
    if (Abc_AigAndLookup(m,b,a) != third) return 2;
    Abc_AigAndDelete(m,third);
    if (Abc_AigAndLookup(m,a,b) != second) return 3;
    Abc_AigRehash(m);
    if (Abc_AigAndLookup(m,b,a) != first) return 4;
    Abc_AigResize(m);
    if (Abc_AigAndLookup(m,b,a) != second) return 5;
    Abc_AigAndDelete(m,second);
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC,&start);
    for (int i = 0; i < repeat; ++i) {
        Abc_ObjXorFaninC(po,0);
        if (Abc_AigAndLookup(m,b,a) != first) return 6;
    }
    clock_gettime(CLOCK_MONOTONIC,&end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("edited output, commuted AND retained, %.1f ns/op\n",ns/repeat);
    Abc_NtkDelete(n);
    return 0;
}

#include "base/abc/abc.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

extern word *Abc_NodeConeTruth(Vec_Ptr_t *, Vec_Ptr_t *, int, Abc_Obj_t *,
                               Vec_Ptr_t *, Vec_Ptr_t *);

/* Matches BenchmarkRefactorTruthSmallCone. */
int main(int argc, char **argv)
{
    int count = argc == 2 ? atoi(argv[1]) : 0;
    if (count <= 0)
        return 1;
    Abc_Ntk_t *n = Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 0);
    Abc_Aig_t *m = (Abc_Aig_t *)n->pManFunc;
    Abc_Obj_t *a = Abc_NtkCreatePi(n), *b = Abc_NtkCreatePi(n);
    for (int i = 0; i < 4096; ++i)
        Abc_NtkCreatePi(n);
    Abc_Obj_t *left = Abc_AigAnd(m, a, b);
    Abc_Obj_t *right = Abc_AigAnd(m, Abc_ObjNot(a), Abc_ObjNot(b));
    Abc_Obj_t *root = Abc_AigAnd(m, Abc_ObjNot(left), Abc_ObjNot(right));
    Abc_ObjAddFanin(Abc_NtkCreatePo(n), root);

    word elementary[] = { ABC_CONST(0xaaaaaaaaaaaaaaaa), ABC_CONST(0xcccccccccccccccc) };
    Vec_Ptr_t *vars = Vec_PtrAlloc(2), *leaves = Vec_PtrAlloc(2);
    Vec_Ptr_t *funcs = Vec_PtrAlloc(100), *visited = Vec_PtrAlloc(100);
    Vec_PtrPush(vars, &elementary[0]);
    Vec_PtrPush(vars, &elementary[1]);
    Vec_PtrPush(leaves, a);
    Vec_PtrPush(leaves, b);

    struct timespec start, stop;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; ++i) {
        word *truth = Abc_NodeConeTruth(vars, funcs, 1, root, leaves, visited);
        if (!truth || truth[0] != ABC_CONST(0x6666666666666666) || Vec_PtrSize(visited) != 3)
            return 2;
    }
    clock_gettime(CLOCK_MONOTONIC, &stop);
    double elapsed = (stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec;
    printf("3 nodes, XOR truth, %.1f ns/op\n", elapsed/count);
    Vec_PtrFreeFree(funcs);
    Vec_PtrFree(visited);
    Vec_PtrFree(leaves);
    Vec_PtrFree(vars);
    Abc_NtkDelete(n);
    return 0;
}

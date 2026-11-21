#include "base/abc/abc.h"
#include <time.h>

int main(void)
{
    Abc_Ntk_t *n = Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1);
    Abc_Obj_t *a = Abc_NtkCreatePi(n), *root = Abc_NtkCreatePi(n);
    for (int i = 0; i < 1024; ++i) root = Abc_AigAnd(n->pManFunc, root, a);
    const int runs = 75000;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < runs; ++i) {
        Vec_Ptr_t *postorder = Abc_NtkDfsNodes(n, &root, 1);
        if (Vec_PtrSize(postorder) != 1024 || Vec_PtrEntryLast(postorder) != root) return 1;
        Vec_PtrFree(postorder);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    printf("%.2f us/native 1024-AND DFS\n",(1e6*(end.tv_sec-start.tv_sec)+1e-3*(end.tv_nsec-start.tv_nsec))/runs);
    Abc_NtkDelete(n);
    return 0;
}

/* Link in place of abcAig.o to exercise the native internal constructors. */
#include "base/abc/abcAig.c"
#include <time.h>

static int check_pins(void)
{
    Abc_Ntk_t *n = Abc_NtkAlloc(ABC_NTK_STRASH,ABC_FUNC_AIG,1);
    Abc_Aig_t *m = n->pManFunc;
    Abc_Obj_t *a = Abc_NtkCreatePi(n), *b = Abc_NtkCreatePi(n);
    Abc_Obj_t *raw = Abc_AigAndCreate(m,Abc_ObjNot(a),a);
    if (!Abc_ObjFaninC0(raw) || Abc_ObjFaninC1(raw) || Abc_ObjFanoutNum(a) != 2) return 0;
    Abc_Obj_t *node = Abc_NtkCreateNode(n);
    node->Level = 57;
    b->Level = 4;
    if (Abc_AigAndCreateFrom(m,b,Abc_ObjNot(a),node) != node ||
        node->Level != 5 || !Abc_ObjFaninC0(node) || Abc_ObjFaninC1(node)) return 0;
    Abc_NtkDelete(n);
    return 1;
}

int main(void)
{
    if (!check_pins()) return 1;
    const int runs = 8000;
    struct timespec a,b;
    clock_gettime(CLOCK_MONOTONIC,&a);
    for (int r = 0; r < runs; ++r) {
        Abc_Ntk_t *n = Abc_NtkAlloc(ABC_NTK_STRASH,ABC_FUNC_AIG,1);
        Abc_Obj_t *x = Abc_NtkCreatePi(n), *root = Abc_NtkCreatePi(n);
        for (int i = 0; i < 1024; ++i) root = Abc_AigAnd(n->pManFunc,root,x);
        if (root->Level != 1024 || Abc_ObjFanoutNum(x) != 1024) return 2;
        Abc_NtkDelete(n);
    }
    clock_gettime(CLOCK_MONOTONIC,&b);
    printf("1024-AND chain, %.2f us/build\n",((b.tv_sec-a.tv_sec)*1e6+(b.tv_nsec-a.tv_nsec)*1e-3)/runs);
    return 0;
}

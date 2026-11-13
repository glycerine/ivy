#include "misc/util/util_hack.h"
#include "aig/aig/aig.h"
#include <stdio.h>

int main(void) {
    Aig_Man_t *p = Aig_ManStart(100), *result;
    Aig_Obj_t *a = Aig_ObjCreateCi(p), *b = Aig_ObjCreateCi(p);
    Aig_Obj_t *c = Aig_ObjCreateCi(p), *ab = Aig_And(p, a, b);
    Aig_Obj_t *redundant = Aig_And(p, ab, Aig_Or(p, a, c));
    Aig_ObjCreateCo(p, redundant);
    result = Aig_ManFraigPartitioned(p, 1, 1000, 0, 0);
    printf("partition-result nodes=%d outputs=%d\n", Aig_ManNodeNum(result), Aig_ManCoNum(result));
    Aig_ManStop(result);
    Aig_ManStop(p);
    {
        Vec_Ptr_t *versions = Vec_PtrAlloc(2);
        int i;
        for (i = 0; i < 2; i++) {
            Aig_Man_t *g = Aig_ManStart(100);
            Aig_Obj_t *x = Aig_ObjCreateCi(g), *y = Aig_ObjCreateCi(g);
            Aig_Obj_t *z = Aig_ObjCreateCi(g);
            Aig_Obj_t *root = i ? Aig_And(g, x, Aig_And(g, y, z)) : Aig_And(g, Aig_And(g, x, y), z);
            Aig_ObjCreateCo(g, root);
            Vec_PtrPush(versions, g);
        }
        result = Aig_ManChoicePartitioned(versions, 2, 1000, 0, 0);
        printf("choice-result nodes=%d outputs=%d choices=%d\n", Aig_ManNodeNum(result), Aig_ManCoNum(result), Aig_ManChoiceNum(result));
        Aig_ManStop(result);
        for (i = 0; i < 2; i++) Aig_ManStop((Aig_Man_t *)Vec_PtrEntry(versions, i));
        Vec_PtrFree(versions);
    }
    return 0;
}

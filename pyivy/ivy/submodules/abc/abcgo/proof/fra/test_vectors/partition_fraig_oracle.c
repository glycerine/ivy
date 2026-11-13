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
    return 0;
}

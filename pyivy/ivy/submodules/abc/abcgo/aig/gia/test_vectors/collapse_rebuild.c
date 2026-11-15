#define ABC_USE_CUDD
#include "aig/gia/giaClp.c"
#include "base/main/main.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"

int main(void) {
    int table, mask, i, inputs[3], root, cube;
    Gia_Man_t *p, *clean, *out;
    Gia_Obj_t *obj;
    Abc_Start();
    for (table = 0; table < 256; ++table) {
        p = Gia_ManStart(128);
        Gia_ManHashStart(p);
        for (i = 0; i < 3; ++i) inputs[i] = Gia_ManAppendCi(p);
        root = 0;
        for (mask = 0; mask < 8; ++mask) if ((table >> mask) & 1) {
            cube = 1;
            for (i = 0; i < 3; ++i)
                cube = Gia_ManHashAnd(p, cube, inputs[i] ^ !((mask >> i) & 1));
            root = Gia_ManHashOr(p, root, cube);
        }
        Gia_ManAppendCo(p, root);
        Gia_ManAppendCo(p, root ^ 1);
        Gia_ManAppendCo(p, inputs[0]);
        Gia_ManAppendCo(p, 1);
        Gia_ManHashStop(p);
        clean = Gia_ManCleanup(p);
        Gia_ManStop(p);
        out = Gia_ManCollapseTest(clean, 0);
        printf("table %x objects %d ands", table, Gia_ManObjNum(out));
        Gia_ManForEachAnd(out, obj, i)
            printf(" %d:%d", Gia_ObjFaninLit0p(out, obj), Gia_ObjFaninLit1p(out, obj));
        printf(" outputs");
        Gia_ManForEachCo(out, obj, i)
            printf(" %d", Gia_ObjFaninLit0p(out, obj));
        printf("\n");
        Gia_ManStop(out);
        Gia_ManStop(clean);
    }
    Abc_Stop();
    return 0;
}

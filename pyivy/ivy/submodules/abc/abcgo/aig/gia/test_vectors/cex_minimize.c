#include "aig/gia/giaCex.c"
#include "base/main/main.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
int main(void) {
    int table, assignment, effort, mask, i, inputs[3], root, cube;
    Gia_Man_t *p, *clean;
    Abc_Cex_t *c, *care;
    Abc_Start();
    for (table = 1; table < 255; ++table) for (assignment = 0; assignment < 8; ++assignment)
    if ((table >> assignment) & 1) for (effort = 0; effort < 2; ++effort) {
        p = Gia_ManStart(64); Gia_ManHashStart(p);
        for (i = 0; i < 3; ++i) inputs[i] = Gia_ManAppendCi(p);
        Gia_ManAppendCi(p);
        root = 0;
        for (mask = 0; mask < 8; ++mask) if ((table >> mask) & 1) {
            cube = 1;
            for (i = 0; i < 3; ++i) cube = Gia_ManHashAnd(p, cube, inputs[i] ^ !((mask >> i) & 1));
            root = Gia_ManHashOr(p, root, cube);
        }
        Gia_ManAppendCo(p, root); Gia_ManAppendCo(p, 0); Gia_ManSetRegNum(p, 1); Gia_ManHashStop(p);
        clean = Gia_ManCleanup(p); Gia_ManStop(p); p = clean;
        c = Abc_CexAlloc(1,3,1); c->iFrame=0; c->iPo=0; c->pData[0]=assignment<<1;
        care = Bmc_CexCareSatBasedMinimizeAig(p,c,effort,0);
        printf("case %x %d %d mask%x\n",table,assignment,effort,care->pData[0]);
        Abc_CexFree(care); Abc_CexFree(c); Gia_ManStop(p);
    }
    Abc_Stop();
    return 0;
}

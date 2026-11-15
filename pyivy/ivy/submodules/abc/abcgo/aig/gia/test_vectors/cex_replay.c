#include "aig/gia/giaCex.c"
#include "base/main/main.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"

static int function(Gia_Man_t *p, int table, int x, int s) {
    int root = 0, cube, mask;
    for (mask = 0; mask < 4; ++mask) if ((table >> mask) & 1) {
        cube = Gia_ManHashAnd(p, x ^ !(mask&1), s ^ !((mask>>1)&1));
        root = Gia_ManHashOr(p, root, cube);
    }
    return root;
}
static void words(unsigned *p, int nBits) {
    int i;
    for (i = 0; i < Abc_BitWordNum(nBits); ++i) printf("%08x", p[i]);
}
int main(void) {
    int transitions, property, fi, seed, init, frames, f, x, s, found, v0, v1, dual;
    int widths[] = {1, 2, 5, 37}, left[] = {8, 2, 14, 0}, right[] = {6, 8, 4, 0};
    unsigned random;
    Gia_Man_t *p, *clean;
    Abc_Cex_t *c, *first, *ext;
    Abc_Start();
    for (transitions = 0; transitions < 16; ++transitions)
    for (property = 0; property < 4; ++property)
    for (fi = 0; fi < 4; ++fi) for (seed = 0; seed < 4; ++seed) for (init = 0; init < 2; ++init) {
        frames = widths[fi]; p = Gia_ManStart(32); Gia_ManHashStart(p);
        x = Gia_ManAppendCi(p); s = Gia_ManAppendCi(p);
        Gia_ManAppendCo(p, function(p, left[property], x, s));
        Gia_ManAppendCo(p, function(p, right[property], x, s));
        Gia_ManAppendCo(p, function(p, transitions, x, s));
        Gia_ManSetRegNum(p, 1); Gia_ManHashStop(p);
        clean = Gia_ManCleanup(p); Gia_ManStop(p); p = clean;
        c = Abc_CexAlloc(1, 1, frames); c->iFrame = frames-1;
        if (init) Abc_InfoSetBit(c->pData, 0);
        random = seed+1;
        for (f = 0; f < frames; ++f) {
            random = random*1664525u+1013904223u;
            if (random>>31) Abc_InfoSetBit(c->pData, 1+f);
        }
        c->iPo = 0; v0 = Gia_ManVerifyCex(p,c,0); dual = Gia_ManVerifyCex(p,c,1);
        c->iPo = 1; v1 = Gia_ManVerifyCex(p,c,0);
        found = Gia_ManFindFailedPoCex(p,c,0);
        printf("case %d %d %d %d %d objects%d verify%d:%d:%d find%d:%d:%d", transitions, property, frames, seed, init, Gia_ManObjNum(p), v0,v1,dual,found,Gia_ManFindFailedPoCex(p,c,1),Gia_ManFindFailedPoCex(p,c,2));
        first = Abc_CexDup(c,c->nRegs); Gia_ManSetFailedPoCex(p,first);
        printf(" first%d:%d:%d",first->iPo,first->iFrame,first->nBits); Abc_CexFree(first);
        if (found >= 0) {
            c->iPo = found;
            ext = Gia_ManCexExtendToIncludeCurrentStates(p,c); printf(" states"); words(ext->pData,ext->nBits); Abc_CexFree(ext);
            ext = Gia_ManCexExtendToIncludeAllObjects(p,c); printf(" legacy-all"); words(ext->pData,ext->nBits); Abc_CexFree(ext);
            if (!init) {
                Gia_ManCounterExampleValueStart(p,c); printf(" probe"); words(p->pData2,frames*Gia_ManObjNum(p)); Gia_ManCounterExampleValueStop(p);
            }
        }
        printf("\n"); Abc_CexFree(c); Gia_ManStop(p);
    }
    Abc_Stop();
    return 0;
}

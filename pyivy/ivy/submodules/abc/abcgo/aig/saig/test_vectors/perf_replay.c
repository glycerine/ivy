#include "misc/util/abc_global.h"
#include "aig/saig/saig.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static Aig_Man_t *counter(void)
{
    Aig_Man_t *g = Aig_ManStart(128);
    Aig_Obj_t *carry = Aig_ObjCreateCi(g), *bad = Aig_ManConst1(g);
    Aig_Obj_t *state[6], *next[6];
    for (int i = 0; i < 6; ++i) {
        state[i] = Aig_ObjCreateCi(g);
        bad = Aig_And(g, bad, state[i]);
    }
    for (int i = 0; i < 6; ++i) {
        next[i] = Aig_Exor(g, state[i], carry);
        carry = Aig_And(g, carry, state[i]);
    }
    Aig_ObjCreateCo(g, bad);
    for (int i = 0; i < 6; ++i) Aig_ObjCreateCo(g, next[i]);
    Aig_ManSetRegNum(g, 6);
    return g;
}

int main(int argc, char **argv)
{
    int repeat = argc == 2 ? atoi(argv[1]) : 0;
    if (repeat <= 0) return 1;
    Aig_Man_t *g = counter();
    Abc_Cex_t *trace = Abc_CexAlloc(6,1,64);
    trace->iFrame = 63;
    for (int bit = 6; bit < trace->nBits; ++bit)
        Abc_InfoSetBit(trace->pData,bit);
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC,&start);
    for (int i = 0; i < repeat; ++i)
        if (!Saig_ManVerifyCex(g,trace)) return 2;
    clock_gettime(CLOCK_MONOTONIC,&end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("64 frames replayed, %.1f ns/op\n", ns/repeat);
    Abc_CexFree(trace);
    Aig_ManStop(g);
    return 0;
}

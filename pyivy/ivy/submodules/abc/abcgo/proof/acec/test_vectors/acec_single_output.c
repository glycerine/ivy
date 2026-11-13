#include <stdio.h>
#include "aig/gia/gia.h"
extern void Acec_InsertFadd(Gia_Man_t *, int *, int *);
extern void *Acec_ProduceBox(Gia_Man_t *, int);
int main(void) {
    Gia_Man_t *p = Gia_ManStart(200);
    int in[7], lo[2], hi[2], j0[3], j1[3];
    for (int i = 0; i < 7; i++) in[i] = Gia_ManAppendCi(p);
    Gia_ManHashStart(p);
    int product = Gia_ManHashAnd(p, in[0], in[1]);
    int digit = Gia_ManHashXor(p, in[2], in[3]);
    int pp = Gia_ManHashXor(p, Gia_ManHashAnd(p, digit, in[4]), in[5]);
    Gia_ManHashStop(p);
    j0[0] = product; j0[1] = pp; j0[2] = in[6];
    Acec_InsertFadd(p, j0, lo);
    j1[0] = lo[1]; j1[1] = in[0]; j1[2] = in[5];
    Acec_InsertFadd(p, j1, hi);
    Gia_ManAppendCo(p, hi[0]);
    printf("box=%d\n", Acec_ProduceBox(p, 0) != NULL);
    return 0;
}

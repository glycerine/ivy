#include <stdio.h>
#include <stdlib.h>
#include "aig/gia/gia.h"
#include "proof/abs/abs.h"

/* Six-input AND: one layer is refined, then the PPI set stays empty. */
int main(int argc, char **argv) {
    Gia_Man_t *g = Gia_ManStart(100);
    int in[6], root;
    for (int i = 0; i < 6; i++) in[i] = Gia_ManAppendCi(g);
    root = in[0];
    for (int i = 1; i < 6; i++) root = Gia_ManAppendAnd(g, root, in[i]);
    Gia_ManAppendCo(g, root);
    Abs_Par_t p;
    Abs_ParSetDefaults(&p);
    p.fAddLayer = 1;
    p.fUseFullProof = atoi(argv[1]);
    p.nFramesMax = 1;
    p.nTimeOut = 1;
    p.nRatioMin = p.nRatioMin2 = 0;
    int result = Gia_ManPerformGlaOld(g, &p, 0);
    printf("SOURCE_STATUS=%d CEX=%d\n", result, g->pCexSeq != NULL);
    return 0;
}

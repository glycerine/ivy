#include <stdio.h>
#include <stdlib.h>
#include "aig/gia/gia.h"
#include "proof/abs/abs.h"

/* Same UNSAT network as the old GLA refinement and rollback regression. */
int main(int argc, char **argv) {
    Gia_Man_t *g = Gia_ManStart(20);
    int x[5];
    for (int i = 0; i < 5; ++i) x[i] = Gia_ManAppendCi(g);
    int left = Gia_ManAppendAnd(g, Gia_ManAppendAnd(g, x[0], x[1]), x[2]);
    int right = Gia_ManAppendAnd(g, Gia_ManAppendAnd(g, x[0] ^ 1, x[3]), x[4]);
    Gia_ManAppendCo(g, Gia_ManAppendAnd(g, left, right));
    Abs_Par_t p;
    Abs_ParSetDefaults(&p);
    p.fUseFullProof = atoi(argv[1]);
    p.fAddLayer = 1;
    p.nTimeOut = 1;
    p.fPropFanout = 0;
    p.nRatioMin = atoi(argv[2]);
    p.nRatioMin2 = 0;
    p.nFramesMax = 3;
    int result = Gia_ManPerformGlaOld(g, &p, 0);
    printf("SOURCE status=%d frame=%d stable=%d cex=%d classes=", result, p.iFrame, p.nFramesNoChange, g->pCexSeq != NULL);
    if (g->vGateClasses) {
        for (int i = 0; i < Vec_IntSize(g->vGateClasses); i++)
            printf("%d", Vec_IntEntry(g->vGateClasses, i));
    }
    printf("\n");
    Gia_ManStop(g);
    return 0;
}

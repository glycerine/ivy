#include "misc/util/abc_global.h"
#include "bdd/bbr/bbr.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static int check_counter(int partitioned)
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
    Saig_ParBbr_t p;
    Bbr_ManSetDefaultParams(&p);
    p.fSilent = 1;
    p.fReorder = p.fReorderImage = 0;
    p.fPartition = partitioned;
    int status = Aig_ManVerifyUsingBdds(g, &p);
    int ok = status == 0 && g->pSeqModel && g->pSeqModel->iFrame == 63 &&
             Saig_ManVerifyCex(g, g->pSeqModel);
    if (!ok)
        fprintf(stderr, "partitioned %d: status %d, counterexample frame %d\n",
                partitioned, status, g->pSeqModel ? g->pSeqModel->iFrame : -1);
    Aig_ManStop(g);
    return ok;
}

int main(int argc, char **argv)
{
    if (argc != 3) return 1;
    int partitioned = atoi(argv[1]), repeat = atoi(argv[2]);
    if (repeat <= 0 || (partitioned != 0 && partitioned != 1)) return 1;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < repeat; ++i)
        if (!check_counter(partitioned)) return 2;
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = 1e9*(end.tv_sec-start.tv_sec) + end.tv_nsec-start.tv_nsec;
    printf("partitioned %d, counterexample replayed at frame 63, %.1f ns/op\n",
           partitioned, ns/repeat);
    return 0;
}

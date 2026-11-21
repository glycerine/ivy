/* Link in place of ivyFraig.o to verify and measure both native simulators. */
#include "aig/ivy/ivyFraig.c"
#include <time.h>

int main(void)
{
    Ivy_Man_t *net = Ivy_ManStart();
    Ivy_Obj_t *a = Ivy_ObjCreatePi(net), *b = Ivy_ObjCreatePi(net), *root = b;
    for (int i = 0; i < 1024; ++i) root = Ivy_And(net, a, root);
    Ivy_ObjCreatePo(net, root);
    Ivy_FraigParams_t params;
    Ivy_FraigParamsDefault(&params);
    params.nSimWords = 32;
    Ivy_FraigMan_t *p = Ivy_FraigStart(net, &params);
    Ivy_Man_t *result = p->pManFraig;
    Ivy_FraigSim_t *original = Ivy_ObjSim(a);
    for (int i = 0; i < 32; i++) {
        original->pData[i] = 0xaaaaaaaaU ^ i;
        Ivy_ObjSim(b)->pData[i] = 0xccccccccU;
    }
    Ivy_Obj_t *first = Ivy_ManObj(net, 3);
    Ivy_FraigSim_t *replacement = calloc(1, sizeof(Ivy_FraigSim_t) + sizeof(unsigned));
    replacement->pData[0] = 0x55555555U;
    p->nSimWords = 1;
    Ivy_ObjSetSim(a, replacement);
    Ivy_ObjSim(first)->pData[1] = 13;
    Ivy_NodeSimulate(p, first);
    if (Ivy_ObjSim(first)->pData[0] != 0x44444444U) return 1;
    first->fPhase ^= 1;
    Ivy_NodeSimulate(p, first);
    if (Ivy_ObjSim(first)->pData[0] != 0xbbbbbbbbU) return 2;
    Ivy_NodeSimulateSim(p, Ivy_ObjSim(first));
    if (Ivy_ObjSim(first)->pData[0] != 0x88888888U || Ivy_ObjSim(first)->pData[1] != 13) return 3;
    first->fPhase ^= 1;
    Ivy_ObjSetSim(a, original);
    ABC_FREE(replacement);
    p->nSimWords = 32;
    const int runs = 20000;
    for (int linked = 0; linked < 2; ++linked) {
        struct timespec x, y;
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &x);
        for (int i = 0; i < runs; ++i) {
            if (linked) Ivy_FraigSimulateOneSim(p);
            else Ivy_FraigSimulateOne(p);
            if (Ivy_ObjSim(root)->pData[31] != ((0xaaaaaaaaU ^ 31) & 0xccccccccU)) return 4;
        }
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &y);
        printf("%s %.2f us/round\n", linked ? "linked" : "graph",
               ((y.tv_sec-x.tv_sec)*1e6 + (y.tv_nsec-x.tv_nsec)*1e-3)/runs);
    }
    Ivy_FraigStop(p);
    Ivy_ManStop(result);
    Ivy_ManStop(net);
    return 0;
}

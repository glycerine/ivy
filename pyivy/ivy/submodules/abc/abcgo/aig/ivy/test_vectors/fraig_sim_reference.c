/* The Go test injects the original ivyFraig.c simulation function unchanged. */
#include <stdio.h>
typedef struct Ivy_FraigSim_t_ Ivy_FraigSim_t;
struct Ivy_FraigSim_t_ {
    int Type;
    Ivy_FraigSim_t *pFanin0, *pFanin1;
    unsigned pData[1];
};
typedef struct { int nSimWords; } Ivy_FraigMan_t;
#include "original_sim.inc"
int main(void) {
    unsigned a, b;
    int opcode;
    Ivy_FraigMan_t manager = {1};
    while (scanf("%d%x%x", &opcode, &a, &b) == 3) {
        Ivy_FraigSim_t left = {0}, right = {0}, out = {0};
        left.pData[0] = a;
        right.pData[0] = b;
        out.Type = opcode;
        out.pFanin0 = &left;
        out.pFanin1 = &right;
        Ivy_NodeSimulateSim(&manager, &out);
        printf("%08x\n", out.pData[0]);
    }
    return 0;
}

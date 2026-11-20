#include "opt/dar/darInt.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

enum { width = 16, levels = 48 };

static Aig_Man_t *graph(void)
{
    Aig_Obj_t *layer[width], *next[width];
    Aig_Man_t *g = Aig_ManStart(1024);
    for (int i = 0; i < width; i++)
        layer[i] = Aig_ObjCreateCi(g);
    for (int k = 0; k < levels; k++) {
        for (int i = 0; i < width; i++) {
            Aig_Obj_t *right = layer[(i + 1 + 2*(k % 5)) % width];
            next[i] = Aig_And(g, Aig_NotCond(layer[i], (i+k)%3 == 0),
                              Aig_NotCond(right, (i+2*k)%5 == 0));
        }
        for (int i = 0; i < width; i++)
            layer[i] = next[i];
    }
    for (int i = 0; i < width; i++)
        Aig_ObjCreateCo(g, layer[i]);
    return g;
}

static uint64_t mix(uint64_t h, unsigned value)
{
    return (h ^ value) * UINT64_C(1099511628211);
}

/* Check every used cut in physical order, including its slot, value and phase. */
static uint64_t digest(Aig_Man_t *g)
{
    uint64_t h = 1;
    Aig_Obj_t *gate;
    Dar_Cut_t *cut;
    int i, slot;
    Aig_ManForEachNode(g, gate, i) {
        h = mix(h, gate->Id);
        Dar_ObjForEachCut(gate, cut, slot) {
            h = mix(h, slot);
            h = mix(h, cut->Value);
            h = mix(h, cut->uTruth);
            h = mix(h, cut->nLeaves);
            for (unsigned j = 0; j < cut->nLeaves; j++)
                h = mix(h, cut->pLeaves[j]);
        }
    }
    return h;
}

int main(int argc, char **argv)
{
    int count = argc == 2 ? atoi(argv[1]) : 0;
    if (count < 1)
        return 1;
    Aig_Man_t *g = graph();
    for (int skip = 0; skip < 2; skip++) {
        struct timespec start, end;
        Aig_MmFixed_t *mem = Dar_ManComputeCuts(g, 10, skip, 0);
        uint64_t hash = digest(g);
        Aig_MmFixedStop(mem, 0);
        clock_gettime(CLOCK_MONOTONIC, &start);
        for (int i = 0; i < count; i++) {
            mem = Dar_ManComputeCuts(g, 10, skip, 0);
            Aig_MmFixedStop(mem, 0);
        }
        clock_gettime(CLOCK_MONOTONIC, &end);
        double us = 1e6*(end.tv_sec-start.tv_sec) + 1e-3*(end.tv_nsec-start.tv_nsec);
        printf("skip %d, %d ANDs, digest %016llx, %.2f us/enumeration\n",
               skip, Aig_ManNodeNum(g), (unsigned long long)hash, us/count);
    }
    Aig_ManStop(g);
    return 0;
}

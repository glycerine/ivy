#include "sat/cnf/cnf.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

extern Cnf_Cut_t *Cnf_CutAlloc(Cnf_Man_t *p, int nLeaves);

/* Same selected binary cuts as BenchmarkScanBinaryAndChain. */
int main(int argc, char **argv)
{
    const int gates = 1024;
    int count = argc == 2 ? atoi(argv[1]) : 0;
    if (count <= 0)
        return 1;
    Aig_Man_t *g = Aig_ManStart(2 * gates + 3);
    Cnf_Man_t *m = Cnf_ManStart();
    m->pManAig = g;
    Aig_Obj_t *last = Aig_ObjCreateCi(g);
    for (int i = 0; i < gates; i++) {
        Aig_Obj_t *input = Aig_ObjCreateCi(g);
        Aig_Obj_t *gate = Aig_And(g, last, input);
        Cnf_Cut_t *cut = Cnf_CutAlloc(m, 2);
        cut->Cost = 3;
        cut->pFanins[0] = last->Id;
        cut->pFanins[1] = input->Id;
        gate->pData = cut;
        last = gate;
    }
    Aig_ObjCreateCo(g, Aig_Not(last));
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; i++) {
        Vec_Ptr_t *mapped = Cnf_ManScanMapping(m, 1, 1);
        if (Vec_PtrSize(mapped) != gates || Vec_PtrEntry(mapped, 0) != last || m->aArea != 3 * gates)
            return 2;
        Vec_PtrFree(mapped);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double us = 1e6 * (end.tv_sec - start.tv_sec) + 1e-3 * (end.tv_nsec - start.tv_nsec);
    printf("%d gates, %d area, %.3f us/scan\n", gates, m->aArea, us / count);
    Aig_ManResetRefs(g);
    Cnf_ManStop(m);
    Aig_ManStop(g);
    return 0;
}

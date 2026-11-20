#include "sat/cnf/cnf.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

extern Cnf_Cut_t *Cnf_CutAlloc(Cnf_Man_t *p, int nLeaves);

/* Same selected binary cuts as BenchmarkScanBinaryAndChain. */
int main(int argc, char **argv)
{
    const int gates = 1024;
    int write = argc == 3 && strcmp(argv[2], "write") == 0;
    int derive = argc == 3 && strcmp(argv[2], "derive") == 0;
    int check = argc == 3 && strcmp(argv[2], "check") == 0;
    int count = argc == 2 || write || derive || check ? atoi(argv[1]) : 0;
    if (count <= 0)
        return 1;
    Aig_Man_t *g = Aig_ManStart(2 * gates + 3);
    Cnf_Man_t *m = Cnf_ManStart();
    m->pManAig = g;
    Aig_Obj_t *last = Aig_ObjCreateCi(g);
    for (int i = 0; i < gates; i++) {
        Aig_Obj_t *input = Aig_ObjCreateCi(g);
        Aig_Obj_t *gate = Aig_And(g, last, input);
        if (derive || check) {
            last = gate;
            continue;
        }
        Cnf_Cut_t *cut = Cnf_CutAlloc(m, 2);
        cut->Cost = 3;
        cut->pFanins[0] = last->Id;
        cut->pFanins[1] = input->Id;
        Cnf_CutTruth(cut)[0] = 0x88888888U;
        gate->pData = cut;
        last = gate;
    }
    Aig_ObjCreateCo(g, Aig_Not(last));
    Vec_Ptr_t *selected = write ? Cnf_ManScanMapping(m, 1, 1) : NULL;
    int clauses = 0, literals = 0;
    if (derive) {
        Cnf_Dat_t *d = Cnf_DeriveWithMan(m, g, 0);
        clauses = d->nClauses;
        literals = d->nLiterals;
        Cnf_DataFree(d);
    }
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; i++) {
        if (check) {
            if (!Aig_ManCheck(g))
                return 5;
            continue;
        }
        if (derive) {
            Cnf_Dat_t *d = Cnf_DeriveWithMan(m, g, 0);
            if (d->nClauses != clauses || d->nLiterals != literals)
                return 4;
            Cnf_DataFree(d);
            continue;
        }
        if (write) {
            Cnf_Dat_t *d = Cnf_ManWriteCnf(m, selected, 0);
            if (d->nVars != 2*gates+3 || d->nClauses != 3*gates+2 || d->nLiterals != 7*gates+2)
                return 3;
            Cnf_DataFree(d);
            continue;
        }
        Vec_Ptr_t *mapped = Cnf_ManScanMapping(m, 1, 1);
        if (Vec_PtrSize(mapped) != gates || Vec_PtrEntry(mapped, 0) != last || m->aArea != 3 * gates)
            return 2;
        Vec_PtrFree(mapped);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double us = 1e6 * (end.tv_sec - start.tv_sec) + 1e-3 * (end.tv_nsec - start.tv_nsec);
    if (check)
        printf("%d gates, %.3f us/check\n", gates, us / count);
    else if (derive)
        printf("%d clauses, %d literals, %.3f us/derive\n", clauses, literals, us / count);
    else
        printf("%d gates, %d area, %.3f us/%s\n", gates, m->aArea, us / count, write ? "write" : "scan");
    if (selected)
        Vec_PtrFree(selected);
    Aig_ManResetRefs(g);
    Cnf_ManStop(m);
    Aig_ManStop(g);
    return 0;
}

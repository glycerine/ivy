#include "sat/cnf/cnf.h"
#include "sat/bsat/satSolver.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char **argv)
{
    enum { variables = 2048, clauses = 8192 };
    int count = argc == 2 ? atoi(argv[1]) : 0;
    if (count < 1)
        return 1;
    Aig_Man_t *g = Aig_ManStart(1);
    Cnf_Dat_t *cnf = Cnf_DataAlloc(g, variables, clauses, 3*clauses);
    for (int i = 0; i < clauses; i++) {
        int v = i%variables, group = i/variables;
        int *p = cnf->pClauses[i] = cnf->pClauses[0] + 3*i;
        p[0] = 2*v;
        p[1] = 2*((v+17)%variables) + (group&1);
        p[2] = 2*((v+71)%variables) + (group>>1);
    }
    for (int load = 0; load < 2; load++) {
        struct timespec start, end;
        clock_gettime(CLOCK_MONOTONIC, &start);
        for (int i = 0; i < count; i++) {
            if (load) {
                sat_solver *s = Cnf_DataWriteIntoSolver(cnf, 1, 0);
                if (!s || sat_solver_nvars(s) != variables)
                    return 2;
                sat_solver_delete(s);
            } else {
                Cnf_Dat_t *d = Cnf_DataDup(cnf);
                if (d->nClauses != clauses || d->nLiterals != 3*clauses)
                    return 2;
                Cnf_DataFree(d);
            }
        }
        clock_gettime(CLOCK_MONOTONIC, &end);
        double us = 1e6*(end.tv_sec-start.tv_sec) + 1e-3*(end.tv_nsec-start.tv_nsec);
        printf("%s, %d clauses, %.3f us/op\n", load ? "load" : "duplicate", clauses, us/count);
    }
    Cnf_DataFree(cnf);
    Aig_ManStop(g);
    return 0;
}

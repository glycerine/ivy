#include "sat/bsat/satSolver.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

/* Build, prove and release the same ordered pigeonhole formula as the Go benchmark. */
int main(int argc, char **argv)
{
    int iterations = argc == 2 || argc == 3 ? atoi(argv[1]) : 0;
    int holes = argc == 3 ? atoi(argv[2]) : 6;
    if (iterations <= 0 || holes < 1 || holes > 7)
        return 1;
    struct timespec start, end;
    long long conflicts = 0, decisions = 0, propagations = 0;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int k = 0; k < iterations; k++) {
        sat_solver *s = sat_solver_new();
        sat_solver_setnvars(s, holes*(holes+1));
        for (int p = 0; p <= holes; p++) {
            lit clause[7];
            for (int h = 0; h < holes; h++)
                clause[h] = 2 * (holes * p + h);
            if (!sat_solver_addclause(s, clause, clause + holes))
                return 1;
        }
        for (int h = 0; h < holes; h++)
            for (int p = 0; p <= holes; p++)
                for (int q = p + 1; q <= holes; q++) {
                    lit clause[2] = {2 * (holes * p + h) + 1, 2 * (holes * q + h) + 1};
                    if (!sat_solver_addclause(s, clause, clause + 2))
                        return 1;
                }
        if (sat_solver_solve(s, NULL, NULL, 0, 0, 0, 0) != l_False)
            return 1;
        conflicts = s->stats.conflicts;
        decisions = s->stats.decisions;
        propagations = s->stats.propagations;
        sat_solver_delete(s);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double seconds = end.tv_sec - start.tv_sec + 1e-9 * (end.tv_nsec - start.tv_nsec);
    printf("%.3f ms/op %lld conflicts %lld decisions %lld propagations\n",
           1000 * seconds / iterations, conflicts, decisions, propagations);
    return 0;
}

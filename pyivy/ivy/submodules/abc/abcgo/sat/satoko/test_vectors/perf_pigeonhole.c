// Matched construction and search for BenchmarkPigeonhole7Into6.
#include "sat/satoko/satoko.h"
#include <time.h>

int main(void)
{
    enum { holes = 6, pigeons = 7, runs = 200 };
    long conflicts = 0, decisions = 0;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < runs; ++i) {
        satoko_t *s = satoko_create();
        satoko_setnvars(s, holes * pigeons);
        for (int p = 0; p < pigeons; ++p) {
            int row[holes];
            for (int h = 0; h < holes; ++h)
                row[h] = (p * holes + h) * 2;
            if (!satoko_add_clause(s, row, holes))
                return 1;
        }
        for (int h = 0; h < holes; ++h)
            for (int p = 0; p < pigeons; ++p)
                for (int q = p + 1; q < pigeons; ++q) {
                    int clash[] = { (p * holes + h) * 2 + 1, (q * holes + h) * 2 + 1 };
                    if (!satoko_add_clause(s, clash, 2))
                        return 1;
                }
        if (satoko_solve(s) != SATOKO_UNSAT)
            return 1;
        conflicts += satoko_stats(s)->n_conflicts_all;
        decisions += satoko_stats(s)->n_decisions;
        satoko_destroy(s);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ns = (end.tv_sec - start.tv_sec) * 1e9 + end.tv_nsec - start.tv_nsec;
    printf("%.0f ns/op, %ld conflicts/op, %ld decisions/op\n", ns / runs, conflicts / runs, decisions / runs);
    return 0;
}

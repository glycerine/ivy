/* Matches BenchmarkSATVariableSetup65536. */
#include "sat/bsat/satSolver.h"
#include <time.h>

int main(void)
{
    const int runs = 400;
    struct timespec a,b;
    clock_gettime(CLOCK_MONOTONIC,&a);
    for (int i=0;i<runs;++i) {
        sat_solver *s = sat_solver_new();
        sat_solver_setnvars(s,65536);
        if (sat_solver_nvars(s)!=65536 || s->assigns[65535]!=3 || s->polarity[65535]!=0 || s->tags[65535]!=0) return 1;
        sat_solver_delete(s);
    }
    clock_gettime(CLOCK_MONOTONIC,&b);
    printf("65536 variables, %.3f ms/setup\n",((b.tv_sec-a.tv_sec)*1e3+(b.tv_nsec-a.tv_nsec)*1e-6)/runs);
    return 0;
}

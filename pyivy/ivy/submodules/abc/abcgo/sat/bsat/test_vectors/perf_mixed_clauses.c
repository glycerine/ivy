/* Matches BenchmarkStoreMixedSATClauses and checks proof partition bitfields. */
#include "sat/bsat/satSolver2.h"
#include <stdio.h>
#include <time.h>

int main(void)
{
    sat_solver2 *s = sat_solver2_new();
    sat_solver2_setnvars(s, 2);
    int lits[] = {0,2,4,6,8,10,12,14};
    int h = Sat_MemAppend(&s->Mem, lits, 2, 0, 1);
    const int values[] = {1,2,3,256,-1,-2,0};
    for (unsigned i = 0; i < sizeof(values)/sizeof(values[0]); i++) {
        clause2_set_partA(s, h, values[i]);
        var_set_partA(s, 1, values[i]);
        if (clause2_is_partA(s,h) != (values[i]&1) || var_is_partA(s,1) != (values[i]&1))
            return 1;
    }
    sat_solver2_delete(s);
    struct timespec start, end;
    const int repeat = 6000;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int run = 0; run < repeat; run++) {
        Sat_Mem_t *mem = Sat_MemAlloc(12);
        int last = 0;
        for (int i = 0; i < 1024; i++)
            last = Sat_MemAppend(mem,lits,2+i%7,i%2,1);
        if (clause_id(Sat_MemClauseHand(mem,last)) != 511) return 2;
        Sat_MemFree(mem);
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &end);
    printf("%.2f process ns/1024 clauses\n",
           ((end.tv_sec-start.tv_sec)*1e9 + end.tv_nsec-start.tv_nsec)/repeat);
    return 0;
}

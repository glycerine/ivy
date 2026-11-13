/* From the ABC root, after building its object files:
 * gcc -DLIN64 -Isrc abcgo/proof/abs/test_vectors/vta_conflict_oracle.c \
 *   src/sat/bsat/{satSolver2,satMem,satStore,satProof}.o \
 *   src/misc/util/utilSort.o -lm -lpthread -lstdc++ -o /tmp/vta-oracle
 * /tmp/vta-oracle
 * budget=1 status=-1 conflicts=2
 * budget=0 status=1 conflicts=7
 * This uses the same sat_solver2_solve call and status mapping as
 * Vta_ManUnsatCore in src/proof/abs/absVta.c. */
#include <stdio.h>
#include "sat/bsat/satSolver2.h"

unsigned enable_dbg_outs = 0;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { return 0; }
char *vnsprintf(const char *format, va_list args) { return NULL; }
int Int2_ManChainStart(Int2_Man_t *p, clause *c) { return 0; }
int Int2_ManChainResolve(Int2_Man_t *p, clause *c, int lit, int var) { return 0; }
void Int2_ManStop(Int2_Man_t *p) {}

int main(void) {
    int budget, pigeon, hole, other, status;
    for (budget = 1; budget >= 0; budget--) {
        sat_solver2 *s = sat_solver2_new();
        lit clause[4], assumption = toLitCond(0, 0);
        sat_solver2_setnvars(s, 13);
        for (pigeon = 0; pigeon < 4; pigeon++) {
            clause[0] = toLitCond(0, 1);
            for (hole = 0; hole < 3; hole++)
                clause[hole+1] = toLitCond(1 + pigeon*3 + hole, 0);
            if (!sat_solver2_addclause(s, clause, clause+4, 1+pigeon)) return 2;
        }
        for (hole = 0; hole < 3; hole++)
            for (pigeon = 0; pigeon < 4; pigeon++)
                for (other = pigeon+1; other < 4; other++) {
                    clause[0] = toLitCond(1+pigeon*3+hole, 1);
                    clause[1] = toLitCond(1+other*3+hole, 1);
                    if (!sat_solver2_addclause(s, clause, clause+2, 5+hole*6+pigeon*4+other)) return 3;
                }
        status = sat_solver2_solve(s, &assumption, &assumption+1, budget, 0, 0, 0);
        printf("budget=%d status=%d conflicts=%d\n", budget,
               status == l_Undef ? -1 : status == l_True ? 0 : 1,
               (int)s->stats.conflicts);
        sat_solver2_delete(s);
    }
    return 0;
}

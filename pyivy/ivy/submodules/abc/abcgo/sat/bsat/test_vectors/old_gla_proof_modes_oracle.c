/* From the ABC root after building its objects:
 * gcc -DLIN64 -Isrc abcgo/sat/bsat/test_vectors/old_gla_proof_modes_oracle.c \
 *   src/sat/bsat/{satSolver2,satMem,satStore,satProof}.o \
 *   src/misc/util/utilSort.o -lm -lpthread -lstdc++ -o /tmp/old-gla-oracle
 * /tmp/old-gla-oracle
 * full=0 core=2
 * full=1 core=2,1
 */
#include <stdio.h>
#include "sat/bsat/satSolver2.h"
#include "sat/bsat/satProof2.h"

unsigned enable_dbg_outs = 0;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { return 0; }
char *vnsprintf(const char *format, va_list args) { return NULL; }
int Int2_ManChainStart(Int2_Man_t *p, clause *c) { return 0; }
int Int2_ManChainResolve(Int2_Man_t *p, clause *c, int lit, int var) { return 0; }
void Int2_ManStop(Int2_Man_t *p) {}

int main(void) {
    int full, i, id;
    for (full = 0; full <= 1; full++) {
        sat_solver2 *s = sat_solver2_new();
        Vec_Int_t *map = Vec_IntStartFull(4), *core;
        lit clause[2] = {toLitCond(1, 1), toLitCond(2, 0)};
        lit assumption = toLitCond(1, 0), unit = toLitCond(2, 1);
        sat_solver2_setnvars(s, 3);
        if (full) s->pPrf1 = Vec_SetAlloc(20);
        if (!sat_solver2_addclause(s, clause, clause+2, 1)) return 2;
        sat_solver2_simplify(s);
        sat_solver2_bookmark(s);
        if (!full) {
            s->pPrf2 = Prf_ManAlloc();
            Prf_ManRestart(s->pPrf2, map, sat_solver2_nlearnts(s), 1);
            Vec_IntWriteEntry(map, 2, 0);
        }
        if (!sat_solver2_addclause(s, &unit, &unit+1, 2)) return 3;
        if (sat_solver2_solve(s, &assumption, &assumption+1, 0, 0, 0, 0) != l_False) return 4;
        core = (Vec_Int_t *)Sat_ProofCore(s);
        printf("full=%d core=", full);
        Vec_IntForEachEntry(core, id, i) printf("%s%d", i ? "," : "", id);
        printf("\n");
        Vec_IntFree(core);
        sat_solver2_delete(s);
        Vec_IntFree(map);
    }
    return 0;
}

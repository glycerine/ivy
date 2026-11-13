/* Source oracle for solver_source_negative_limits_test.go.
 * Link against ABC's satSolver, satSolver2, satMem, satStore, satProof,
 * and utilSort objects. No interpolation manager is used in these cases.
 */
#include <stdio.h>
#include "sat/bsat/satSolver.h"
#include "sat/bsat/satSolver2.h"

unsigned enable_dbg_outs = 0;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { return 0; }
char *vnsprintf(const char *format, va_list args) { return NULL; }
int Int2_ManChainStart(Int2_Man_t *p, clause *c) { return 0; }
int Int2_ManChainResolve(Int2_Man_t *p, clause *c, int lit, int var) { return 0; }
void Int2_ManStop(Int2_Man_t *p) {}

int main(void) {
    int kind, position, warm, status, k;
    for (kind = 0; kind < 2; kind++)
        for (position = 0; position < 4; position++)
            for (warm = 0; warm < 2; warm++) {
                ABC_INT64_T limits[4] = {0, 0, 0, 0};
                limits[position] = -1;
                if (kind == 0) {
                    sat_solver *s = sat_solver_new();
                    sat_solver_setnvars(s, 2);
                    if (warm) sat_solver_solve(s, NULL, NULL, 0, 0, 0, 0);
                    status = sat_solver_solve(s, NULL, NULL, limits[0], limits[1], limits[2], limits[3]);
                    sat_solver_delete(s);
                } else {
                    sat_solver2 *s = sat_solver2_new();
                    sat_solver2_setnvars(s, 2);
                    if (warm) sat_solver2_solve(s, NULL, NULL, 0, 0, 0, 0);
                    status = sat_solver2_solve(s, NULL, NULL, limits[0], limits[1], limits[2], limits[3]);
                    sat_solver2_delete(s);
                }
                if (status != l_Undef) return 1;
            }
    for (kind = 0; kind < 2; kind++)
        for (position = 0; position < 2; position++)
            for (k = -4; k <= -2; k++) {
                if (kind == 0) {
                    sat_solver *s = sat_solver_new();
                    s->stats.conflicts = s->stats.propagations = 3;
                    status = sat_solver_solve(s, NULL, NULL, position ? 0 : k, position ? k : 0, 0, 0);
                    sat_solver_delete(s);
                } else {
                    sat_solver2 *s = sat_solver2_new();
                    s->stats.conflicts = s->stats.propagations = 3;
                    status = sat_solver2_solve(s, NULL, NULL, position ? 0 : k, position ? k : 0, 0, 0);
                    sat_solver2_delete(s);
                }
                if (status != (k == -3 ? l_True : l_Undef)) return 2;
            }
    for (kind = 0; kind < 2; kind++) {
        lit assumptions[2] = {0, 1};
        int sat_status, unsat_status;
        if (kind == 0) {
            sat_solver *s = sat_solver_new();
            sat_solver_setnvars(s, 1);
            sat_solver_solve(s, NULL, NULL, 0, 0, 0, 0);
            sat_status = sat_solver_solve(s, NULL, NULL, 0, -1, 0, 0);
            unsat_status = sat_solver_solve(s, assumptions, assumptions + 2, -1, 0, 0, 0);
            sat_solver_delete(s);
        } else {
            sat_solver2 *s = sat_solver2_new();
            sat_solver2_setnvars(s, 1);
            sat_solver2_solve(s, NULL, NULL, 0, 0, 0, 0);
            sat_status = sat_solver2_solve(s, NULL, NULL, 0, -1, 0, 0);
            unsat_status = sat_solver2_solve(s, assumptions, assumptions + 2, -1, 0, 0, 0);
            sat_solver2_delete(s);
        }
        if (sat_status != l_True || unsat_status != l_False) return 3;
    }
    puts("negative-limit source oracle passed");
    return 0;
}

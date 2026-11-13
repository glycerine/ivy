/* Linux oracle; build with ABC objects as in vta_conflict_oracle.c.
 * ABC's LIN64 Abc_Clock uses CLOCK_THREAD_CPUTIME_ID. Sleeping therefore
 * does not consume the solver's 100 ms CPU budget. */
#include <stdio.h>
#include <time.h>
#include "sat/bsat/satSolver2.h"

unsigned enable_dbg_outs = 0;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { return 0; }
char *vnsprintf(const char *format, va_list args) { return NULL; }
int Int2_ManChainStart(Int2_Man_t *p, clause *c) { return 0; }
int Int2_ManChainResolve(Int2_Man_t *p, clause *c, int lit, int var) { return 0; }
void Int2_ManStop(Int2_Man_t *p) {}

int main(void) {
    sat_solver2 *s = sat_solver2_new();
    struct timespec delay = {0, 250000000};
    lit assumption = toLitCond(0, 0);
    int status;
    sat_solver2_setnvars(s, 1);
    sat_solver2_set_runtime_limit(s, Abc_Clock() + CLOCKS_PER_SEC / 10);
    while (nanosleep(&delay, &delay) != 0) {}
    status = sat_solver2_solve(s, &assumption, &assumption+1, 0, 0, 0, 0);
    printf("status=%d\n", status == l_Undef ? -1 : status == l_True ? 0 : 1);
    sat_solver2_delete(s);
    return 0;
}

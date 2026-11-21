/* Read valid named and initialized AIGER circuits, then time the reset
   transformation used by BenchmarkDupZeroUndcChain. */
#include "aig/gia/gia.h"
#include <time.h>

static int check_read(const char *wire, int simple, int skip, int ands, int pos, int regs)
{
    int size = (int)strlen(wire);
    char *copy = Abc_UtilStrsav((char *)wire);
    Gia_Man_t *p = Gia_AigerReadFromMemory(copy, size, simple, skip, 1);
    ABC_FREE(copy);
    if (!p) return 0;
    int valid = Gia_ManAndNum(p) == ands && Gia_ManPoNum(p) == pos && Gia_ManRegNum(p) == regs;
    if (!valid)
        printf("read got %d ANDs, %d outputs, %d registers\n", Gia_ManAndNum(p), Gia_ManPoNum(p), Gia_ManRegNum(p));
    Gia_ManStop(p);
    return valid;
}

int main(void)
{
    const char *invariant = "aig 3 2 0 2 1\n2\n6\n\x02\x02o1 invariant:skip\n";
    if (!check_read(invariant, 0, 0, 0, 1, 0) ||
        !check_read(invariant, 1, 0, 1, 1, 0) ||
        !check_read(invariant, 0, 1, 1, 1, 0) ||
        !check_read("aig 1 0 1 1 0\n0 2\n1\n", 0, 0, 0, 1, 2) ||
        !check_read("aig 5 2 1 2 2\n0 1\n8\n10\n\x04\x02\x06\x02", 0, 1, 2, 2, 1) ||
        !check_read("aig 4 2 0 2 2\n6\n8\n\x02\x02\x04\x02o1 constraint:enable\n", 0, 1, 2, 2, 0))
        return 1;
    Gia_Man_t *p = Gia_ManStart(4110);
    int x = Gia_ManAppendCi(p), r = Gia_ManAppendCi(p), r1 = Gia_ManAppendCi(p);
    Gia_ManAppendCi(p);
    int root = r;
    for (int i = 0; i < 4096; ++i)
        root = Gia_ManAppendAnd(p, x ^ (i & 1), root);
    int drivers[] = {root, x, r1 ^ 1, 0};
    for (int i = 0; i < 4; ++i) Gia_ManAppendCo(p, drivers[i]);
    Gia_ManSetRegNum(p, 3);
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    int runs = 16000;
    for (int i = 0; i < runs; ++i) {
        Gia_Man_t *q = Gia_ManDupZeroUndc(p, "X1X", 0, 0, 0);
        if (Gia_ManAndNum(q) != 4099 || Gia_ManRegNum(q) != 4 || Gia_ManPiNum(q) != 3)
            return 2;
        Gia_ManStop(q);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    Gia_ManStop(p);
    printf("%.3f us/reset transform\n", ((end.tv_sec-start.tv_sec)*1e6 + (end.tv_nsec-start.tv_nsec)*1e-3)/runs);
    return 0;
}

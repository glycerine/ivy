#include "base/abc/abc.h"
#include "base/io/ioAbc.h"
#include "base/main/main.h"
#include "proof/fraig/fraig.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

/* Matches BenchmarkI10NativeResynthesisCEC. Load the same pair before timing
   and construct, prove, and destroy a fresh miter in each iteration. */
int main(int argc, char **argv)
{
    int count = argc == 4 ? atoi(argv[3]) : 0;
    if (count <= 0) {
        fprintf(stderr, "usage: perf_cec original.aig resynthesized.aig iterations\n");
        return 1;
    }
    Abc_Start();
    Abc_Ntk_t *left = Io_ReadAiger(argv[1], 1);
    Abc_Ntk_t *right = Io_ReadAiger(argv[2], 1);
    if (!left || !right)
        return 1;
    struct timespec start, end;
    int nodes = 0;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; i++) {
        Abc_Ntk_t *miter = Abc_NtkMiter(left, right, 1, 0, 0, 0);
        if (!miter || Abc_NtkMiterIsConstant(miter) != -1)
            return 1;
        nodes = Abc_NtkNodeNum(miter);
        Prove_Params_t params;
        Prove_ParamsSetDefault(&params);
        params.nItersMax = 5;
        if (Abc_NtkIvyProve(&miter, &params) != 1)
            return 1;
        Abc_NtkDelete(miter);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double elapsed = (end.tv_sec - start.tv_sec) + 1e-9 * (end.tv_nsec - start.tv_nsec);
    printf("%.3f ms/op, %d miter ANDs/op, equivalent\n", 1000 * elapsed / count, nodes);
    Abc_NtkDelete(left);
    Abc_NtkDelete(right);
    Abc_Stop();
    return 0;
}

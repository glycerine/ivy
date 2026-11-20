#include "base/abc/abc.h"
#include "base/io/ioAbc.h"
#include "base/main/main.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

/* Matches BenchmarkI10ReadBalanceRewrite, including network teardown. */
int main(int argc, char **argv)
{
    int iterations = argc == 3 ? atoi(argv[2]) : 0;
    int nodes = 0, depth = 0;
    struct timespec start, end;
    if (iterations <= 0) {
        fprintf(stderr, "usage: perf_rewrite input.aig iterations\n");
        return 1;
    }
    Abc_Start();
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < iterations; i++) {
        Abc_Ntk_t *network = Io_ReadAiger(argv[1], 1);
        if (!network)
            return 1;
        for (int pass = 0; pass < 2; pass++) {
            Abc_Ntk_t *balanced = Abc_NtkBalance(network, 0, 0, 1);
            Abc_NtkDelete(network);
            network = balanced;
        }
        if (!Abc_NtkRewrite(network, 0, 0, 0, 0, 0))
            return 1;
        nodes = Abc_NtkNodeNum(network);
        depth = Abc_AigLevel(network);
        Abc_NtkDelete(network);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double seconds = end.tv_sec - start.tv_sec + 1e-9 * (end.tv_nsec - start.tv_nsec);
    printf("%.3f ms/op  %d ANDs/op  %d levels/op\n", 1000 * seconds / iterations, nodes, depth);
    Abc_Stop();
    return 0;
}

#include "base/abc/abc.h"
#include "base/io/ioAbc.h"
#include "base/main/main.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

extern void Abc_NodeBalancePermute(Abc_Ntk_t *, Vec_Ptr_t *, int);

static int check_single_pair(void)
{
    Abc_Obj_t a = {0}, b = {0};
    Vec_Ptr_t *pair = Vec_PtrAlloc(2);
    Vec_PtrPush(pair, &a);
    Vec_PtrPush(pair, Abc_ObjNot(&b));
    Abc_NodeBalancePermute(NULL, pair, 0);
    int ok = Vec_PtrEntry(pair, 0) == &a && Vec_PtrEntry(pair, 1) == Abc_ObjNot(&b);
    Vec_PtrFree(pair);
    return ok;
}

/* Matches BenchmarkI10ReadBalanceRewrite, including network teardown. */
int main(int argc, char **argv)
{
    int balance_only = argc == 4 && strcmp(argv[3], "balance") == 0;
    int iterations = argc == 3 || balance_only ? atoi(argv[2]) : 0;
    int nodes = 0, depth = 0;
    struct timespec start, end;
    if (iterations <= 0) {
        fprintf(stderr, "usage: perf_rewrite input.aig iterations [balance]\n");
        return 1;
    }
    Abc_Start();
    if (!check_single_pair())
        return 2;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < iterations; i++) {
        Abc_Ntk_t *network = Io_ReadAiger(argv[1], 1);
        if (!network)
            return 1;
        for (int pass = 0; pass < (balance_only ? 1 : 2); pass++) {
            Abc_Ntk_t *balanced = Abc_NtkBalance(network, 0, 0, 1);
            Abc_NtkDelete(network);
            network = balanced;
        }
        if (!balance_only && !Abc_NtkRewrite(network, 0, 0, 0, 0, 0))
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

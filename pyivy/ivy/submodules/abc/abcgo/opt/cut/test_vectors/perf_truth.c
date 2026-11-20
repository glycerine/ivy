#include "opt/cut/cutInt.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

/* Matches BenchmarkCutTruthMerge. Replicate each child's function over its
   unused inputs before calling the native truth-table API. */
static void fill_truth(Cut_Cut_t *cut, unsigned seed)
{
    unsigned *table = Cut_CutReadTruth(cut);
    unsigned domain = 1u << cut->nLeaves;
    for (int i = 0; i < Cut_TruthWords(cut->nVarsMax); i++) {
        unsigned value = 0;
        for (int b = 0; b < 32; b++) {
            unsigned index = (32u * i + b) % domain;
            unsigned word = seed + 0x9e3779b9u * (index / 32);
            value |= ((word >> (index % 32)) & 1u) << b;
        }
        table[i] = value;
    }
}

int main(int argc, char **argv)
{
    int n = argc == 3 ? atoi(argv[1]) : 0;
    int count = argc == 3 ? atoi(argv[2]) : 0;
    if (n < 3 || n > CUT_SIZE_MAX || count <= 0) {
        fprintf(stderr, "usage: perf_truth variables iterations\n");
        return 1;
    }
    Cut_Params_t params = {0};
    params.nVarsMax = n;
    params.fTruth = 1;
    Cut_Man_t *manager = Cut_ManStart(&params);
    Cut_Cut_t *out = Cut_CutAlloc(manager);
    Cut_Cut_t *left = Cut_CutAlloc(manager);
    Cut_Cut_t *right = Cut_CutAlloc(manager);
    for (int i = 0; i < n; i++) {
        out->pLeaves[out->nLeaves++] = i;
        if (i % 3 != 1)
            left->pLeaves[left->nLeaves++] = i;
        if (i % 3 != 0)
            right->pLeaves[right->nLeaves++] = i;
    }
    fill_truth(left, 0xf185349au);
    fill_truth(right, 0x8b234ce7u);
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; i++)
        Cut_TruthCompute(manager, out, left, right, 1, 0);
    clock_gettime(CLOCK_MONOTONIC, &end);
    uint32_t hash = 2166136261u;
    for (int i = 0; i < Cut_TruthWords(n); i++)
        hash = (hash ^ Cut_CutReadTruth(out)[i]) * 16777619u;
    double ns = 1e9 * (end.tv_sec - start.tv_sec) + end.tv_nsec - start.tv_nsec;
    printf("%d vars, %.2f ns/op, truth hash %08x\n", n, ns / count, hash);
    Cut_ManStop(manager);
    return 0;
}

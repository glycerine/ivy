/* Seven ANDs with complement phases, also used by BenchmarkDecompositionTruth32. */
#include "base/abc/abc.h"
#include "bool/dec/dec.h"
#include "bool/kit/kit.h"
#include <stdio.h>
#include <time.h>

int main(void)
{
    Dec_Graph_t *g = Dec_GraphCreate(4);
    Dec_Edge_t root = Dec_EdgeCreate(0, 0);
    for (int i = 0; i < 7; ++i) {
        root.fCompl ^= 1;
        root = Dec_GraphAddNodeAnd(g, root, Dec_EdgeCreate(1+i%3, i%2));
    }
    root.fCompl ^= 1;
    Dec_GraphSetRoot(g, root);
    unsigned truth = Dec_GraphDeriveTruth(g);
    struct timespec start, stop;
    const int repeat = 4000000;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < repeat; ++i)
        if (Dec_GraphDeriveTruth(g) != truth) return 1;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    printf("%08x %.2f process ns/truth\n", truth,
           ((stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec)/repeat);
    Dec_GraphFree(g);

    Kit_Graph_t *literal = Kit_GraphCreateLeaf(2, 4, 1);
    for (int i = 0; i < 4; i++) literal->pNodes[i].iFunc = i + 10;
    if (Kit_GraphToTruth(literal) != 0x0f0f0f0fU) return 2;
    for (int i = 0; i < 4; i++)
        if (literal->pNodes[i].iFunc != i+10) return 3;
    Kit_GraphFree(literal);
    return 0;
}

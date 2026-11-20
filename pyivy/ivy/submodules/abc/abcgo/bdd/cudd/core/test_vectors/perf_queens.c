#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

/* Row-major variable order and pair insertion order match the Go benchmark. */
static int queens(int side, int *nodes)
{
    int vars = side*side;
    DdManager *dd = Cudd_Init(vars, 0, 256, 16384, 0);
    DdNode *f = Cudd_ReadOne(dd);
    Cudd_Ref(f);
    for (int r = 0; r < side; r++) {
        DdNode *row = Cudd_ReadLogicZero(dd);
        Cudd_Ref(row);
        for (int c = 0; c < side; c++) {
            DdNode *x = Cudd_bddOr(dd, row, Cudd_bddIthVar(dd, r*side+c));
            Cudd_Ref(x);
            Cudd_RecursiveDeref(dd, row);
            row = x;
        }
        DdNode *x = Cudd_bddAnd(dd, f, row);
        Cudd_Ref(x);
        Cudd_RecursiveDeref(dd, row);
        Cudd_RecursiveDeref(dd, f);
        f = x;
    }
    for (int i = 0; i < vars; i++) {
        for (int j = i+1; j < vars; j++) {
            int dr = j/side - i/side, dc = abs(j%side-i%side);
            if (dr && dc && dr != dc)
                continue;
            DdNode *pair = Cudd_bddAnd(dd, Cudd_bddIthVar(dd, i), Cudd_bddIthVar(dd, j));
            Cudd_Ref(pair);
            DdNode *x = Cudd_bddAnd(dd, f, Cudd_Not(pair));
            Cudd_Ref(x);
            Cudd_RecursiveDeref(dd, pair);
            Cudd_RecursiveDeref(dd, f);
            f = x;
        }
    }
    int solutions = (int)Cudd_CountMinterm(dd, f, vars);
    *nodes = Cudd_DagSize(f);
    Cudd_RecursiveDeref(dd, f);
    Cudd_Quit(dd);
    return solutions;
}

int main(int argc, char **argv)
{
    int side = argc == 3 ? atoi(argv[1]) : 0;
    int repeat = argc == 3 ? atoi(argv[2]) : 0;
    if (side < 1 || side > 8 || repeat < 1)
        return 1;
    struct timespec start, end;
    int solutions = 0, nodes = 0;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < repeat; i++)
        solutions = queens(side, &nodes);
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ms = 1e3*(end.tv_sec - start.tv_sec) + 1e-6*(end.tv_nsec-start.tv_nsec);
    printf("%d queens, %d solutions, %d DAG nodes, %.4f ms/op\n", side, solutions, nodes, ms/repeat);
    return 0;
}

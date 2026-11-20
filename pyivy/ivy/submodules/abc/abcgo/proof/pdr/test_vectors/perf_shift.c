#include "aig/aig/aig.h"
#include "base/main/main.h"
#include "proof/pdr/pdr.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char **argv)
{
    enum { stages = 12 };
    int count = argc == 2 ? atoi(argv[1]) : 0, frame = -1;
    if (count <= 0)
        return 1;
    Abc_Start();
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int trial = 0; trial < count; trial++) {
        Aig_Man_t *g = Aig_ManStart(64);
        g->pName = Abc_UtilStrsav("paired_shift");
        Aig_Obj_t *input = Aig_ObjCreateCi(g), *left[stages], *right[stages];
        for (int i = 0; i < stages; i++) {
            left[i] = Aig_ObjCreateCi(g);
            right[i] = Aig_ObjCreateCi(g);
        }
        Aig_ObjCreateCo(g, Aig_Exor(g, left[stages-1], right[stages-1]));
        Aig_ObjCreateCo(g, input);
        Aig_ObjCreateCo(g, input);
        for (int i = 1; i < stages; i++) {
            Aig_ObjCreateCo(g, left[i-1]);
            Aig_ObjCreateCo(g, right[i-1]);
        }
        Aig_ManSetRegNum(g, 2*stages);
        Pdr_Par_t p;
        Pdr_ManSetDefaultParams(&p);
        p.fSilent = 1;
        if (Pdr_ManSolve(g, &p) != 1)
            return 2;
        frame = p.iFrame;
        Aig_ManStop(g);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ms = 1e3 * (end.tv_sec - start.tv_sec) + 1e-6 * (end.tv_nsec - start.tv_nsec);
    printf("%d stages, last frame %d, %.3f ms/proof\n", stages, frame, ms / count);
    Abc_Stop();
    return 0;
}

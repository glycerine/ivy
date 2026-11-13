#include "aig/aig/aig.h"
#include "aig/saig/saig.h"
#include <stdlib.h>
#include <sys/resource.h>

int main(int argc, char **argv)
{
    struct rlimit no_core = {0, 0};
    setrlimit(RLIMIT_CORE, &no_core);
    if (argc != 5) return 2;
    int frames = atoi(argv[1]), maximum = atoi(argv[2]);
    int registers = atoi(argv[3]), internal = atoi(argv[4]);
    Aig_Man_t *g = Aig_ManStart(50);
    Aig_Obj_t *x = Aig_ObjCreateCi(g), *q = NULL;
    if (registers) q = Aig_ObjCreateCi(g);
    Aig_ObjCreateCo(g, registers ? (internal ? Aig_And(g, x, q) : q) : x);
    if (registers) {
        Aig_ObjCreateCo(g, x);
        Aig_ManSetRegNum(g, 1);
    }
    Aig_Man_t *result = Saig_ManTimeframeSimplify(g, frames, maximum, 1, 0);
    if (result) Aig_ManStop(result);
    Aig_ManStop(g);
    return result ? 0 : 3;
}

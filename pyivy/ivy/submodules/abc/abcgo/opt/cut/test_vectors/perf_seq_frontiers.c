/* One previously seen merge must leave room for the next sequential cut. */
#include "opt/cut/cutInt.h"
#include <time.h>

static Cut_Cut_t *family(Cut_Man_t *p, int size)
{
    Cut_Cut_t *c = Cut_CutAlloc(p);
    c->nLeaves = size;
    for (int i = 0; i < size; ++i) {
        c->pLeaves[i] = (16+i) << CUT_SHIFT;
        c->uSign |= Cut_NodeSign(c->pLeaves[i]);
    }
    return c;
}

static int check_choices(int sequential)
{
    Cut_Params_t params = {0};
    params.nVarsMax = 4; params.nKeepMax = 2; params.nIdsMax = 24;
    params.fFilter = 1; params.fSeq = sequential; params.nCutSet = 1;
    Cut_Man_t *p = Cut_ManStart(&params);
    Cut_Cut_t *top = Cut_CutCreateTriv(p,8), *other = Cut_CutCreateTriv(p,9);
    Cut_Cut_t *first = family(p,3), *last = family(p,2);
    top->pNext = first;
    other->pNext = last;
    Cut_NodeWriteCutsNew(p,8,top);
    Cut_NodeWriteCutsNew(p,9,other);
    if (sequential) {
        Cut_Cut_t *phase = Cut_CutCreateTriv(p,8);
        phase->fSimul = 1;
        Cut_NodeWriteCutsOld(p,8,phase);
    }
    Vec_Int_t *nodes = Vec_IntAlloc(2);
    Vec_IntPush(nodes,8); Vec_IntPush(nodes,9);
    Cut_Cut_t *result = sequential ? Cut_NodeUnionCutsSeq(p,nodes,-1,1) : Cut_NodeUnionCuts(p,nodes);
    int ok = result == top && top->pNext == first && first->pNext == NULL && first->fCompl == sequential;
    if (sequential && ok) {
        Cut_Cut_t *larger = family(p,4);
        Cut_NodeWriteCutsTemp(p,0,larger);
        params.nKeepMax = 10;
        ok = Cut_NodeUnionCutsSeq(p,nodes,0,0) == NULL && Cut_NodeReadCutsNew(p,8) == top;
    }
    Vec_IntFree(nodes);
    Cut_ManStop(p);
    return ok;
}

int main(void)
{
    Cut_Params_t params = {0};
    params.nVarsMax = 4; params.nKeepMax = 2; params.nIdsMax = 24; params.fSeq = 1;
    Cut_Man_t *p = Cut_ManStart(&params);
    Cut_Cut_t *a = Cut_CutCreateTriv(p,16), *b = Cut_CutCreateTriv(p,17);
    a->pNext = Cut_CutCreateTriv(p,18);
    Cut_NodeWriteCutsNew(p,0,a);
    Cut_NodeWriteCutsNew(p,1,b);
    Cut_NodeWriteCutsOld(p,8,family(p,2));
    struct timespec x,y;
    const int count = 4000000;
    clock_gettime(CLOCK_MONOTONIC,&x);
    for (int i = 0; i < count; ++i) {
        Cut_NodeComputeCutsSeq(p,8,0,1,0,0,0,0,0,-1);
        Cut_Cut_t *c = Cut_NodeReadCutsNew(p,8);
        if (!c || c->pNext || c->nLeaves != 2 || c->pLeaves[0] != (17<<CUT_SHIFT) || c->pLeaves[1] != (18<<CUT_SHIFT)) return 1;
        Cut_NodeFreeCuts(p,8);
    }
    clock_gettime(CLOCK_MONOTONIC,&y);
    if (p->timeMerge <= 0) return 4;
    printf("%.2f ns/sequential frontier with one free slot\n", ((y.tv_sec-x.tv_sec)*1e9+(y.tv_nsec-x.tv_nsec))/count);
    params.nKeepMax = 3;
    Cut_CutRecycleList(p,Cut_NodeReadCutsOld(p,8));
    Cut_Cut_t *top = Cut_CutCreateTriv(p,8);
    top->pNext = family(p,3);
    Cut_NodeWriteCutsOld(p,8,top);
    Cut_NodeComputeCutsSeq(p,8,0,1,0,0,0,0,0,-1);
    if (top->pNext || !Cut_NodeReadCutsNew(p,8)) return 2;
    Cut_ManStop(p);
    return check_choices(0) && check_choices(1) ? 0 : 3;
}

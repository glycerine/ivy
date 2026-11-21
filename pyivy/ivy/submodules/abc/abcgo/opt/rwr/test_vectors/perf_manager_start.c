/* Timed warm-frame startup and teardown of ABC's standard rewrite library. */
#include "base/main/main.h"
#include "bool/dec/dec.h"
#include "opt/rwr/rwr.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static unsigned long long mix(unsigned long long hash, unsigned value)
{
    return (hash ^ value) * 1099511628211ULL;
}

static unsigned long long library_digest(Rwr_Man_t *p)
{
    unsigned long long h = mix(14695981039346656037ULL, p->nTravIds);
    Rwr_Node_t *node;
    int i;
    Vec_PtrForEachEntry(Rwr_Node_t *, p->vForest, node, i) {
        h = mix(h, node->Id);
        h = mix(h, node->TravId);
        h = mix(h, node->Volume);
        h = mix(h, node->Level);
        h = mix(h, node->uTruth);
    }
    for (int c = 0; c < Vec_VecSize(p->vClasses); c++) {
        Vec_Ptr_t *members = Vec_VecEntry(p->vClasses, c);
        h = mix(h, p->pMapInv[c]);
        h = mix(h, Vec_PtrSize(members));
        Vec_PtrForEachEntry(Rwr_Node_t *, members, node, i) {
            Dec_Graph_t *g = (Dec_Graph_t *)node->pNext;
            h = mix(h, node->Id);
            h = mix(h, g->fConst);
            h = mix(h, g->nLeaves);
            h = mix(h, g->nSize);
            h = mix(h, Dec_EdgeToInt(g->eRoot));
            for (int j = g->nLeaves; j < g->nSize; j++) {
                Dec_Node_t *n = Dec_GraphNode(g, j);
                h = mix(h, Dec_EdgeToInt(n->eEdge0));
                h = mix(h, Dec_EdgeToInt(n->eEdge1));
                h = mix(h, n->fNodeOr | n->fCompl0<<1 | n->fCompl1<<2);
            }
        }
    }
    return h;
}

int main(int argc, char **argv)
{
    int repeat = argc > 1 ? atoi(argv[1]) : 1000;
    if (repeat < 1) return 1;
    Abc_Start();
    Rwr_Man_t *a = Rwr_ManStart(0), *b = Rwr_ManStart(0);
    if (a->puCanons != b->puCanons || a->pPhases != b->pPhases ||
        a->pPerms != b->pPerms || a->pMap != b->pMap ||
        a->pPractical == b->pPractical || a->vForest == b->vForest)
        return 2;
    printf("library digest %016llx\n", library_digest(a));
    Rwr_ManStop(a);
    Rwr_ManStop(b);
    struct timespec start, end;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < repeat; i++) {
        Rwr_Man_t *p = Rwr_ManStart(0);
        if (Vec_PtrSize(p->vForest) != 1785 || Vec_VecSize(p->vClasses) != 222)
            return 3;
        Rwr_ManStop(p);
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &end);
    printf("%.2f process us/manager\n",
           ((end.tv_sec-start.tv_sec)*1e6+(end.tv_nsec-start.tv_nsec)*1e-3)/repeat);
    Abc_Stop();
    return 0;
}

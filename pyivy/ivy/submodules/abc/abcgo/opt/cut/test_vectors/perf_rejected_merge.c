/* Native cut filtering on the same dominated four-leaf candidate as Go. */
#include "opt/cut/cutInt.h"
#include <time.h>

static Cut_Cut_t *pair(Cut_Man_t *p, int a, int b)
{
    Cut_Cut_t *c = Cut_CutAlloc(p);
    c->nLeaves = 2;
    c->pLeaves[0] = a;
    c->pLeaves[1] = b;
    c->uSign = Cut_NodeSign(a) | Cut_NodeSign(b);
    return c;
}

int main(void)
{
    Cut_Params_t params = {0};
    params.nVarsMax = 4; params.nKeepMax = 1000; params.fFilter = params.fTruth = 1;
    Cut_Man_t *p = Cut_ManStart(&params);
    Cut_List_t list;
    Cut_ListStart(&list);
    Cut_Cut_t *old = pair(p,1,2), *a = pair(p,1,3), *b = pair(p,2,4);
    Cut_ListAdd(&list,old);
    p->nNodeCuts = 1;
    const int count = 18000000;
    struct timespec x,y;
    clock_gettime(CLOCK_MONOTONIC,&x);
    for (int i = 0; i < count; ++i) Cut_NodeDoComputeCuts(p,&list,8,0,0,a,b,0,0);
    clock_gettime(CLOCK_MONOTONIC,&y);
    if (list.pHead[2] != old || old->pNext || p->nCutsFilter != count) return 1;
    printf("%.2f ns/rejected four-leaf merge\n",((y.tv_sec-x.tv_sec)*1e9+(y.tv_nsec-x.tv_nsec))/count);
    Cut_ManStop(p);
    return 0;
}

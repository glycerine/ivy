/* Matches the 16-input tree in BenchmarkRewriteOrderedNodeCuts. */
#include "base/main/main.h"
#include "base/abc/abc.h"
#include "opt/rwr/rwr.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char **argv)
{
    int repeat = argc > 1 ? atoi(argv[1]) : 1000000;
    int pinned = argc > 2 ? atoi(argv[2]) : 0;
    if (repeat < 1) return 1;
    Abc_Start();
    Abc_Ntk_t *net = Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1);
    Abc_Obj_t *level[16];
    for (int i = 0; i < 16; i++)
        level[i] = Abc_NtkCreatePi(net);
    for (int size = 16; size > 1; size /= 2) {
        for (int i = 0; i < size; i++) {
            if (pinned && size == 4)
                Abc_ObjAddFanin(Abc_NtkCreatePo(net), level[i]);
            if (i % 2 == 1)
                level[i / 2] = Abc_AigAnd(net->pManFunc, level[i-1], level[i]);
        }
    }
    Abc_ObjAddFanin(Abc_NtkCreatePo(net), level[0]);
    Cut_Params_t params = {0};
    params.nVarsMax = 4;
    params.nKeepMax = 250;
    params.fTruth = params.fFilter = 1;
    params.nIdsMax = Abc_NtkObjNumMax(net);
    Cut_Man_t *cuts = Cut_ManStart(&params);
    Abc_Obj_t *pi;
    int i;
    Abc_NtkForEachPi(net,pi,i)
        Cut_NodeSetTriv(cuts, pi->Id);
    Abc_NodeGetCutsRecursive(cuts, level[0], 0, 0);
    Rwr_Man_t *rwr = Rwr_ManStart(0);
    int gain = Rwr_NodeRewrite(rwr, cuts, level[0], 0, 1, 0);
    printf("pinned %d gain %d good %d subgraphs %d phase %d", pinned, gain,
           rwr->nCutsGood, rwr->nSubgraphs, rwr->fCompl);
    if (gain >= 0) {
        Abc_Obj_t *fanin;
        Vec_PtrForEachEntry(Abc_Obj_t *, rwr->vFanins, fanin, i)
            printf(" %d/%d", Abc_ObjRegular(fanin)->Id, Abc_ObjIsComplement(fanin));
    }
    puts("");
    struct timespec start, end;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (i = 0; i < repeat; i++)
        if (Rwr_NodeRewrite(rwr, cuts, level[0], 0, 1, 0) != gain) return 2;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &end);
    printf("%.2f process ns/node\n",
           ((end.tv_sec-start.tv_sec)*1e9 + end.tv_nsec-start.tv_nsec)/repeat);
    Rwr_ManStop(rwr);
    Cut_ManStop(cuts);
    Abc_NtkDelete(net);
    Abc_Stop();
    return 0;
}

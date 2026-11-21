/* Match BenchmarkDeleteAlternateNodes or BenchmarkDeleteDanglingAndChain. */
#include "base/main/main.h"
#include "base/abc/abc.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

int main(int argc, char **argv)
{
    int repeat = argc > 1 ? atoi(argv[1]) : 2000;
    const int count = 4096;
    int chain = argc > 2 && !strcmp(argv[2], "chain");
    if (repeat < 1) return 1;
    Abc_Start();
    struct timespec start, stop;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int r = 0; r < repeat; r++) {
        Abc_Ntk_t *net = Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1);
        if (chain) {
            Abc_Obj_t *root = Abc_NtkCreatePi(net);
            for (int i = 0; i < 1024; ++i)
                root = Abc_AigAnd(net->pManFunc, root, Abc_NtkCreatePi(net));
            Abc_AigDeleteNode(net->pManFunc, root);
            if (Abc_NtkNodeNum(net) != 0 || Abc_NtkObjNum(net) != 1026) return 4;
            Abc_NtkDelete(net);
            continue;
        }
        for (int i = 0; i < count; i++) Abc_NtkCreateNode(net);
        for (int id = 1; id < count; id += 2)
            Abc_NtkDeleteObj(Abc_NtkObj(net, id));
        Abc_Obj_t *node;
        int i, live = 0;
        Abc_NtkForEachNode(net, node, i) {
            if (Abc_ObjId(node) != 2*(live+1)) return 2;
            live++;
        }
        if (live != count/2 || Abc_NtkNodeNum(net) != live) return 3;
        Abc_NtkDelete(net);
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    printf("%.2f process us/network, %d surviving nodes\n",
           ((stop.tv_sec-start.tv_sec)*1e6 + (stop.tv_nsec-start.tv_nsec)*1e-3)/repeat,
           chain ? 0 : count/2);
    Abc_Stop();
    return 0;
}

/* Compare four-leaf graph construction and verify native level truncation. */
#include "base/main/main.h"
#include "base/abc/abc.h"
#include "bool/dec/dec.h"
#include "bool/kit/kit.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

extern int Dec_GraphToNetworkCount(Abc_Obj_t *, Dec_Graph_t *, int, int);

static int check_levels(unsigned input, int limit, int count,
                        unsigned leaf, unsigned first, unsigned root)
{
    Abc_Ntk_t *net = Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1);
    Abc_Obj_t *a = Abc_NtkCreatePi(net), *b = Abc_NtkCreatePi(net);
    Abc_Obj_t *old = Abc_NtkCreatePi(net);
    Dec_Graph_t *g = Dec_GraphCreate(2);
    Dec_Edge_t gate = Dec_GraphAddNodeAnd(g, Dec_EdgeCreate(0, 0), Dec_EdgeCreate(1, 0));
    Dec_GraphSetRoot(g, Dec_GraphAddNodeAnd(g, gate, Dec_EdgeCreate(1, 0)));
    a->Level = input;
    g->pNodes[0].pFunc = a;
    g->pNodes[1].pFunc = b;
    int got = Dec_GraphToNetworkCount(old, g, 2, limit);
    int ok = got == count && g->pNodes[0].Level == leaf &&
             g->pNodes[2].Level == first && g->pNodes[3].Level == root;
    Dec_GraphFree(g);
    Abc_NtkDelete(net);
    return ok;
}

int main(int argc, char **argv)
{
    int repeat = argc > 1 ? atoi(argv[1]) : 1000000;
    int gates = argc > 2 ? atoi(argv[2]) : 7;
    if (repeat < 1 || gates < 1) return 1;
    Abc_Start();
    if (!check_levels(16384, 2, 2, 0, 1, 2) ||
        !check_levels(16383, 16384, 2, 16383, 0, 1) ||
        !check_levels(16383, 16383, -1, 16383, 0, 0))
        return 2;
    if (Kit_EdgeToInt(Kit_EdgeCreate((1<<30)+7, 2)) != 14 ||
        Kit_EdgeToInt(Kit_EdgeCreate(-1, -1)) != 0x7fffffffU ||
        Kit_EdgeToInt(Kit_IntToEdge(0x8000000fU)) != 15)
        return 4;
    for (int kit = 0; kit < 2; kit++) {
        struct timespec start, stop;
        unsigned long checksum = 0;
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
        for (int k = 0; k < repeat; k++) {
            if (kit) {
                Kit_Graph_t *g = Kit_GraphCreate(4);
                Kit_Edge_t edge = Kit_EdgeCreate(0, 0);
                for (int i = 0; i < gates; i++)
                    edge = Kit_GraphAddNodeAnd(g, edge, Kit_EdgeCreate(1+i%3, i%2));
                Kit_GraphSetRoot(g, edge);
                checksum += g->eRoot.Node;
                Kit_GraphFree(g);
            } else {
                Dec_Graph_t *g = Dec_GraphCreate(4);
                Dec_Edge_t edge = Dec_EdgeCreate(0, 0);
                for (int i = 0; i < gates; i++)
                    edge = Dec_GraphAddNodeAnd(g, edge, Dec_EdgeCreate(1+i%3, i%2));
                Dec_GraphSetRoot(g, edge);
                checksum += g->eRoot.Node;
                Dec_GraphFree(g);
            }
        }
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
        if (checksum != (unsigned long)repeat*(3+gates)) return 3;
        printf("%s %d gates %.2f process ns/graph\n", kit ? "KIT" : "DEC", gates,
               ((stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec)/repeat);
    }
    Abc_Stop();
    return 0;
}

/* Compare four-leaf graph construction and verify native level truncation. */
#include "base/main/main.h"
#include "base/abc/abc.h"
#include "bool/dec/dec.h"
#include "bool/kit/kit.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
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

static int bench_leaf_depth(int repeat, int gates, int diamonds)
{
    Kit_Graph_t *graph = Kit_GraphCreate(diamonds ? 3 : 2);
    Kit_Edge_t edge = Kit_EdgeCreate(0, 0);
    for (int i = 0; i < gates; i++) {
        if (diamonds) {
            Kit_Edge_t left = Kit_GraphAddNodeAnd(graph, edge, Kit_EdgeCreate(1, 0));
            Kit_Edge_t right = Kit_GraphAddNodeAnd(graph, edge, Kit_EdgeCreate(2, 1));
            edge = Kit_GraphAddNodeAnd(graph, left, right);
        } else {
            edge = Kit_GraphAddNodeAnd(graph, edge, Kit_EdgeCreate(1, 0));
        }
    }
    Kit_Node_t *leaf = Kit_GraphNode(graph, 0), *root = Kit_GraphNode(graph, edge.Node);
    struct timespec start, stop;
    long long total = 0;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < repeat; i++)
        total += Kit_GraphLeafDepth_rec(graph, root, leaf);
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    Kit_GraphFree(graph);
    if (total != (long long)gates * repeat * (diamonds ? 2 : 1))
        return 7;
    printf("KIT leaf depth %d %s %.2f process ns/traversal\n", gates, diamonds ? "diamonds" : "gates",
           ((stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec)/repeat);
    return 0;
}

int main(int argc, char **argv)
{
    int repeat = argc > 1 ? atoi(argv[1]) : 1000000;
    int gates = argc > 2 ? atoi(argv[2]) : 7;
    if (repeat < 1 || gates < 1) return 1;
    if (argc > 3 && !strcmp(argv[3], "depth"))
        return bench_leaf_depth(repeat, gates, 0);
    if (argc > 3 && !strcmp(argv[3], "diamonds"))
        return bench_leaf_depth(repeat, gates, 1);
    Abc_Start();
    if (!check_levels(16384, 2, 2, 0, 1, 2) ||
        !check_levels(16383, 16384, 2, 16383, 0, 1) ||
        !check_levels(16383, 16383, -1, 16383, 0, 0))
        return 2;
    if (Kit_EdgeToInt(Kit_EdgeCreate((1<<30)+7, 2)) != 14 ||
        Kit_EdgeToInt(Kit_EdgeCreate(-1, -1)) != 0x7fffffffU ||
        Kit_EdgeToInt(Kit_IntToEdge(0x8000000fU)) != 15)
        return 4;
    if (Dec_EdgeToInt(Dec_EdgeCreate((1<<30)+7, 2)) != 14 ||
        Dec_EdgeToInt(Dec_EdgeCreate(-1, -1)) != 0x7fffffffU ||
        Dec_EdgeToInt(Dec_IntToEdge(0x8000000fU)) != 15)
        return 5;
    Kit_Graph_t *zero = Kit_GraphCreateConst0(), *one = Kit_GraphCreateConst1();
    if (zero->pNodes || one->pNodes || !Kit_GraphIsConst0(zero) || !Kit_GraphIsConst1(one))
        return 6;
    Kit_GraphFree(zero);
    Kit_GraphFree(one);
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

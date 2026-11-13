#include <stdio.h>

#include "base/main/main.h"
#include "aig/aig/aig.h"
#include "opt/dar/dar.h"

static void dump(Aig_Man_t *graph, const char *label)
{
    Aig_Obj_t *node;
    int i;
    printf("%s ci=%d nodes=%d co=%d\n", label, Aig_ManCiNum(graph),
           Aig_ManNodeNum(graph), Aig_ManCoNum(graph));
    printf("const=%d\n", Aig_ManConst1(graph)->fPhase);
    Aig_ManForEachCi(graph, node, i)
        printf("i%d=%d\n", node->Id, node->fPhase);
    Aig_ManForEachNode(graph, node, i)
        printf("n%d=%d,%d,%d\n", node->Id,
               2 * Aig_ObjFaninId0(node) + Aig_ObjFaninC0(node),
               2 * Aig_ObjFaninId1(node) + Aig_ObjFaninC1(node), node->fPhase);
    Aig_ManForEachCo(graph, node, i)
        printf("o%d=%d,%d\n", node->Id,
               2 * Aig_ObjFaninId0(node) + Aig_ObjFaninC0(node), node->fPhase);
}

static Aig_Man_t *duplicate_and_stop(Aig_Man_t *graph)
{
    Aig_Man_t *result = Aig_ManDupDfs(graph);
    Aig_ManStop(graph);
    return result;
}

int main(void)
{
    Aig_Man_t *graph, *next;
    Aig_Obj_t *inputs[8], *outputs[4], *parity, *left, *right;
    Dar_RwrPar_t rewrite;
    Dar_RefPar_t refactor;
    int i;
    Abc_Start();
    graph = Aig_ManStart(500);
    for (i = 0; i < 8; i++)
        inputs[i] = Aig_ObjCreateCi(graph);
    parity = inputs[0];
    for (i = 1; i <= 24; i++) {
        // Separate statements give both implementations the same input graph.
        left = Aig_And(graph, parity, Aig_Not(inputs[i % 8]));
        right = Aig_And(graph, Aig_Not(parity), inputs[i % 8]);
        parity = Aig_Or(graph, left, right);
        if (i % 6 == 0)
            outputs[i / 6 - 1] = parity;
    }
    for (i = 0; i < 4; i++)
        Aig_ObjCreateCo(graph, outputs[i]);
    dump(graph, "initial");
    graph = duplicate_and_stop(graph);
    dump(graph, "dup");
    next = Dar_ManBalance(graph, 0);
    Aig_ManStop(graph);
    graph = next;
    dump(graph, "balance1");
    Dar_ManDefaultRwrParams(&rewrite);
    rewrite.fUpdateLevel = 0;
    Dar_ManRewrite(graph, &rewrite);
    graph = duplicate_and_stop(graph);
    dump(graph, "rewrite1");
    Dar_ManDefaultRefParams(&refactor);
    refactor.fUpdateLevel = 0;
    Dar_ManRefactor(graph, &refactor);
    graph = duplicate_and_stop(graph);
    dump(graph, "refactor");
    next = Dar_ManBalance(graph, 0);
    Aig_ManStop(graph);
    graph = next;
    dump(graph, "balance2");
    Dar_ManRewrite(graph, &rewrite);
    graph = duplicate_and_stop(graph);
    dump(graph, "rewrite2");
    Aig_ManStop(graph);
    Abc_Stop();
    return 0;
}

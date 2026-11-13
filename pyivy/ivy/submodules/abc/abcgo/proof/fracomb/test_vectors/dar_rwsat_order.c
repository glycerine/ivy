#include <stdio.h>

#include "base/main/main.h"
#include "aig/aig/aig.h"
#include "opt/dar/dar.h"

static void dump(Aig_Man_t *graph, const char *label)
{
    Aig_Obj_t *node;
    int i;
    printf("%s in=%d and=%d out=%d\n", label, Aig_ManCiNum(graph),
           Aig_ManNodeNum(graph), Aig_ManCoNum(graph));
    Aig_ManForEachNode(graph, node, i)
        printf("n%d=%d,%d\n", node->Id,
               2 * Aig_ObjFaninId0(node) + Aig_ObjFaninC0(node),
               2 * Aig_ObjFaninId1(node) + Aig_ObjFaninC1(node));
    Aig_ManForEachCo(graph, node, i)
        printf("o%d=%d\n", node->Id,
               2 * Aig_ObjFaninId0(node) + Aig_ObjFaninC0(node));
}

int main(void)
{
    Aig_Man_t *graph, *result;
    Aig_Obj_t *a, *b, *c, *x0, *x1, *x, *y0, *y1;
    Abc_Start();
    graph = Aig_ManStart(50);
    a = Aig_ObjCreateCi(graph);
    b = Aig_ObjCreateCi(graph);
    c = Aig_ObjCreateCi(graph);
    // Separate statements fix the construction order that nested C calls
    // otherwise leave unspecified.
    x0 = Aig_And(graph, a, Aig_Not(b));
    x1 = Aig_And(graph, Aig_Not(a), b);
    x = Aig_Or(graph, x0, x1);
    y0 = Aig_And(graph, x, Aig_Not(c));
    y1 = Aig_And(graph, Aig_Not(x), c);
    Aig_ObjCreateCo(graph, Aig_Or(graph, y0, y1));
    dump(graph, "initial");
    result = Dar_ManRwsat(graph, 0, 0);
    dump(result, "unbalanced");
    Aig_ManStop(result);
    result = Dar_ManRwsat(graph, 1, 0);
    dump(result, "balanced");
    Aig_ManStop(result);
    Aig_ManStop(graph);

    graph = Aig_ManStart(20);
    a = Aig_ObjCreateCi(graph);
    b = Aig_ObjCreateCi(graph);
    x0 = Aig_And(graph, a, b);
    x1 = Aig_And(graph, a, Aig_Not(b));
    Aig_ObjCreateCo(graph, Aig_Or(graph, x0, x1));
    dump(graph, "factor_initial");
    result = Dar_ManRwsat(graph, 1, 0);
    dump(result, "factor_balanced");
    Aig_ManStop(result);
    Aig_ManStop(graph);
    Abc_Stop();
    return 0;
}

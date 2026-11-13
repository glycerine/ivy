#include "bool/kit/kit.h"
#include <stdio.h>
int main(void) {
    unsigned truth;
    Vec_Int_t *memory = Vec_IntAlloc(1 << 12);
    while (scanf("%x", &truth) == 1) {
        int phase;
        for (phase = 0; phase < 2; phase++) {
            unsigned t = phase ? ~truth : truth;
            int r = Kit_TruthIsop(&t, 5, memory, 0);
            if (r < 0) { printf("-1 -1\n"); continue; }
            Kit_Graph_t *g = Kit_SopFactor(memory, phase, 5, memory);
            printf("%d %u", Kit_GraphNodeNum(g), Kit_EdgeToInt(g->eRoot));
            Kit_Node_t *n; int i;
            Kit_GraphForEachNode(g, n, i)
                printf(" %u %u", Kit_EdgeToInt(n->eEdge0), Kit_EdgeToInt(n->eEdge1));
            printf("\n");
            Kit_GraphFree(g);
        }
    }
    Vec_IntFree(memory);
    return 0;
}

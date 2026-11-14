#include "aig/aig/aigFact.c"

int main(void)
{
    Aig_Man_t p;
    Aig_Obj_t o[6], *item;
    Vec_Ptr_t *nodes, *support;
    int i;
    memset(&p, 0, sizeof(p)); memset(o, 0, sizeof(o));
    p.nTravIds = 7;
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 3; i++) o[i].Type = AIG_OBJ_CI;
    o[4].Type = o[5].Type = AIG_OBJ_AND;
    for (i = 0; i < 6; i++) o[i].Id = i;
    o[4].pFanin0 = o+1; o[4].pFanin1 = o+2;
    o[5].pFanin0 = o+3; o[5].pFanin1 = o+4;
    o[1].TravId = o[4].TravId = o[5].TravId = 7;
    nodes = Vec_PtrAlloc(2);
    Vec_PtrPush(nodes, o+4); Vec_PtrPush(nodes, o+5);
    support = Aig_SuppMinCollectSupport(&p, nodes);
    printf("support:");
    Vec_PtrForEachEntry(Aig_Obj_t *, support, item, i) printf(" %d", item->Id);
    printf("\nmarks:");
    for (i = 0; i < 6; i++) printf(" %d", o[i].TravId);
    puts("");
    Vec_PtrFree(nodes); Vec_PtrFree(support);
    return 0;
}

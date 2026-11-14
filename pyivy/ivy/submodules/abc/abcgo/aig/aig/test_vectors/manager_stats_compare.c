#include "aig/aig/aig.h"

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t objects[9];
    int i;
    memset(&g, 0, sizeof(g));
    memset(objects, 0, sizeof(objects));
    g.pName = "stats";
    g.pConst1 = objects;
    g.vObjs = Vec_PtrAlloc(9);
    g.vCis = Vec_PtrAlloc(2);
    g.vCos = Vec_PtrAlloc(1);
    objects[0].Type = AIG_OBJ_CONST1;
    objects[1].Type = objects[2].Type = AIG_OBJ_CI;
    objects[3].Type = AIG_OBJ_AND;
    objects[3].pFanin0 = objects + 1;
    objects[3].pFanin1 = objects + 2;
    objects[3].Level = 1;
    objects[3].nRefs = 1;
    objects[4].Type = AIG_OBJ_CO;
    objects[4].pFanin0 = objects + 3;
    objects[4].Level = 99;
    Vec_PtrPush(g.vCis, objects + 1);
    Vec_PtrPush(g.vCis, objects + 2);
    Vec_PtrPush(g.vCos, objects + 4);
    for (i = 0; i < 5; i++)
    {
        objects[i].Id = i;
        g.nObjs[objects[i].Type]++;
        Vec_PtrPush(g.vObjs, objects + i);
    }
    printf("clean %d\n", Aig_ManHasNoGaps(&g));
    Aig_ManPrintStats(&g);
    Vec_PtrPush(g.vObjs, NULL);
    g.nDeleted = 1;
    printf("deleted %d\n", Aig_ManHasNoGaps(&g));
    Aig_ManPrintStats(&g);
    g.nRegs = 1;
    Aig_ManPrintStats(&g);
    objects[6].Type = AIG_OBJ_EXOR;
    objects[6].pFanin0 = objects + 1;
    objects[6].pFanin1 = objects + 2;
    objects[7].Type = AIG_OBJ_BUF;
    objects[7].pFanin0 = objects + 6;
    objects[8].Type = AIG_OBJ_AND;
    objects[8].pFanin0 = objects + 1;
    objects[8].pFanin1 = objects + 3;
    objects[3].nRefs++;
    g.pEquivs = ABC_CALLOC(Aig_Obj_t *, 9);
    g.pEquivs[3] = objects + 8;
    for (i = 6; i < 9; i++)
    {
        objects[i].Id = i;
        g.nObjs[objects[i].Type]++;
        Vec_PtrPush(g.vObjs, objects + i);
    }
    printf("typed %d\n", Aig_ManHasNoGaps(&g));
    Aig_ManPrintStats(&g);
    ABC_FREE(g.pEquivs);
    Vec_PtrFree(g.vObjs);
    Vec_PtrFree(g.vCis);
    Vec_PtrFree(g.vCos);
    return 0;
}

#include "aig/aig/aig.h"

/* Build only the graph representation here. The actual writers and their DFS
   are linked directly from production C, without substitute algorithms. */
int main(int argc, char **argv)
{
    int nci, nnodes, nco, nregs, total, i, a, b, count = 0;
    char filename[4096];
    assert(argc == 2);
    while (scanf("%d%d%d%d", &nci, &nnodes, &nco, &nregs) == 4)
    {
        Aig_Man_t g;
        Aig_Obj_t *objects;
        memset(&g, 0, sizeof(g));
        total = 1 + nci + nnodes + nco;
        objects = ABC_CALLOC(Aig_Obj_t, total);
        g.pConst1 = objects;
        g.pName = "dump";
        g.nTravIds = 1;
        g.nRegs = nregs;
        g.nTruePis = nci - nregs;
        g.nTruePos = nco - nregs;
        g.vObjs = Vec_PtrAlloc(total);
        g.vCis = Vec_PtrAlloc(nci);
        g.vCos = Vec_PtrAlloc(nco);
        objects[0].Type = AIG_OBJ_CONST1;
        for (i = 1; i <= nci; i++)
        {
            objects[i].Type = AIG_OBJ_CI;
            Vec_PtrPush(g.vCis, objects + i);
        }
        for (; i <= nci + nnodes; i++)
        {
            assert(scanf("%d%d", &a, &b) == 2);
            objects[i].Type = AIG_OBJ_AND;
            objects[i].pFanin0 = Aig_NotCond(objects + a/2, a & 1);
            objects[i].pFanin1 = Aig_NotCond(objects + b/2, b & 1);
        }
        for (; i < total; i++)
        {
            assert(scanf("%d", &a) == 1);
            objects[i].Type = AIG_OBJ_CO;
            objects[i].pFanin0 = Aig_NotCond(objects + a/2, a & 1);
            Vec_PtrPush(g.vCos, objects + i);
        }
        for (i = 0; i < total; i++)
        {
            objects[i].Id = i;
            g.nObjs[objects[i].Type]++;
            Vec_PtrPush(g.vObjs, objects + i);
            if (objects[i].pFanin0) Aig_ObjFanin0(objects + i)->nRefs++;
            if (objects[i].pFanin1) Aig_ObjFanin1(objects + i)->nRefs++;
        }
        assert(snprintf(filename, sizeof(filename), "%s/%d.blif", argv[1], count) < sizeof(filename));
        Aig_ManDumpBlif(&g, filename, NULL, NULL);
        assert(snprintf(filename, sizeof(filename), "%s/%d.v", argv[1], count) < sizeof(filename));
        Aig_ManDumpVerilog(&g, filename);
        Vec_PtrFree(g.vObjs);
        Vec_PtrFree(g.vCis);
        Vec_PtrFree(g.vCos);
        ABC_FREE(objects);
        count++;
    }
    printf("dumped %d\n", count);
    return 0;
}

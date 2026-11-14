#include "aig/aig/aigRetF.c"

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t o[11];
    int mode, pass, i;
    for (mode = 0; mode < 3; mode++)
    {
        memset(&g, 0, sizeof(g)); memset(o, 0, sizeof(o));
        g.nTravIds = 1; g.nRegs = mode == 2 ? 4 : 3;
        g.nTruePis = mode == 2 ? 0 : 1; g.nTruePos = mode == 2 ? 0 : 1;
        g.pConst1 = o;
        g.vObjs = Vec_PtrAlloc(11); g.vCis = Vec_PtrAlloc(4); g.vCos = Vec_PtrAlloc(4);
        o[0].Type = AIG_OBJ_CONST1;
        for (i = 1; i <= 4; i++) { o[i].Type = AIG_OBJ_CI; Vec_PtrPush(g.vCis, o+i); }
        o[5].Type = o[6].Type = AIG_OBJ_AND;
        o[5].pFanin0 = o+2; o[5].pFanin1 = o+3;
        o[6].pFanin0 = o+3; o[6].pFanin1 = o+4;
        for (i = 7; i <= 10; i++) { o[i].Type = AIG_OBJ_CO; Vec_PtrPush(g.vCos, o+i); }
        o[7].pFanin0 = o+5;
        o[8].pFanin0 = mode == 0 ? o+1 : o+3;
        o[9].pFanin0 = o+4;
        o[10].pFanin0 = mode == 0 ? Aig_Not(o+3) : o+1;
        for (i = 0; i < 11; i++) { o[i].Id = i; g.nObjs[o[i].Type]++; Vec_PtrPush(g.vObjs, o+i); }
        for (pass = 0; pass < 2; pass++)
        {
            Aig_ManRetimeMark(&g);
            printf("mode %d pass %d generation %d:", mode, pass, g.nTravIds);
            for (i = 0; i < 11; i++) printf(" %d/%d", o[i].fMarkB, o[i].TravId);
            puts("");
            for (i = 1; i <= 4; i++) assert(o[i].pNext == NULL);
            for (i = 7; i <= 10; i++) assert(o[i].pNext == NULL);
            for (i = 0; i < 11; i++) o[i].fMarkB = 0;
        }
        Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos);
    }
    return 0;
}

#include "aig/aig/aigDfs.c"

int main(void)
{
    Aig_Man_t p;
    Aig_Obj_t o[9];
    int i, choice;
    memset(&p, 0, sizeof(p));
    memset(o, 0, sizeof(o));
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 4; i++) o[i].Type = AIG_OBJ_CI;
    o[5].Type = AIG_OBJ_EXOR;
    o[6].Type = o[7].Type = AIG_OBJ_AND;
    o[8].Type = AIG_OBJ_CO;
    o[5].pFanin0 = &o[1]; o[5].pFanin1 = &o[2];
    o[6].pFanin0 = &o[3]; o[6].pFanin1 = &o[5];
    o[7].pFanin0 = &o[3]; o[7].pFanin1 = &o[4];
    o[8].pFanin0 = &o[6];
    o[5].Level = Aig_ObjLevelNew(&o[5]);
    o[6].Level = Aig_ObjLevelNew(&o[6]);
    o[7].Level = Aig_ObjLevelNew(&o[7]);
    printf("structural %u %u %d\n", o[5].Level, o[6].Level, Aig_ObjLevelNew(&o[5]));
    p.pConst1 = o;
    p.vObjs = Vec_PtrAlloc(9);
    p.vCis = Vec_PtrAlloc(4);
    p.vCos = Vec_PtrAlloc(1);
    for (i = 0; i < 9; i++)
    {
        o[i].Id = i;
        Vec_PtrPush(p.vObjs, &o[i]);
        if (o[i].Type == AIG_OBJ_CI) Vec_PtrPush(p.vCis, &o[i]);
    }
    Vec_PtrPush(p.vCos, &o[8]);
    choice = Aig_ManChoiceLevel(&p);
    printf("choice %d %u %u %u\n", choice, o[5].Level, o[6].Level, o[7].Level);
    o[1].Level = 4;
    o[5].Level = Aig_ObjLevelNew(&o[5]);
    o[6].Level = Aig_ObjLevelNew(&o[6]);
    printf("arrival %u %u\n", o[5].Level, o[6].Level);
    Vec_PtrFree(p.vObjs);
    Vec_PtrFree(p.vCis);
    Vec_PtrFree(p.vCos);
    return 0;
}

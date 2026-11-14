#include "aig/aig/aigTiming.c"

static void prepare(Aig_Man_t *p, Aig_Obj_t *objects, int count)
{
    int i;
    memset(p, 0, sizeof(*p));
    p->vObjs = Vec_PtrAlloc(count);
    p->vCos = Vec_PtrAlloc(3);
    p->pConst1 = objects;
    for (i = 0; i < count; i++)
    {
        objects[i].Id = i;
        p->nObjs[objects[i].Type]++;
        Vec_PtrPush(p->vObjs, &objects[i]);
        if (objects[i].Type == AIG_OBJ_CO)
            Vec_PtrPush(p->vCos, &objects[i]);
        if (objects[i].pFanin0)
            Aig_ObjFanin0(&objects[i])->nRefs++;
        if (objects[i].pFanin1)
            Aig_ObjFanin1(&objects[i])->nRefs++;
    }
    Aig_ManFanoutStart(p);
}

static void release(Aig_Man_t *p)
{
    Aig_ManFanoutStop(p);
    Vec_PtrFree(p->vObjs);
    Vec_PtrFree(p->vCos);
    Vec_VecFree(p->vLevels);
    if (p->vLevelR)
        Vec_IntFree(p->vLevelR);
}

int main(void)
{
    Aig_Man_t manager;
    Aig_Obj_t o[12];
    int mode, i;
    for (mode = 0; mode < 3; mode++)
    {
        memset(o, 0, sizeof(o));
        o[0].Type = AIG_OBJ_CONST1;
        for (i = 1; i <= 4; i++) o[i].Type = AIG_OBJ_CI;
        for (i = 5; i <= 7; i++) o[i].Type = AIG_OBJ_AND;
        o[8].Type = AIG_OBJ_CO;
        o[5].pFanin0 = &o[1]; o[5].pFanin1 = &o[2];
        o[6].pFanin0 = &o[3]; o[6].pFanin1 = &o[5];
        o[7].pFanin0 = &o[2]; o[7].pFanin1 = &o[4];
        o[8].pFanin0 = &o[6];
        o[5].Level = mode == 2 ? 5 : 1;
        o[6].Level = o[8].Level = o[5].Level + 1;
        o[7].Level = mode == 0 ? 17 : 1;
        o[6].fMarkA = mode == 0;
        prepare(&manager, o, 9);
        o[1].Level = mode == 2 ? 0 : 4;
        Aig_ManUpdateLevel(&manager, &o[5]);
        printf("forward%d %u %u %u %u %u\n", mode, o[5].Level,
            o[6].Level, o[7].Level, o[6].fMarkA, o[8].Level);
        release(&manager);
    }
    memset(o, 0, sizeof(o));
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 4; i++) o[i].Type = AIG_OBJ_CI;
    for (i = 5; i <= 8; i++) o[i].Type = AIG_OBJ_AND;
    for (i = 9; i <= 11; i++) o[i].Type = AIG_OBJ_CO;
    o[5].pFanin0 = &o[1]; o[5].pFanin1 = &o[2];
    o[6].pFanin0 = &o[3]; o[6].pFanin1 = &o[4];
    o[7].pFanin0 = &o[3]; o[7].pFanin1 = &o[5];
    o[8].pFanin0 = &o[4]; o[8].pFanin1 = &o[7];
    o[9].pFanin0 = &o[5]; o[10].pFanin0 = &o[6]; o[11].pFanin0 = &o[8];
    o[5].Level = o[6].Level = 1; o[7].Level = 2; o[8].Level = 3;
    prepare(&manager, o, 12);
    manager.vLevelR = Vec_IntStart(12);
    Aig_ObjSetReverseLevel(&manager, &o[5], 1);
    Aig_ObjSetReverseLevel(&manager, &o[6], 13);
    Aig_ManUpdateReverseLevel(&manager, &o[8]);
    printf("reverse %d %d %d %d %u%u%u\n", Aig_ObjReverseLevel(&manager, &o[8]),
        Aig_ObjReverseLevel(&manager, &o[7]), Aig_ObjReverseLevel(&manager, &o[5]),
        Aig_ObjReverseLevel(&manager, &o[6]), o[5].fMarkA, o[7].fMarkA, o[8].fMarkA);
    release(&manager);
    return 0;
}

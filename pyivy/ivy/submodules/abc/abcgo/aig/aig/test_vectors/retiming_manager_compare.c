#include "aig/aig/aigRet.c"

unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t o[9];
    Rtm_Man_t *p;
    Rtm_Obj_t *obj;
    Rtm_Edg_t *edge;
    int i, k, j;
    memset(&g, 0, sizeof(g)); memset(o, 0, sizeof(o));
    g.pConst1 = o; g.nRegs = 2; g.nTruePis = 2; g.nTruePos = 1;
    g.vObjs = Vec_PtrAlloc(9); g.vCis = Vec_PtrAlloc(4); g.vCos = Vec_PtrAlloc(3);
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 4; i++) { o[i].Type = AIG_OBJ_CI; Vec_PtrPush(g.vCis, o+i); }
    o[5].Type = AIG_OBJ_AND; o[5].pFanin0 = Aig_Not(o+3); o[5].pFanin1 = Aig_Not(o+4);
    for (i = 6; i <= 8; i++) { o[i].Type = AIG_OBJ_CO; Vec_PtrPush(g.vCos, o+i); }
    o[6].pFanin0 = o+5; o[7].pFanin0 = o+1; o[8].pFanin0 = o+2;
    for (i = 0; i < 9; i++)
    {
        o[i].Id = i; g.nObjs[o[i].Type]++; Vec_PtrPush(g.vObjs, o+i);
        if (o[i].pFanin0) Aig_ObjFanin0(o+i)->nRefs++;
        if (o[i].pFanin1) Aig_ObjFanin1(o+i)->nRefs++;
    }
    p = Rtm_ManFromAig(&g);
    printf("mapping:");
    for (i = 0; i < 9; i++) printf(" %d", ((Rtm_Obj_t *)o[i].pData)->Id);
    puts("");
    Rtm_ManForEachObj(p, obj, i)
    {
        printf("%d type %d capacity %d/%d fanins:", obj->Id, obj->Type, obj->Num, obj->Temp);
        for (k = 0; k < (int)obj->nFanins; k++) printf(" %d/%d", Rtm_ObjFanin(obj,k)->Id, k ? obj->fCompl1 : obj->fCompl0);
        printf(" fanouts:");
        for (k = 0; k < (int)obj->nFanouts; k++) printf(" %d", Rtm_ObjFanout(obj,k)->Id);
        puts(""); obj->Num = 0;
    }
    printf("autonomous %d ", Rtm_ManMarkAutoFwd(p)); printf("%d\n", Rtm_ManMarkAutoBwd(p));
    for (i = 3; i <= 4; i++) Rtm_ObjAddFirst(p,Rtm_ObjEdge((Rtm_Obj_t *)o[i].pData,0),RTM_VAL_ZERO);
    for (i = 3; i <= 5; i++) Rtm_ObjRetimeFwd(p,(Rtm_Obj_t *)o[i].pData);
    printf("max %d\n", Rtm_ManLatchMax(p));
    Rtm_ManForEachObj(p, obj, i)
    Rtm_ObjForEachFaninEdge(obj, edge, k)
    {
        printf("edge %d/%d %d:",obj->Id,k,edge->nLats);
        for (j = 0; j < (int)edge->nLats; j++) printf(" %d",Rtm_ObjGetOne(p,edge,j));
        puts("");
    }
    Rtm_ManFree(p);
    Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos);
    return 0;
}

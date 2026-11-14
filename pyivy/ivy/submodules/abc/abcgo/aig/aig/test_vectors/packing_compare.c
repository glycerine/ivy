#include "aig/aig/aigPack.c"

static void snapshot(Aig_ManPack_t *p, int result)
{
    Vec_Int_t *nodes;
    int i;
    Aig_ManPackSimulate(p);
    printf("%d %d %d %d:", result, Aig_ManPackCountCares(p), p->nPatTotal, p->nPatSkip);
    for (i = 0; i < Vec_WrdSize(p->vPiPats); i++)
        printf(" %016llx/%016llx", (unsigned long long)Vec_WrdEntry(p->vPiPats,i), (unsigned long long)Vec_WrdEntry(p->vPiCare,i));
    printf(" | ");
    for (i = 0; i < Vec_WrdSize(p->vSigns); i++) printf("%016llx ", (unsigned long long)Vec_WrdEntry(p->vSigns,i));
    printf("| ");
    nodes = Aig_ManPackConstNodes(p);
    for (i = 0; i < Vec_IntSize(nodes); i++) printf("%d ", Vec_IntEntry(nodes,i));
    puts(""); Vec_IntFree(nodes);
}

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t o[19];
    Aig_ManPack_t *p;
    Vec_Int_t *lits;
    int i, action, lane, count, literal, result;
    memset(&g,0,sizeof(g)); memset(o,0,sizeof(o));
    g.pConst1 = o;
    g.vObjs = Vec_PtrAlloc(19); g.vCis = Vec_PtrAlloc(7); g.vCos = Vec_PtrAlloc(4);
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 7; i++) { o[i].Type = AIG_OBJ_CI; Vec_PtrPush(g.vCis,o+i); }
    for (i = 8; i <= 14; i++) o[i].Type = AIG_OBJ_AND;
    o[8].pFanin0=o+1; o[8].pFanin1=o+2;
    o[9].pFanin0=Aig_Not(o+1); o[9].pFanin1=o+2;
    o[10].pFanin0=o+3; o[10].pFanin1=Aig_Not(o+4);
    o[11].pFanin0=Aig_Not(o+5); o[11].pFanin1=Aig_Not(o+6);
    o[12].pFanin0=o+8; o[12].pFanin1=Aig_Not(o+9);
    o[13].pFanin0=Aig_Not(o+10); o[13].pFanin1=Aig_Not(o+11);
    o[14].pFanin0=o+12; o[14].pFanin1=o+13;
    for (i=15; i<19; i++) { o[i].Type=AIG_OBJ_CO; Vec_PtrPush(g.vCos,o+i); }
    o[15].pFanin0=o+14; o[16].pFanin0=Aig_Not(o+8);
    o[17].pFanin0=Aig_Not(o); o[18].pFanin0=o;
    for (i=0; i<19; i++) { o[i].Id=i; g.nObjs[o[i].Type]++; Vec_PtrPush(g.vObjs,o+i); }
    p=Aig_ManPackAlloc(&g); lits=Vec_IntAlloc(16);
    Aig_ManRandom(1); Aig_ManPackSetRandom(p); snapshot(p,-1);
    while (scanf("%d%d%d",&action,&lane,&count)==3)
    {
        Vec_IntClear(lits);
        for (i=0; i<count; i++) { assert(scanf("%d",&literal)==1); Vec_IntPush(lits,literal); }
        result=-1;
        if (action==0) result=Aig_ManPackAddPatternTry(p,lane,lits);
        else if (action==1) Aig_ManPackAddPattern(p,lits);
        else if (action==2) Aig_ManPackSetRandom(p);
        else abort();
        snapshot(p,result);
    }
    puts("care diagnostic:"); Aig_ManPackPrintCare(p);
    puts("statistics:"); Aig_ManPackPrintStats(p);
    Aig_ManPackFree(p);
    Vec_IntFree(lits); Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos);
    return 0;
}

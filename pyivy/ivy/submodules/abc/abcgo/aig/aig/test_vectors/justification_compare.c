#include "aig/aig/aigJust.c"

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t *o;
    Vec_Int_t *support;
    int nci, nnodes, nco, nqueries, total, i, j, a, b, mode, id, value, heur, count, lit, result;
    while (scanf("%d%d%d%d", &nci, &nnodes, &nco, &nqueries) == 4)
    {
        memset(&g, 0, sizeof(g)); g.nTravIds = 1;
        total = 1+nci+nnodes+nco; o = ABC_CALLOC(Aig_Obj_t, total);
        g.pConst1 = o; o[0].Type = AIG_OBJ_CONST1; o[0].fPhase = 1;
        g.vObjs = Vec_PtrAlloc(total); g.vCis = Vec_PtrAlloc(nci); g.vCos = Vec_PtrAlloc(nco);
        for (i = 1; i <= nci; i++) { o[i].Type = AIG_OBJ_CI; o[i].CioId = i-1; Vec_PtrPush(g.vCis, o+i); }
        for (i = 1+nci; i < 1+nci+nnodes; i++)
        {
            assert(scanf("%d%d", &a, &b) == 2);
            o[i].Type = AIG_OBJ_AND;
            o[i].pFanin0 = Aig_NotCond(o+a/2, a&1);
            o[i].pFanin1 = Aig_NotCond(o+b/2, b&1);
        }
        for (i = 1+nci+nnodes; i < total; i++)
        {
            assert(scanf("%d", &a) == 1);
            o[i].Type = AIG_OBJ_CO; o[i].CioId = i-(1+nci+nnodes);
            o[i].pFanin0 = Aig_NotCond(o+a/2, a&1); Vec_PtrPush(g.vCos, o+i);
        }
        for (i = 0; i < total; i++) { o[i].Id=i; g.nObjs[o[i].Type]++; Vec_PtrPush(g.vObjs,o+i); }
        support = Vec_IntAlloc(nci);
        for (j = 0; j < nqueries; j++)
        {
            assert(scanf("%d%d%d%d%d", &mode, &id, &value, &heur, &count) == 5);
            Vec_IntClear(support);
            for (i=0; i<count; i++) { assert(scanf("%d", &lit)==1); Vec_IntPush(support,lit); }
            if (mode == 0) result = Aig_ObjFindSatAssign(&g,o+id,value,support);
            else if (mode == 1)
            {
                Aig_ManIncrementTravId(&g);
                result = Aig_NtkFindSatAssign_rec(&g,o+id,value,support,heur);
            }
            else { assert(mode == 2); result = Aig_ObjTerSimulate(&g,o+id,support); }
            printf("%d %d %d:", mode, result, g.nTravIds);
            for (i=0; i<Vec_IntSize(support); i++) printf(" %d",Vec_IntEntry(support,i));
            printf(" | ");
            for (i=0; i<total; i++) printf("%d/%d ",Aig_ObjGetTerValue(o+i),o[i].TravId);
            puts("");
        }
        Vec_IntFree(support); Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos); free(o);
    }
    return 0;
}

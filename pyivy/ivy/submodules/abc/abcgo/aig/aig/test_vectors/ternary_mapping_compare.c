#include "aig/aig/aigTsim.c"

int main(void)
{
    Aig_Man_t p;
    Aig_Obj_t o[38], *e;
    Vec_Ptr_t *mapping;
    int i;
    memset(&p, 0, sizeof(p)); memset(o, 0, sizeof(o));
    p.nRegs = 18; p.nTruePis = 1; p.nTruePos = 0; p.pConst1 = o;
    p.vObjs = Vec_PtrAlloc(38); p.vCis = Vec_PtrAlloc(19); p.vCos = Vec_PtrAlloc(18);
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 19; i++) { o[i].Type = AIG_OBJ_CI; Vec_PtrPush(p.vCis, o+i); }
    for (i = 0; i < 18; i++)
    {
        e = o+20+i; e->Type = AIG_OBJ_CO;
        switch (i % 4)
        {
        case 0: e->pFanin0 = o+2+i; break;
        case 1: e->pFanin0 = Aig_Not(o+2+i); break;
        case 2: e->pFanin0 = o+1; break;
        case 3: e->pFanin0 = o; break;
        }
        Vec_PtrPush(p.vCos, e);
    }
    for (i = 0; i < 38; i++) { o[i].Id = i; Vec_PtrPush(p.vObjs, o+i); p.nObjs[o[i].Type]++; }
    mapping = Aig_ManTernarySimulate(&p, 0, 0);
    printf("mapping:");
    Vec_PtrForEachEntry(Aig_Obj_t *, mapping, e, i) printf(" %d", 2*Aig_Regular(e)->Id+Aig_IsComplement(e));
    printf("\nvalues:");
    for (i = 0; i < 38; i++) printf(" %d", Aig_ObjGetXsim(o+i));
    puts("");
    Vec_PtrFree(mapping); Vec_PtrFree(p.vObjs); Vec_PtrFree(p.vCis); Vec_PtrFree(p.vCos);
    return 0;
}

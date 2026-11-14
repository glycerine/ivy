#include "aig/aig/aigTsim.c"

int main(void)
{
    int mode, i;
    for (mode = 0; mode < 2; mode++)
    {
        Aig_Man_t p;
        Aig_Obj_t o[6], *e;
        Vec_Ptr_t *mapping;
        memset(&p, 0, sizeof(p)); memset(o, 0, sizeof(o));
        p.pConst1 = o; p.nRegs = mode; p.nTruePis = 2-mode; p.nTruePos = 1;
        p.vObjs = Vec_PtrAlloc(6); p.vCis = Vec_PtrAlloc(2); p.vCos = Vec_PtrAlloc(2);
        o[0].Type = AIG_OBJ_CONST1;
        o[1].Type = o[2].Type = AIG_OBJ_CI;
        o[3].Type = AIG_OBJ_AND; o[3].pFanin0 = o+1; o[3].pFanin1 = o+2;
        o[4].Type = AIG_OBJ_CO; o[4].pFanin0 = o+3; Aig_ObjSetXsim(o+4, AIG_XVS1);
        o[5].Type = AIG_OBJ_CO; o[5].pFanin0 = o+2;
        Vec_PtrPush(p.vCis, o+1); Vec_PtrPush(p.vCis, o+2);
        Vec_PtrPush(p.vCos, o+4); if (mode) Vec_PtrPush(p.vCos, o+5);
        for (i = 0; i < 5+mode; i++) { o[i].Id = i; Vec_PtrPush(p.vObjs, o+i); p.nObjs[o[i].Type]++; }
        mapping = Aig_ManTernarySimulate(&p, 0, 0);
        printf("mode %d mapping:", mode);
        if (mapping) Vec_PtrForEachEntry(Aig_Obj_t *, mapping, e, i) printf(" %d", 2*Aig_Regular(e)->Id+Aig_IsComplement(e));
        printf("\nvalues:");
        for (i = 0; i < 5+mode; i++) printf(" %d", Aig_ObjGetXsim(o+i));
        puts("");
        if (mapping) Vec_PtrFree(mapping);
        Vec_PtrFree(p.vObjs); Vec_PtrFree(p.vCis); Vec_PtrFree(p.vCos);
    }
    return 0;
}

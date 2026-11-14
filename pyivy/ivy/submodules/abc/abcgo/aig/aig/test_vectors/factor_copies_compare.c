#include "aig/aig/aigFact.c"

/* Console environment only; algorithms below are the original C routines. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

static int lit(Aig_Obj_t *o) { return 2*Aig_Regular(o)->Id + Aig_IsComplement(o); }

int main(void)
{
    Aig_Man_t p;
    Aig_Obj_t o[6], *result;
    Vec_Ptr_t *support, *nodes, *cofactors;
    int i;
    memset(&p, 0, sizeof(p)); memset(o, 0, sizeof(o));
    p.nTravIds = 1; p.pConst1 = o;
    p.nTableSize = 17; p.pTable = ABC_CALLOC(Aig_Obj_t *, 17);
    p.vObjs = Vec_PtrAlloc(16); p.pMemObjs = Aig_MmFixedStart(sizeof(Aig_Obj_t), 64);
    o[0].Type = AIG_OBJ_CONST1; o[0].fPhase = 1;
    for (i = 1; i <= 3; i++) o[i].Type = AIG_OBJ_CI;
    o[4].Type = o[5].Type = AIG_OBJ_AND;
    o[4].pFanin0 = o+1; o[4].pFanin1 = o+2;
    o[5].pFanin0 = o+3; o[5].pFanin1 = o+4;
    for (i = 0; i < 6; i++) { o[i].Id = i; p.nObjs[o[i].Type]++; Vec_PtrPush(p.vObjs, o+i); }
    for (i = 4; i < 6; i++) { Aig_ObjFanin0(o+i)->nRefs++; Aig_ObjFanin1(o+i)->nRefs++; }
    Aig_TableInsert(&p, o+4); Aig_TableInsert(&p, o+5);
    result = Aig_ManDeriveNewCone_rec(&p, o+5);
    printf("recursive %d", lit(result));
    for (i = 1; i <= 5; i++) printf(" %d@%d", lit((Aig_Obj_t *)o[i].pData), o[i].TravId);
    puts("");
    support = Vec_PtrAlloc(3); nodes = Vec_PtrAlloc(2); cofactors = Vec_PtrAlloc(1);
    for (i = 1; i <= 3; i++) Vec_PtrPush(support, o+i);
    Vec_PtrPush(nodes, o+4); Vec_PtrPush(nodes, o+5);
    Vec_PtrPush(cofactors, Aig_Not(o+1));
    result = Aig_ManDeriveNewCone(&p, cofactors, o+5);
    printf("derive %d generation %d\n", lit(result), p.nTravIds);
    Vec_PtrClear(cofactors); Vec_PtrPush(cofactors, o+1);
    result = Aig_SuppMinReconstruct(&p, cofactors, nodes, support);
    printf("reconstruct %d", lit(result));
    for (i = 1; i <= 5; i++) printf(" %d", lit((Aig_Obj_t *)o[i].pData));
    puts("");
    Vec_PtrFree(support); Vec_PtrFree(nodes); Vec_PtrFree(cofactors);
    Vec_PtrFree(p.vObjs); free(p.pTable); Aig_MmFixedStop(p.pMemObjs, 0);
    return 0;
}

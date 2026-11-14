#include "aig/saig/saigSimMv.c"

/* Standalone console environment. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t o[14], *obj;
    Vec_Ptr_t *mapping;
    int i;
    memset(&g, 0, sizeof(g)); memset(o, 0, sizeof(o));
    g.pConst1 = o; g.nRegs = 6; g.nTruePis = 0; g.nTruePos = 1;
    g.vObjs = Vec_PtrAlloc(14); g.vCis = Vec_PtrAlloc(6); g.vCos = Vec_PtrAlloc(7);
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 6; i++) { o[i].Type = AIG_OBJ_CI; o[i].CioId = i-1; Vec_PtrPush(g.vCis, o+i); }
    for (i = 7; i < 14; i++) { o[i].Type = AIG_OBJ_CO; o[i].CioId = i-7; Vec_PtrPush(g.vCos, o+i); }
    o[7].pFanin0 = o+6; o[8].pFanin0 = Aig_Not(o+1);
    for (i = 9; i < 14; i++) o[i].pFanin0 = o+i-8;
    for (i = 0; i < 14; i++) { o[i].Id = i; Vec_PtrPush(g.vObjs, o+i); g.nObjs[o[i].Type]++; }
    mapping = Saig_MvManSimulate(&g, 1, 2, 0, 1);
    printf("driver:");
    if (mapping) Vec_PtrForEachEntry(Aig_Obj_t *, mapping, obj, i) printf(" %d", 2*Aig_Regular(obj)->Id+Aig_IsComplement(obj));
    puts("");
    if (mapping) Vec_PtrFree(mapping);
    Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos);
    return 0;
}

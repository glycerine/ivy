#include "aig/saig/saigSimMv.c"

/* Standalone console environment, without an ABC frame or bridge. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

static void print_vec(const char *name, Vec_Int_t *v)
{
    int i, value;
    printf("%s:", name);
    Vec_IntForEachEntry(v, value, i) printf(" %d", value);
    puts("");
}

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t o[16], *obj;
    Saig_MvMan_t *p;
    Vec_Int_t *x, *constants, *binary, *oscillators, *skip;
    Vec_Ptr_t *mapping;
    int i, f, state, ci = 0, co = 0;
    memset(&g, 0, sizeof(g)); memset(o, 0, sizeof(o));
    g.pConst1 = o; g.nRegs = 5; g.nTruePis = 3; g.nTruePos = 1;
    g.vObjs = Vec_PtrAlloc(16); g.vCis = Vec_PtrAlloc(8); g.vCos = Vec_PtrAlloc(6);
    o[0].Type = AIG_OBJ_CONST1;
    for (i = 1; i <= 9; i++) o[i].Type = AIG_OBJ_CI;
    o[3].Type = AIG_OBJ_AND; o[3].pFanin0 = o+1; o[3].pFanin1 = o+2;
    for (i = 10; i <= 15; i++) o[i].Type = AIG_OBJ_CO;
    o[10].pFanin0 = o+3; o[11].pFanin0 = o+5; o[12].pFanin0 = Aig_Not(o+6);
    o[13].pFanin0 = o[14].pFanin0 = o+4; o[15].pFanin0 = o+3;
    for (i = 0; i < 16; i++)
    {
        o[i].Id = i; g.nObjs[o[i].Type]++; Vec_PtrPush(g.vObjs, o+i);
        if (Aig_ObjIsCi(o+i)) { o[i].CioId = ci++; Vec_PtrPush(g.vCis, o+i); }
        if (Aig_ObjIsCo(o+i)) { o[i].CioId = co++; Vec_PtrPush(g.vCos, o+i); }
    }
    p = Saig_MvManStart(&g, 3);
    for (i = 0; i < 5; i++) ((Saig_MvObj_t *)Vec_PtrEntry(p->vFlops, i))->Value = 1;
    printf("initial %d\n", Saig_MvSaveState(p));
    for (f = 0; f < 8; f++)
    {
        Saig_MvSimulateFrame(p, f < 2, 0);
        state = Saig_MvSaveState(p);
        printf("frame %d state %d objects %d:", f, state, p->nObjs);
        for (i = 0; i < 16; i++) printf(" %d", p->pAigOld[i].Value);
        puts("");
        if (state >= 0) break;
    }
    x = Saig_MvManFindXFlops(p); constants = Saig_MvManFindConstBinaryFlops(p, &binary);
    oscillators = Saig_MvManFindOscilators(p, &skip); Vec_IntFree(skip);
    skip = Saig_MvManCreateNextSkip(p);
    print_vec("undefined", x); print_vec("constants", constants); print_vec("binary", binary);
    print_vec("oscillators", oscillators); print_vec("skip", skip);
    mapping = Saig_MvManDeriveMap(p, 0);
    printf("mapping:");
    Vec_PtrForEachEntry(Aig_Obj_t *, mapping, obj, i) printf(" %d",2*Aig_Regular(obj)->Id+Aig_IsComplement(obj));
    puts(""); Vec_PtrFree(mapping);
    Vec_IntFree(x); Vec_IntFree(constants); Vec_IntFree(binary); Vec_IntFree(oscillators); Vec_IntFree(skip);
    Saig_MvManStop(p);
    mapping = Saig_MvManSimulate(&g, 2, 3, 0, 0);
    printf("driver:");
    Vec_PtrForEachEntry(Aig_Obj_t *, mapping, obj, i) printf(" %d",2*Aig_Regular(obj)->Id+Aig_IsComplement(obj));
    puts(""); Vec_PtrFree(mapping);
    Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos);
    return 0;
}

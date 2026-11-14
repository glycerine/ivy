#include "aig/aig/aigDoms.c"

/* Standalone console environment; no ABC frame or bridge is running. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    int mode, i, j, k;
    for (mode = 0; mode < 3; mode++)
    {
        Aig_Man_t p;
        Aig_Obj_t o[11];
        Aig_Sto_t *s;
        memset(&p, 0, sizeof(p)); memset(o, 0, sizeof(o));
        p.vObjs = Vec_PtrAlloc(11); p.vCis = Vec_PtrAlloc(4); p.vCos = Vec_PtrAlloc(3);
        p.pConst1 = o; p.nRegs = 1; p.nTruePis = 3; p.nTruePos = 2; p.nTravIds = 7;
        o[0].Type = AIG_OBJ_CONST1;
        for (i = 1; i <= 4; i++) o[i].Type = AIG_OBJ_CI;
        for (i = 5; i <= 7; i++) o[i].Type = AIG_OBJ_AND;
        o[8].Type = o[9].Type = o[10].Type = AIG_OBJ_CO;
        o[5].pFanin0 = o+1; o[5].pFanin1 = o+4;
        o[6].pFanin0 = o+2; o[6].pFanin1 = o+4;
        o[7].pFanin0 = o+1; o[7].pFanin1 = o+2;
        o[8].pFanin0 = o+7; o[9].pFanin0 = o+5; o[10].pFanin0 = o+6;
        for (i = 0; i < 11; i++)
        {
            o[i].Id = i; o[i].TravId = 5; p.nObjs[o[i].Type]++;
            Vec_PtrPush(p.vObjs, o+i);
            if (Aig_ObjIsCi(o+i)) Vec_PtrPush(p.vCis, o+i);
            if (Aig_ObjIsCo(o+i)) Vec_PtrPush(p.vCos, o+i);
            if (o[i].pFanin0) Aig_ObjFanin0(o+i)->nRefs++;
            if (o[i].pFanin1) Aig_ObjFanin1(o+i)->nRefs++;
        }
        Aig_ManFanoutStart(&p);
        /* Rotate LO's head without changing the graph. */
        Aig_ObjRemoveFanout(&p, o+4, o+5); Aig_ObjAddFanout(&p, o+4, o+5);
        s = mode == 0 ? Aig_ManComputeDomsNodes(&p, 2) :
            mode == 1 ? Aig_ManComputeDomsFlops(&p, 2) : Aig_ManComputeDomsPis(&p, 2);
        printf("BEGIN %d\nstate %d %d %d %d %d\n", mode, p.nTravIds,
            s->nDomNodes, s->nDomsTotal, s->nDomsFilter1, s->nDomsFilter2);
        for (i = 0; i < 11; i++)
        {
            Vec_Ptr_t *v = (Vec_Ptr_t *)Vec_PtrEntry(s->vDoms, i);
            printf("%d@%d:", i, o[i].TravId);
            if (v) for (j = 0; j < Vec_PtrSize(v); j++)
            {
                Aig_Dom_t *d = (Aig_Dom_t *)Vec_PtrEntry(v,j);
                printf("["); for (k = 0; k < d->nNodes; k++) printf("%s%d", k ? "," : "", d->pNodes[k]); printf("]");
            }
            puts("");
        }
        puts("END");
        if (mode == 1)
        {
            int volume, marked, refs[11];
            Vec_Int_t *los;
            Aig_Dom_t *d = (Aig_Dom_t *)malloc(sizeof(Aig_Dom_t) + sizeof(int));
            d->nNodes = 1; d->pNodes[0] = 5;
            o[5].fMarkA = 1;
            volume = Aig_ObjDomVolume(s, d);
            o[5].fMarkA = 0;
            free(d);
            los = Aig_ObjDomCollectLos(s); marked = o[4].fMarkA;
            Vec_IntFree(los);
            Aig_ObjPoLogicDeref(s);
            for (i = 0; i < 11; i++) refs[i] = o[i].nRefs;
            Aig_ObjPoLogicRef(s);
            o[1].fMarkA = 1;
            Aig_ObjDomFindGood(s);
            printf("DIAG BEGIN\nvolume %d lo-mark %d generation %d\n", volume, marked, p.nTravIds);
            for (i = 0; i < 11; i++) printf("%d refs %d restored %d mark %d trav %d\n", i, refs[i], o[i].nRefs, o[i].fMarkA, o[i].TravId);
            puts("DIAG END");
        }
        Aig_ManDomStop(s); Aig_ManFanoutStop(&p);
        Vec_PtrFree(p.vObjs); Vec_PtrFree(p.vCis); Vec_PtrFree(p.vCos);
    }
    return 0;
}

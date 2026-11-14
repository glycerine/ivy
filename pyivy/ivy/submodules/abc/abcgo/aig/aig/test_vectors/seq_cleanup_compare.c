#include "aig/aig/aig.h"

/* Console environment only. Allocation, pin updates, hash tables and cleanup
   all use the original production C routines. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

static int literal(Aig_Obj_t *o)
{
    return o ? 2*Aig_Regular(o)->Id + Aig_IsComplement(o) : -1;
}

int main(void)
{
    int count, fanouts, registers, basic, kind, a, b, id, pass, i;
    while (scanf("%d%d%d%d", &count, &fanouts, &registers, &basic) == 4)
    {
        Aig_Man_t *g = Aig_ManStart(count + 1);
        Aig_Obj_t **objects = ABC_ALLOC(Aig_Obj_t *, count);
        objects[0] = Aig_ManConst1(g);
        if (fanouts) Aig_ManFanoutStart(g);
        for (id = 1; id < count; id++)
        {
            Aig_Obj_t *o, *f0, *f1;
            assert(scanf("%d%d%d", &kind, &a, &b) == 3);
            f0 = a < 0 ? NULL : Aig_NotCond(objects[a/2], a & 1);
            f1 = b < 0 ? NULL : Aig_NotCond(objects[b/2], b & 1);
            if (kind == 1) o = Aig_ObjCreateCi(g);
            else if (kind == 5) o = Aig_ObjCreateCo(g, f0);
            else
            {
                o = Aig_ManFetchMemory(g);
                o->Type = kind == 2 ? AIG_OBJ_AND : kind == 4 ? AIG_OBJ_EXOR : AIG_OBJ_BUF;
                g->nObjs[o->Type]++;
                Aig_ObjConnect(g, o, f0, f1);
                if (Aig_ObjIsBuf(o)) Vec_PtrPush(g->vBufs, o);
            }
            assert(o->Id == id);
            objects[id] = o;
        }

        Aig_ManSetRegNum(g, registers);
        g->nTruePis = 77; g->nTruePos = 88;
        g->vFlopNums = Vec_IntAlloc(registers);
        for (i = 0; i < registers; i++) Vec_IntPush(g->vFlopNums, 100+i);
        for (i = 0; i < Aig_ManCiNum(g); i++) Aig_ManCi(g,i)->CioId = 10+i;
        for (i = 0; i < Aig_ManCoNum(g); i++) { Aig_ManCo(g,i)->CioId = 20+i; Aig_ManCo(g,i)->Level = 99; }
        for (pass = 0; pass < 2; pass++)
        {
            int removed = basic ? Aig_ManSeqCleanupBasic(g) : Aig_ManSeqCleanup(g);
            printf("removed %d regs %d true %d/%d gen %d hash %d | ci", removed,
                g->nRegs, g->nTruePis, g->nTruePos, g->nTravIds, Aig_TableCountEntries(g));
            for (i = 0; i < Aig_ManCiNum(g); i++) printf(" %d", Aig_ManCi(g,i)->Id);
            printf(" | co");
            for (i = 0; i < Aig_ManCoNum(g); i++) printf(" %d", Aig_ManCo(g,i)->Id);
            printf(" | flops");
            for (i = 0; i < Vec_IntSize(g->vFlopNums); i++) printf(" %d", Vec_IntEntry(g->vFlopNums,i));
            printf(" | objects");
            for (id = 0; id < count; id++)
            {
                Aig_Obj_t *o = Aig_ManObj(g,id);
                if (!o) printf(" dead");
                else
                {
                    int cio = (Aig_ObjIsCi(o) || Aig_ObjIsCo(o)) ? o->CioId : -1;
                    printf(" %d/%d/%d/%d/%d/%d", o->nRefs, literal(o->pFanin0), literal(o->pFanin1), o->Level, o->TravId, cio);
                    if (fanouts)
                    {
                        int first = Aig_ObjFanout0Int(g,id), pin = first;
                        printf("[");
                        if (first) do { printf("%d,",pin); pin=Aig_ObjFanoutNext(g,pin); } while (pin!=first);
                        printf("]");
                    }
                }
            }
            puts("");
        }
        ABC_FREE(objects);
        Aig_ManStop(g);
    }
    return 0;
}

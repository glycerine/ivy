#include "aig/aig/aig.h"

/* Console environment only. Allocation, support propagation and register
   grouping all use the original production C routines. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    int count, registers, kind, a, b, id, i, j, found;
    while (scanf("%d%d", &count, &registers) == 2)
    {
        Aig_Man_t *g = Aig_ManStart(count + 1);
        Aig_Obj_t **objects = ABC_ALLOC(Aig_Obj_t *, count);
        Vec_Ptr_t *supports;
        objects[0] = Aig_ManConst1(g);
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

        /* Permutations make support indices independent of object IDs. */
        for (i = 0; i < Aig_ManCiNum(g); i++)
        {
            assert(scanf("%d", &id) == 1);
            Vec_PtrWriteEntry(g->vCis,i,objects[id]);
        }
        for (i = 0; i < Aig_ManCoNum(g); i++)
        {
            assert(scanf("%d", &id) == 1);
            Vec_PtrWriteEntry(g->vCos,i,objects[id]);
        }
        Aig_ManSetRegNum(g,registers);
        for (id = 0; id < count; id++) objects[id]->TravId = 71 + id;
        supports = Aig_ManSupports(g);
        for (i = 1; i < Vec_PtrSize(supports); i++)
            assert(Vec_IntSize((Vec_Int_t *)Vec_PtrEntry(supports,i-1)) >= Vec_IntSize((Vec_Int_t *)Vec_PtrEntry(supports,i)));
        /* Equal-size rows have no C qsort tie contract. Compare by their
           explicit CO indices without discarding the row-size order check. */
        for (i = 0; i < Aig_ManCoNum(g); i++)
        {
            found = 0;
            for (j = 0; j < Vec_PtrSize(supports); j++)
            {
                Vec_Int_t *row = (Vec_Int_t *)Vec_PtrEntry(supports,j);
                int k;
                if (Vec_IntEntryLast(row) != i) continue;
                printf("row %d:",i);
                for (k = 0; k < Vec_IntSize(row)-1; k++) printf(" %d",Vec_IntEntry(row,k));
                puts(""); found++;
            }
            assert(found == 1);
        }
        Vec_VecFree((Vec_Vec_t *)supports);
        supports = Aig_ManSupportsRegisters(g);
        for (i = 0; i < registers; i++)
        {
            Vec_Int_t *row = (Vec_Int_t *)Vec_PtrEntry(supports,i);
            printf("reg %d:",i);
            for (j = 0; j < Vec_IntSize(row); j++) printf(" %d",Vec_IntEntry(row,j));
            puts("");
        }
        Vec_VecFree((Vec_Vec_t *)supports);
        Aig_ManComputeSccs(g);
        printf("state %d:",g->nTravIds);
        for (id = 0; id < count; id++)
            printf(" %d/%d",objects[id]->TravId,objects[id]->nRefs);
        printf(" | ci");
        for (i = 0; i < Aig_ManCiNum(g); i++) printf(" %d",Aig_ObjCioId(Aig_ManCi(g,i)));
        printf(" | co");
        for (i = 0; i < Aig_ManCoNum(g); i++) printf(" %d",Aig_ObjCioId(Aig_ManCo(g,i)));
        puts("");
        ABC_FREE(objects);
        Aig_ManStop(g);
    }
    return 0;
}

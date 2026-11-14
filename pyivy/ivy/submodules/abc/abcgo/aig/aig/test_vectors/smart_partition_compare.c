#include "aig/aig/aig.h"

unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    int size, limit, i, j, id;
    while (scanf("%d%d", &size, &limit) == 2)
    {
        Aig_Man_t *g = Aig_ManStart(1000);
        Aig_Obj_t *large = Aig_ManConst1(g), *small = Aig_ManConst1(g);
        Vec_Ptr_t *parts, *supports;
        for (i = 0; i < 512; i++) Aig_ObjCreateCi(g);
        for (i = 0; i < size; i++) large = Aig_And(g,large,Aig_ManCi(g,i));
        small = Aig_And(g,small,Aig_ManCi(g,0));
        for (i = 0; i < 9; i++) small = Aig_And(g,small,Aig_ManCi(g,256+i));
        Aig_ObjCreateCo(g,large); Aig_ObjCreateCo(g,small);
        parts = Aig_ManPartitionSmart(g,limit,0,&supports);
        printf("parts %d:\n",Vec_PtrSize(parts));
        for (i = 0; i < Vec_PtrSize(parts); i++)
        {
            Vec_Int_t *row = (Vec_Int_t *)Vec_PtrEntry(parts,i);
            printf("outputs");
            for (j = 0; j < Vec_IntSize(row); j++) printf(" %d",Vec_IntEntry(row,j));
            printf(" | inputs");
            row = (Vec_Int_t *)Vec_PtrEntry(supports,i);
            for (j = 0; j < Vec_IntSize(row); j++) printf(" %d",Vec_IntEntry(row,j));
            puts("");
        }
        /* C's compactor shrinks parts twice instead of shrinking supports.
           Its consumers observe the first parts->nSize support rows. */
        assert(Vec_PtrSize(supports) >= Vec_PtrSize(parts));
        Vec_VecFree((Vec_Vec_t *)parts);
        Vec_VecFree((Vec_Vec_t *)supports);
        Aig_ManStop(g);
    }
    return 0;
}

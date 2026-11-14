#include "aig/gia/gia.h"

extern int *Gia_ManMapMiniLut2MiniAig(Gia_Man_t *, Gia_Man_t *, Gia_Man_t *, Vec_Int_t *, Vec_Int_t *);
extern Vec_Int_t *Gia_ManMapEquivAfterScorr(Gia_Man_t *, Vec_Int_t *);

int main(void)
{
    int signs, phase, i;
    for (signs = 0; signs < 8; ++signs)
        for (phase = 0; phase < 8; ++phase)
        {
            Gia_Man_t *g = Gia_ManStart(1), *a = Gia_ManStart(1), *b = Gia_ManStart(1);
            int a0 = Gia_ManAppendCi(a), a1 = Gia_ManAppendCi(a), a2 = Gia_ManAppendCi(a);
            int ax = Gia_ManAppendAnd(a,a0^(signs&1),a1^((signs>>1)&1));
            int ay = Gia_ManAppendAnd(a,ax,a2^((signs>>2)&1));
            int ad = Gia_ManAppendAnd(a,a0^(signs&1),a1^((signs>>1)&1));
            int ao = Gia_ManAppendCo(a,ay), an = Gia_ManAppendCo(a,ay^1);
            int b0 = Gia_ManAppendCi(b), b1 = Gia_ManAppendCi(b), b2 = Gia_ManAppendCi(b);
            int bx = Gia_ManAppendAnd(b,b1^((signs>>1)&1),b2^((signs>>2)&1));
            int by = Gia_ManAppendAnd(b,b0^(signs&1),bx);
            int unused = Gia_ManAppendAnd(b,b0,b2);
            int bd = Gia_ManAppendAnd(b,b0^(signs&1),b1^((signs>>1)&1));
            int bo = Gia_ManAppendCo(b,by), bn = Gia_ManAppendCo(b,by^1);
            int amap[] = {0,a0,a1,a2,-1,ax^(phase&1),ay^((phase>>1)&1),ao,an,ad^((phase&1)^1)};
            int bmap[] = {0,1,b0,b1,b2,bx,by^((phase>>2)&1),unused,-1,bo,bn,bd};
            Vec_Int_t *va = Vec_IntAlloc(10), *vb = Vec_IntAlloc(12), *eq;
            int *names;
            for (i = 0; i < 10; ++i) { Gia_ManAppendCi(g); Vec_IntPush(va,amap[i]); }
            for (i = 0; i < 12; ++i) Vec_IntPush(vb,bmap[i]);
            g->pReprs = ABC_CALLOC(Gia_Rpr_t,g->nObjs);
            a->pReprs = ABC_CALLOC(Gia_Rpr_t,a->nObjs);
            for (i = 0; i < g->nObjs; ++i) g->pReprs[i].iRepr = GIA_VOID;
            for (i = 0; i < a->nObjs; ++i) a->pReprs[i].iRepr = GIA_VOID;
            Gia_ObjSetRepr(g,7,5); Gia_ObjSetRepr(g,9,4); Gia_ObjSetRepr(g,10,4);
            Gia_ObjSetRepr(a,ad>>1,ax>>1);
            Gia_ManFillValue(a); Gia_ManFillValue(b);
            Gia_ManConst0(a)->Value = Gia_ManConst0(b)->Value = 0;
            for (i = 1; i <= 3; ++i) Gia_ManObj(a,i)->Value = Gia_ManObj(b,i)->Value = 2*i;
            Gia_ManObj(a,ax>>1)->Value = 8; Gia_ManObj(a,ay>>1)->Value = 10; Gia_ManObj(a,ad>>1)->Value = 18;
            Gia_ManObj(b,bx>>1)->Value = 12; Gia_ManObj(b,by>>1)->Value = 14; Gia_ManObj(b,bd>>1)->Value = 20;
            names = Gia_ManMapMiniLut2MiniAig(g,a,b,va,vb);
            eq = Gia_ManMapEquivAfterScorr(a,va);
            printf("%d %d names",signs,phase);
            for (i = 0; i < 12; ++i) printf(" %d",names[i]);
            printf(" equiv");
            for (i = 0; i < 10; ++i) printf(" %d",Vec_IntEntry(eq,i));
            puts("");
            ABC_FREE(names); Vec_IntFree(eq); Vec_IntFree(va); Vec_IntFree(vb);
            Gia_ManStop(g); Gia_ManStop(a); Gia_ManStop(b);
        }
    return 0;
}

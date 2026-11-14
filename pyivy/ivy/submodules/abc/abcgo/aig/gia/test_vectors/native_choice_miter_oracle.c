#include "aig/gia/gia.h"
#include <stdio.h>

extern Gia_Man_t *Hcd_ManChoiceMiter(Vec_Ptr_t *vGias);
extern Gia_Man_t *Hcd_ManEquivToChoices(Gia_Man_t *, int);

static int equiv_choices(int hcd)
{
    int phase, i, mask, vals[128];
    Gia_Obj_t *o;
    for (phase = 0; phase < 8; ++phase)
    {
        Gia_Man_t *p = Gia_ManStart(32), *q;
        int a = Gia_ManAppendCi(p)^(phase&1);
        int b = Gia_ManAppendCi(p)^((phase>>1)&1);
        int c = Gia_ManAppendCi(p)^((phase>>2)&1);
        int ab = Gia_ManAppendAnd(p,a,b), head = Gia_ManAppendAnd(p,ab,c);
        int ac = Gia_ManAppendAnd(p,a,c), tail = Gia_ManAppendAnd(p,ac,b);
        Gia_ManAppendCo(p,head);
        Gia_ManAppendCo(p,tail);
        Gia_ManAppendCo(p,tail^1);
        p->pReprs = ABC_CALLOC(Gia_Rpr_t,p->nObjs);
        p->pNexts = ABC_CALLOC(int,p->nObjs);
        for (i = 0; i < p->nObjs; ++i) p->pReprs[i].iRepr = GIA_VOID;
        Gia_ObjSetRepr(p,tail>>1,head>>1);
        Gia_ObjSetNext(p,head>>1,tail>>1);
        q = hcd ? Hcd_ManEquivToChoices(p,1) : Gia_ManEquivToChoices(p,1);
        printf("eq %d %d %d %d\n",phase,q->nObjs,Gia_ManCoNum(q),Gia_ManCountChoices(q));
        Gia_ManForEachObj(q,o,i)
        {
            int kind = i == 0 ? 0 : Gia_ObjIsCi(o) ? 1 : Gia_ObjIsCo(o) ? 6 : 2;
            int f0 = kind >= 2 ? Gia_ObjFaninLit0(o,i) : -1;
            int f1 = kind == 2 ? Gia_ObjFaninLit1(o,i) : -1;
            printf("obj %d %d %d %d %u %d\n",i,kind,f0,f1,
                   q->pReprs ? q->pReprs[i].iRepr : GIA_VOID,
                   q->pNexts ? q->pNexts[i] : 0);
        }
        for (mask = 0; mask < 8; ++mask)
        {
            vals[0] = 0;
            Gia_ManForEachObj1(q,o,i)
            {
                if (Gia_ObjIsCi(o)) vals[i] = (mask>>Gia_ObjCioId(o))&1;
                else
                {
                    vals[i] = vals[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o);
                    if (Gia_ObjIsAnd(o)) vals[i] &= vals[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o);
                }
            }
            printf("truth %d",mask);
            Gia_ManForEachCo(q,o,i) printf(" %d",vals[Gia_ObjId(q,o)]);
            puts("");
        }
        Gia_ManStop(q);
        Gia_ManStop(p);
    }
    return 0;
}

static Gia_Man_t *snapshot(int phase, int real)
{
    Gia_Man_t *g = Gia_ManStart(24);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g);
    int c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g), x, m;
    if (real)
    {
        /* Optional reproducer of the source copier's AND-only limitation. */
        g->pMuxes = ABC_CALLOC(unsigned, g->nObjsAlloc);
        x = Gia_ManAppendXorReal(g, a, b^1);
        m = Gia_ManAppendMuxReal(g, c, b, d^1);
    }
    else
    {
        /* Supply ordinary AND graphs to the source choice algorithm, using
           its existing expansions of the real gates' Boolean functions. */
        Gia_ManHashAlloc(g);
        x = Gia_ManHashXor(g, a, b^1);
        m = Gia_ManHashMux(g, c, b, d^1);
        Gia_ManHashStop(g);
    }
    Gia_ManAppendCo(g, x^(phase&1));
    Gia_ManAppendCo(g, m^((phase>>1)&1));
    return g;
}

int main(int argc, char **argv)
{
    Gia_Man_t *g[3], *m;
    Vec_Ptr_t *v = Vec_PtrAlloc(3);
    Gia_Obj_t *o;
    int i, mask, vals[128], real = 0, hcd = 0, equiv = 0;
    for (i = 1; i < argc; ++i)
    {
        real |= strcmp(argv[i], "--real") == 0;
        hcd |= strcmp(argv[i], "--hcd") == 0;
        equiv |= strcmp(argv[i], "--equiv") == 0;
    }
    if (equiv) { Vec_PtrFree(v); return equiv_choices(hcd); }
    for (i = 0; i < 3; ++i)
    {
        g[i] = snapshot(i, real);
        Vec_PtrPush(v, g[i]);
    }
    m = hcd ? Hcd_ManChoiceMiter(v) : Gia_ManChoiceMiter(v);
    printf("nodes:%d,outputs:%d\n", Gia_ManObjNum(m), Gia_ManCoNum(m));
    for (mask = 0; mask < 16; ++mask)
    {
        vals[0] = 0;
        Gia_ManForEachObj1(m, o, i)
        {
            if (Gia_ObjIsCi(o))
                vals[i] = (mask >> Gia_ObjCioId(o)) & 1;
            else
            {
                vals[i] = vals[Gia_ObjFaninId0(o,i)] ^ Gia_ObjFaninC0(o);
                if (Gia_ObjIsAnd(o))
                    vals[i] &= vals[Gia_ObjFaninId1(o,i)] ^ Gia_ObjFaninC1(o);
            }
        }
        printf("case:%d", mask);
        Gia_ManForEachCo(m, o, i)
            printf(",%d", vals[Gia_ObjId(m,o)]);
        puts("");
    }
    Gia_ManStop(m);
    for (i = 0; i < 3; ++i)
        Gia_ManStop(g[i]);
    Vec_PtrFree(v);
    return 0;
}

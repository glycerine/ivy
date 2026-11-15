#include "aig/gia/gia.h"

static Gia_Man_t *graph(int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g);
    int x = Gia_ManAppendAnd(g,a^(signs&1),b);
    int y = Gia_ManAppendAnd(g,x,c), z = Gia_ManAppendAnd(g,y,d);
    int u,v,w;
    Gia_ManAppendCo(g,z^((signs>>1)&1));
    u = Gia_ManAppendAnd(g,a,c^((signs>>2)&1));
    v = Gia_ManAppendAnd(g,b^((signs>>3)&1),d);
    w = Gia_ManAppendAnd(g,u,v);
    Gia_ManAppendCo(g,w); Gia_ManAppendCo(g,u^((signs>>4)&1));
    Gia_ManAppendCo(g,d); Gia_ManAppendCo(g,y); Gia_ManAppendCo(g,1);
    Gia_ManSetRegNum(g,registers);
    g->pName = Abc_UtilStrsav("levels"); g->pSpec = Abc_UtilStrsav("levels.aig");
    return g;
}


extern Gia_Man_t *Gia_ManDupAbs(Gia_Man_t *, Vec_Int_t *, Vec_Int_t *);

int main(int argc, char **argv)
{
    Gia_Man_t *(*copies[])(Gia_Man_t *) = {Gia_ManDupOrderDfs,Gia_ManDupOrderDfsReverse,Gia_ManDupOrderAiger};
    Gia_Obj_t *obj;
    int mode, signs, regs, i, mask, packed;
    if (argc > 1 && !strcmp(argv[1],"--representatives"))
    {
        for (mode = 0; mode < 3; ++mode)
            for (regs = 0; regs <= 2; ++regs)
            {
                Gia_Man_t *g = Gia_ManStart(1), *out;
                int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), x, negative, zero, y;
                Gia_ManAppendCi(g);
                x = Gia_ManAppendAnd(g,a,b); negative = Gia_ManAppendAnd(g,a^1,b);
                zero = Gia_ManAppendAnd(g,x,negative);
                Gia_ManAppendCo(g,zero);
                y = Gia_ManAppendAnd(g,a,b);
                Gia_ManAppendCo(g,x); Gia_ManAppendCo(g,y); Gia_ManSetRegNum(g,regs);
                g->pReprs = ABC_CALLOC(Gia_Rpr_t,g->nObjs);
                for (i = 0; i < g->nObjs; ++i) Gia_ObjSetRepr(g,i,GIA_VOID);
                Gia_ObjSetRepr(g,y>>1,x>>1); Gia_ObjSetRepr(g,zero>>1,0);
                g->pNexts = Gia_ManDeriveNexts(g);
                out = copies[mode](g);
                printf("%d %d %d",mode,regs,out->nObjs);
                for (i = 0; i < out->nObjs; ++i) printf(" %d:%d",Gia_ObjRepr(out,i),Gia_ObjNext(out,i));
                puts(""); Gia_ManStop(out); Gia_ManStop(g);
            }
        return 0;
    }
    if (argc > 1 && !strcmp(argv[1],"--constraints"))
    {
        Gia_Man_t *g = graph(0,2), *out;
        Vec_Int_t *forward = Vec_IntAlloc(0), *inverse = Vec_IntStartFull(2);
        g->nConstrs = 1;
        printf("%d",g->nConstrs);
        for (mode = 0; mode < 3; ++mode)
        {
            out = copies[mode](g); printf(" %d",out->nConstrs); Gia_ManStop(out);
        }
        out = Gia_ManDupAbs(g,forward,inverse); printf(" %d\n",out->nConstrs);
        Gia_ManStop(out); Gia_ManStop(g); Vec_IntFree(forward); Vec_IntFree(inverse);
        return 0;
    }
    for (mode = 0; mode < 3; ++mode)
        for (signs = 0; signs < 32; ++signs)
            for (regs = 0; regs <= 4; regs += 2)
            {
                Gia_Man_t *g = graph(signs,0), *out;
                int *values;
                Gia_ManAppendCi(g); Gia_ManSetRegNum(g,regs);
                Gia_ManFillValue(g);
                out = copies[mode](g);
                printf("%d %d %d %d %d %d %d",mode,signs,regs,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
                Gia_ManForEachObj(out,obj,i)
                {
                    int kind = 0, a = 0, b = 0, ordinal = -1;
                    if (Gia_ObjIsCi(obj)) { kind = 1; ordinal = Gia_ObjCioId(obj); }
                    else if (Gia_ObjIsCo(obj)) { kind = 6; ordinal = Gia_ObjCioId(obj); a = Gia_ObjFaninLit0p(out,obj); }
                    else if (Gia_ObjIsAnd(obj)) { kind = 2; a = Gia_ObjFaninLit0p(out,obj); b = Gia_ObjFaninLit1p(out,obj); }
                    printf(" %d:%d:%d:%d",kind,a,b,ordinal);
                }
                Gia_ManForEachObj(g,obj,i) printf(" %u",obj->Value);
                values = ABC_CALLOC(int,out->nObjs);
                for (mask = 0; mask < 32; ++mask)
                {
                    Gia_ManForEachCi(out,obj,i) values[Gia_ObjId(out,obj)] = (mask>>i)&1;
                    Gia_ManForEachAnd(out,obj,i)
                        values[i] = (values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
                    packed = 0;
                    Gia_ManForEachCo(out,obj,i)
                        packed |= (values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj)) << i;
                    printf(" %d",packed);
                }
                puts(""); ABC_FREE(values); Gia_ManStop(out); Gia_ManStop(g);
            }
    return 0;
}

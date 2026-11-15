#include "aig/gia/gia.h"
extern void Gia_ManDupRemapCis(Gia_Man_t *, Gia_Man_t *);
extern void Gia_ManDupRemapCos(Gia_Man_t *, Gia_Man_t *);

static void literals(void)
{
    unsigned options[7] = {0,1,2,3,8,9,~0u};
    int a,b,c,i;
    for (a = 0; a < 7; ++a) for (b = 0; b < 7; ++b) for (c = 0; c < 7; ++c)
    {
        Gia_Man_t *g = Gia_ManStart(1);
        Vec_Int_t *lits = Vec_IntAlloc(11);
        for (i = 0; i < 3; ++i) Gia_ManAppendCi(g);
        Gia_ManConst0(g)->Value = 0;
        Gia_ManCi(g,0)->Value = options[a]; Gia_ManCi(g,1)->Value = options[b]; Gia_ManCi(g,2)->Value = options[c];
        Vec_IntPush(lits,-1); Vec_IntPush(lits,-7); Vec_IntPush(lits,(-2147483647-1));
        for (i = 0; i < 8; ++i) Vec_IntPush(lits,i);
        Gia_ManDupRemapLiterals(lits,g);
        printf("%d %d %d",a,b,c);
        for (i = 0; i < Vec_IntSize(lits); ++i) printf(" %d",Vec_IntEntry(lits,i));
        puts(""); Vec_IntFree(lits); Gia_ManStop(g);
    }
}

static void permute(int perms[24][4], int *count, int order[4], int size, int used)
{
    int i;
    if (size == 4) { memcpy(perms[(*count)++],order,4*sizeof(int)); return; }
    for (i = 0; i < 4; ++i) if (!(used&(1<<i)))
    {
        order[size] = i; permute(perms,count,order,size+1,used|(1<<i));
    }
}

static void terminals(void)
{
    int perms[24][4], order[4], count = 0, pi, po, regs, i, position, mask, packed;
    Gia_Obj_t *obj;
    permute(perms,&count,order,0,0);
    for (pi = 0; pi < count; ++pi) for (po = 0; po < count; ++po) for (regs = 0; regs <= 4; regs += 2)
    {
        Gia_Man_t *g = Gia_ManStart(1), *out = Gia_ManStart(1);
        int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g);
        int x = Gia_ManAppendAnd(g,a,b), physical[4], *values;
        Gia_ManAppendCo(g,a^1); Gia_ManAppendCo(g,x); Gia_ManAppendCo(g,c); Gia_ManAppendCo(g,d);
        Gia_ManSetRegNum(g,regs);
        for (i = 0; i < 4; ++i) physical[i] = Gia_ManAppendCi(out);
        Gia_ManConst0(g)->Value = 0;
        for (i = 0; i < 4; ++i) Gia_ManCi(g,i)->Value = physical[perms[pi][i]];
        Gia_ManObj(g,x>>1)->Value = Gia_ManAppendAnd(out,Gia_ManCi(g,0)->Value,Gia_ManCi(g,1)->Value);
        for (position = 0; position < 4; ++position)
            for (i = 0; i < 4; ++i) if (perms[po][i] == position)
            {
                obj = Gia_ManCo(g,i); obj->Value = Gia_ManAppendCo(out,Gia_ObjFanin0Copy(obj));
            }
        Gia_ManSetRegNum(out,regs);
        Gia_ManDupRemapCis(out,g); Gia_ManDupRemapCos(out,g);
        printf("%d %d %d",pi,po,regs);
        Gia_ManForEachCi(out,obj,i) printf(" %d:%d",Gia_ObjId(out,obj),Gia_ObjCioId(obj));
        Gia_ManForEachCo(out,obj,i) printf(" %d:%d",Gia_ObjId(out,obj),Gia_ObjCioId(obj));
        values = ABC_CALLOC(int,out->nObjs);
        for (mask = 0; mask < 16; ++mask)
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
}

static void equiv(void)
{
    int rotation, reverse, phase, drop, i;
    for (rotation = 0; rotation < 8; ++rotation) for (reverse = 0; reverse < 2; ++reverse)
        for (phase = 0; phase < 4; ++phase) for (drop = 0; drop < 2; ++drop)
        {
            Gia_Man_t *g = Gia_ManStart(1), *out = Gia_ManStart(1);
            for (i = 0; i < 8; ++i) { Gia_ManAppendCi(g); Gia_ManAppendCi(out); }
            g->pReprs = ABC_CALLOC(Gia_Rpr_t,g->nObjs);
            for (i = 0; i < g->nObjs; ++i) Gia_ObjSetRepr(g,i,GIA_VOID);
            Gia_ObjSetRepr(g,3,1); Gia_ObjSetRepr(g,5,1); Gia_ObjSetRepr(g,4,2);
            Gia_ObjSetRepr(g,6,0); Gia_ObjSetRepr(g,7,0);
            g->pNexts = Gia_ManDeriveNexts(g);
            Gia_ManConst0(g)->Value = 0;
            for (i = 1; i <= 8; ++i)
            {
                int index = (i-1+rotation)%8, polarity = 0;
                if (reverse) index = 7-index;
                if (phase == 1) polarity = 1;
                else if (phase == 2) polarity = i&1;
                else if (phase == 3) polarity = (i>>1)&1;
                Gia_ManObj(g,i)->Value = ((index+1)<<1)|polarity;
            }
            if (drop) Gia_ManObj(g,8)->Value = ~0u;
            Gia_ManDupRemapEquiv(out,g);
            printf("%d %d %d %d",rotation,reverse,phase,drop);
            for (i = 0; i < out->nObjs; ++i) printf(" %d:%d",Gia_ObjRepr(out,i),Gia_ObjNext(out,i));
            puts(""); Gia_ManStop(out); Gia_ManStop(g);
        }
}

int main(int argc, char **argv)
{
    if (argc != 2) return 1;
    if (!strcmp(argv[1],"literals")) literals();
    else if (!strcmp(argv[1],"terminals")) terminals();
    else if (!strcmp(argv[1],"equiv")) equiv();
    else return 1;
    return 0;
}

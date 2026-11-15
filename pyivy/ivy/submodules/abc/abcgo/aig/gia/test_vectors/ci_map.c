#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupDfsCiMap(Gia_Man_t *,int *,Vec_Int_t *);

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


static int cases(int specs[131][5])
{
    int count = 0, a,b,c,d,e,i,phases,constants;
    for (a = -1; a <= 3; ++a) for (b = -1; b <= 5; ++b)
    {
        int *s = specs[count++]; s[0] = a; s[1] = b; s[2] = s[3] = s[4] = -1;
    }
    for (phases = 0; phases < 32; ++phases)
    {
        int *s = specs[count++]; s[0] = 2+(phases&1);
        for (i = 1; i < 5; ++i) s[i] = 2+2*(i-1)+((phases>>i)&1);
    }
    for (constants = 0; constants < 32; ++constants)
    {
        int *s = specs[count++]; for (i = 0; i < 5; ++i) s[i] = (constants>>i)&1;
    }
    for (b = 0; b < 2; ++b) for (c = 0; c < 2; ++c) for (d = 0; d < 2; ++d) for (e = 0; e < 4; ++e)
    {
        int *s = specs[count++]; s[0] = -1; s[1] = 4+b; s[2] = 2+c; s[3] = 4+d; s[4] = 6+e;
    }
    return count;
}

int main(int argc, char **argv)
{
    const int signsList[8] = {0,1,2,4,8,16,21,31};
    int specs[131][5], count = cases(specs), pattern, ciCase, regs, mode, i, mask, packed;
    Gia_Obj_t *obj;
    if (argc > 1 && !strcmp(argv[1],"constraints"))
    {
        Gia_Man_t *g = graph(0,0), *out;
        int mapping[5] = {-1,-1,-1,-1,-1};
        Gia_ManAppendCi(g); Gia_ManSetRegNum(g,2); g->nConstrs = 1;
        out = Gia_ManDupDfsCiMap(g,mapping,NULL);
        printf("%d %d\n",g->nConstrs,out->nConstrs); Gia_ManStop(out); Gia_ManStop(g); return 0;
    }
    if (count != 131) return 1;
    for (pattern = 0; pattern < 8; ++pattern) for (ciCase = 0; ciCase < count; ++ciCase)
        for (regs = 0; regs <= 4; regs += 2) for (mode = 0; mode < 3; ++mode)
        {
            int signs = signsList[pattern], mapping[5], *values;
            Gia_Man_t *g = graph(signs,0), *out;
            Vec_Int_t *lits = NULL;
            Gia_ManAppendCi(g); Gia_ManSetRegNum(g,regs); g->nConstrs = 1;
            for (i = 0; i < 5; ++i)
            {
                int entry = specs[ciCase][i];
                mapping[i] = entry;
                if (entry >= 2) mapping[i] = (Gia_ObjId(g,Gia_ManCi(g,(entry-2)/2))<<1)|((entry-2)&1);
            }
            if (mode)
            {
                lits = Vec_IntAlloc(3+regs);
                if (mode == 1)
                {
                    int first = Gia_ObjFaninLit0p(g,Gia_ManCo(g,0));
                    Vec_IntPush(lits,first^1); Vec_IntPush(lits,first); Vec_IntPush(lits,first^1);
                }
                for (i = 0; i < regs; ++i) Vec_IntPush(lits,Gia_ObjFaninLit0p(g,Gia_ManRi(g,i)));
            }
            out = Gia_ManDupDfsCiMap(g,mapping,lits);
            printf("%d %d %d %d %d %d %d %d",signs,ciCase,regs,mode,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
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
            puts(""); ABC_FREE(values); if (lits) Vec_IntFree(lits); Gia_ManStop(out); Gia_ManStop(g);
        }
    return 0;
}

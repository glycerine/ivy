#include "aig/gia/gia.h"
extern Gia_Man_t *Gia_ManDupOuts(Gia_Man_t *);

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


static void truth(Gia_Man_t *out)
{
    Gia_Obj_t *obj;
    int i, mask, packed, *values = ABC_CALLOC(int,out->nObjs);
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
    puts("");
    ABC_FREE(values);
}

int main(void)
{
    const int perms[6][4] = {{0,1,2,3},{3,2,1,0},{1,0,3,2},{2,3,0,1},{1,3,0,2},{3,0,2,1}};
    const int gateOrders[2][6] = {{0,1,2,3,4,5},{0,3,4,1,5,2}};
    const int selections[4][6] = {{0,1,2,3,4,5},{5,4,3,2,1,0},{5,0,0,3}, {0}};
    const int sizes[4] = {6,6,4,0};
    int signs, pi, ordering, si, regs, i, k;
    Gia_Obj_t *obj;
    for (signs = 0; signs < 32; ++signs)
        for (pi = 0; pi < 6; ++pi)
            for (ordering = 0; ordering < 2; ++ordering)
                for (si = 0; si < 4; ++si)
                    for (regs = 0; regs <= 4; regs += 2)
                    {
                        Gia_Man_t *g, *out;
                        Vec_Int_t *cis, *ands, *cos;
                        int ids[6];
                        if (regs > sizes[si]) continue;
                        g = graph(signs,2); g->nConstrs = 1;
                        cis = Vec_IntAlloc(4); ands = Vec_IntAlloc(6); cos = Vec_IntAlloc(6);
                        for (i = 0; i < 4; ++i) Vec_IntPush(cis,Gia_ObjId(g,Gia_ManCi(g,perms[pi][i])));
                        k = 0;
                        Gia_ManForEachAnd(g,obj,i) ids[k++] = i;
                        for (i = 0; i < 6; ++i) Vec_IntPush(ands,ids[gateOrders[ordering][i]]);
                        for (i = 0; i < sizes[si]; ++i) Vec_IntPush(cos,Gia_ObjId(g,Gia_ManCo(g,selections[si][i])));
                        out = Gia_ManDupFromVecs(g,cis,ands,cos,regs);
                        printf("0 %d %d %d %d %d %d %d %d %d",signs,pi,ordering,si,regs,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
                        truth(out);
                        Vec_IntFree(cis); Vec_IntFree(ands); Vec_IntFree(cos);
                        Gia_ManStop(out); Gia_ManStop(g);
                    }
    for (signs = 0; signs < 32; ++signs)
        for (regs = 0; regs <= 4; regs += 2)
        {
            Gia_Man_t *g = graph(signs,regs), *out;
            g->nConstrs = 1;
            out = Gia_ManDupOuts(g);
            printf("1 %d %d %d %d %d %d",signs,regs,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
            truth(out);
            Gia_ManStop(out); Gia_ManStop(g);
        }
    return 0;
}

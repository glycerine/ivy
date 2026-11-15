#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupDfsOnePo(Gia_Man_t *,int);

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


int main(void)
{
    int signs, regs, choice, i, mask, packed;
    Gia_Obj_t *obj;
    for (signs = 0; signs < 32; ++signs) for (regs = 0; regs <= 4; regs += 2)
        for (choice = 0; choice < 6-regs; ++choice)
        {
            Gia_Man_t *g = graph(signs,0), *out;
            int *values;
            Gia_ManAppendCi(g); Gia_ManSetRegNum(g,regs); g->nConstrs = 1;
            Gia_ManForEachObj(g,obj,i) { obj->fMark0 = i&1; obj->fMark1 = (i>>1)&1; }
            out = Gia_ManDupDfsOnePo(g,choice);
            printf("%d %d %d %d %d %d %d",signs,regs,choice,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
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

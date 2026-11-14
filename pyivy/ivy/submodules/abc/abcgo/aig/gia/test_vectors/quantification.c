#include "aig/gia/gia.h"

extern Gia_Man_t *Gia_ManDupExist2(Gia_Man_t *, int);

static Gia_Man_t *graph(int signs, int registers, int single)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g);
    int x = Gia_ManAppendXor(g,a^(signs&1),b);
    int buffer = Gia_ManAppendBuf(g,x^((signs>>1)&1));
    int mux = Gia_ManAppendMux(g,c^((signs>>2)&1),buffer,d);
    Gia_ManAppendCo(g,Gia_ManAppendAnd(g,mux,a^1));
    if (!single)
    {
        Gia_ManAppendCo(g,Gia_ManAppendAnd(g,c,x));
        Gia_ManAppendCo(g,mux^1);
    }
    if (registers) { Gia_ManSetRegNum(g,1); g->nConstrs = 1; }
    return g;
}

int main(void)
{
    int regs, signs, variable, op, mask, i, value, positiveName;
    for (regs = 0; regs < 2; ++regs)
        for (signs = 0; signs < 8; ++signs)
            for (variable = 0; variable < 4-regs; ++variable)
                for (op = 0; op < 8; ++op)
                {
                    Gia_Man_t *g, *out, *negative = NULL;
                    Gia_Obj_t *obj;
                    int *values;
                    if (regs && op >= 3) continue;
                    g = graph(signs,regs,op == 4 || op == 5);
                    switch (op)
                    {
                    case 0: out = Gia_ManDupCofactorVar(g,variable,0); break;
                    case 1: out = Gia_ManDupCofactorVar(g,variable,1); break;
                    case 2: out = Gia_ManDupExist2(g,variable); break;
                    case 3: out = Gia_ManDupUniv(g,variable); break;
                    case 4: out = Gia_ManDupExist(g,variable); break;
                    case 5:
                        negative = graph(signs^7,0,1);
                        g->pName = Abc_UtilStrsav("positive"); g->pSpec = Abc_UtilStrsav("positive.aig");
                        negative->pName = Abc_UtilStrsav("negative"); negative->pSpec = Abc_UtilStrsav("negative.aig");
                        out = Gia_ManDupMux(variable,g,negative); break;
                    case 6: out = Gia_ManDupCofactorObj(g,Gia_ObjFaninId0p(g,Gia_ManCo(g,0)),variable&1); break;
                    default: out = Gia_ManDupCofactorObj(g,Gia_ObjId(g,Gia_ManCi(g,variable)),signs&1); break;
                    }
                    positiveName = out->pName && out->pSpec && !strcmp(out->pName,"positive") && !strcmp(out->pSpec,"positive.aig");
                    printf("%d %d %d %d %d %d %d %d %d",regs,signs,variable,op,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out),out->nConstrs,positiveName);
                    values = ABC_CALLOC(int,out->nObjs);
                    for (mask = 0; mask < 16; ++mask)
                    {
                        Gia_ManForEachCi(out,obj,i) values[Gia_ObjId(out,obj)] = (mask>>i)&1;
                        Gia_ManForEachAnd(out,obj,i)
                            values[i] = (values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
                        value = 0;
                        Gia_ManForEachCo(out,obj,i)
                            value |= (values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
                        printf(" %d",value);
                    }
                    puts("");
                    ABC_FREE(values); Gia_ManStop(out); Gia_ManStop(g);
                    if (negative) Gia_ManStop(negative);
                }
    return 0;
}

#include "aig/gia/gia.h"

int main(void)
{
    int variable, phase, reverse, signs;
    for (variable = 0; variable <= 1; ++variable)
        for (phase = 0; phase <= 1; ++phase)
            for (reverse = 0; reverse <= 1; ++reverse)
                for (signs = 0; signs < 4; ++signs)
                {
                    Gia_Man_t *g = Gia_ManStart(1);
                    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), control = phase;
                    int t = reverse ? b : a, e = reverse ? a : b;
                    int root, mask, observed = 0, inputs[4];
                    Gia_Obj_t *obj;
                    if (variable) control = Gia_ManAppendCi(g)^phase;
                    t ^= signs&1; e ^= (signs>>1)&1;
                    g->pMuxes = ABC_CALLOC(unsigned,g->nObjsAlloc);
                    root = Gia_ManAppendMuxReal(g,control,t,e);
                    obj = Gia_ManObj(g,Abc_Lit2Var(root));
                    for (mask = 0; mask < (variable ? 8 : 4); ++mask)
                    {
                        int id = Abc_Lit2Var(root), tv, ev, value;
                        inputs[0] = 0; inputs[1] = mask&1; inputs[2] = (mask>>1)&1; inputs[3] = (mask>>2)&1;
                        tv = inputs[Gia_ObjFaninId1(obj,id)]^Gia_ObjFaninC1(obj);
                        ev = inputs[Gia_ObjFaninId0(obj,id)]^Gia_ObjFaninC0(obj);
                        if (Gia_ObjIsMux(g,obj))
                        {
                            int c = g->pMuxes[id];
                            value = (inputs[Abc_Lit2Var(c)]^Abc_LitIsCompl(c)) ? tv : ev;
                        }
                        else value = tv & ev;
                        observed |= value<<mask;
                    }
                    printf("%d %d %d %d %d %d\n",variable,phase,reverse,signs,Gia_ObjIsMux(g,obj)!=0,observed);
                    Gia_ManStop(g);
                }
    return 0;
}

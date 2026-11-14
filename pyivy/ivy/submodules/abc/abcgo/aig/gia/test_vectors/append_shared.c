#include "aig/gia/gia.h"

/* HashStart requires a structurally unique AND graph without physical buffers.
   Use the equivalent ordinary-AND expansion; Go coverage keeps real gates. */
static Gia_Man_t *graph(int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), ro[2], i, x, buffer, mux;
    int control = a, other = b;
    for (i = 0; i < registers; ++i) ro[i] = Gia_ManAppendCi(g);
    if (registers) control = ro[0];
    if (registers > 1) other = ro[registers-1];
    Gia_ManHashAlloc(g);
    x = Gia_ManHashXor(g, a^(signs&1), b);
    buffer = x^((signs>>1)&1);
    mux = Gia_ManHashMux(g, control^((signs>>2)&1), buffer, other);
    Gia_ManAppendCo(g, mux);
    Gia_ManAppendCo(g, a^1);
    for (i = 0; i < registers; ++i)
        Gia_ManAppendCo(g, Gia_ManHashXor(g, ro[i], mux));
    Gia_ManHashStop(g);
    Gia_ManSetRegNum(g, registers);
    return g;
}

int main(void)
{
    int s1, s2, regs, mode, mask, i, packed;
    for (s1 = 0; s1 < 8; ++s1)
        for (s2 = 0; s2 < 8; ++s2)
            for (regs = 0; regs <= 2; ++regs)
                for (mode = 0; mode < 3; ++mode)
                {
                    Gia_Man_t *a = graph(s1,regs), *b = graph(s2,regs), *out;
                    Gia_Obj_t *obj;
                    int *values;
                    if (mode == 0) out = Gia_ManDup2(a,b);
                    else
                    {
                        out = Gia_ManDup(a);
                        if (s1&1) Gia_ManHashStart(out);
                        if (mode == 1) Gia_ManDupAppendShare(out,b);
                        else Gia_ManDupAppend(out,b);
                    }
                    values = ABC_CALLOC(int, out->nObjs);
                    printf("%d %d %d %d %d %d %d", s1,s2,regs,mode,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out));
                    for (mask = 0; mask < (1 << Gia_ManCiNum(out)); ++mask)
                    {
                        Gia_ManForEachCi(out,obj,i) values[Gia_ObjId(out,obj)] = (mask>>i)&1;
                        Gia_ManForEachAnd(out,obj,i)
                            values[i] = (values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
                        packed = 0;
                        Gia_ManForEachCo(out,obj,i)
                            packed |= (values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj)) << i;
                        printf(" %d", packed);
                    }
                    puts("");
                    ABC_FREE(values);
                    Gia_ManStop(out); Gia_ManStop(a); Gia_ManStop(b);
                }
    return 0;
}

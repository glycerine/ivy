#include "aig/gia/gia.h"

static Gia_Man_t *graph(int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), ro[2], i, x, buffer, mux;
    int control = a, other = b;
    for (i = 0; i < registers; ++i) ro[i] = Gia_ManAppendCi(g);
    if (registers) control = ro[0];
    if (registers > 1) other = ro[registers-1];
    x = Gia_ManAppendXor(g, a^(signs&1), b);
    buffer = Gia_ManAppendBuf(g, x^((signs>>1)&1));
    mux = Gia_ManAppendMux(g, control^((signs>>2)&1), buffer, other);
    Gia_ManAppendCo(g, mux);
    Gia_ManAppendCo(g, a^1);
    for (i = 0; i < registers; ++i)
        Gia_ManAppendCo(g, Gia_ManAppendXor(g, ro[i], mux));
    Gia_ManSetRegNum(g, registers);
    return g;
}

int main(void)
{
    int s1, s2, r1, r2, mask, i, packed;
    for (s1 = 0; s1 < 8; ++s1)
        for (s2 = 0; s2 < 8; ++s2)
            for (r1 = 0; r1 <= 2; ++r1)
                for (r2 = 0; r2 <= 2; ++r2)
                {
                    Gia_Man_t *a = graph(s1,r1), *b = graph(s2,r2), *out = Gia_ManDupAppendNew(a,b);
                    Gia_Obj_t *obj;
                    int *values = ABC_CALLOC(int, out->nObjs);
                    printf("%d %d %d %d %d %d %d", s1,s2,r1,r2,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out));
                    for (mask = 0; mask < (1 << (2+r1+r2)); ++mask)
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

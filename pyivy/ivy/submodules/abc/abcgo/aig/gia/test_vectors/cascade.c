#include "aig/gia/gia.h"

static Gia_Man_t *first(int signs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g);
    int x = Gia_ManAppendXor(g, a^(signs&1), b);
    int buffer = Gia_ManAppendBuf(g, x^((signs>>1)&1));
    int m = Gia_ManAppendMux(g, c^((signs>>2)&1), buffer, d);
    Gia_ManAppendCo(g, Gia_ManAppendAnd(g, m, a^1));
    Gia_ManAppendCo(g, Gia_ManAppendAnd(g, c, x));
    Gia_ManAppendCo(g, m^1);
    return g;
}

static Gia_Man_t *second(int signs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g);
    int x = Gia_ManAppendXor(g, a^(signs&1), b);
    int buffer = Gia_ManAppendBuf(g, x^((signs>>1)&1));
    int m = Gia_ManAppendMux(g, c^((signs>>2)&1), buffer, a^1);
    Gia_ManAppendCo(g, m);
    Gia_ManAppendCo(g, x^1);
    Gia_ManAppendAnd(g, b, c);
    return g;
}

int main(void)
{
    int s1, s2, mask, i, packed;
    for (s1 = 0; s1 < 8; ++s1)
        for (s2 = 0; s2 < 8; ++s2)
        {
            Gia_Man_t *a = first(s1), *b = second(s2), *out = Gia_ManDupOntop(a, b);
            Gia_Obj_t *obj;
            int *values = ABC_CALLOC(int, out->nObjs);
            printf("%d %d %d %d %d", s1, s2, Gia_ManCiNum(out), Gia_ManCoNum(out), Gia_ManRegNum(out));
            for (mask = 0; mask < 16; ++mask)
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

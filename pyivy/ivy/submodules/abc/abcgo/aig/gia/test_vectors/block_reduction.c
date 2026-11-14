#include "aig/gia/gia.h"

static Gia_Man_t *graph(int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int ci[6], i, x, b, m, y;
    for (i = 0; i < 6; ++i) ci[i] = Gia_ManAppendCi(g);
    x = Gia_ManAppendXor(g, ci[0]^(signs&1), ci[2]);
    b = Gia_ManAppendBuf(g, x^((signs>>1)&1));
    m = Gia_ManAppendMux(g, ci[4]^((signs>>2)&1), b, ci[1]);
    Gia_ManAppendCo(g, m);
    y = Gia_ManAppendAnd(g, m, ci[3]^1);
    Gia_ManAppendCo(g, x);
    Gia_ManAppendCo(g, b^1);
    Gia_ManAppendCo(g, y);
    Gia_ManAppendCo(g, ci[5]);
    Gia_ManAppendCo(g, 1);
    Gia_ManSetRegNum(g, registers);
    return g;
}

int main(void)
{
    int signs, bi, ri, mask, i, packed;
    int blocks[] = {1, 2, 3, 6};
    for (signs = 0; signs < 8; ++signs)
        for (bi = 0; bi < 4; ++bi)
            for (ri = 0; ri < 3; ++ri)
            {
                int block = blocks[bi], registers = ri == 0 ? 0 : ri == 1 ? block : 6;
                Gia_Man_t *g = graph(signs, registers);
                Gia_Man_t *out = Gia_ManDupBlock(g, block);
                Gia_Obj_t *obj;
                int *values = ABC_CALLOC(int, out->nObjs);
                printf("%d %d %d %d %d %d", signs, block, registers, Gia_ManCiNum(out), Gia_ManCoNum(out), Gia_ManRegNum(out));
                for (mask = 0; mask < (1 << (6/block)); ++mask)
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
                Gia_ManStop(out);
                Gia_ManStop(g);
            }
    return 0;
}

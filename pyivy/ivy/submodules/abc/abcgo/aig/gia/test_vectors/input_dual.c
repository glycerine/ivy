#include "aig/gia/gia.h"

extern Gia_Man_t *Gia_ManDupOnsetOffset(Gia_Man_t *);

static Gia_Man_t *drop_graph(int keep, int signs, int last)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int all[6], *ci, i, m = signs&1;
    for (i = 0; i < 6; ++i) all[i] = Gia_ManAppendCi(g);
    ci = all + (last ? 6-keep : 0);
    Gia_ManHashAlloc(g);
    if (keep)
    {
        m = ci[0]^(signs&1);
        if (keep > 1) m = Gia_ManHashXor(g,m,ci[1]);
        m ^= (signs>>1)&1; /* Transparent equivalent of Go's physical buffer. */
        if (keep > 2) m = Gia_ManHashMux(g,ci[2]^((signs>>2)&1),m,ci[0]^1);
        for (i = 3; i < keep; ++i) m = Gia_ManHashXor(g,m,ci[i]);
    }
    Gia_ManHashStop(g);
    Gia_ManAppendCo(g,m);
    Gia_ManAppendCo(g,m^1);
    Gia_ManAppendCo(g,(signs>>2)&1);
    return g;
}

static Gia_Man_t *dual_graph(int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int ci[6], i, x, buffer, m, y;
    for (i = 0; i < 6; ++i) ci[i] = Gia_ManAppendCi(g);
    x = Gia_ManAppendXor(g,ci[0]^(signs&1),ci[2]);
    buffer = x^((signs>>1)&1);
    m = Gia_ManAppendMux(g,ci[4]^((signs>>2)&1),buffer,ci[1]);
    Gia_ManAppendCo(g,m);
    y = Gia_ManAppendAnd(g,m,ci[3]^1);
    Gia_ManAppendCo(g,x);
    Gia_ManAppendCo(g,buffer^1);
    Gia_ManAppendCo(g,y);
    Gia_ManAppendCo(g,ci[5]);
    Gia_ManAppendCo(g,1);
    Gia_ManSetRegNum(g,registers);
    return g;
}

static void record(Gia_Man_t *g, int mode, int signs, int parameter)
{
    Gia_Obj_t *obj;
    int mask, i, packed;
    int *values = ABC_CALLOC(int,g->nObjs);
    printf("%d %d %d %d %d %d",mode,signs,parameter,Gia_ManCiNum(g),Gia_ManCoNum(g),Gia_ManRegNum(g));
    for (mask = 0; mask < (1 << Gia_ManCiNum(g)); ++mask)
    {
        Gia_ManForEachCi(g,obj,i) values[Gia_ObjId(g,obj)] = (mask>>i)&1;
        Gia_ManForEachAnd(g,obj,i)
            values[i] = (values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
        packed = 0;
        Gia_ManForEachCo(g,obj,i)
            packed |= (values[Gia_ObjFaninId0p(g,obj)]^Gia_ObjFaninC0(obj)) << i;
        printf(" %d",packed);
    }
    puts("");
    ABC_FREE(values);
}

int main(void)
{
    int signs, keep, mode, registers;
    for (signs = 0; signs < 8; ++signs)
        for (keep = 0; keep <= 6; ++keep)
            for (mode = 0; mode < 2; ++mode)
            {
                Gia_Man_t *g = drop_graph(keep,signs,mode == 1);
                Gia_Man_t *out = mode == 0 ? Gia_ManDupRemovePis(g,6-keep) : Gia_ManDupLastPis(g,keep);
                record(out,mode,signs,keep);
                Gia_ManStop(out); Gia_ManStop(g);
            }
    for (signs = 0; signs < 8; ++signs)
        for (registers = 0; registers <= 6; registers += 3)
        {
            Gia_Man_t *g = dual_graph(signs,registers), *out = Gia_ManDupOnsetOffset(g);
            record(out,2,signs,registers);
            Gia_ManStop(out); Gia_ManStop(g);
        }
    return 0;
}

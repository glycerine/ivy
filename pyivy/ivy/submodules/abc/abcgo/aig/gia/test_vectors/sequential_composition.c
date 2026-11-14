#include "aig/gia/gia.h"

extern Gia_Man_t *Gia_ManDupInvMiter(Gia_Man_t *, Gia_Man_t *);

static Gia_Man_t *source_graph(int signs)
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
    Gia_ManSetRegNum(g,3);
    return g;
}

static Gia_Man_t *cone_graph(int inputs, int signs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int ci[6], i, x, buffer, m;
    for (i = 0; i < inputs; ++i) ci[i] = Gia_ManAppendCi(g);
    x = Gia_ManAppendXor(g,ci[0]^(signs&1),ci[1]);
    buffer = x^((signs>>1)&1);
    m = Gia_ManAppendMux(g,ci[2]^((signs>>2)&1),buffer,ci[0]^1);
    if (inputs == 6)
    {
        m = Gia_ManAppendXor(g,m,ci[3]);
        m = Gia_ManAppendMux(g,ci[4],m,ci[5]);
    }
    Gia_ManAppendCo(g,m);
    return g;
}

static void record(Gia_Man_t *out, int mode, int signs, int coneSigns, int onlyRegs, int count)
{
    Gia_Obj_t *obj;
    int mask, i, packed;
    int *values = ABC_CALLOC(int,out->nObjs);
    printf("%d %d %d %d %d %d %d %d",mode,signs,coneSigns,onlyRegs,count,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out));
    for (mask = 0; mask < (1 << Gia_ManCiNum(out)); ++mask)
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
    int signs, cs, onlyRegs, count;
    for (signs = 0; signs < 8; ++signs)
    {
        Gia_Man_t *g = source_graph(signs), *out = Gia_ManDupSelf(g);
        record(out,0,signs,0,0,0);
        Gia_ManStop(out); Gia_ManStop(g);
    }
    for (signs = 0; signs < 8; ++signs)
        for (cs = 0; cs < 8; ++cs)
        {
            Gia_Man_t *g = source_graph(signs), *inv = cone_graph(3,cs);
            Gia_Man_t *out = Gia_ManDupInvMiter(g,inv);
            record(out,1,signs,cs,0,0);
            Gia_ManStop(out); Gia_ManStop(g); Gia_ManStop(inv);
        }
    for (signs = 0; signs < 8; ++signs)
        for (cs = 0; cs < 8; ++cs)
            for (onlyRegs = 0; onlyRegs <= 1; ++onlyRegs)
                for (count = 0; count <= 2; ++count)
                {
                    Gia_Man_t *g = source_graph(signs), *cone = cone_graph(onlyRegs ? 3 : 6,cs);
                    Gia_Man_t *cones[] = {cone,cone,NULL};
                    Gia_Man_t *out = Gia_ManDupAppendCones(g,cones,count,onlyRegs);
                    record(out,2,signs,cs,onlyRegs,count);
                    Gia_ManStop(out); Gia_ManStop(g); Gia_ManStop(cone);
                }
    for (signs = 0; signs < 8; ++signs)
    {
        Gia_Man_t *g = cone_graph(3,signs), *cones[] = {g,g};
        Gia_Man_t *out = Gia_ManDupAppendCones(g,cones,2,0);
        record(out,3,signs,signs,0,2);
        Gia_ManStop(out); Gia_ManStop(g);
    }
    return 0;
}

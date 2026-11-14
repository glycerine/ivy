#include "aig/gia/gia.h"

/* Equivalent ordinary-AND graphs omit physical buffers: WithConstraints
   copies AND nodes with AppendAnd, whose fanins must be distinct. */
static Gia_Man_t *source_graph(int signs, int registers)
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
    if (Gia_ManPoNum(g)) g->nConstrs = 1;
    return g;
}

static Gia_Man_t *property_graph(int signs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g);
    int x = Gia_ManAppendXor(g,a^(signs&1),b);
    int buffer = x^((signs>>1)&1);
    int m = Gia_ManAppendMux(g,c^((signs>>2)&1),buffer,a^1);
    Gia_ManAppendCo(g,m);
    Gia_ManAppendCo(g,x^1);
    Gia_ManAppendAnd(g,b,c);
    return g;
}

static void record(Gia_Man_t *out, int mode, int signs, int parameter, int registers)
{
    Gia_Obj_t *obj;
    int mask, i, packed;
    int *values = ABC_CALLOC(int,out->nObjs);
    printf("%d %d %d %d %d %d %d %d",mode,signs,parameter,registers,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out),out->nConstrs);
    for (mask = 0; mask < 64; ++mask)
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
    int signs, parameter, registers, i;
    for (signs = 0; signs < 8; ++signs)
        for (parameter = 0; parameter < 8; ++parameter)
        {
            Gia_Man_t *g = source_graph(signs,3), *p = property_graph(parameter);
            Gia_Man_t *out = Gia_ManDupWithNewPo(g,p);
            record(out,0,signs,parameter,3);
            Gia_ManStop(out); Gia_ManStop(g); Gia_ManStop(p);
        }
    for (signs = 0; signs < 8; ++signs)
        for (registers = 0; registers <= 6; registers += 3)
            for (parameter = 0; parameter < (1 << (6-registers)); ++parameter)
            {
                Gia_Man_t *g = source_graph(signs,registers), *out;
                Vec_Int_t *types = Vec_IntAlloc(6-registers);
                for (i = 0; i < 6-registers; ++i) Vec_IntPush(types,(parameter>>i)&1);
                out = Gia_ManDupWithConstraints(g,types);
                record(out,1,signs,parameter,registers);
                Vec_IntFree(types); Gia_ManStop(out); Gia_ManStop(g);
            }
    return 0;
}

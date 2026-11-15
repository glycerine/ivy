#include "aig/gia/gia.h"

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


static void emit(int mode, int signs, int index, int regs, int first, int last, Vec_Int_t *selected)
{
    Gia_Man_t *g = graph(signs,0), *out;
    Gia_Obj_t *obj;
    int i, mask, packed, *values;
    Gia_ManAppendCi(g); Gia_ManSetRegNum(g,regs); g->nConstrs = 1;
    if (mode == 0) out = Gia_ManDupOutputGroup(g,first,last);
    else if (mode == 1) out = Gia_ManDupSelectedOutputs(g,selected);
    else
    {
        Vec_Int_t *present = Vec_IntStart(6);
        for (i = 0; i < Vec_IntSize(selected); ++i)
        {
            int co = Vec_IntEntry(selected,i);
            Vec_IntWriteEntry(present,co,co%2 ? 1 : -7);
        }
        out = Gia_ManDupOutputVec(g,present); Vec_IntFree(present);
    }
    printf("%d %d %d %d %d %d %d",mode,signs,index,regs,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out));
    Gia_ManForEachCi(g,obj,i) printf(" %u",obj->Value);
    Gia_ManForEachCo(g,obj,i) printf(" %u",obj->Value);
    values = ABC_CALLOC(int,out->nObjs);
    for (mask = 0; mask < (1<<Gia_ManCiNum(out)); ++mask)
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

int main(void)
{
    const int extra[3][4] = {{5,0,0,3},{3,1,5,0},{5,5,5,5}};
    Vec_Int_t *selected = Vec_IntAlloc(6);
    int signs, index, regs, first, last, mask, mode, i;
    for (signs = 0; signs < 32; ++signs)
    {
        index = 0;
        for (regs = 0; regs <= 4; regs += 2)
            for (first = 0; first <= 6; ++first)
                for (last = first; last <= 6; ++last)
                    emit(0,signs,index++,regs,first,last,selected);
        for (mask = 0; mask < 64; ++mask)
        {
            Vec_IntClear(selected);
            for (i = 0; i < 6; ++i) if (mask&(1<<i)) Vec_IntPush(selected,i);
            emit(1,signs,index++,0,0,0,selected);
            Vec_IntClear(selected);
            for (i = 5; i >= 0; --i) if (mask&(1<<i)) Vec_IntPush(selected,i);
            emit(1,signs,index++,0,0,0,selected);
        }
        for (mask = 0; mask < 64; ++mask)
        {
            Vec_IntClear(selected);
            for (i = 0; i < 6; ++i) if (mask&(1<<i)) Vec_IntPush(selected,i);
            emit(2,signs,index++,0,0,0,selected);
        }
        for (mode = 0; mode < 3; ++mode)
        {
            Vec_IntClear(selected);
            for (i = 0; i < 4; ++i) Vec_IntPush(selected,extra[mode][i]);
            emit(1,signs,index++,0,0,0,selected);
        }
    }
    Vec_IntFree(selected);
    return 0;
}

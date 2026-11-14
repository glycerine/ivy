#include "aig/gia/gia.h"

extern Gia_Man_t *Gia_ManDupPiPerm(Gia_Man_t *);

static Gia_Man_t *graph(int mode, int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    if (mode == 0)
    {
        int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g);
        int x = Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
        int y, z, w, q, duplicate, r;
        Gia_ManAppendCo(g,x);
        y = Gia_ManAppendAnd(g,c^((signs>>2)&1),d);
        z = Gia_ManAppendAnd(g,x^((signs>>3)&1),y);
        w = Gia_ManAppendAnd(g,0,a);
        q = Gia_ManAppendAnd(g,z^((signs>>4)&1),w^1);
        duplicate = Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
        r = Gia_ManAppendAnd(g,x,a^1);
        Gia_ManAppendCo(g,q); Gia_ManAppendCo(g,z); Gia_ManAppendCo(g,w);
        Gia_ManAppendCo(g,a^1); Gia_ManAppendCo(g,duplicate); Gia_ManAppendCo(g,r);
        Gia_ManSetRegNum(g,registers);
    }
    else if (mode == 1 || mode == 2)
    {
        int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g);
        Gia_ManAppendCo(g,mode == 1 ? Gia_ManAppendAnd(g,a,b^1) : Gia_ManAppendAnd(g,0,a));
        if (mode == 2) Gia_ManAppendCo(g,b);
    }
    else if (mode == 3) { Gia_ManAppendCo(g,0); Gia_ManAppendCo(g,1); }
    return g;
}

static void run(int mode, int signs, int registers)
{
    Gia_Man_t *src = graph(mode,signs,registers), *out;
    Gia_Obj_t *obj;
    int *values, i, mask, packed;
    unsigned next;
    Gia_ManRandom(0); Gia_ManRandomW(0);
    out = Gia_ManDupPiPerm(src);
    next = Gia_ManRandom(0);
    printf("%d %d %d %d %d %d %d %u",mode,signs,registers,Gia_ManObjNum(out),Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out),next);
    values = ABC_CALLOC(int,out->nObjs);
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
    ABC_FREE(values); Gia_ManStop(out); Gia_ManStop(src);
}

int main(void)
{
    int signs, registers, mode;
    for (signs = 0; signs < 32; ++signs)
        for (registers = 0; registers <= 4; registers += 2)
            run(0,signs,registers);
    for (mode = 1; mode <= 4; ++mode) run(mode,0,0);
    return 0;
}

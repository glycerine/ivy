#include "aig/gia/gia.h"

extern Gia_Man_t *Gia_ManDupSpreadFlop(Gia_Man_t *, Vec_Int_t *);

static Gia_Man_t *graph(int signs)
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

static void run(int mode, int signs, int p1, int p2, int *order, int length)
{
    Gia_Man_t *src = graph(signs), *out = NULL;
    Vec_Int_t *v = Vec_IntAlloc(length);
    Gia_Obj_t *obj;
    int i, mask, packed;
    int *values;
    for (i = 0; i < length; ++i) Vec_IntPush(v,order[i]);
    switch (mode)
    {
    case 0: out = Gia_ManDupPerm(src,v); break;
    case 1: out = Gia_ManDupPermFlop(src,v); break;
    case 2: out = Gia_ManDupPermFlopGap(src,v); break;
    case 3: out = Gia_ManDupSpreadFlop(src,v); break;
    case 4: out = Gia_ManPermuteInputs(src,p1,p2); break;
    case 5: out = Gia_ManDupZero(src); break;
    }
    values = ABC_CALLOC(int,out->nObjs);
    printf("%d %d %d %d %d %d %d",mode,signs,p1,p2,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out));
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
    ABC_FREE(values); Vec_IntFree(v); Gia_ManStop(out); Gia_ManStop(src);
}

int main(void)
{
    int perms[6][3] = {{0,1,2},{0,2,1},{1,0,2},{1,2,0},{2,0,1},{2,1,0}};
    int signs, p, pattern, i, count, spread[5], mask[5], index, ppis, extra;
    for (signs = 0; signs < 8; ++signs)
    {
        for (p = 0; p < 6; ++p)
        {
            run(0,signs,p,0,perms[p],3);
            run(1,signs,p,0,perms[p],3);
        }
        for (pattern = 0; pattern < 32; ++pattern)
        {
            count = 0;
            for (i = 0; i < 5; ++i) count += (pattern>>i)&1;
            if (count != 3) continue;
            for (i = 0; i < 5; ++i) spread[i] = (pattern&(1<<i)) ? (i%2 ? 99-i : -2-i) : -1;
            run(3,signs,pattern,0,spread,5);
            for (p = 0; p < 6; ++p)
            {
                index = 0;
                for (i = 0; i < 5; ++i) mask[i] = spread[i] == -1 ? -1 : perms[p][index++];
                run(2,signs,pattern,p,mask,5);
            }
        }
        for (ppis = 0; ppis <= 3; ++ppis)
            for (extra = 0; ppis+extra <= 3; ++extra)
                run(4,signs,ppis,extra,NULL,0);
        run(5,signs,0,0,NULL,0);
    }
    return 0;
}

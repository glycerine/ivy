#include "aig/gia/gia.h"

static Gia_Man_t *copies_graph(int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), ro[2], i, x, buffer, mux;
    int control = a, other = b;
    for (i = 0; i < registers; ++i) ro[i] = Gia_ManAppendCi(g);
    if (registers) control = ro[0];
    if (registers > 1) other = ro[registers-1];
    x = Gia_ManAppendXor(g,a^(signs&1),b);
    buffer = x^((signs>>1)&1);
    mux = Gia_ManAppendMux(g,control^((signs>>2)&1),buffer,other);
    Gia_ManAppendCo(g,mux);
    Gia_ManAppendCo(g,a^1);
    for (i = 0; i < registers; ++i)
        Gia_ManAppendCo(g,Gia_ManAppendXor(g,ro[i],mux));
    Gia_ManSetRegNum(g,registers);
    return g;
}

static Gia_Man_t *classes_graph(int signs)
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

static int evaluate(Gia_Man_t *g, int mask)
{
    Gia_Obj_t *obj;
    int i, packed = 0;
    int *values = ABC_CALLOC(int,g->nObjs);
    Gia_ManForEachCi(g,obj,i) values[Gia_ObjId(g,obj)] = (mask>>i)&1;
    Gia_ManForEachAnd(g,obj,i)
        values[i] = (values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
    Gia_ManForEachCo(g,obj,i)
        packed |= (values[Gia_ObjFaninId0p(g,obj)]^Gia_ObjFaninC0(obj)) << i;
    ABC_FREE(values);
    return packed;
}

static void record(Gia_Man_t *g, int mode, int signs, int p1, int p2)
{
    int mask;
    printf("%d %d %d %d %d %d %d",mode,signs,p1,p2,Gia_ManCiNum(g),Gia_ManCoNum(g),Gia_ManRegNum(g));
    for (mask = 0; mask < (1 << Gia_ManCiNum(g)); ++mask)
        printf(" %d",evaluate(g,mask));
    puts("");
}

int main(int argc, char **argv)
{
    int signs, regs, times, pattern, selected, i;
    int labels[] = {-2,5,9}, selections[] = {-2,5,9,77};
    if (argc > 1 && !strcmp(argv[1],"--dfs-order-witness"))
    {
        Gia_Man_t *g = Gia_ManStart(1), *dfs, *out;
        int b;
        Gia_ManAppendCi(g);
        b = Gia_ManAppendCi(g);
        Gia_ManAppendCo(g,b);
        dfs = Gia_ManDupOrderDfs(g);
        out = Gia_ManDupTimes(dfs,2);
        printf("%d\n",evaluate(out,1));
        Gia_ManStop(out); Gia_ManStop(dfs); Gia_ManStop(g);
        return 0;
    }
    for (signs = 0; signs < 8; ++signs)
        for (regs = 0; regs <= 2; ++regs)
            for (times = 1; times <= 3; ++times)
            {
                Gia_Man_t *g = copies_graph(signs,regs), *out = Gia_ManDupTimes(g,times);
                record(out,0,signs,times,regs);
                Gia_ManStop(out); Gia_ManStop(g);
            }
    for (signs = 0; signs < 8; ++signs)
        for (pattern = 0; pattern < 27; ++pattern)
            for (selected = 0; selected < 4; ++selected)
            {
                Gia_Man_t *g = classes_graph(signs), *out;
                int p = pattern;
                g->vFlopClasses = Vec_IntAlloc(3);
                for (i = 0; i < 3; ++i) { Vec_IntPush(g->vFlopClasses,labels[p%3]); p /= 3; }
                out = Gia_ManDupFlopClass(g,selections[selected]);
                record(out,1,signs,pattern,selected);
                Gia_ManStop(out); Gia_ManStop(g);
            }
    return 0;
}

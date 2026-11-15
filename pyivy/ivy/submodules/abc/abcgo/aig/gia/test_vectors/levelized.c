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

int main(int argc, char **argv)
{
    int signs, registers, i, mask, packed;
    Gia_Obj_t *obj;
    if (argc > 1 && !strcmp(argv[1],"--constraints"))
    {
        Gia_Man_t *g = graph(0,0), *level, *dfs;
        g->nConstrs = 1;
        level = Gia_ManDupLevelized(g); dfs = Gia_ManDupOrderDfs(g);
        printf("%d %d %d\n",g->nConstrs,level->nConstrs,dfs->nConstrs);
        Gia_ManStop(level); Gia_ManStop(dfs); Gia_ManStop(g);
        return 0;
    }
    for (signs = 0; signs < 32; ++signs)
        for (registers = 0; registers <= 4; registers += 2)
        {
            Gia_Man_t *g = graph(signs,registers), *out = Gia_ManDupLevelized(g);
            int *values = ABC_CALLOC(int,out->nObjs);
            printf("%d %d %d %d %d %d",signs,registers,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out));
            Gia_ManForEachAnd(g,obj,i) printf(" %u",obj->Value);
            for (mask = 0; mask < 16; ++mask)
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
            ABC_FREE(values); Gia_ManStop(out); Gia_ManStop(g);
        }
    return 0;
}

#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupDfs2(Gia_Man_t *);

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


extern int Gia_ManDupDfs2_rec(Gia_Man_t *, Gia_Man_t *, Gia_Obj_t *);

static void values(Gia_Man_t *g, Gia_Man_t *out, int rows)
{
    Gia_Obj_t *obj;
    int i, mask, packed, *data = ABC_CALLOC(int,out->nObjs);
    Gia_ManForEachObj(g,obj,i) printf(" %u",obj->Value);
    for (mask = 0; mask < rows; ++mask)
    {
        Gia_ManForEachCi(out,obj,i) data[Gia_ObjId(out,obj)] = (mask>>i)&1;
        Gia_ManForEachAnd(out,obj,i)
            data[i] = (data[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (data[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
        packed = 0;
        Gia_ManForEachCo(out,obj,i)
            packed |= (data[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj)) << i;
        printf(" %d",packed);
    }
    puts(""); ABC_FREE(data);
}

static Gia_Man_t *representatives(int signs, int regs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), x,y,z,u,zero,one,i;
    Gia_ManAppendCi(g); Gia_ManAppendCi(g);
    x = Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    y = Gia_ManAppendAnd(g,x^1,1); z = Gia_ManAppendAnd(g,y,1);
    u = Gia_ManAppendAnd(g,(a^(signs&1))^1,b^((signs>>1)&1));
    zero = Gia_ManAppendAnd(g,x,u); one = Gia_ManAppendAnd(g,zero^1,1);
    Gia_ManAppendCo(g,x^((signs>>2)&1)); Gia_ManAppendCo(g,y^((signs>>3)&1));
    Gia_ManAppendCo(g,z^((signs>>4)&1)); Gia_ManAppendCo(g,zero); Gia_ManAppendCo(g,one);
    Gia_ManSetRegNum(g,regs); Gia_ManSetPhase(g);
    g->pReprsOld = ABC_ALLOC(int,g->nObjs);
    for (i = 0; i < g->nObjs; ++i) g->pReprsOld[i] = -1;
    g->pReprsOld[y>>1] = x>>1; g->pReprsOld[z>>1] = y>>1;
    g->pReprsOld[zero>>1] = 0; g->pReprsOld[one>>1] = 0;
    return g;
}

int main(int argc, char **argv)
{
    int mode, signs, regs, i;
    Gia_Obj_t *obj;
    if (argc != 2) return 1;
    if (!strcmp(argv[1],"constraints"))
    {
        Gia_Man_t *g = graph(0,2), *first, *second;
        g->nConstrs = 1; first = Gia_ManDupDfs(g); second = Gia_ManDupDfs2(g);
        printf("%d %d %d\n",g->nConstrs,first->nConstrs,second->nConstrs);
        Gia_ManStop(first); Gia_ManStop(second); Gia_ManStop(g); return 0;
    }
    for (mode = 0; mode < 2; ++mode)
        for (signs = 0; signs < 32; ++signs)
            for (regs = 0; regs <= 4; regs += 2)
            {
                Gia_Man_t *g, *out;
                if (!strcmp(argv[1],"full"))
                {
                    g = graph(signs,0); Gia_ManAppendCi(g); Gia_ManSetRegNum(g,regs);
                    out = mode ? Gia_ManDupDfs2(g) : Gia_ManDupDfs(g);
                    printf("%d %d %d %d %d %d %d",mode,signs,regs,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
                    values(g,out,32);
                }
                else if (!strcmp(argv[1],"representatives"))
                {
                    g = representatives(signs,regs);
                    if (!mode) out = Gia_ManDupDfs2(g);
                    else
                    {
                        Gia_ManFillValue(g); out = Gia_ManStart(g->nObjs); Gia_ManConst0(g)->Value = 0;
                        Gia_ManForEachCi(g,obj,i) obj->Value = Gia_ManAppendCi(out);
                        Gia_ManHashAlloc(out);
                        Gia_ManForEachCo(g,obj,i) Gia_ManDupDfs2_rec(out,g,obj);
                        Gia_ManHashStop(out); Gia_ManSetRegNum(out,regs);
                    }
                    printf("%d %d %d %d",mode,signs,regs,out->nObjs); values(g,out,16);
                }
                else return 1;
                Gia_ManStop(out); Gia_ManStop(g);
            }
    return 0;
}

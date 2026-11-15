#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupDfsNode(Gia_Man_t *,Gia_Obj_t *);
extern Gia_Man_t *Gia_ManDupDfsLitArray(Gia_Man_t *,Vec_Int_t *);

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


static void record(int mode, int signs, int regs, int index, Gia_Man_t *g, Gia_Man_t *out)
{
    Gia_Obj_t *obj;
    int i, mask, packed, *values = ABC_CALLOC(int,out->nObjs);
    printf("%d %d %d %d %d %d %d",mode,signs,regs,index,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out));
    Gia_ManForEachObj(g,obj,i) printf(" %u",obj->Value);
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
    puts(""); ABC_FREE(values);
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
    int signs, regs, mode, i, index, count, candidates[12];
    Gia_Obj_t *obj;
    if (argc > 1 && !strcmp(argv[1],"old"))
    {
        for (signs = 0; signs < 32; ++signs) for (regs = 0; regs <= 4; regs += 2)
        {
            Gia_Man_t *g = representatives(signs,regs), *out;
            Vec_Int_t *lits = Vec_IntAlloc(9);
            int mask, packed, *values;
            Gia_ManForEachCo(g,obj,i) Vec_IntPush(lits,Gia_ObjFaninLit0p(g,obj));
            Vec_IntPush(lits,0); Vec_IntPush(lits,1); Vec_IntPush(lits,Vec_IntEntry(lits,2)^1); Vec_IntPush(lits,Vec_IntEntry(lits,0));
            out = Gia_ManDupDfsLitArray(g,lits);
            printf("%d %d %d",signs,regs,out->nObjs);
            Gia_ManForEachObj(g,obj,i) printf(" %u",obj->Value);
            values = ABC_CALLOC(int,out->nObjs);
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
            puts(""); ABC_FREE(values); Vec_IntFree(lits); Gia_ManStop(out); Gia_ManStop(g);
        }
        return 0;
    }
    for (signs = 0; signs < 32; ++signs) for (regs = 0; regs <= 4; regs += 2)
        for (mode = 0; mode < 2; ++mode)
        {
            Gia_Man_t *g = graph(signs,0), *out;
            Gia_ManAppendCi(g); Gia_ManSetRegNum(g,regs); g->nConstrs = 1;
            count = 0; candidates[count++] = 0;
            Gia_ManForEachCi(g,obj,i) candidates[count++] = Gia_ObjId(g,obj);
            Gia_ManForEachAnd(g,obj,i) candidates[count++] = i;
            if (!mode)
                for (index = 0; index < 6; ++index)
                {
                    out = Gia_ManDupDfsNode(g,Gia_ManObj(g,candidates[6+index]));
                    record(mode,signs,regs,index,g,out); Gia_ManStop(out);
                }
            else
                for (index = 0; index < 2*count+4; ++index)
                {
                    Vec_Int_t *lits = Vec_IntAlloc(6);
                    int tokens[6], size = 0, last = (count-1)<<1;
                    if (index < 2*count) { tokens[0] = index; size = 1; }
                    else if (index == 2*count+1)
                    {
                        tokens[0] = 1; tokens[1] = 10; tokens[2] = last|1;
                        tokens[3] = 12; tokens[4] = last; tokens[5] = 3; size = 6;
                    }
                    else if (index == 2*count+2)
                    { tokens[0] = last; tokens[1] = last|1; tokens[2] = last; size = 3; }
                    else if (index == 2*count+3)
                    { tokens[0] = 12; tokens[1] = 14; tokens[2] = 13; size = 3; }
                    for (i = 0; i < size; ++i) Vec_IntPush(lits,(candidates[tokens[i]>>1]<<1)|(tokens[i]&1));
                    out = Gia_ManDupDfsLitArray(g,lits);
                    record(mode,signs,regs,index,g,out); Vec_IntFree(lits); Gia_ManStop(out);
                }
            Gia_ManStop(g);
        }
    return 0;
}

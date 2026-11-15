#include "aig/gia/gia.h"
extern Gia_Man_t *Gia_ManDupAbs(Gia_Man_t *, Vec_Int_t *, Vec_Int_t *);

static Gia_Man_t *graph(int signs)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), r0 = Gia_ManAppendCi(g), r1 = Gia_ManAppendCi(g), r2 = Gia_ManAppendCi(g);
    int x = Gia_ManAppendAnd(g,a^(signs&1),r0^((signs>>1)&1));
    int y = Gia_ManAppendAnd(g,b^((signs>>2)&1),r1);
    int z = Gia_ManAppendAnd(g,x,y), u;
    Gia_ManAppendCo(g,z^((signs>>3)&1));
    u = Gia_ManAppendAnd(g,r2^((signs>>4)&1),a);
    Gia_ManAppendCo(g,u^1); Gia_ManAppendCo(g,1);
    Gia_ManAppendCo(g,r1^1); Gia_ManAppendCo(g,x);
    Gia_ManAppendCo(g,Gia_ManAppendAnd(g,u,b));
    Gia_ManAppendAnd(g,a,b);
    Gia_ManSetRegNum(g,3);
    g->pName = Abc_UtilStrsav("abstract"); g->pSpec = Abc_UtilStrsav("abstract.aig");
    return g;
}

static void cases(int signs, int *index, int *promoted, int count, int used)
{
    Gia_Man_t *g = graph(signs), *out;
    Vec_Int_t *forward = Vec_IntAlloc(3), *inverse = Vec_IntStartFull(3);
    Gia_Obj_t *obj;
    int i, mask, packed, reg, *values;
    for (i = 0; i < count; ++i)
    {
        Vec_IntPush(forward,promoted[i]);
        Vec_IntWriteEntry(inverse,promoted[i],i);
    }
    out = Gia_ManDupAbs(g,forward,inverse);
    printf("%d %d %d %d %d %d %d",signs,(*index)++,out->nObjs,Gia_ManCiNum(out),Gia_ManPiNum(out),Gia_ManCoNum(out),out->nRegs);
    values = ABC_CALLOC(int,out->nObjs);
    for (mask = 0; mask < 32; ++mask)
    {
        Gia_ManForEachCi(out,obj,i) values[Gia_ObjId(out,obj)] = (mask>>i)&1;
        Gia_ManForEachAnd(out,obj,i)
            values[i] = (values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
        packed = 0;
        Gia_ManForEachCo(out,obj,i)
            packed |= (values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj)) << i;
        printf(" %d",packed);
    }
    puts(""); ABC_FREE(values); Gia_ManStop(out); Gia_ManStop(g); Vec_IntFree(forward); Vec_IntFree(inverse);
    for (reg = 0; reg < 3; ++reg)
        if (!(used & (1<<reg)))
        {
            promoted[count] = reg;
            cases(signs,index,promoted,count+1,used|(1<<reg));
        }
}

int main(void)
{
    int signs, index, promoted[3];
    for (signs = 0; signs < 32; ++signs)
    {
        index = 0;
        cases(signs,&index,promoted,0,0);
    }
    return 0;
}

#include "aig/gia/gia.h"

static Gia_Man_t *graph(int signs, int registers)
{
    Gia_Man_t *g = Gia_ManStart(1);
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g), d = Gia_ManAppendCi(g);
    int x = Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    int z, y, r;
    Gia_ManAppendCo(g,x^1);
    z = Gia_ManAppendAnd(g,x,c);
    y = Gia_ManAppendAnd(g,c^((signs>>2)&1),d^((signs>>3)&1));
    r = Gia_ManAppendAnd(g,z,y);
    Gia_ManAppendAnd(g,a,c);
    Gia_ManAppendCo(g,r^1); Gia_ManAppendCo(g,a); Gia_ManAppendCo(g,z^1); Gia_ManAppendCo(g,1);
    Gia_ManSetRegNum(g,registers);
    g->pName = Abc_UtilStrsav("slice"); g->pSpec = Abc_UtilStrsav("original.aig");
    return g;
}

int main(void)
{
    int signs, registers, limit, i, mask, marks, packed;
    Gia_Obj_t *obj;
    for (signs = 0; signs < 16; ++signs)
        for (registers = 0; registers <= 4; registers += 2)
            for (limit = -1; limit <= 5; ++limit)
            {
                Gia_Man_t *g = graph(signs,registers), *out;
                int *values;
                Gia_ManForEachObj(g,obj,i) { obj->fMark0 = (i+1)&1; obj->fMark1 = 1; }
                out = Gia_ManDupSliced(g,limit);
                marks = 0;
                Gia_ManForEachObj(g,obj,i) marks += obj->fMark0 || obj->fMark1;
                printf("%d %d %d %d %d %d %d %d %d %d",signs,registers,limit,Gia_ManObjNum(out),Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out),marks,out->pName && !strcmp(out->pName,"slice"),out->pSpec != NULL);
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
                puts("");
                ABC_FREE(values); Gia_ManStop(out); Gia_ManStop(g);
            }
    return 0;
}

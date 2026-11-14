#include "aig/gia/gia.h"

int main(void)
{
    int signs, mask, i, packed;
    for (signs = 0; signs < 16; ++signs)
    {
        Gia_Man_t *src = Gia_ManStart(1), *dst = Gia_ManStart(1);
        int a = Gia_ManAppendCi(src), b = Gia_ManAppendCi(src), c = Gia_ManAppendCi(src);
        int m = Gia_ManAppendMux(src,c^((signs>>2)&1),b^(signs&1),a^((signs>>1)&1));
        int d = Gia_ManAppendCi(dst);
        Gia_Obj_t *obj;
        int *values;
        Gia_ManAppendCo(src,m^((signs>>3)&1));
        Gia_ManAppendCo(src,Gia_ManAppendAnd(src,m,a^1));
        Gia_ManAppendCo(dst,d^1);
        Gia_ManDupAppend(dst,src);
        values = ABC_CALLOC(int,dst->nObjs);
        printf("%d %d %d",signs,Gia_ManCiNum(dst),Gia_ManCoNum(dst));
        for (mask = 0; mask < 16; ++mask)
        {
            Gia_ManForEachCi(dst,obj,i) values[Gia_ObjId(dst,obj)] = (mask>>i)&1;
            Gia_ManForEachAnd(dst,obj,i)
                values[i] = (values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed = 0;
            Gia_ManForEachCo(dst,obj,i)
                packed |= (values[Gia_ObjFaninId0p(dst,obj)]^Gia_ObjFaninC0(obj)) << i;
            printf(" %d",packed);
        }
        puts("");
        ABC_FREE(values);
        Gia_ManStop(src); Gia_ManStop(dst);
    }
    return 0;
}

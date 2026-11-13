#include "aig/gia/gia.h"
#include <stdio.h>
int main(void) {
 Gia_Man_t *p=Gia_ManStart(32), *q=Gia_ManStart(32); Gia_Obj_t *o; int i;
 int a=Gia_ManAppendCi(p), b=Gia_ManAppendCi(p); int x=Gia_ManAppendXorReal(p,a,b);
 Gia_ObjSetValue(Gia_ManConst0(p),0);
 Gia_ManForEachCi(p,o,i) Gia_ObjSetValue(o,Gia_ManAppendCi(q));
 Gia_ManHashAlloc(q);
 Gia_ManForEachAnd(p,o,i) Gia_ObjSetValue(o,Gia_ManHashAnd(q,Gia_ObjFanin0Copy(o),Gia_ObjFanin1Copy(o)));
 int expected=Gia_ManHashAnd(q,Gia_ObjValue(Gia_ManObj(p,a>>1)),Gia_ObjValue(Gia_ManObj(p,b>>1)));
 printf("native_xor_copy=%d and=%d\n", Gia_ObjValue(Gia_ManObj(p,x>>1)),expected);
 return 0;
}

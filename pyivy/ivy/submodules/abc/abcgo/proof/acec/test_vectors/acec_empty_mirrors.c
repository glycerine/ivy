#include <stdio.h>
#include "aig/gia/gia.h"
extern Gia_Man_t *Acec_DetectXorBuildNew(Gia_Man_t*,Vec_Int_t*);
int main(void){Gia_Man_t*p=Gia_ManStart(16);int a=Gia_ManAppendCi(p),b=Gia_ManAppendCi(p);int x=Gia_ManAppendAnd(p,a,b);Gia_ManAppendCo(p,x);Vec_Int_t*r=Vec_IntAlloc(0);Gia_Man_t*n=Acec_DetectXorBuildNew(p,r);Gia_Obj_t*o=Gia_ManCo(n,0);printf("driver=%d complement=%d\n",Gia_ObjFaninId0p(n,o),Gia_ObjFaninC0(o));return 0;}

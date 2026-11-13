#include <stdio.h>
#include "aig/gia/gia.h"
extern void Gia_AcecMarkHadd(Gia_Man_t *,int *);
int main(void) {Gia_Man_t*p=Gia_ManStart(16);int a=Gia_ManAppendCi(p),b=Gia_ManAppendCi(p),x=Gia_ManAppendXorReal(p,a,b),carry=Gia_ManAppendAnd(p,a,b);int sigs[]={Abc_Lit2Var(x),Abc_Lit2Var(carry)};Gia_AcecMarkHadd(p,sigs);for(int i=0;i<Gia_ManObjNum(p);i++){Gia_Obj_t*o=Gia_ManObj(p,i);printf("%d:%d%d ",i,o->fMark0,o->fMark1);}puts("");return 0;}

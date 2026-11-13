#include <stdio.h>
#include "aig/gia/gia.h"
#include "misc/tim/tim.h"
extern Gia_Man_t *Gia_ManDupWithArtificalFaddBoxes(Gia_Man_t *,int,int);
extern int Gia_ManFindAnnotatedDelay(Gia_Man_t *,int,int *,int);
int main(void){for(int inv=0;inv<2;inv++)for(int fan=0;fan<2;fan++)for(int trick=0;trick<2;trick++){
 Gia_Man_t*g=Gia_ManStart(32);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),d=Gia_ManAppendCi(g);int x=Gia_ManAppendAnd(g,a,b),y=Gia_ManAppendAnd(g,x^inv,c),z=Gia_ManAppendAnd(g,y,d);Gia_ManAppendCo(g,z);Gia_ManAppendCo(g,y);
 Gia_Obj_t *yo=Gia_ManObj(g,Abc_Lit2Var(y)),*zo=Gia_ManObj(g,Abc_Lit2Var(z));if(Gia_ObjFaninId0(yo,Abc_Lit2Var(y))==Abc_Lit2Var(x))yo->fMark0=1;else yo->fMark1=1;if(Gia_ObjFaninId0(zo,Abc_Lit2Var(z))==Abc_Lit2Var(y))zo->fMark0=1;else zo->fMark1=1;
 Gia_Man_t*p=Gia_ManDupWithArtificalFaddBoxes(g,fan,trick);p->vLevels=Vec_IntStart(Gia_ManObjNum(p));
 int count0,count1;int delay0=Gia_ManFindAnnotatedDelay(p,10,&count0,0),delay1=Gia_ManFindAnnotatedDelay(p,10,&count1,1);
 printf("%d %d %d: ci=%d co=%d boxes=%d delay=%d/%d ignore=%d/%d\n",inv,fan,trick,Gia_ManCiNum(p),Gia_ManCoNum(p),Gia_ManBoxNum(p),delay0,count0,delay1,count1);
 Gia_ManStop(p);Gia_ManStop(g);
 }return 0;}

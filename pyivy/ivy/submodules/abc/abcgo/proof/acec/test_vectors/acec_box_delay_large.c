#include <stdio.h>
#include "aig/gia/gia.h"
#include "misc/tim/tim.h"
extern Gia_Man_t *Gia_ManDupWithArtificalFaddBoxes(Gia_Man_t *,int,int);
extern int Gia_ManFindAnnotatedDelay(Gia_Man_t *,int,int *,int);
int main(void){for(int fan=0;fan<2;fan++)for(int trick=0;trick<2;trick++){
 Gia_Man_t*g=Gia_ManStart(128);int in[9];for(int i=0;i<9;i++)in[i]=Gia_ManAppendCi(g);int prior=Gia_ManAppendAnd(g,in[0],in[1]),penultimate=prior;
 for(int i=2;i<9;i++){int node=Gia_ManAppendAnd(g,prior,in[i]);Gia_Obj_t*o=Gia_ManObj(g,Abc_Lit2Var(node));if(Gia_ObjFaninId0(o,Abc_Lit2Var(node))==Abc_Lit2Var(prior))o->fMark0=1;else o->fMark1=1;penultimate=prior;prior=node;}
 Gia_ManAppendCo(g,prior);Gia_ManAppendCo(g,penultimate);
 Gia_Man_t*p=Gia_ManDupWithArtificalFaddBoxes(g,fan,trick);p->vLevels=Vec_IntStart(Gia_ManObjNum(p));int count0,count1;int d0=Gia_ManFindAnnotatedDelay(p,10,&count0,0),d1=Gia_ManFindAnnotatedDelay(p,10,&count1,1);
 printf("%d %d: ci=%d co=%d boxes=%d delay=%d/%d ignore=%d/%d\n",fan,trick,Gia_ManCiNum(p),Gia_ManCoNum(p),Gia_ManBoxNum(p),d0,count0,d1,count1);Gia_ManStop(p);Gia_ManStop(g);
 }return 0;}

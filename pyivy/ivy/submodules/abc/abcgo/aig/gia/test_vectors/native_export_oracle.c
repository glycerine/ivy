#include "aig/gia/gia.h"
#include "aig/gia/giaAig.h"
#include "aig/aig/aig.h"
#include <stdio.h>
static Gia_Man_t *make(void) {
 Gia_Man_t *g=Gia_ManStart(24);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),d=Gia_ManAppendCi(g),x,m;
 g->pMuxes=ABC_CALLOC(unsigned,g->nObjsAlloc);x=Gia_ManAppendXorReal(g,a,b^1);m=Gia_ManAppendMuxReal(g,c,b,d^1);Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,m);return g;
}
int main(void){Gia_Man_t *g=make();Aig_Man_t *a=Gia_ManToAig(g,0),*s=Gia_ManToAigSkip(g,1),*d=Gia_ManToAigSimple(g);Aig_Man_t *all[3]={a,s,d};int i,j,mask,ci,vals[100];Aig_Obj_t *o;
 for(i=0;i<3;i++)for(mask=0;mask<16;mask++){ci=0;vals[0]=1;Aig_ManForEachObj(all[i],o,j){if(j==0)continue;if(Aig_ObjIsCi(o))vals[j]=(mask>>ci++)&1;else {vals[j]=vals[Aig_ObjId(Aig_ObjFanin0(o))]^Aig_ObjFaninC0(o);if(Aig_ObjIsNode(o))vals[j]&=vals[Aig_ObjId(Aig_ObjFanin1(o))]^Aig_ObjFaninC1(o);}}printf("native %d %d",i,mask);Aig_ManForEachCo(all[i],o,j)printf(" %d",vals[Aig_ObjId(o)]);puts("");}
 for(i=0;i<3;i++)Aig_ManStop(all[i]);Gia_ManStop(g);return 0;}

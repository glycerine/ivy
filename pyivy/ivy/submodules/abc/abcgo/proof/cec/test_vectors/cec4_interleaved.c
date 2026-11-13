#include "proof/cec/cecSatG2.c"
int main(void) {
 int a,b,c,x,y,i,status; Gia_Man_t *g=Gia_ManStart(16); Cec_ParFra_t pars; Cec4_Man_t *p;
 a=Gia_ManAppendCi(g);b=Gia_ManAppendCi(g);x=Gia_ManAppendAnd(g,a,b);c=Gia_ManAppendCi(g);y=Gia_ManAppendAnd(g,a,c);
 Gia_ManSetPhase(g);g->pReprs=ABC_CALLOC(Gia_Rpr_t,Gia_ManObjNum(g)); for(i=0;i<Gia_ManObjNum(g);i++)Gia_ObjSetRepr(g,i,GIA_VOID);Gia_ObjSetRepr(g,y>>1,x>>1);
 Cec4_ManSetParams(&pars);pars.jType=0;pars.nWords=1;pars.nBTLimit=100;p=Cec4_ManCreate(g,&pars);p->pNew=Cec4_ManStartNew(g);
 Gia_ManObj(g,x>>1)->Value=Gia_ManAppendAnd(p->pNew,2,4);Gia_ManObj(g,y>>1)->Value=Gia_ManAppendAnd(p->pNew,2,6);
 Cec4_ManSimAlloc(g,1);g->iPatsPi=1;Cec4_ObjSimSetInputBit(g,c>>1,1);g->iPatsPi=0;
 status=Cec4_ManSweepNode(p,y>>1,x>>1);
 printf("status=%d old_c=%d old_gate=%d pattern=",status,(int)((Cec4_ObjSim(g,c>>1)[0]>>1)&1),(int)((Cec4_ObjSim(g,x>>1)[0]>>1)&1));
 Vec_IntForEachEntry(p->vPat,a,i)printf("%s%d",i?",":"",a); puts("");return 0;
}

#include "proof/cec/cecSatG2.c"
static int unknown_case(void) {
 int i,j,h, a[6][5], cand, formula, clause, status;
 Gia_Man_t *g=Gia_ManStart(64); Cec_ParFra_t pars; Cec4_Man_t *p;
 for(i=0;i<30;i++) Gia_ManAppendCi(g);
 cand=Gia_ManAppendAnd(g,2,4)>>1;
 for(i=0;i<300;i++) Gia_ManAppendAnd(g,2,4);
 Gia_ManSetPhase(g);g->pReprs=ABC_CALLOC(Gia_Rpr_t,Gia_ManObjNum(g));
 for(i=0;i<Gia_ManObjNum(g);i++) Gia_ObjSetRepr(g,i,GIA_VOID);
 Gia_ObjSetRepr(g,cand,0);
 Cec4_ManSetParams(&pars);pars.jType=0;pars.nWords=1;pars.nBTLimit=1;
 p=Cec4_ManCreate(g,&pars);p->pNew=Cec4_ManStartNew(g);
 for(i=0;i<6;i++)for(j=0;j<5;j++)a[i][j]=2*(1+5*i+j);
 formula=1;
 for(i=0;i<6;i++) {clause=0;for(j=0;j<5;j++)clause=Gia_ManHashOr(p->pNew,clause,a[i][j]);formula=Gia_ManHashAnd(p->pNew,formula,clause);}
 for(h=0;h<5;h++)for(i=0;i<6;i++)for(j=0;j<i;j++) {clause=Gia_ManHashOr(p->pNew,a[i][h]^1,a[j][h]^1);formula=Gia_ManHashAnd(p->pNew,formula,clause);}
 Vec_IntFillExtra(&p->pNew->vCopies2,Gia_ManObjNum(p->pNew),-1);
 Gia_ManObj(g,cand)->Value=formula;Cec4_ManSimAlloc(g,1);
 status=Cec4_ManSweepNode(p,cand,0);
 printf("unknown=%d candidate=%d marked=%d\n",status,cand,Vec_BitEntry(p->vFails,cand));
 return 0;
}
static int pack_case(int cursor) {
 Gia_Man_t *g=Gia_ManStart(4);Vec_Int_t *lits=Vec_IntAlloc(1);int result;
 Gia_ManAppendCi(g);Cec4_ManSimAlloc(g,1);Vec_IntPush(lits,2);g->iPatsPi=cursor;
 result=Cec4_ManPackAddPatternTry(g,1,lits);printf("pack=%d\n",result);return 0;
}
int main(int argc, char **argv) {
 if(argc>1 && !strcmp(argv[1],"pack")) return pack_case(atoi(argv[2]));
 if(argc>1 && !strcmp(argv[1],"unknown")) return unknown_case();
 int mode=atoi(argv[1]), a,b,c,x,y,i,lit,ret;
 Gia_Man_t *g=Gia_ManStart(16); Cec_ParFra_t pars; Cec4_Man_t *p;
 a=Gia_ManAppendCi(g);b=Gia_ManAppendCi(g);c=Gia_ManAppendCi(g);
 x=Gia_ManAppendAnd(g,a,b);y=Gia_ManAppendAnd(g,a,b^1);
 Gia_ManSetPhase(g); g->pReprs=ABC_CALLOC(Gia_Rpr_t,Gia_ManObjNum(g));
 for(i=0;i<Gia_ManObjNum(g);i++) Gia_ObjSetRepr(g,i,GIA_VOID);
 Gia_ObjSetRepr(g,y>>1,x>>1);
 Cec4_ManSetParams(&pars);pars.jType=mode;pars.nWords=1;pars.nBTLimit=100;pars.fVerbose=0;
 p=Cec4_ManCreate(g,&pars);p->pNew=Cec4_ManStartNew(g);
 Gia_ManObj(g,x>>1)->Value=Gia_ManAppendAnd(p->pNew,a,b);
 Gia_ManObj(g,y>>1)->Value=Gia_ManAppendAnd(p->pNew,a,b^1);
 Cec4_ManSimAlloc(g,1); Cec4_ObjGetCnfVar(p,c>>1);
 g->iPatsPi=1; Cec4_ObjSimSetInputBit(g,c>>1,1);g->iPatsPi=0;
 ret=Cec4_ManSweepNode(p,y>>1,x>>1);
 printf("result=%d c=%d pattern=",ret,(int)((Cec4_ObjSim(g,c>>1)[0]>>1)&1));
 Vec_IntForEachEntry(p->vPat,lit,i) printf("%s%d",i?",":"",lit);
 printf("\n");return 0;
}

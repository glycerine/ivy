/* Compile from tests against the unmodified original C implementation. */
#include "aig/gia/giaSwitch.c"
#include "opt/dar/darLib.c"
#include "aig/aig/aigMffc.c"
#include <stdint.h>
static float from_bits(unsigned x) {float f;memcpy(&f,&x,4);return f;}
static unsigned to_bits(float f) {unsigned x;memcpy(&x,&f,4);return x;}
int main(void) {
 char mode;
 while(scanf(" %c",&mode)==1) {
  if(mode=='S') {
   int n,ni,no,nr,nw,nf,np,trans,factor;scanf("%d%d%d%d%d%d%d%d%d",&n,&ni,&no,&nr,&nw,&nf,&np,&trans,&factor);
   Gia_Man_t g={0};Gia_ParSwi_t pars={0};Gia_ManSwi_t p={0};
   g.nObjs=n;g.nRegs=nr;g.nFront=n;g.pObjs=calloc(n,sizeof(Gia_Obj_t));g.vCis=Vec_IntAlloc(ni);g.vCos=Vec_IntAlloc(no);
   int ci=0,co=0;
   for(int i=0;i<n;i++) {int kind,a,b;scanf("%d%d%d",&kind,&a,&b);Gia_Obj_t *o=g.pObjs+i;o->Value=i;
    if(kind==0){o->iDiff0=o->iDiff1=GIA_NONE;}
    else if(kind==1){o->fTerm=1;o->iDiff0=GIA_NONE;o->iDiff1=ci++;Vec_IntPush(g.vCis,i);}
    else if(kind==2||kind==4){o->iDiff0=a>>1;o->fCompl0=(a&1)^((a>>1)==0);o->iDiff1=kind==4?0:b>>1;o->fCompl1=kind==4?1:(b&1)^((b>>1)==0);}
    else if(kind==3){o->fTerm=1;o->Value=GIA_NONE;o->iDiff0=a>>1;o->fCompl0=(a&1)^((a>>1)==0);o->iDiff1=co++;Vec_IntPush(g.vCos,i);}
    else return 8;
   }
   pars.nWords=nw;pars.nIters=nf;pars.nPref=np;pars.fProbTrans=trans;
   p.pAig=&g;p.pPars=&pars;p.nWords=nw;p.pDataSim=calloc(n*nw,sizeof(unsigned));p.pDataSimCis=calloc(ni*nw,sizeof(unsigned));p.pDataSimCos=calloc(no*nw,sizeof(unsigned));p.pData1=calloc(n,sizeof(int));
   unsigned long long *coCounts=calloc(no,sizeof(unsigned long long));
   Gia_ManRandom(1);Gia_ManSwiSimInfoInit(&p);
   for(int frame=0;frame<nf;frame++) {
    Gia_ManSwiSimulateRound(&p,frame>=np);
    if(frame>=np)for(int i=0;i<no;i++)for(int w=0;w<nw;w++){unsigned v=p.pDataSimCos[i*nw+w];coCounts[i]+=trans?2*Gia_WordCountOnes((v^(v>>16))&0xffff):Gia_WordCountOnes(v);}
    if(frame+1<nf){if(trans)Gia_ManSwiSimInfoTransferShift(&p,factor);else Gia_ManSwiSimInfoTransfer(&p,factor);}
   }
   printf("%d",n);
   for(int i=0;i<n;i++){float prob;if(i==0)prob=trans?0:1;else if(Gia_ObjIsCo(g.pObjs+i))prob=Gia_ManSwiComputeProbOne(coCounts[Gia_ObjCioId(g.pObjs+i)],nw*(nf-np));else prob=Gia_ManSwiComputeProbOne(p.pData1[i],nw*(nf-np));printf(" %u",to_bits(prob));}puts("");
   free(coCounts);free(g.pObjs);Vec_IntFree(g.vCis);Vec_IntFree(g.vCos);free(p.pDataSim);free(p.pDataSimCis);free(p.pDataSimCos);free(p.pData1);
  } else if(mode=='E') {
   int n,root;scanf("%d%d",&n,&root);Dar_Lib_t lib={0};lib.pObjs=calloc(n,sizeof(Dar_LibObj_t));lib.pDatas=calloc(n,sizeof(Dar_LibDat_t));s_DarLib=&lib;Aig_Obj_t dummy={0};
   for(int i=0;i<n;i++){int a,b,state;unsigned prob;scanf("%d%d%d%u",&a,&b,&state,&prob);lib.pObjs[i].Num=i;lib.pObjs[i].fTerm=i<4;lib.pObjs[i].Fan0=a>>1;lib.pObjs[i].Fan1=b>>1;lib.pObjs[i].fCompl0=a&1;lib.pObjs[i].fCompl1=b&1;lib.pDatas[i].TravId=-1;lib.pDatas[i].dProb=from_bits(prob);if(state){lib.pDatas[i].pFunc=&dummy;lib.pDatas[i].fMffc=state==2;}}
   float power;int area=Dar_LibEval_rec(lib.pObjs+root,1,100000,100000,&power);printf("%d %u\n",area,to_bits(power));free(lib.pObjs);free(lib.pDatas);s_DarLib=NULL;
  } else if(mode=='M') {
   int n,root;scanf("%d%d",&n,&root);Aig_Obj_t *objects=calloc(n,sizeof(Aig_Obj_t));float *probs=calloc(n,sizeof(float));
   for(int i=0;i<n;i++){int kind,a,b,refs;unsigned prob;scanf("%d%d%d%d%u",&kind,&a,&b,&refs,&prob);objects[i].Id=i;objects[i].Type=kind==1?AIG_OBJ_CI:kind==4?AIG_OBJ_BUF:AIG_OBJ_AND;objects[i].nRefs=refs;objects[i].pFanin0=Aig_NotCond(objects+(a>>1),a&1);objects[i].pFanin1=Aig_NotCond(objects+(b>>1),b&1);probs[i]=from_bits(prob);}
   float power;int area=Aig_NodeDeref_rec(objects+root,0,&power,probs);printf("%d %u\n",area,to_bits(power));free(objects);free(probs);
  } else return 9;
 }
 return 0;
}

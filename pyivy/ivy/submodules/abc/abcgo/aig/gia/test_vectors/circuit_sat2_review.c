#include "aig/gia/giaCSat2.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static Gia_Man_t *graph(int seed,int *lits) {
 int i,a,b;unsigned random;Gia_Man_t *g;
  g=Gia_ManStart(128);Gia_ManHashStart(g);
  for(i=0;i<5;++i)lits[i]=Gia_ManAppendCi(g);
  if(seed<8){
  random=seed+17;
  for(i=5;i<29;++i){
   random=random*1664525u+1013904223u;a=lits[random%i];
   random=random*1664525u+1013904223u;a^=random>>31;
   random=random*1664525u+1013904223u;b=lits[random%i];
   random=random*1664525u+1013904223u;b^=random>>31;
   lits[i]=i%3==0?Gia_ManHashXor(g,a,b):Gia_ManHashAnd(g,a,b);
  }
  }else{
   int v[5];for(i=0;i<5;++i)v[i]=lits[i]^(((seed-8)>>i)&1);
   lits[5]=Gia_ManHashAnd(g,v[0],v[1]);lits[6]=Gia_ManHashAnd(g,v[0],v[1]^1);
   lits[7]=Gia_ManHashOr(g,lits[5],lits[6]);lits[8]=Gia_ManHashAnd(g,v[0]^1,lits[7]);
   lits[9]=Gia_ManHashXor(g,v[0],v[1]);lits[10]=Gia_ManHashXor(g,lits[9],v[2]);
   lits[11]=Gia_ManHashXor(g,v[1],v[2]);lits[12]=Gia_ManHashXor(g,v[0],lits[11]);lits[13]=Gia_ManHashAnd(g,lits[10],lits[12]^1);
   lits[14]=Gia_ManHashXor(g,v[3],v[4]);lits[15]=Gia_ManHashXor(g,lits[10],lits[14]);
   lits[16]=Gia_ManHashXor(g,v[2],lits[14]);lits[17]=Gia_ManHashXor(g,v[1],lits[16]);lits[18]=Gia_ManHashXor(g,v[0],lits[17]);
   lits[19]=Gia_ManHashAnd(g,lits[15],lits[18]^1);lits[20]=Gia_ManHashOr(g,lits[15],lits[18]^1);
   lits[21]=Gia_ManHashAnd(g,lits[13]^1,lits[8]^1);lits[22]=Gia_ManHashAnd(g,lits[9],lits[11]);
   lits[23]=Gia_ManHashXor(g,lits[22],lits[19]);lits[24]=Gia_ManHashXor(g,lits[20],lits[9]);
   lits[25]=Gia_ManHashAnd(g,lits[24],lits[9]);lits[26]=Gia_ManHashOr(g,lits[25],lits[22]);
   lits[27]=Gia_ManHashAnd(g,lits[26],lits[22]^1);lits[28]=Gia_ManHashOr(g,lits[27],lits[8]);
  }
  Gia_ManHashStop(g);return g;
}

static void miters(void) {
 int seed,bi,i,value,lits[29],budgets[]={-1,0,1,3,1000};Gia_Man_t *g;Gia_Obj_t *obj;Vec_Str_t *statuses;Vec_Int_t *store;
 for(seed=0;seed<12;++seed)for(bi=0;bi<5;++bi){
  g=graph(seed,lits);Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);
  for(i=0;i<29;++i){Gia_ManAppendCo(g,lits[i]);Gia_ManAppendCo(g,lits[i]^1);}
  if(seed&1){Vec_IntReverseOrder(g->vCos);for(i=0;i<Gia_ManCoNum(g);++i)Gia_ObjSetCioId(Gia_ManCo(g,i),i);}
  if(seed&2){Vec_IntReverseOrder(g->vCis);for(i=0;i<Gia_ManCiNum(g);++i)Gia_ObjSetCioId(Gia_ManCi(g,i),i);}
  Gia_ManForEachObj(g,obj,i)obj->fPhase=1;
  store=Cbs2_ManSolveMiterNc(g,budgets[bi],&statuses,0);
  printf("miter %d %d statuses",seed,budgets[bi]);
  for(i=0;i<Vec_StrSize(statuses);++i)printf(" %d",(unsigned char)Vec_StrEntry(statuses,i));
  printf(" store");Vec_IntForEachEntry(store,value,i)printf(" %d",value);
  printf(" drivers");Gia_ManForEachCo(g,obj,i)printf(" %d",Gia_ObjFaninLit0p(g,obj));
  printf(" phases");Gia_ManForEachObj(g,obj,i)printf("%d",obj->fPhase);printf("\n");
  Vec_IntFree(store);Vec_StrFree(statuses);Gia_ManStop(g);
 }
}

static void stats(void) {
 Gia_Man_t *g=Gia_ManStart(16);Cbs2_Man_t *p;int a=Gia_ManAppendCi(g);
 Gia_ManAppendCo(g,a);p=Cbs2_ManAlloc(g);
 p->nSatTotal=4;p->nSatUnsat=2;p->nSatSat=1;p->nSatUndec=1;p->nConfUnsat=5;p->nConfSat=3;p->nConfUndec=7;
 p->timeTotal=4*CLOCKS_PER_SEC;p->timeSatUnsat=2*CLOCKS_PER_SEC;p->timeSatSat=CLOCKS_PER_SEC;p->timeSatUndec=CLOCKS_PER_SEC/2;
 Cbs2_ManSatPrintStats(p);Cbs2_ManStop(p);p=Cbs2_ManAlloc(g);Cbs2_ManSatPrintStats(p);Cbs2_ManStop(p);Gia_ManStop(g);
}

static int learned(Cbs2_Man_t *p,int *lits,int n) {
 int h=p->pClauses.iHead,i;Cbs2_QuePush(&p->pClauses,n);
 for(i=0;i<n;++i)Cbs2_QuePush(&p->pClauses,lits[i]);
 Cbs2_QuePush(&p->pClauses,0);Cbs2_QuePush(&p->pClauses,0);p->pClauses.iHead=p->pClauses.iTail;
 Cbs2_ManWatchClause(p,h,lits[0]);Cbs2_ManWatchClause(p,h,lits[1]);return h;
}
static void snapshot(Cbs2_Man_t *p,const char *label,int h) {
 int i;printf("watch %s h%d head%d tail%d clauses",label,h,p->pClauses.iHead,p->pClauses.iTail);
 for(i=1;i<p->pClauses.iTail;++i)printf(" %d",p->pClauses.pData[i]);
 printf(" trail");for(i=0;i<p->pProp.iTail;++i)printf(" %d",p->pProp.pData[i]);
 printf(" watches");for(i=0;i<Vec_IntSize(&p->vWatches);++i)printf(" %d",Vec_IntEntry(&p->vWatches,i));
 printf(" assigns");for(i=0;i<Vec_StrSize(&p->vAssign);++i)printf(" %d",Vec_StrEntry(&p->vAssign,i));
 printf(" levels");for(i=0;i<Vec_IntSize(&p->vLevReason);++i)printf(" %d",Vec_IntEntry(&p->vLevReason,i));
 printf(" marks");for(i=0;i<Vec_StrSize(&p->vMark);++i)printf(" %d",Vec_StrEntry(&p->vMark,i));
 printf(" calls%d conflicts%d\n",p->nPropCalls[1],p->nClauseConf);
}
static void watch(void) {
 Gia_Man_t *g=Gia_ManStart(16);Cbs2_Man_t *p;int a,b,c,d,e,lits[4],h;
 a=Gia_ManAppendCi(g);b=Gia_ManAppendCi(g);c=Gia_ManAppendCi(g);d=Gia_ManAppendCi(g);e=Gia_ManAppendCi(g);
 p=Cbs2_ManAlloc(g);p->pClauses.nSize=2;p->pClauses.pData=ABC_REALLOC(int,p->pClauses.pData,2);
 lits[0]=a;lits[1]=b;lits[2]=c;lits[3]=d;learned(p,lits,4);lits[1]=e;learned(p,lits,2);
 Cbs2_ManAssign(p,a^1,0,0,0);h=Cbs2_ManPropagateClauses(p,0,a^1);snapshot(p,"a",h);
 Cbs2_ManAssign(p,b^1,1,0,0);h=Cbs2_ManPropagateClauses(p,1,b^1);snapshot(p,"b",h);
 Cbs2_ManAssign(p,c^1,2,0,0);h=Cbs2_ManPropagateClauses(p,2,c^1);snapshot(p,"unit",h);
 Cbs2_ManCancelUntil(p,4);Cbs2_ManAssign(p,d^1,3,0,0);h=Cbs2_ManPropagateClauses(p,3,c^1);snapshot(p,"conflict",h);
 Cbs2_ManPrintCube(p,3,h);Cbs2_ManPrintClause(p,3,h+Cbs2_ClauseSize(p,h)+1);
 Cbs2_ManCancelUntil(p,0);Cbs2_ManCleanWatch(p);Cbs2_ManStop(p);Gia_ManStop(g);
}

static void fanouts(void) {
 Gia_Man_t *g=Gia_ManStart(16);Cbs2_Man_t *p;int a,b,c,d,ab,r0,r1,status,i,v;
 a=Gia_ManAppendCi(g);b=Gia_ManAppendCi(g);c=Gia_ManAppendCi(g);d=Gia_ManAppendCi(g);
 ab=Gia_ManAppendAnd(g,a,b);r0=Gia_ManAppendAnd(g,ab,c);r1=Gia_ManAppendAnd(g,ab,d);
 Gia_ManCreateRefs(g);p=Cbs2_ManAlloc(g);Gia_ManIncrementTravId(g);
 Cbs2_ManCreateFanout_rec(p,r0>>1);Cbs2_ManCreateFanout_rec(p,r1>>1);
 Cbs2_ManPrintFanouts(p);printf("heads");Vec_IntForEachEntry(&p->vFanout0,v,i)printf(" %d",v);
 printf(" links");Vec_IntForEachEntry(&p->vFanoutN,v,i)printf(" %d",v);printf("\n");
 status=Cbs2_ManSolve2(p,r0,r1);printf("prepared status%d model",status);Vec_IntForEachEntry(p->vModel,v,i)printf(" %d",v);printf("\n");
 Cbs2_ManPrintFanouts(p);
 Cbs2_ManDeleteFanout_rec(p,r0>>1);Cbs2_ManDeleteFanout_rec(p,r1>>1);Cbs2_ManCheckFanouts(p);
 Cbs2_ManStop(p);Gia_ManStop(g);
}

static void verbose(void) {
 Gia_Man_t *g=Gia_ManStart(16);Vec_Str_t *status;Vec_Int_t *store;int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g);
 Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);Gia_ManAppendCo(g,Gia_ManAppendAnd(g,a,b));
 store=Cbs2_ManSolveMiterNc(g,0,&status,1);Vec_IntFree(store);Vec_StrFree(status);Gia_ManStop(g);
}

int main(int argc,char **argv) {
 if(argc>1){if(argv[1][0]=='m')miters();else if(argv[1][0]=='w')watch();else if(argv[1][0]=='f')fanouts();else if(argv[1][0]=='v')verbose();else stats();return 0;}
 setvbuf(stdout,NULL,_IONBF,0);
 int seed,fanout,bi,ji,q,phase,mode,status,k,value,lits[29];
 int budgets[]={0,1,3,1000},frontiers[]={0,1,3,100};Gia_Man_t *g;Cbs2_Man_t *p;
 for(seed=0;seed<12;++seed){
  g=graph(seed,lits);Gia_ManCreateRefs(g);
  for(fanout=0;fanout<2;++fanout)for(bi=0;bi<4;++bi)for(ji=0;ji<4;++ji){
   p=Cbs2_ManAlloc(g);p->Pars.fUseFanout=fanout;
   p->Pars.nBTLimit=budgets[bi];p->Pars.nJustLimit=frontiers[ji];
   for(q=0;q<29;++q)if(lits[q]>=2)for(phase=0;phase<2;++phase)for(mode=0;mode<3;mode+=2){
    if(mode==2 && q%5)continue;
    if(fanout){Gia_ManIncrementTravId(g);Cbs2_ManCreateFanout_rec(p,lits[q]>>1);}
    status=mode==0?Cbs2_ManSolve(p,lits[q]^phase):Cbs2_ManSolve2(p,lits[q]^phase,0);
    printf("case %d %d %d %d %d %d %d status%d bt%d nc%d just%d totals%d:%d prop%d:%d:%d fails%d:%d clauses%d model",seed,fanout,budgets[bi],frontiers[ji],q,phase,mode,status,p->Pars.nBTThis,p->Pars.nBTThisNc,p->Pars.nJustThis,p->Pars.nBTTotal,p->Pars.nJustTotal,p->nPropCalls[0],p->nPropCalls[1],p->nPropCalls[2],p->nFails[0],p->nFails[1],p->nClauseConf);
    if(status==0)Vec_IntForEachEntry(p->vModel,value,k)printf(" %d",value);printf("\n");
    if(fanout){Cbs2_ManDeleteFanout_rec(p,lits[q]>>1);Cbs2_ManCheckFanouts(p);}
   }
   Cbs2_ManStop(p);
  }
  Gia_ManStop(g);
 }
 return 0;
}

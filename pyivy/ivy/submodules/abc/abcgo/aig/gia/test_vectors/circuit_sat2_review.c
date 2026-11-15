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

int main(int argc,char **argv) {
 if(argc>1){if(argv[1][0]=='m')miters();else stats();return 0;}
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

#include "aig/gia/giaCSat.c"
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
 int seed,bi,prove,i,value,lits[29],budgets[]={-1,0,1,3,1000};Gia_Man_t *g;Gia_Obj_t *obj;Vec_Str_t *statuses;Vec_Int_t *store;
 for(seed=0;seed<12;++seed)for(bi=0;bi<5;++bi)for(prove=0;prove<2;++prove){
  g=graph(seed,lits);Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);
  for(i=0;i<29;++i){Gia_ManAppendCo(g,lits[i]);Gia_ManAppendCo(g,lits[i]^1);}
  if(seed&1){Vec_IntReverseOrder(g->vCos);for(i=0;i<Gia_ManCoNum(g);++i)Gia_ObjSetCioId(Gia_ManCo(g,i),i);}
  if(seed&2){Vec_IntReverseOrder(g->vCis);for(i=0;i<Gia_ManCiNum(g);++i)Gia_ObjSetCioId(Gia_ManCi(g,i),i);}
  Gia_ManForEachObj(g,obj,i)obj->fPhase=1;
  store=Cbs_ManSolveMiterNc(g,budgets[bi],&statuses,prove,0);
  printf("miter %d %d %d statuses",seed,budgets[bi],prove);
  for(i=0;i<Vec_StrSize(statuses);++i)printf(" %d",(unsigned char)Vec_StrEntry(statuses,i));
  printf(" store");Vec_IntForEachEntry(store,value,i)printf(" %d",value);
  printf(" drivers");Gia_ManForEachCo(g,obj,i)printf(" %d",Gia_ObjFaninLit0p(g,obj));
  printf(" phases");Gia_ManForEachObj(g,obj,i)printf("%d",obj->fPhase);printf("\n");
  Vec_IntFree(store);Vec_StrFree(statuses);Gia_ManStop(g);
 }
}
static void stats(void) {
 Gia_Man_t *g=Gia_ManStart(16);Cbs_Man_t *p;int a=Gia_ManAppendCi(g);
 Gia_ManAppendCo(g,a);p=Cbs_ManAlloc(g);
 p->nSatTotal=4;p->nSatUnsat=2;p->nSatSat=1;p->nSatUndec=1;p->nConfUnsat=5;p->nConfSat=3;p->nConfUndec=7;
 p->timeTotal=4*CLOCKS_PER_SEC;p->timeSatUnsat=2*CLOCKS_PER_SEC;p->timeSatSat=CLOCKS_PER_SEC;p->timeSatUndec=CLOCKS_PER_SEC/2;
 Cbs_ManSatPrintStats(p);Cbs_ManStop(p);p=Cbs_ManAlloc(g);Cbs_ManSatPrintStats(p);Cbs_ManStop(p);Gia_ManStop(g);
}
static void reason(void) {
 Gia_Man_t *g=Gia_ManStart(16);Cbs_Man_t *p;int a,b,c,ab,bc,h,i;Gia_Obj_t *obj;
 a=Gia_ManAppendCi(g);b=Gia_ManAppendCi(g);c=Gia_ManAppendCi(g);ab=Gia_ManAppendAnd(g,a,b);bc=Gia_ManAppendAnd(g,b,c);
 Gia_ManCreateRefs(g);Gia_ManFillValue(g);p=Cbs_ManAlloc(g);
 Cbs_ManAssign(p,Gia_Not(Gia_ManObj(g,ab>>1)),0,NULL,NULL);
 Cbs_ManAssign(p,Gia_ManObj(g,bc>>1),0,NULL,NULL);
 Cbs_ManAssign(p,Gia_ManObj(g,a>>1),1,NULL,NULL);
 Cbs_ManAssign(p,Gia_Not(Gia_ManObj(g,b>>1)),1,Gia_ManObj(g,ab>>1),Gia_ManObj(g,a>>1));
 h=Cbs_ManPropagateOne(p,Gia_ManObj(g,bc>>1),1);
 printf("reason h%d head%d size%d level%d reasons%d:%d levels",h,Gia_ObjId(g,p->pClauses.pData[h]),Cbs_ManClauseSize(p,h),Cbs_ClauseDecLevel(p,h),Gia_ObjId(g,Cbs_VarReason0(p,Gia_ManObj(g,b>>1))),Gia_ObjId(g,Cbs_VarReason1(p,Gia_ManObj(g,b>>1))));
 Vec_IntForEachEntry(p->vLevReas,a,i)printf(" %d",a);printf(" marks");Gia_ManForEachObj(g,obj,i)printf("%d%d",obj->fMark0,obj->fMark1);printf("\n");
 Cbs_ManPrintClause(p,1,h);Cbs_ManPrintClauseNew(p,1,h);
 Cbs_ManCancelUntil(p,0);Cbs_ManStop(p);Gia_ManStop(g);
}
int main(int argc,char **argv) {
 if(argc>1){if(argv[1][0]=='m')miters();else if(argv[1][0]=='r')reason();else stats();return 0;}
 int seed,heuristic,bi,ji,query,phase,mode,other,status,k,value,lits[29];
 int budgets[]={0,1,3,1000},frontiers[]={0,1,3,100};Gia_Man_t *g;Cbs_Man_t *p;
 for(seed=0;seed<12;++seed){
  g=graph(seed,lits);Gia_ManCreateRefs(g);Gia_ManCleanMark0(g);Gia_ManCleanMark1(g);Gia_ManFillValue(g);
  for(heuristic=0;heuristic<3;++heuristic)for(bi=0;bi<4;++bi)for(ji=0;ji<4;++ji){
   p=Cbs_ManAlloc(g);p->Pars.fUseHighest=heuristic==0;p->Pars.fUseLowest=heuristic==1;p->Pars.fUseMaxFF=heuristic==2;
   p->Pars.nBTLimit=budgets[bi];p->Pars.nJustLimit=frontiers[ji];
   for(query=0;query<29;++query)if(lits[query]>=2)for(phase=0;phase<2;++phase)for(mode=0;mode<3;++mode){
    other=lits[(query+7)%29]^(phase^1);
    if(mode==1 && (other<2 || (other>>1)==(lits[query]>>1)))continue;
    if(mode==2 && query%5)continue;
    if(mode==0)status=Cbs_ManSolve(p,Gia_NotCond(Gia_ManObj(g,lits[query]>>1),(lits[query]^phase)&1));
    else status=Cbs_ManSolve2(p,Gia_NotCond(Gia_ManObj(g,lits[query]>>1),(lits[query]^phase)&1),mode==1?Gia_NotCond(Gia_ManObj(g,other>>1),other&1):NULL);
    printf("case %d %d %d %d %d %d %d status%d bt%d nc%d just%d totals%d:%d model",seed,heuristic,budgets[bi],frontiers[ji],query,phase,mode,status,p->Pars.nBTThis,p->Pars.nBTThisNc,p->Pars.nJustThis,p->Pars.nBTTotal,p->Pars.nJustTotal);
    if(status==0)Vec_IntForEachEntry(p->vModel,value,k)printf(" %d",value);
    printf("\n");
   }
   Cbs_ManStop(p);
  }
  Gia_ManStop(g);
 }
 return 0;
}

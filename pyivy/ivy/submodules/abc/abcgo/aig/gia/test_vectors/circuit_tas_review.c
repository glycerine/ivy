#include "aig/gia/giaCTas.c"
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
  store=Tas_ManSolveMiterNc(g,budgets[bi],&statuses,0);
  printf("miter %d %d statuses",seed,budgets[bi]);
  for(i=0;i<Vec_StrSize(statuses);++i)printf(" %d",(unsigned char)Vec_StrEntry(statuses,i));
  printf(" store");Vec_IntForEachEntry(store,value,i)printf(" %d",value);
  printf(" drivers");Gia_ManForEachCo(g,obj,i)printf(" %d",Gia_ObjFaninLit0p(g,obj));
  printf(" phases");Gia_ManForEachObj(g,obj,i)printf("%d",obj->fPhase);printf("\n");
  Vec_IntFree(store);Vec_StrFree(statuses);Gia_ManStop(g);
 }
}


static void stats(void) {
 Gia_Man_t *g=Gia_ManStart(16);Tas_Man_t *p;int a=Gia_ManAppendCi(g);
 Gia_ManAppendCo(g,a);p=Tas_ManAlloc(g,1000);
 p->nSatTotal=4;p->nSatUnsat=2;p->nSatSat=1;p->nSatUndec=1;p->nConfUnsat=5;p->nConfSat=3;p->nConfUndec=7;
 p->timeTotal=4*CLOCKS_PER_SEC;p->timeSatUnsat=2*CLOCKS_PER_SEC;p->timeSatSat=CLOCKS_PER_SEC;p->timeSatUndec=CLOCKS_PER_SEC/2;
 Tas_ManSatPrintStats(p);Tas_ManStop(p);p=Tas_ManAlloc(g,1000);Tas_ManSatPrintStats(p);Tas_ManStop(p);Gia_ManStop(g);
}

static void patterns(void) {
 Vec_Ptr_t *info=Vec_PtrAlloc(5),*pres=Vec_PtrAlloc(5);Vec_Int_t *model=Vec_IntAlloc(5);
 unsigned data[5][3],present[5][3],random;int seed,trial,i,j,count,bit,result,lits[5],bits[]={0,1,31,32,33,63,64,95};
 for(i=0;i<5;++i){Vec_PtrPush(info,data[i]);Vec_PtrPush(pres,present[i]);}
 for(seed=0;seed<8;++seed){
  random=seed+103;
  for(i=0;i<5;++i)for(j=0;j<3;++j){random=random*1664525u+1013904223u;data[i][j]=random;random=random*1664525u+1013904223u;present[i][j]=(seed&1)?random:0;}
  for(trial=0;trial<64;++trial){
   count=trial%6;bit=bits[(trial+seed)%8];
   for(i=0;i<count;++i){random=random*1664525u+1013904223u;lits[i]=2*((trial+2*i)%5)+(random>>31);}
   result=Tas_StorePatternTry(info,pres,bit,lits,count);
   printf("try %d %d %d %d %d",seed,trial,bit,count,result);
   for(i=0;i<5;++i)for(j=0;j<3;++j)printf(" %u:%u",data[i][j],present[i][j]);printf("\n");
  }
 }
 memset(data,0,sizeof(data));memset(present,0,sizeof(present));
 for(trial=0;trial<64;++trial){
  Vec_IntClear(model);for(i=0;i<5;++i)Vec_IntPush(model,2*i+(((trial>>i)&1)^1));
  result=Tas_StorePattern(info,pres,model);printf("pack %d %d",trial,result);
  for(i=0;i<5;++i)printf(" %u:%u",data[i][0],present[i][0]);printf("\n");
 }
 Vec_IntFree(model);Vec_PtrFree(info);Vec_PtrFree(pres);
}

static void watch_learned(Tas_Man_t *p,int *lits,int n) {
 Tas_Cls_t *clause=Tas_ManAllocCls(p,n+3);clause->nLits=n;memcpy(clause->pLits,lits,n*sizeof(int));
 Tas_ManWatchClause(p,clause,lits[0]);Tas_ManWatchClause(p,clause,lits[1]);
}
static void watch_snapshot(Tas_Man_t *p,const char *label,int h) {
 int i;Gia_Obj_t *obj;printf("watch %s h%d head%d tail%d clauses",label,h,p->pClauses.iHead,p->pClauses.iTail);
 for(i=1;i<p->pClauses.iTail;++i)printf(" %d",p->pClauses.pData[i]?Gia_ObjId(p->pAig,p->pClauses.pData[i]):-1);
 printf(" store");for(i=16;i<p->pStore.iCur;++i)printf(" %d",p->pStore.pData[i]);
 printf(" trail");for(i=0;i<p->pProp.iTail;++i)printf(" %d",Tas_VarToLit(p,p->pProp.pData[i]));
 printf(" watches");for(i=0;i<2*Gia_ManObjNum(p->pAig);++i)printf(" %d",p->pWatches[i]);
 printf(" levels");for(i=0;i<Vec_IntSize(p->vLevReas);++i)printf(" %d",Vec_IntEntry(p->vLevReas,i));
 printf(" marks");Gia_ManForEachObj(p->pAig,obj,i)printf(" %d%d%d:%u",obj->fMark0,obj->fMark1,obj->fPhase,obj->Value);
 printf(" count%d\n",p->nClauses);
}
static void watch(void) {
 Gia_Man_t *g=Gia_ManStart(16);Tas_Man_t *p;int a,b,c,d,e,lits[4],h;
 a=Gia_ManAppendCi(g);b=Gia_ManAppendCi(g);c=Gia_ManAppendCi(g);d=Gia_ManAppendCi(g);e=Gia_ManAppendCi(g);
 Gia_ManFillValue(g);p=Tas_ManAlloc(g,1000);
 p->pStore.nSize=16;p->pStore.pData=ABC_REALLOC(int,p->pStore.pData,16);
 p->pClauses.nSize=2;p->pClauses.pData=ABC_REALLOC(Gia_Obj_t *,p->pClauses.pData,2);
 lits[0]=a;lits[1]=b;lits[2]=c;lits[3]=d;watch_learned(p,lits,4);lits[1]=e;watch_learned(p,lits,2);
 Tas_ManAssign(p,Gia_ObjFromLit(g,a^1),0,NULL,NULL);h=Tas_ManPropagateWatch(p,0,a^1);watch_snapshot(p,"a",h);
 Tas_ManAssign(p,Gia_ObjFromLit(g,b^1),1,NULL,NULL);h=Tas_ManPropagateWatch(p,1,b^1);watch_snapshot(p,"b",h);
 Tas_ManAssign(p,Gia_ObjFromLit(g,c^1),2,NULL,NULL);h=Tas_ManPropagateWatch(p,2,c^1);watch_snapshot(p,"unit",h);
 Tas_ManCancelUntil(p,4);Tas_ManAssign(p,Gia_ObjFromLit(g,d^1),3,NULL,NULL);h=Tas_ManPropagateWatch(p,3,c^1);
 Tas_ManCreateCls(p,h);watch_snapshot(p,"conflict",h);Tas_ManPrintClause(p,3,h);Tas_ManPrintClauseNew(p,3,h);
 Tas_ManCancelUntil(p,0);Tas_ManStop(p);Gia_ManStop(g);
}

static void packed(void) {
 int seeds[]={0,3,8,11},budgets[]={-1,0,1,1000},counts[]={17,1000,1001,2000,2001};
 int si,bi,ni,i,k,n,lits[29],pool[58],a;unsigned random;Gia_Man_t *g,*old;Gia_Obj_t *obj;Vec_Ptr_t *roots,*sim;
 for(si=0;si<4;++si)for(bi=0;bi<4;++bi)for(ni=0;ni<5;++ni){
  g=graph(seeds[si],lits);n=0;for(i=0;i<29;++i)if(lits[i]>=2){pool[n++]=lits[i];pool[n++]=lits[i]^1;}
  for(i=0;i<counts[ni];++i)Gia_ManAppendCo(g,pool[(7*i+seeds[si])%n]);
  if(seeds[si]&1){Vec_IntReverseOrder(g->vCos);for(i=0;i<Gia_ManCoNum(g);++i)Gia_ObjSetCioId(Gia_ManCo(g,i),i);}
  if(seeds[si]&2){Vec_IntReverseOrder(g->vCis);for(i=0;i<Gia_ManCiNum(g);++i)Gia_ObjSetCioId(Gia_ManCi(g,i),i);}
  old=Gia_ManStart(counts[ni]+1);roots=Vec_PtrAlloc(counts[ni]);
  for(i=0;i<counts[ni];++i){a=Gia_ManAppendCi(old);Vec_PtrPush(roots,Gia_ManObj(old,a>>1));}
  old->pReprs=ABC_CALLOC(Gia_Rpr_t,Gia_ManObjNum(old));sim=Vec_PtrAllocSimInfo(5,1);random=seeds[si]+211;
  for(i=0;i<5;++i){random=random*1664525u+1013904223u;((unsigned *)Vec_PtrEntry(sim,i))[0]=random;}
  Tas_ManSolveMiterNc2(g,budgets[bi],old,roots,sim);
  printf("packed %d %d %d proved",seeds[si],budgets[bi],counts[ni]);
  for(i=0;i<counts[ni];++i)printf("%d",Gia_ObjProved(old,i+1));
  printf(" sim");for(i=0;i<5;++i)printf(" %u",((unsigned *)Vec_PtrEntry(sim,i))[0]);printf("\n");
  Vec_PtrFree(sim);Vec_PtrFree(roots);Gia_ManStop(old);Gia_ManStop(g);
 }
}

int main(int argc,char **argv) {
 if(argc>1){if(argv[1][0]=='m')miters();else if(argv[1][0]=='p')patterns();else if(argv[1][0]=='w')watch();else if(argv[1][0]=='P')packed();else stats();return 0;}
 int seed,heuristic,bi,ji,q,phase,mode,other,status,k,value,lits[29];
 int budgets[]={0,1,3,1000},frontiers[]={0,1,3,100};Gia_Man_t *g;Tas_Man_t *p;Vec_Ptr_t *roots=Vec_PtrAlloc(3);
 setvbuf(stdout,NULL,_IONBF,0);
 for(seed=0;seed<12;++seed){
  g=graph(seed,lits);Gia_ManCreateRefs(g);Gia_ManCleanMark0(g);Gia_ManCleanMark1(g);Gia_ManFillValue(g);Gia_ManCleanPhase(g);
  for(heuristic=0;heuristic<4;++heuristic)for(bi=0;bi<4;++bi)for(ji=0;ji<4;++ji){
   p=Tas_ManAlloc(g,budgets[bi]);p->Pars.nJustLimit=frontiers[ji];
   p->Pars.fUseHighest=heuristic==0;p->Pars.fUseLowest=heuristic==1;p->Pars.fUseMaxFF=heuristic==2;p->Pars.fUseActive=heuristic==3;
   for(q=0;q<29;++q)if(lits[q]>=2)for(phase=0;phase<2;++phase)for(mode=0;mode<3;++mode){
    other=lits[(q+7)%29]^(phase^1);
    if(mode==1 && (other<2 || (other>>1)==(lits[q]>>1)))continue;
    if(mode==2 && q%5)continue;
    if(mode==0)status=Tas_ManSolve(p,Gia_NotCond(Gia_ManObj(g,lits[q]>>1),(lits[q]^phase)&1),NULL);
    else if(mode==1)status=Tas_ManSolve(p,Gia_NotCond(Gia_ManObj(g,lits[q]>>1),(lits[q]^phase)&1),Gia_NotCond(Gia_ManObj(g,other>>1),other&1));
    else {Vec_PtrClear(roots);Vec_PtrPush(roots,Gia_NotCond(Gia_ManObj(g,lits[q]>>1),(lits[q]^phase)&1));status=Tas_ManSolveArray(p,roots);}
    printf("case %d %d %d %d %d %d %d status%d bt%d nc%d just%d totals%d:%d model",seed,heuristic,budgets[bi],frontiers[ji],q,phase,mode,status,p->Pars.nBTThis,p->Pars.nBTThisNc,p->Pars.nJustThis,p->Pars.nBTTotal,p->Pars.nJustTotal);
    if(status==0)Vec_IntForEachEntry(p->vModel,value,k)printf(" %d",value);printf("\n");
   }
   Tas_ManStop(p);
  }
  Gia_ManStop(g);
 }
 Vec_PtrFree(roots);return 0;
}

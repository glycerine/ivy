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


static void stats(void) {
 Gia_Man_t *g=Gia_ManStart(16);Tas_Man_t *p;int a=Gia_ManAppendCi(g);
 Gia_ManAppendCo(g,a);p=Tas_ManAlloc(g,1000);
 p->nSatTotal=4;p->nSatUnsat=2;p->nSatSat=1;p->nSatUndec=1;p->nConfUnsat=5;p->nConfSat=3;p->nConfUndec=7;
 p->timeTotal=4*CLOCKS_PER_SEC;p->timeSatUnsat=2*CLOCKS_PER_SEC;p->timeSatSat=CLOCKS_PER_SEC;p->timeSatUndec=CLOCKS_PER_SEC/2;
 Tas_ManSatPrintStats(p);Tas_ManStop(p);p=Tas_ManAlloc(g,1000);Tas_ManSatPrintStats(p);Tas_ManStop(p);Gia_ManStop(g);
}

int main(int argc,char **argv) {
 if(argc>1){stats();return 0;}
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

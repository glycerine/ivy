#include "aig/gia/giaCex.c"
#include "base/main/main.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static int function(Gia_Man_t *p,int table,int x,int s) {
 int root=0,mask,cube;
 for(mask=0;mask<4;++mask)if((table>>mask)&1){cube=Gia_ManHashAnd(p,x^!(mask&1),s^!((mask>>1)&1));root=Gia_ManHashOr(p,root,cube);}
 return root;
}
int main(void) {
 int transition,pi,fi,seed,effort,frames,f,i,x,s;
 int properties[]={8,6,14,2},widths[]={2,5,37};unsigned random;
 Gia_Man_t *p,*clean,*unrolled;Abc_Cex_t *c,*care;Cnf_Dat_t *cnf;sat_solver *sat;int lit,status;
 Abc_Start();
 for(transition=0;transition<16;++transition)for(pi=0;pi<4;++pi)for(fi=0;fi<3;++fi)for(seed=0;seed<4;++seed){
  frames=widths[fi];p=Gia_ManStart(32);Gia_ManHashStart(p);x=Gia_ManAppendCi(p);s=Gia_ManAppendCi(p);
  Gia_ManAppendCo(p,function(p,properties[pi],x,s));Gia_ManAppendCo(p,function(p,transition,x,s));
  Gia_ManSetRegNum(p,1);Gia_ManHashStop(p);clean=Gia_ManCleanup(p);Gia_ManStop(p);p=clean;
  c=Abc_CexAlloc(1,1,frames);c->iFrame=frames-1;random=seed+1;
  for(f=0;f<frames;++f){random=random*1664525u+1013904223u;if(random>>31)Abc_InfoSetBit(c->pData,1+f);}
  if(!Gia_ManVerifyCex(p,c,0)){Abc_CexFree(c);Gia_ManStop(p);continue;}
  unrolled=Gia_ManFramesForCexMin(p,c->iFrame+1);
  cnf=(Cnf_Dat_t*)Mf_ManGenerateCnf(unrolled,8,0,0,0,0);sat=(sat_solver*)Cnf_DataWriteIntoSolver(cnf,1,0);lit=Abc_Var2Lit(1,1);
  status=sat_solver_addclause(sat,&lit,&lit+1);
  sat_solver_delete(sat);Cnf_DataFree(cnf);Gia_ManStop(unrolled);
  if(!status){
   printf("forced-true %d %d %d %d\n",transition,properties[pi],frames,seed);
   Abc_CexFree(c);Gia_ManStop(p);continue;
  }
  for(effort=0;effort<2;++effort){
   care=Bmc_CexCareSatBasedMinimizeAig(p,c,effort,0);
   printf("case %d %d %d %d %d mask",transition,properties[pi],frames,seed,effort);
   for(i=0;i<Abc_BitWordNum(care->nBits);++i)printf("%08x",care->pData[i]);printf("\n");Abc_CexFree(care);
  }
  Abc_CexFree(c);Gia_ManStop(p);
 }
 Abc_Stop();return 0;
}

/* Expose the original private manager and routines, without reimplementation. */
#include "aig/gia/giaCCof.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static Gia_Man_t *source(int regs){Gia_Man_t *g=Gia_ManStart(16);int i;for(i=0;i<regs;i++)Gia_ManAppendCi(g);for(i=0;i<regs;i++)Gia_ManAppendCo(g,Gia_ManCiLit(g,i));Gia_ManSetRegNum(g,regs);return g;}
static Gia_Man_t *frames(int regs,int pis,int mode,int phase,int *property){
 Gia_Man_t *g=Gia_ManStart(128);int ci[5],i,state,lit,term;
 for(i=0;i<regs+pis;i++)ci[i]=Gia_ManAppendCi(g);Gia_ManHashStart(g);
 state=(mode==0 || mode==3)?1:0;
 for(i=0;i<regs;i++){
  lit=ci[i]^((phase>>i)&1);
  if(mode==2){term=Gia_ManHashAnd(g,lit,ci[regs+i%pis]^((phase>>(i+1))&1));state=Gia_ManHashOr(g,state,term);}
  else if(mode==0 || mode==3)state=Gia_ManHashAnd(g,state,lit);
  else state=Gia_ManHashOr(g,state,lit);
 }
 if(mode!=2){lit=ci[regs]^(phase&1);if(mode==3 && pis>1)lit=Gia_ManHashXor(g,lit,ci[regs+1]);state=Gia_ManHashAnd(g,state,lit);}
 *property=state;return g;
}
int main(void){int regs,pis,mode,phase,property,status,again,i,assignment,*values,any;unsigned image;Gia_Obj_t *o;Ccf_Man_t p;
 for(regs=1;regs<=3;regs++)for(pis=1;pis<=2;pis++)for(mode=0;mode<4;mode++)for(phase=0;phase<8;phase++){
  memset(&p,0,sizeof(p));p.pGia=source(regs);p.pFrames=frames(regs,pis,mode,phase,&property);p.vCopies=Vec_IntAlloc(128);p.pSat=sat_solver_new();
  Gia_ManCofExtendSolver(&p);status=Gia_ManCofGetReachable(&p,property);again=Gia_ManCofGetReachable(&p,property);
  values=ABC_CALLOC(int,p.pFrames->nObjs);image=0;
  for(assignment=0;assignment<(1<<regs);assignment++){
   Gia_ManForEachCi(p.pFrames,o,i)values[Gia_ObjId(p.pFrames,o)]=i<regs?((assignment>>i)&1):0;
   Gia_ManForEachAnd(p.pFrames,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
   any=0;Gia_ManForEachCo(p.pFrames,o,i)any|=values[Gia_ObjFaninId0p(p.pFrames,o)]^Gia_ObjFaninC0(o);
   image|=any<<assignment;
  }
  printf("case %d %d %d %d %d %d %x\n",regs,pis,mode,phase,status,again,image);
  ABC_FREE(values);sat_solver_delete(p.pSat);Vec_IntFree(p.vCopies);Gia_ManStop(p.pFrames);Gia_ManStop(p.pGia);
 }
 return 0;
}

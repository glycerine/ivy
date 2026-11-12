#include "misc/util/util_hack.h"
#include "aig/aig/aig.h"
#include "proof/dch/dch.h"
#include <stdio.h>
static Aig_Man_t *variant(int which) {
 Aig_Man_t *g=Aig_ManStart(100);Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g),*c=Aig_ObjCreateCi(g),*r;
 if(which==0)r=Aig_And(g,Aig_And(g,a,b),c);
 else if(which==1)r=Aig_And(g,a,Aig_And(g,b,c));
 else if(which==2)r=Aig_And(g,a,c);
 else r=Aig_Or(g,Aig_And(g,Aig_And(g,a,b),c),Aig_And(g,Aig_And(g,a,Aig_Not(b)),c));
 Aig_ObjCreateCo(g,r);return g;
}
static int eval(Aig_Obj_t *edge,unsigned bits){
 Aig_Obj_t *n=Aig_Regular(edge);int value;
 if(Aig_ObjIsConst1(n))value=1;
 else if(Aig_ObjIsCi(n))value=(bits>>Aig_ObjCioId(n))&1;
 else if(Aig_ObjIsCo(n))value=eval(Aig_ObjChild0(n),bits);
 else value=eval(Aig_ObjChild0(n),bits)&eval(Aig_ObjChild1(n),bits);
 return value^Aig_IsComplement(edge);
}
int main(void){
 for(int which=0;which<3;which++){
  Aig_Man_t *a=variant(which?2:0),*b=variant(which?3:1);
  Vec_Ptr_t *v=Vec_PtrAlloc(2);Vec_PtrPush(v,a);Vec_PtrPush(v,b);
  Aig_Man_t *total=Dch_DeriveTotalAig(v);
  Dch_Pars_t params;Dch_ManSetDefaultParams(&params);params.fSkipRedSupp=which==2;
  Aig_Man_t *result=Dch_ComputeChoices(total,&params);
  unsigned truth=0;for(unsigned bits=0;bits<8;bits++)truth|=eval(Aig_ManCo(result,0),bits)<<bits;
  printf("%d %d %d %d %u\n",which,Aig_ManNodeNum(total),Aig_ManNodeNum(result),Aig_ManChoiceNum(result),truth);
  Aig_ManStop(result);Aig_ManStop(total);Aig_ManStop(a);Aig_ManStop(b);Vec_PtrFree(v);
 }
 return 0;
}

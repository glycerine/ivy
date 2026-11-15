#define ABC_USE_CUDD
#include "aig/gia/giaClp.c"

#include "base/main/main.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static DdNode *function(DdManager *dd,int n,int table){
 DdNode *sum=Cudd_Not(dd->one),*cube,*next;int mask,i;Cudd_Ref(sum);
 for(mask=0;mask<(1<<n);mask++)if((table>>mask)&1){
  cube=dd->one;Cudd_Ref(cube);
  for(i=0;i<n;i++){next=Cudd_bddAnd(dd,cube,Cudd_NotCond(Cudd_bddIthVar(dd,i),!((mask>>i)&1)));Cudd_Ref(next);Cudd_RecursiveDeref(dd,cube);cube=next;}
  next=Cudd_bddOr(dd,sum,cube);Cudd_Ref(next);Cudd_RecursiveDeref(dd,sum);Cudd_RecursiveDeref(dd,cube);sum=next;
 }
 return sum;
}
int main(void){int n,table,order,phase,i,j,sequence[3],root,values[128],mask,packed;DdManager *dd;DdNode *f;Gia_Man_t *g;Gia_Obj_t *obj;Vec_Int_t *leaves;Vec_Str_t *sop,*cube;char *text;
 Abc_Start();
 for(n=2;n<=3;n++)for(table=1;table<(1<<(1<<n))-1;table++)for(order=0;order<3;order++){
  dd=Cudd_Init(n,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);
  for(i=0;i<n;i++){sequence[i]=order==0?i:order==1?n-1-i:(i+1)%n;}
  assert(Cudd_ShuffleHeap(dd,sequence));Cudd_zddVarsFromBddVars(dd,2);f=function(dd,n,table);
  for(phase=0;phase<(1<<n);phase++){
   g=Gia_ManStart(64);leaves=Vec_IntAlloc(3);sop=Vec_StrAlloc(0);cube=Vec_StrAlloc(0);
   for(i=0;i<n;i++)Vec_IntPush(leaves,Gia_ManAppendCi(g)^((phase>>i)&1));Gia_ManHashStart(g);
   root=Gia_ManRebuildIsop(dd,f,g,leaves,sop,cube);
   printf("case %d %x %d %d %d sop",n,table,order,phase,root);
   text=Vec_StrArray(sop);for(i=0;text[i];i++)printf("%02x",(unsigned char)text[i]);printf(" nodes");
   Gia_ManForEachAnd(g,obj,i)printf(" %d:%d",Gia_ObjFaninLit0p(g,obj),Gia_ObjFaninLit1p(g,obj));
   packed=0;
   for(mask=0;mask<(1<<n);mask++){
    memset(values,0,sizeof(values));Gia_ManForEachCi(g,obj,j)values[Gia_ObjId(g,obj)]=(mask>>j)&1;
    Gia_ManForEachAnd(g,obj,i)values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
    packed|=(values[root>>1]^(root&1))<<mask;
   }
   printf(" truth %x\n",packed);Gia_ManHashStop(g);Gia_ManStop(g);Vec_IntFree(leaves);Vec_StrFree(sop);Vec_StrFree(cube);
  }
  Cudd_RecursiveDeref(dd,f);Cudd_Quit(dd);
 }
 Abc_Stop();return 0;
}

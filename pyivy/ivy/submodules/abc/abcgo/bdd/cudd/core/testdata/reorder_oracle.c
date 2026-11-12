#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdint.h>

static DdNode *from_table(DdManager *dd,unsigned mask,int level,unsigned assignment) {
 if(level==3)return mask&(1u<<assignment)?Cudd_ReadOne(dd):Cudd_ReadLogicZero(dd);
 DdNode *t=from_table(dd,mask,level+1,assignment|(1u<<level));Cudd_Ref(t);
 DdNode *e=from_table(dd,mask,level+1,assignment);Cudd_Ref(e);
 DdNode *r=Cudd_bddIte(dd,Cudd_bddIthVar(dd,level),t,e);Cudd_Ref(r);
 Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;
}
static unsigned table(DdNode *root) {
 unsigned mask=0;
 for(unsigned a=0;a<8;a++){
  DdNode *n=root;
  while(!Cudd_IsConstant(n)){
   int comp=Cudd_IsComplement(n);DdNode *r=Cudd_Regular(n);
   n=(a&(1u<<r->index))?cuddT(r):cuddE(r);n=Cudd_NotCond(n,comp);
  }
  if(!Cudd_IsComplement(n))mask|=1u<<a;
 }
 return mask;
}
int main(void){
 for(int method=0;method<2;method++)for(unsigned mask=0;mask<256;mask++){
  DdManager *dd=Cudd_Init(3,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);
  DdNode *root=from_table(dd,mask,0,0);Cudd_Ref(root);
  int before=Cudd_DagSize(root);
  int status=Cudd_ReduceHeap(dd,method?CUDD_REORDER_SYMM_SIFT:CUDD_REORDER_SIFT,1);
  printf("%d %u %d %d %d %u %d%d%d\n",method,mask,before,Cudd_DagSize(root),status,table(root),dd->invperm[0],dd->invperm[1],dd->invperm[2]);
  Cudd_RecursiveDeref(dd,root);Cudd_Quit(dd);
 }
 return 0;
}

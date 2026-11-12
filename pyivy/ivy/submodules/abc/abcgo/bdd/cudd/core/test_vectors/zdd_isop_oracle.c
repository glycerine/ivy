#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdint.h>
static unsigned truth(DdManager *dd,DdNode *f) {
 unsigned result=0;
 for(unsigned assignment=0;assignment<8;assignment++){
  DdNode *n=f;
  while(!Cudd_IsConstant(n)){unsigned index=Cudd_NodeReadIndex(n);DdNode *next=(assignment&(1u<<index))?Cudd_T(n):Cudd_E(n);n=Cudd_NotCond(next,Cudd_IsComplement(n));}
  if(n==Cudd_ReadOne(dd))result|=1u<<assignment;
 }
 return result;
}
static uint64_t coverbits(DdManager *dd,DdNode *n,unsigned set) {
 if(n==Cudd_ReadZero(dd))return 0;if(n==Cudd_ReadOne(dd))return UINT64_C(1)<<set;
 return coverbits(dd,Cudd_E(n),set)|coverbits(dd,Cudd_T(n),set|(1u<<Cudd_NodeReadIndex(n)));
}
static DdNode *function(DdManager *dd,unsigned table,unsigned index,unsigned assignment) {
 if(index==3){DdNode *n=(table&(1u<<assignment))?Cudd_ReadOne(dd):Cudd_ReadLogicZero(dd);Cudd_Ref(n);return n;}
 DdNode *t=function(dd,table,index+1,assignment|(1u<<index));
 DdNode *e=function(dd,table,index+1,assignment);
 DdNode *n=Cudd_bddIte(dd,Cudd_bddIthVar(dd,index),t,e);Cudd_Ref(n);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);return n;
}
int main(void) {
 DdManager *dd=Cudd_Init(3,6,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);Cudd_DisableGarbageCollection(dd);
#ifdef REORDER_BDD
 int order[3]={2,0,1};if(!Cudd_ShuffleHeap(dd,order))return 1;
#endif
#ifdef REORDER_ZDD
#ifdef REVERSE_LITERAL_PAIRS
 int orderZ[6]={1,0,3,2,5,4};
#else
 int orderZ[6]={4,5,0,1,2,3};
#endif
 if(!Cudd_zddShuffleHeap(dd,orderZ))return 1;
#endif
 DdNode *functions[256];for(unsigned i=0;i<256;i++)functions[i]=function(dd,i,0,0);
 uint64_t hash=UINT64_C(14695981039346656037),each[6];for(unsigned i=0;i<6;i++)each[i]=hash;
 for(unsigned code=0;code<6561;code++){
  unsigned remaining=code,lower=0,upper=0;
  for(unsigned assignment=0;assignment<8;assignment++){unsigned state=remaining%3;remaining/=3;if(state)upper|=1u<<assignment;if(state==2)lower|=1u<<assignment;}
  DdNode *cover,*chosen=Cudd_zddIsop(dd,functions[lower],functions[upper],&cover);if(!chosen)return 1;Cudd_Ref(chosen);Cudd_Ref(cover);
  DdNode *fromCover=Cudd_MakeBddFromZddCover(dd,cover);if(!fromCover)return 1;Cudd_Ref(fromCover);
  DdNode *bddOnly=Cudd_bddIsop(dd,functions[lower],functions[upper]);if(!bddOnly)return 1;Cudd_Ref(bddOnly);
  DdNode *complement=Cudd_zddComplement(dd,cover);if(!complement)return 1;Cudd_Ref(complement);
  DdNode *complementBDD=Cudd_MakeBddFromZddCover(dd,complement);if(!complementBDD)return 1;Cudd_Ref(complementBDD);
  uint64_t values[6]={truth(dd,chosen),coverbits(dd,cover,0),truth(dd,fromCover),truth(dd,bddOnly),coverbits(dd,complement,0),truth(dd,complementBDD)};
  for(unsigned i=0;i<6;i++){hash^=values[i];hash*=UINT64_C(1099511628211);each[i]^=values[i];each[i]*=UINT64_C(1099511628211);}
  Cudd_RecursiveDeref(dd,complementBDD);Cudd_RecursiveDerefZdd(dd,complement);Cudd_RecursiveDeref(dd,bddOnly);Cudd_RecursiveDeref(dd,fromCover);Cudd_RecursiveDerefZdd(dd,cover);Cudd_RecursiveDeref(dd,chosen);
 }
 printf("all %016llx\n",(unsigned long long)hash);for(unsigned i=0;i<6;i++)printf("op%u %016llx\n",i,(unsigned long long)each[i]);
 Cudd_Quit(dd);return 0;
}

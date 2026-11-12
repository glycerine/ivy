#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdint.h>
static uint64_t table(DdManager *dd,DdNode *n,unsigned set) {
 if(n==NULL)return UINT64_MAX;
 if(n==Cudd_ReadZero(dd))return 0;
 if(n==Cudd_ReadOne(dd))return UINT64_C(1)<<set;
 return table(dd,Cudd_E(n),set)|table(dd,Cudd_T(n),set|(1u<<Cudd_NodeReadIndex(n)));
}
static DdNode *cover(DdManager *dd,unsigned mask) {
 unsigned cubes[27],count=0;for(unsigned set=0;set<64;set++)if((set&3)!=3&&((set>>2)&3)!=3&&((set>>4)&3)!=3)cubes[count++]=set;
 DdNode *f=Cudd_ReadZero(dd);Cudd_Ref(f);
 for(unsigned term=0;term<27;term++)if(mask&(1u<<term)) {
  DdNode *c=Cudd_ReadOne(dd);Cudd_Ref(c);
  for(int i=0;i<6;i++)if(cubes[term]&(1u<<i)){DdNode *next=Cudd_zddChange(dd,c,i);Cudd_Ref(next);Cudd_RecursiveDerefZdd(dd,c);c=next;}
  DdNode *next=Cudd_zddUnion(dd,f,c);Cudd_Ref(next);Cudd_RecursiveDerefZdd(dd,f);Cudd_RecursiveDerefZdd(dd,c);f=next;
 }
 return f;
}
int main(void) {
 DdManager *dd=Cudd_Init(0,6,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);
 Cudd_DisableGarbageCollection(dd);
 int order[6]={4,5,0,1,2,3};if(!Cudd_zddShuffleHeap(dd,order))return 1;
 uint64_t hash=UINT64_C(14695981039346656037),each[6];for(unsigned i=0;i<6;i++)each[i]=hash;
 uint32_t seed=UINT32_C(0x9e3779b9);unsigned different=0;
 DdNode *(*ops[6])(DdManager*,DdNode*,DdNode*)={Cudd_zddUnateProduct,Cudd_zddProduct,Cudd_zddDivide,Cudd_zddDivideF,Cudd_zddWeakDiv,Cudd_zddWeakDivF};
 for(unsigned iteration=0;iteration<2000;iteration++) {
  seed=seed*UINT32_C(1664525)+UINT32_C(1013904223);unsigned a=(seed^(seed>>16))&UINT32_C(0x7ffffff);
  seed=seed*UINT32_C(1664525)+UINT32_C(1013904223);unsigned b=(seed^(seed>>16))&UINT32_C(0x7ffffff);if(!b)b=1;
  unsigned cubes[27],count=0;for(unsigned set=0;set<64;set++)if((set&3)!=3&&((set>>2)&3)!=3&&((set>>4)&3)!=3)cubes[count++]=set;
  if(iteration%3==0){unsigned mask=0;for(unsigned i=0;i<27;i++)if(a&(1u<<i))for(unsigned j=0;j<27;j++)if(b&(1u<<j)){unsigned cube=cubes[i]|cubes[j];if((cube&3)==3||((cube>>2)&3)==3||((cube>>4)&3)==3)continue;for(unsigned k=0;k<27;k++)if(cubes[k]==cube)mask|=1u<<k;}a=mask;}
  if(iteration%3==1){unsigned literal=1u<<((iteration/3)%6);for(unsigned k=0;k<27;k++)if(cubes[k]==literal)b=1u<<k;}
  DdNode *f=cover(dd,a),*g=cover(dd,b);uint64_t values[6];
  for(unsigned i=0;i<6;i++){DdNode *result=ops[i](dd,f,g);if(!result){fprintf(stderr,"unexpected null %u %u\n",iteration,i);return 1;}Cudd_Ref(result);values[i]=table(dd,result,0);Cudd_RecursiveDerefZdd(dd,result);hash^=values[i];hash*=UINT64_C(1099511628211);each[i]^=values[i];each[i]*=UINT64_C(1099511628211);}
  if(values[4]!=values[5])different++;
  Cudd_RecursiveDerefZdd(dd,f);Cudd_RecursiveDerefZdd(dd,g);
 }
 printf("all %016llx\n",(unsigned long long)hash);for(unsigned i=0;i<6;i++)printf("op%u %016llx\n",i,(unsigned long long)each[i]);printf("weak variants differ on %u pairs\n",different);
 Cudd_Quit(dd);return 0;
}

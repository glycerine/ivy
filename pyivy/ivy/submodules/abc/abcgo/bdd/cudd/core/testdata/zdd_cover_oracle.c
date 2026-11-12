#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdint.h>
static unsigned table(DdManager *dd,DdNode *n,unsigned set) {
 if(n==NULL)return 65536;
 if(n==Cudd_ReadZero(dd))return 0;
 if(n==Cudd_ReadOne(dd))return 1u<<set;
 return table(dd,Cudd_E(n),set)|table(dd,Cudd_T(n),set|(1u<<Cudd_NodeReadIndex(n)));
}
static DdNode *cover(DdManager *dd,unsigned mask) {
 static unsigned cubes[9]={0,1,2,4,5,6,8,9,10};
 DdNode *f=Cudd_ReadZero(dd);Cudd_Ref(f);
 for(unsigned term=0;term<9;term++)if(mask&(1u<<term)) {
  DdNode *c=Cudd_ReadOne(dd);Cudd_Ref(c);
  for(int i=0;i<4;i++)if(cubes[term]&(1u<<i)){DdNode *next=Cudd_zddChange(dd,c,i);Cudd_Ref(next);Cudd_RecursiveDerefZdd(dd,c);c=next;}
  DdNode *next=Cudd_zddUnion(dd,f,c);Cudd_Ref(next);Cudd_RecursiveDerefZdd(dd,f);Cudd_RecursiveDerefZdd(dd,c);f=next;
 }
 return f;
}
int main(void) {
 DdManager *dd=Cudd_Init(0,4,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);
 DdNode *covers[512];for(unsigned i=0;i<512;i++)covers[i]=cover(dd,i);
 uint64_t hash=UINT64_C(14695981039346656037), each[6];for(unsigned i=0;i<6;i++)each[i]=hash;
 unsigned different=0;
 for(unsigned a=0;a<512;a++)for(unsigned b=1;b<512;b++) {
  DdNode *f=covers[a],*g=covers[b];
  DdNode *(*ops[6])(DdManager*,DdNode*,DdNode*)={Cudd_zddUnateProduct,Cudd_zddProduct,Cudd_zddDivide,Cudd_zddDivideF,Cudd_zddWeakDiv,Cudd_zddWeakDivF};
  unsigned values[6];
  for(unsigned i=0;i<6;i++){DdNode *result=ops[i](dd,f,g);if(result)Cudd_Ref(result);values[i]=table(dd,result,0);if(result)Cudd_RecursiveDerefZdd(dd,result);}

  if(values[4]!=values[5])different++;
  for(unsigned i=0;i<6;i++){hash^=values[i];hash*=UINT64_C(1099511628211);each[i]^=values[i];each[i]*=UINT64_C(1099511628211);}
 }
 printf("all %016llx\n",(unsigned long long)hash);
 for(unsigned i=0;i<6;i++)printf("op%u %016llx\n",i,(unsigned long long)each[i]);
 printf("weak variants differ on %u pairs\n",different);
 Cudd_Quit(dd);return 0;
}

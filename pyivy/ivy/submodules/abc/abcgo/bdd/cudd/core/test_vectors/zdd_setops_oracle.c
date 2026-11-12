#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdio.h>
#include <stdint.h>
static unsigned table(DdManager *dd,DdNode *n,unsigned set) {
 if(n==Cudd_ReadZero(dd))return 0;
 if(n==Cudd_ReadOne(dd))return 1u<<set;
 return table(dd,Cudd_E(n),set)|table(dd,Cudd_T(n),set|(1u<<Cudd_NodeReadIndex(n)));
}
static DdNode *family(DdManager *dd,unsigned mask) {
 DdNode *f=Cudd_ReadZero(dd);Cudd_Ref(f);
 for(unsigned set=0;set<8;set++)if(mask&(1u<<set)) {
  DdNode *c=Cudd_ReadOne(dd);Cudd_Ref(c);
  for(int i=0;i<3;i++)if(set&(1u<<i)){DdNode *next=Cudd_zddChange(dd,c,i);Cudd_Ref(next);Cudd_RecursiveDerefZdd(dd,c);c=next;}
  DdNode *next=Cudd_zddUnion(dd,f,c);Cudd_Ref(next);Cudd_RecursiveDerefZdd(dd,f);Cudd_RecursiveDerefZdd(dd,c);f=next;
 }
 return f;
}
int main(void) {
 DdManager *dd=Cudd_Init(0,3,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);
 DdNode *families[256];for(unsigned i=0;i<256;i++)families[i]=family(dd,i);
 uint64_t hash=UINT64_C(14695981039346656037);
 for(unsigned a=0;a<256;a++)for(unsigned b=0;b<256;b++) {
  DdNode *f=families[a],*g=families[b],*h=families[(a+3*b)&255];
  unsigned values[5]={table(dd,Cudd_zddUnion(dd,f,g),0),table(dd,Cudd_zddIntersect(dd,f,g),0),table(dd,Cudd_zddDiff(dd,f,g),0),table(dd,Cudd_zddIte(dd,f,g,h),0),Cudd_zddDiffConst(dd,f,g)==Cudd_ReadZero(dd)};
  for(unsigned i=0;i<5;i++){hash^=values[i];hash*=UINT64_C(1099511628211);}
 }
 printf("%016llx\n",(unsigned long long)hash);
 Cudd_Quit(dd);return 0;
}

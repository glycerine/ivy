#include "misc/util/util_hack.h"
#include "bdd/dsd/dsdInt.h"
#include <inttypes.h>
static DdNode *from_table(DdManager *dd,uint64_t mask,int n,int i,uint64_t assignment) {
 if(i==n)return (mask>>assignment)&1?dd->one:Cudd_Not(dd->one);
 DdNode *t=from_table(dd,mask,n,i+1,assignment|(UINT64_C(1)<<i));Cudd_Ref(t);
 DdNode *e=from_table(dd,mask,n,i+1,assignment);Cudd_Ref(e);
 DdNode *r=Cudd_bddIte(dd,dd->vars[i],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;
}
static uint64_t table(DdManager *dd,DdNode *f,int n) {
 uint64_t mask=0;for(int a=0;a<(1<<n);a++){DdNode *g=f;while(!Cudd_IsConstant(g)){DdNode *r=Cudd_Regular(g);g=Cudd_NotCond((a>>r->index)&1?cuddT(r):cuddE(r),Cudd_IsComplement(g));}if(g==dd->one)mask|=UINT64_C(1)<<a;}return mask;
}
static void print_tree(DdManager *dd,Dsd_Node_t *p,int n) {
 Dsd_Node_t *r=Dsd_Regular(p);printf("%d:%d:%016" PRIx64 ":%016" PRIx64 "[",Dsd_IsComplement(p),r->Type,table(dd,r->G,n),r->Type==DSD_NODE_PRIME?table(dd,Dsd_TreeGetPrimeFunction(dd,r),n):UINT64_C(0));
 if(r->Type!=DSD_NODE_BUF)for(int i=0;i<r->nDecs;i++){if(i)putchar(',');print_tree(dd,r->pDecs[i],n);}putchar(']');
}
int main(void) {
 int n,reverse;uint64_t mask;while(scanf("%d %d %" SCNx64,&n,&reverse,&mask)==3) {
  DdManager *dd=Cudd_Init(n,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);
  if(reverse){int perm[6];for(int i=0;i<n;i++)perm[i]=n-1-i;Cudd_ShuffleHeap(dd,perm);}
  DdNode *f=from_table(dd,mask,n,0,0);Cudd_Ref(f);Dsd_Manager_t *m=Dsd_ManagerStart(dd,n,0);Dsd_Node_t *p=Dsd_DecomposeOne(m,f);
  print_tree(dd,p,n);putchar('\n');Dsd_ManagerStop(m);Cudd_RecursiveDeref(dd,f);Cudd_Quit(dd);
 }
 return 0;
}

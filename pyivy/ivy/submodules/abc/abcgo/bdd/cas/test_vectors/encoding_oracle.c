/* Fixture for GetSingleOutputFunction from original casCore.c. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdint.h>
extern DdNode *GetSingleOutputFunction(DdManager *, DdNode **, int, DdNode **, int, int);
static DdNode *truth(DdManager *dd, unsigned table, int level) {
 DdNode *t,*e,*r;
 if(level==3)return (table&1)?dd->one:Cudd_Not(dd->one);
 unsigned width=1u<<(2-level), mask=(1u<<width)-1;
 t=truth(dd,table>>width,level+1);Cudd_Ref(t);
 e=truth(dd,table&mask,level+1);Cudd_Ref(e);
 r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);
 Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;
}
int main(void) {
 uint64_t hash=UINT64_C(14695981039346656037);int a,b,v;
 for(a=0;a<256;a++)for(b=0;b<16;b++) {
  DdManager *dd=Cudd_Init(5,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *outs[3],*f,*r;
  outs[0]=truth(dd,a,0);Cudd_Ref(outs[0]);outs[1]=truth(dd,b*17,0);Cudd_Ref(outs[1]);outs[2]=dd->one;
  f=GetSingleOutputFunction(dd,outs,3,dd->vars+3,2,0);Cudd_Ref(f);
  for(v=0;v<32;v++){int inputs[5],i;for(i=0;i<5;i++)inputs[i]=(v>>(4-i))&1;r=Cudd_Eval(dd,f,inputs);hash^=(unsigned)cuddV(r);hash*=UINT64_C(1099511628211);}
  Cudd_Quit(dd);
 }
 printf("%016llx\n",(unsigned long long)hash);return 0;
}

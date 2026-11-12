/* Fixture generated entirely through original extraBddUnate.c APIs. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
static uint64_t hash=UINT64_C(14695981039346656037);
static void number(uint64_t v){int i;for(i=0;i<8;i++){hash^=(v>>(8*i))&255;hash*=UINT64_C(1099511628211);}}
static DdNode *truth(DdManager *dd,unsigned table,int level){DdNode *t,*e,*r;unsigned width,mask;if(level==3)return (table&1)?dd->one:Cudd_Not(dd->one);width=1u<<(2-level);mask=(1u<<width)-1;t=truth(dd,table>>width,level+1);Cudd_Ref(t);e=truth(dd,table&mask,level+1);Cudd_Ref(e);r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;}
int main(void){unsigned table;for(table=0;table<256;table++){DdManager *dd=Cudd_Init(3,6,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *f=truth(dd,table,0);Extra_UnateInfo_t *fast,*slow;int i;Cudd_Ref(f);Cudd_zddVarsFromBddVars(dd,2);fast=Extra_UnateComputeFast(dd,f);slow=Extra_UnateComputeSlow(dd,f);for(i=0;i<3;i++)number(Extra_bddCheckUnateNaive(dd,f,i));number(fast->nVars);number(fast->nVarsMax);number(fast->nUnate);for(i=0;i<fast->nVars;i++){number(fast->pVars[i].iVar);number(fast->pVars[i].Pos);number(fast->pVars[i].Neg);}number(slow->nUnate);Extra_UnateInfoDissolve(fast);Extra_UnateInfoDissolve(slow);Cudd_Quit(dd);}printf("%016llx\n",(unsigned long long)hash);return 0;}

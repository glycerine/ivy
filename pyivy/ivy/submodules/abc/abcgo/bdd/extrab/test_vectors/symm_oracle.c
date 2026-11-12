/* Exact original extraBddSymm.c recursion fixture. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
static uint64_t hash=UINT64_C(14695981039346656037);
static void byte(unsigned b){hash^=b;hash*=UINT64_C(1099511628211);}
static DdNode *truth(DdManager *dd,unsigned table,int level){DdNode *t,*e,*r;unsigned width,mask;if(level==3)return(table&1)?dd->one:Cudd_Not(dd->one);width=1u<<(2-level);mask=(1u<<width)-1;t=truth(dd,table>>width,level+1);Cudd_Ref(t);e=truth(dd,table&mask,level+1);Cudd_Ref(e);r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;}
static unsigned equations(DdManager *dd,DdNode *f,unsigned mask){if(f==dd->zero)return 0;if(f==dd->one)return 1u<<mask;return equations(dd,cuddE(f),mask)|equations(dd,cuddT(f),mask|(1u<<(f->index/2)));}
static void zcollect(DdManager *dd,DdNode *f){Cudd_Ref(f);byte(equations(dd,f,0));Cudd_RecursiveDerefZdd(dd,f);}
int main(void){DdManager *dd=Cudd_Init(3,6,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *functions[256],*vars;int a,b,i,j;Cudd_zddVarsFromBddVars(dd,2);vars=Cudd_bddAnd(dd,dd->vars[0],dd->vars[1]);Cudd_Ref(vars);{DdNode *temp=Cudd_bddAnd(dd,vars,dd->vars[2]);Cudd_Ref(temp);Cudd_RecursiveDeref(dd,vars);vars=temp;}for(a=0;a<256;a++){functions[a]=truth(dd,a,0);Cudd_Ref(functions[a]);}for(a=0;a<256;a++){Extra_SymmInfo_t *p=Extra_SymmPairsCompute(dd,functions[a]);byte(p->nVars);byte(p->nSymms);for(i=0;i<p->nVars;i++){byte(p->pVars[i]);for(j=i+1;j<p->nVars;j++)byte(p->pSymms[i][j]);}Extra_SymmPairsDissolve(p);for(i=0;i<3;i++)for(j=i+1;j<3;j++)byte(Extra_bddCheckVarsSymmetric(dd,functions[a],i,j));for(b=0;b<256;b++)zcollect(dd,Extra_zddGetSymmetricVars(dd,functions[a],functions[b],vars));}printf("%016llx\n",(unsigned long long)hash);Cudd_Quit(dd);return 0;}

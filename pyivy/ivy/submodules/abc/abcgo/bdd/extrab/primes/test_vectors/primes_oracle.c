/* Calls the original Extra_zddPrimes in extraBddMisc.c. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
static DdNode *truth(DdManager *dd,unsigned table,int level){DdNode *t,*e,*r;unsigned width,mask;if(level==3)return (table&1)?dd->one:Cudd_Not(dd->one);width=1u<<(2-level);mask=(1u<<width)-1;t=truth(dd,table>>width,level+1);Cudd_Ref(t);e=truth(dd,table&mask,level+1);Cudd_Ref(e);r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;}
static int member(DdManager *dd,DdNode *f,unsigned mask){while(f!=dd->one&&f!=dd->zero){unsigned bit=1u<<f->index;if(mask&bit){mask^=bit;f=cuddT(f);}else f=cuddE(f);}return f==dd->one&&mask==0;}
int main(void){uint64_t hash=UINT64_C(14695981039346656037);unsigned table;for(table=0;table<256;table++){DdManager *dd=Cudd_Init(3,6,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *f=truth(dd,table,0),*p;int code;Cudd_Ref(f);Cudd_zddVarsFromBddVars(dd,2);p=Extra_zddPrimes(dd,f);Cudd_Ref(p);for(code=0;code<27;code++){int c=code,v;unsigned mask=0;for(v=0;v<3;v++){int state=c%3;c/=3;if(state)mask|=1u<<(2*v+state-1);}hash^=member(dd,p,mask);hash*=UINT64_C(1099511628211);}Cudd_Quit(dd);}printf("%016llx\n",(unsigned long long)hash);return 0;}

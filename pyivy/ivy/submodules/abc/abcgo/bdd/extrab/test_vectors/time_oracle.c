/* Original timed operators with timeout disabled. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
static uint64_t hash=UINT64_C(14695981039346656037);
static void byte(unsigned b){hash^=b;hash*=UINT64_C(1099511628211);}
static DdNode *truth(DdManager *dd,unsigned table,int level){DdNode *t,*e,*r;unsigned width,mask;if(level==3)return(table&1)?dd->one:Cudd_Not(dd->one);width=1u<<(2-level);mask=(1u<<width)-1;t=truth(dd,table>>width,level+1);Cudd_Ref(t);e=truth(dd,table&mask,level+1);Cudd_Ref(e);r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;}
static unsigned bits(DdManager *dd,DdNode *f){unsigned result=0;int a,v,input[6];for(a=0;a<8;a++){for(v=0;v<6;v++)input[v]=v<3?((a>>(2-v))&1):0;if(Cudd_Eval(dd,f,input)==dd->one)result|=1u<<a;}return result;}
static void bcollect(DdManager *dd,DdNode *f){Cudd_Ref(f);byte(bits(dd,f));Cudd_RecursiveDeref(dd,f);}
int main(void){DdManager *dd=Cudd_Init(3,6,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0),*dst=Cudd_Init(3,6,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *functions[256];int a,b,perm[3]={2,0,1};for(a=0;a<256;a++){functions[a]=truth(dd,a,0);Cudd_Ref(functions[a]);}for(a=0;a<256;a++){bcollect(dst,Extra_TransferPermuteTime(dd,dst,functions[a],perm,0));for(b=0;b<256;b++){bcollect(dd,Extra_bddAndTime(dd,functions[a],functions[b],0));bcollect(dd,Extra_bddAndAbstractTime(dd,functions[a],functions[b],dd->vars[1],0));}}printf("%016llx\n",(unsigned long long)hash);Cudd_Quit(dd);Cudd_Quit(dst);return 0;}

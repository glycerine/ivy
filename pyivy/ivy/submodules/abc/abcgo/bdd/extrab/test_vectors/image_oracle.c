/* Original tree-scheduled and monolithic image implementations. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include "bdd/extrab/extraBdd.h"
#include <stdint.h>
static uint64_t hash=UINT64_C(14695981039346656037);
static void byte(unsigned b){hash^=b;hash*=UINT64_C(1099511628211);}
static DdNode *truth(DdManager *dd,unsigned table,int level,int n){DdNode *t,*e,*r;unsigned width,mask;if(level==n)return(table&1)?dd->one:Cudd_Not(dd->one);width=1u<<(n-level-1);mask=(1u<<width)-1;t=truth(dd,table>>width,level+1,n);Cudd_Ref(t);e=truth(dd,table&mask,level+1,n);Cudd_Ref(e);r=Cudd_bddIte(dd,dd->vars[level],t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;}
static unsigned bits(DdManager *dd,DdNode *f){unsigned result=0;int a,v,input[3];for(a=0;a<8;a++){for(v=0;v<3;v++)input[v]=(a>>(2-v))&1;if(Cudd_Eval(dd,f,input)==dd->one)result|=1u<<a;}return result;}
int main(void){uint64_t seed=17;int sample;for(sample=0;sample<128;sample++){DdManager *dd=Cudd_Init(3,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *parts[2],*initial,*care,*keep[1];Extra_ImageTree_t *tree;Extra_ImageTree2_t *mono;int count=sample%3,i,table;for(i=0;i<count;i++){seed^=seed<<13;seed^=seed>>7;seed^=seed<<17;parts[i]=truth(dd,seed&255,0,3);Cudd_Ref(parts[i]);}initial=Cudd_bddXor(dd,dd->vars[0],dd->vars[1]);Cudd_Ref(initial);keep[0]=dd->vars[2];tree=Extra_bddImageStart(dd,initial,count,parts,1,keep,0);mono=Extra_bddImageStart2(dd,initial,count,parts,1,keep,0);byte(bits(dd,Extra_bddImageRead(tree)));byte(bits(dd,Extra_bddImageRead2(mono)));for(table=0;table<16;table++){care=truth(dd,table,0,2);Cudd_Ref(care);byte(bits(dd,Extra_bddImageCompute(tree,care)));byte(bits(dd,Extra_bddImageCompute2(mono,care)));Cudd_RecursiveDeref(dd,care);}Extra_bddImageTreeDelete(tree);Extra_bddImageTreeDelete2(mono);Cudd_Quit(dd);}printf("%016llx\n",(unsigned long long)hash);return 0;}

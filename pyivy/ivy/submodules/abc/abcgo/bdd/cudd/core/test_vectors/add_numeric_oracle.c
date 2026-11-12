/* Independent fixture generator for original CUDD ADD arithmetic families. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdint.h>
#include <string.h>
static double leaves[]={-3.,-2.,-1.,0.,1.,2.,3.,4.};
static uint64_t hash=UINT64_C(14695981039346656037);
static void byte(unsigned b){hash^=b;hash*=UINT64_C(1099511628211);}
static void value(double d){uint64_t bits;int i;memcpy(&bits,&d,8);for(i=0;i<8;i++)byte((bits>>(8*i))&255);}
static DdNode *build(DdManager *dd,int level,int start,int end){
 DdNode *t,*e,*r;if(end-start==1)return Cudd_addConst(dd,leaves[start]);
 t=build(dd,level+1,(start+end)/2,end);Cudd_Ref(t);e=build(dd,level+1,start,(start+end)/2);Cudd_Ref(e);
 r=Cudd_addIte(dd,Cudd_addIthVar(dd,level),t,e);Cudd_Ref(r);Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;
}
static void collect(DdManager *dd,DdNode *f){int a,i,inputs[4];if(!f){byte(255);return;}byte(0);Cudd_Ref(f);for(a=0;a<16;a++){for(i=0;i<4;i++)inputs[i]=(a>>(3-i))&1;value(cuddV(Cudd_Eval(dd,f,inputs)));}Cudd_RecursiveDeref(dd,f);}
int main(void){DdManager *dd=Cudd_Init(4,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *f,*boolean,*cube,*temp;int mask,i;double eps[]={0.,.5,1.,4.};
 f=build(dd,0,0,8);Cudd_Ref(f);boolean=Cudd_addBddPattern(dd,f);Cudd_Ref(boolean);temp=Cudd_BddToAdd(dd,boolean);Cudd_Ref(temp);Cudd_RecursiveDeref(dd,boolean);boolean=temp;
 for(mask=0;mask<16;mask++){cube=dd->one;Cudd_Ref(cube);for(i=0;i<4;i++)if(mask&(1<<i)){temp=Cudd_addApply(dd,Cudd_addTimes,cube,Cudd_addIthVar(dd,i));Cudd_Ref(temp);Cudd_RecursiveDeref(dd,cube);cube=temp;}collect(dd,Cudd_addExistAbstract(dd,f,cube));collect(dd,Cudd_addUnivAbstract(dd,f,cube));collect(dd,Cudd_addOrAbstract(dd,boolean,cube));Cudd_RecursiveDeref(dd,cube);}
 collect(dd,Cudd_addFindMax(dd,f));collect(dd,Cudd_addFindMin(dd,f));
 for(i=0;i<4;i++)collect(dd,Cudd_addScalarInverse(dd,f,Cudd_addConst(dd,eps[i])));
 for(i=0;i<5;i++)collect(dd,Cudd_addIthBit(dd,f,i));collect(dd,Cudd_addIthBit(dd,f,31));
 printf("%016llx\n",(unsigned long long)hash);Cudd_Quit(dd);return 0;
}

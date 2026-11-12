/* Original CUDD numeric-to-Boolean bridge fixture, no port code involved. */
#include "misc/util/util_hack.h"
#include "bdd/cudd/cuddInt.h"
#include <stdint.h>
#include <math.h>
static double leaves[]={-2147483648.,-17.75,-1.25,0.,.75,2.5,17.,2147483647.};
static uint64_t hash=UINT64_C(14695981039346656037);
static DdNode *build(DdManager *dd,int level,int start,int end){
 DdNode *t,*e,*r;
 if(end-start==1)return Cudd_addConst(dd,leaves[start]);
 t=build(dd,level+1,(start+end)/2,end);Cudd_Ref(t);
 e=build(dd,level+1,start,(start+end)/2);Cudd_Ref(e);
 r=Cudd_addIte(dd,Cudd_addIthVar(dd,level),t,e);Cudd_Ref(r);
 Cudd_RecursiveDeref(dd,t);Cudd_RecursiveDeref(dd,e);Cudd_Deref(r);return r;
}
static void collect(DdManager *dd,DdNode *f){
 int a,i,inputs[3];Cudd_Ref(f);
 for(a=0;a<8;a++){for(i=0;i<3;i++)inputs[i]=(a>>(2-i))&1;hash^=(Cudd_Eval(dd,f,inputs)==dd->one);hash*=UINT64_C(1099511628211);}
 Cudd_RecursiveDeref(dd,f);
}
int main(void){
 double thresholds[]={-20.,-1.,0.,.75,17.,100.,NAN,-INFINITY,INFINITY};
 double intervals[][2]={{-20.,20.},{-1.,0.},{0.,0.},{2.,-2.},{NAN,0.},{0.,NAN}};
 DdManager *dd=Cudd_Init(3,0,CUDD_UNIQUE_SLOTS,CUDD_CACHE_SLOTS,0);DdNode *f;int i;
 f=build(dd,0,0,8);Cudd_Ref(f);
 for(i=0;i<9;i++){collect(dd,Cudd_addBddThreshold(dd,f,thresholds[i]));collect(dd,Cudd_addBddStrictThreshold(dd,f,thresholds[i]));}
 for(i=0;i<6;i++)collect(dd,Cudd_addBddInterval(dd,f,intervals[i][0],intervals[i][1]));
 for(i=0;i<32;i++)collect(dd,Cudd_addBddIthBit(dd,f,i));
 printf("%016llx\n",(unsigned long long)hash);Cudd_Quit(dd);return 0;
}

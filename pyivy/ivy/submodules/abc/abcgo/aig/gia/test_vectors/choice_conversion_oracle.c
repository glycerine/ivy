#include "aig/gia/gia.h"
#include "aig/gia/giaAig.h"
#include "aig/aig/aig.h"
#include <stdio.h>
#include <stdlib.h>

static Gia_Man_t *snapshot(int which) {
 Gia_Man_t *g=Gia_ManStart(16);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g);
 int x=Gia_ManAppendAnd(g,a,b^1),y=Gia_ManAppendAnd(g,b,c);
 Gia_ManAppendCo(g,x^(which&1));Gia_ManAppendCo(g,y^((which>>1)&1));return g;
}
int main(void) {
 Gia_Man_t *g[3],*m; Aig_Man_t *a;Vec_Ptr_t *v=Vec_PtrAlloc(3);Gia_Obj_t *o;Aig_Obj_t *n;int i,k,mask,ci;int gv[128],av[128];
 for(i=0;i<3;i++){g[i]=snapshot(i);Vec_PtrPush(v,g[i]);}m=Gia_ManChoiceMiter(v);
 printf("gia %d %d %d\n",Gia_ManObjNum(m),Gia_ManCiNum(m),Gia_ManCoNum(m));
 Gia_ManForEachObj(m,o,i){int kind=Gia_ObjIsConst0(o)?0:Gia_ObjIsCi(o)?1:Gia_ObjIsCo(o)?3:2;
  int f0=kind>=2?Gia_ObjFaninLit0(o,i):-1, f1=kind==2?Gia_ObjFaninLit1(o,i):-1;
  printf("g %d %d %d %d\n",i,kind,f0,f1);o->Value=100+2*i;}
 Gia_ManLevelNum(m);Gia_ManForEachCi(m,o,i) Gia_ObjSetLevelId(m,Gia_ObjId(m,o),7+i);
 a=Gia_ManToAigSkip(m,3);
 Gia_ManForEachObj(m,o,i) printf("value %d %u\n",i,o->Value);
 printf("aig %d %d %d\n",Aig_ManObjNum(a),Aig_ManCiNum(a),Aig_ManCoNum(a));
 Aig_ManForEachObj(a,n,i){int kind=Aig_ObjIsConst1(n)?0:Aig_ObjIsCi(n)?1:Aig_ObjIsCo(n)?3:2;
  int f0=kind>=2?Abc_Var2Lit(Aig_ObjId(Aig_ObjFanin0(n)),Aig_ObjFaninC0(n)):-1;
  int f1=kind==2?Abc_Var2Lit(Aig_ObjId(Aig_ObjFanin1(n)),Aig_ObjFaninC1(n)):-1;
  printf("a %d %d %d %d\n",i,kind,f0,f1);}
 Aig_ManForEachCi(a,n,i) printf("level %d %d\n",i,Aig_ObjLevel(n));
 for(mask=0;mask<8;mask++){
  gv[0]=0;Gia_ManForEachObj1(m,o,i){if(Gia_ObjIsCi(o))gv[i]=(mask>>Gia_ObjCioId(o))&1;else {gv[i]=gv[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o);if(Gia_ObjIsAnd(o))gv[i]&=gv[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o);}}
  av[0]=1;ci=0;Aig_ManForEachObj(a,n,i){if(i==0)continue;if(Aig_ObjIsCi(n))av[i]=(mask>>ci++)&1;else {av[i]=av[Aig_ObjId(Aig_ObjFanin0(n))]^Aig_ObjFaninC0(n);if(Aig_ObjIsNode(n))av[i]&=av[Aig_ObjId(Aig_ObjFanin1(n))]^Aig_ObjFaninC1(n);}}
  printf("truth %d",mask);Gia_ManForEachCo(m,o,i)printf(" %d",gv[Gia_ObjId(m,o)]);printf(" |");Aig_ManForEachCo(a,n,i)printf(" %d",av[Aig_ObjId(n)]);puts("");
 }
 Aig_ManStop(a);Gia_ManStop(m);for(k=0;k<3;k++)Gia_ManStop(g[k]);Vec_PtrFree(v);return 0;
}

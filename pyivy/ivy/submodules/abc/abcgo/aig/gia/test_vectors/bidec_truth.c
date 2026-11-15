#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern unsigned *Gia_ManConvertAigToTruth(Gia_Man_t *,Gia_Obj_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t *);
static Gia_Man_t *graph(int n,int signs){
 Gia_Man_t *g=Gia_ManStart(128);int ci[16],i,root=0,lit;
 for(i=0;i<n;i++)ci[i]=Gia_ManAppendCi(g);
 if(n)root=ci[0]^(signs&1);Gia_ManHashStart(g);
 for(i=1;i<n;i++){
  lit=ci[i]^((signs>>(i%2))&1);
  if(i%3==0)root=Gia_ManHashXor(g,root,lit);
  else if(i%3==1)root=Gia_ManHashAnd(g,root,lit);
  else root=Gia_ManHashMux(g,lit,root,ci[i-1]^1);
 }
 Gia_ManHashStop(g);Gia_ManAppendCo(g,root);return g;
}
static int root(Gia_Man_t *g,int selector){switch(selector){
 case 0:return Gia_ObjFaninLit0p(g,Gia_ManCo(g,0));
 case 1:return Gia_ObjFaninLit0p(g,Gia_ManCo(g,0))^1;
 case 2:return Gia_ManCiNum(g)?Gia_Obj2Lit(g,Gia_ManCi(g,0)):0;
 case 3:return Gia_ManCiNum(g)?Gia_Obj2Lit(g,Gia_ManCi(g,Gia_ManCiNum(g)-1))^1:1;
 case 4:return 0;default:return 1;}}
int main(void){int n,signs,rev,selector,i,marker;Gia_Man_t *g;Vec_Int_t *leaves,*storage,*visited;unsigned *truth;
 for(n=0;n<=15;n++)for(signs=0;signs<4;signs++)for(rev=0;rev<2;rev++)for(selector=0;selector<6;selector++){
  g=graph(n,signs);Gia_ManCleanTruth(g);leaves=Vec_IntDup(g->vCis);if(rev)Vec_IntReverseOrder(leaves);
  storage=Vec_IntAlloc(1<<16);visited=Vec_IntAlloc(64);
  truth=Gia_ManConvertAigToTruth(g,Gia_ObjFromLit(g,root(g,selector)),leaves,storage,visited);
  printf("case %d %d %d %d",n,signs,rev,selector);for(i=0;i<Abc_TruthWordNum(n);i++)printf(" %08x",truth[i]);puts("");
  Vec_IntForEachEntry(g->vTruths,marker,i)assert(marker==-1);
  Vec_IntFree(leaves);Vec_IntFree(storage);Vec_IntFree(visited);Gia_ManStop(g);
 }
 return 0;
}

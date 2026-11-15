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
static int resynthesis(void){
 int n,signs,i,mask,packed,*values;Gia_Man_t *g,*out;Gia_Obj_t *o;
 for(n=2;n<=8;n++)for(signs=0;signs<4;signs++){
  g=graph(n,signs);Gia_ManAppendCo(g,root(g,1));Gia_ManAppendCo(g,Gia_ManCiLit(g,0)^(signs&1));Gia_ManSetRegNum(g,1);
  g->vMapping=Vec_IntStart(g->nObjs);
  Gia_ManForEachAnd(g,o,i){Vec_IntWriteEntry(g->vMapping,i,Vec_IntSize(g->vMapping));Vec_IntPush(g->vMapping,2);Vec_IntPush(g->vMapping,Gia_ObjFaninId0(o,i));Vec_IntPush(g->vMapping,Gia_ObjFaninId1(o,i));Vec_IntPush(g->vMapping,i);}
  out=Gia_ManPerformBidec(g,0);
  printf("resynthesis %d %d %d %d %d",n,signs,Gia_ManCiNum(out),Gia_ManCoNum(out),Gia_ManRegNum(out));
  values=ABC_CALLOC(int,out->nObjs);
  for(mask=0;mask<(1<<n);mask++){
   Gia_ManForEachCi(out,o,i)values[Gia_ObjId(out,o)]=(mask>>i)&1;
   Gia_ManForEachAnd(out,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
   packed=0;Gia_ManForEachCo(out,o,i)packed|=(values[Gia_ObjFaninId0p(out,o)]^Gia_ObjFaninC0(o))<<i;printf(" %d",packed);
  }
  puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(g);
 }
 return 0;
}
int main(int argc,char **argv){
 if(argc>1 && !strcmp(argv[1],"resynthesis"))return resynthesis();int n,signs,rev,selector,i,marker;Gia_Man_t *g;Vec_Int_t *leaves,*storage,*visited;unsigned *truth;
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

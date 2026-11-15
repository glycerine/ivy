#include "aig/gia/gia.h"
#include "sat/bsat/satSolver.h"
static int safe_batch;
static int review_solve(sat_solver *,lit *,lit *,ABC_INT64_T,ABC_INT64_T,ABC_INT64_T,ABC_INT64_T);
#define sat_solver_solve review_solve
#include "aig/gia/giaFalse.c"
#undef sat_solver_solve
// Default mode is unmodified C. Safe mode changes only the demonstrated
// original-PI transition bug, using the existing solver and same total budget.
static int review_solve(sat_solver *s,lit *begin,lit *end,ABC_INT64_T conflicts,ABC_INT64_T propagations,ABC_INT64_T globalConflicts,ABC_INT64_T globalPropagations) {
 int status=sat_solver_solve(s,begin,end,conflicts,propagations,globalConflicts,globalPropagations),n,i;lit *core;
 if(!safe_batch || status!=l_False || begin==end || begin[0]!=0)return status;
 n=sat_solver_final(s,&core);
 for(i=0;i<n;++i)if(Abc_Lit2Var(core[i])==0){
  ABC_INT64_T cap=conflicts>0 && (!globalConflicts || conflicts<globalConflicts)?conflicts:globalConflicts;
  return sat_solver_solve(s,begin+1,end,conflicts,propagations,cap,globalPropagations);
 }
 return status;
}
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static unsigned rng;
static unsigned next_random(void){rng=rng*1664525u+1013904223u;return rng>>8;}
static Gia_Man_t *graph(unsigned seed,int *root){
 Gia_Man_t *g=Gia_ManStart(128);int pool[32],n=0,i,a,b,ia,ib; rng=seed;
 for(i=0;i<4;++i)pool[n++]=Gia_ManAppendCi(g);
 for(i=0;i<12;++i){ia=next_random()%n;ib=next_random()%n;if(ia==ib)ib=(ib+1)%n;a=pool[ia]^(next_random()&1);b=pool[ib]^(next_random()&1);pool[n++]=Gia_ManAppendAnd(g,a,b);}
 *root=pool[n-1]>>1;Gia_ManAppendCo(g,pool[n-1]);Gia_ManAppendCo(g,pool[n-2]^1);return g;
}
static void inspect_single(void){unsigned seed;int root,i,bits,packed,values[4096];Gia_Man_t *g,*out;Vec_Int_t *path;Gia_Obj_t *o;
 for(seed=0;seed<64;++seed){g=graph(seed,&root);Gia_ManLevelNum(g);path=Gia_ManCollectPath(g,Gia_ManObj(g,root));printf("path %u",seed);Vec_IntForEachEntry(path,i,bits)printf(" %d",i);puts("");Vec_IntFree(path);
 out=Gia_ManCheckOne(g,0,root,0,0,0);printf("result %u %d",seed,out?Gia_ManObjNum(out):-1);
 if(out){assert(Gia_ManObjNum(out)<4096);for(bits=0;bits<16;++bits){memset(values,0,sizeof(values));Gia_ManForEachCi(out,o,i)values[Gia_ObjId(out,o)]=(bits>>i)&1;
 Gia_ManForEachAnd(out,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
 packed=0;Gia_ManForEachCo(out,o,i)packed|=(values[Gia_ObjFaninId0p(out,o)]^Gia_ObjFaninC0(o))<<i;printf(" %d",packed);}Gia_ManStop(out);}puts("");Gia_ManStop(g);
 }return;
}

static void inspect_batch(void) {
 unsigned seed;int root,i,j,id;Gia_Man_t *g;Gia_Obj_t *o;Vec_Wec_t *hooks;Vec_Int_t *path;
 for(seed=0;seed<64;++seed){
  g=graph(seed,&root);Gia_ManLevelNum(g);hooks=Vec_WecStart(Gia_ManObjNum(g));
  Gia_ManForEachObj(g,o,i)o->Value=71+17*i;
  for(i=0;i<Gia_ManCoNum(g);++i)Gia_ManCheckFalseOne(g,i,0,hooks,0,0);
  printf("batch %u",seed);Vec_WecForEachLevel(hooks,path,i)if(Vec_IntSize(path)){printf(" [%d",i);Vec_IntForEachEntry(path,id,j)printf(" %d",id);printf("]");}puts("");
  printf("scratch");Gia_ManForEachObj(g,o,i)printf(" %u",o->Value);puts("");
  Vec_WecFree(hooks);Gia_ManStop(g);
 }
}
static void inspect_priority(void){
 Gia_Man_t *g=Gia_ManStart(32),*out;int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),x=Gia_ManAppendAnd(g,a,b),root=Gia_ManAppendAnd(g,x,a^1),i;
 for(i=0;i<5;++i)Gia_ManAppendCo(g,root^(i&1));out=Gia_ManCheckFalse2(g,0,0,1,0);Gia_ManStop(out);Gia_ManStop(g);
}
static void inspect_verbose(void){
 Gia_Man_t *g,*out;int mode,a,b,x,root,i,verbose,very;
 for(verbose=0;verbose<2;++verbose)for(very=0;very<2;++very)for(mode=0;mode<3;++mode){
  g=Gia_ManStart(32);a=Gia_ManAppendCi(g);b=Gia_ManAppendCi(g);x=Gia_ManAppendAnd(g,a,b);root=mode==2?x:Gia_ManAppendAnd(g,x,a^1);
  for(i=0;i<(mode==1?5:1);++i)Gia_ManAppendCo(g,root^(i&1));Gia_ManLevelNum(g);printf("mode %d %d %d\n",mode,verbose,very);
  out=mode==1?Gia_ManCheckFalse2(g,0,0,verbose,very):Gia_ManCheckOne(g,0,root>>1,0,verbose,very);
  if(out)Gia_ManStop(out);Gia_ManStop(g);
 }
}
static int output_bits(Gia_Man_t *g,int bits){
 int values[4096]={0},i,packed=0;Gia_Obj_t *o;assert(Gia_ManObjNum(g)<4096);
 Gia_ManForEachCi(g,o,i)values[Gia_ObjId(g,o)]=(bits>>i)&1;
 Gia_ManForEachAnd(g,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
 Gia_ManForEachCo(g,o,i)packed|=(values[Gia_ObjFaninId0p(g,o)]^Gia_ObjFaninC0(o))<<i;return packed;
}
static Gia_Man_t *reconvergent_graph(void){
 Gia_Man_t *g=Gia_ManStart(32);int a=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),d=Gia_ManAppendCi(g);
 int ad=Gia_ManAppendAnd(g,a,d),left=Gia_ManAppendAnd(g,c^1,ad),right=Gia_ManAppendAnd(g,c,ad^1),xnor=Gia_ManAppendAnd(g,left^1,right^1);
 int low=Gia_ManAppendAnd(g,c,d^1),high=Gia_ManAppendAnd(g,d,xnor);Gia_ManAppendCo(g,Gia_ManAppendAnd(g,low^1,high^1));return g;
}
static void inspect_reconvergence(void){
 Gia_Man_t *g,*out;int mode,bits;
 for(mode=0;mode<2;++mode){safe_batch=mode;g=reconvergent_graph();out=Gia_ManCheckFalse2(g,32,0,0,0);printf("safety %d",mode);for(bits=0;bits<8;++bits)printf(" %d/%d",output_bits(g,bits),output_bits(out,bits));puts("");Gia_ManStop(out);Gia_ManStop(g);}
}
int main(int argc,char **argv){
 if(argc>1 && !strcmp(argv[1],"--batch"))inspect_batch();
 else if(argc>1 && !strcmp(argv[1],"--safe-batch")){safe_batch=1;inspect_batch();}
 else if(argc>1 && !strcmp(argv[1],"--priority"))inspect_priority();
 else if(argc>1 && !strcmp(argv[1],"--verbose"))inspect_verbose();
 else if(argc>1 && !strcmp(argv[1],"--reconvergence"))inspect_reconvergence();
 else inspect_single();return 0;
}

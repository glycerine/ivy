#include "aig/gia/gia.h"
#include "aig/gia/giaEsop.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static unsigned rng;
static unsigned random_next(void) {rng=rng*1664525u+1013904223u;return rng>>8;}
static int random_edge(int *pool,int n){int lit=pool[random_next()%n];return lit^(random_next()&1);}
static Gia_Man_t *make_graph(unsigned seed) {
 Gia_Man_t *g=Gia_ManStart(256);int pool[32]={0,1},n=2,i,a,b,c,op;
 rng=seed;for(i=0;i<4;++i)pool[n++]=Gia_ManAppendCi(g);Gia_ManHashAlloc(g);
 for(i=0;i<12;++i){a=random_edge(pool,n);b=random_edge(pool,n);c=random_edge(pool,n);op=random_next()%3;
 pool[n++]=op==0?Gia_ManHashAnd(g,a,b):op==1?Gia_ManHashXor(g,a,b):Gia_ManHashMux(g,a,b,c);}
 for(i=0;i<3;++i)Gia_ManAppendCo(g,pool[n-1-i]^(seed&1));Gia_ManHashStop(g);return g;
}
static void inspect_covers(void){unsigned seed;int i,k,lit;Vec_Wec_t *cover;Vec_Int_t *cube;Gia_Man_t *g;
 for(seed=0;seed<64;++seed){g=make_graph(seed);cover=NULL;Eso_ManCompute(g,0,&cover);
 printf("cover %u levels%d\n",seed,Vec_WecSize(cover));
 Vec_WecForEachLevel(cover,cube,i){printf("cube");Vec_IntForEachEntry(cube,lit,k)printf(" %d",lit);puts("");}
 Vec_WecFree(cover);Gia_ManStop(g);
 }
 return;
}

static void inspect_cubes(void) {
 Gia_Man_t *g=Gia_ManStart(16);Eso_Man_t *p;Vec_Int_t *cubes[81],*out=Vec_IntAlloc(4),*esop=Vec_IntAlloc(8);
 int ids[81],a,b,i,n,x,equal,unequal,product;
 for(i=0;i<4;++i)Gia_ManAppendCi(g);p=Eso_ManAlloc(g);
 for(a=0;a<81;++a){
  cubes[a]=Vec_IntAlloc(4);n=a;
  for(i=0;i<4;++i){x=n%3;n/=3;if(x)Vec_IntPush(cubes[a],2*i+(x==2));}
  ids[a]=a?Hsh_VecManAdd(p->pHash,cubes[a]):p->Cube1;
 }
 for(a=0;a<81;++a)for(b=0;b<81;++b){
  equal=unequal=-2;
  if(Vec_IntSize(cubes[a])==Vec_IntSize(cubes[b]))equal=Eso_ManFindDistOneLitEqual(Vec_IntArray(cubes[a]),Vec_IntArray(cubes[b]),Vec_IntSize(cubes[a]));
  if(Vec_IntSize(cubes[a])+1==Vec_IntSize(cubes[b]))unequal=Eso_ManFindDistOneLitNotEqual(Vec_IntArray(cubes[a]),Vec_IntArray(cubes[b]),Vec_IntSize(cubes[a]));
  product=!a?ids[b]:!b?ids[a]:Eso_ManComputeAnd(p,cubes[a],cubes[b],out);
  Eso_ManMinimizeAdd(p,ids[a]);Eso_ManMinimizeAdd(p,ids[b]);Eso_ManMinimizeCopy(p,esop);
  printf("pair %d %d eq%d neq%d and%d xor",a,b,equal,unequal,product);
  Vec_IntForEachEntry(esop,x,i)printf(" %d",x);puts("");
 }
 for(a=0;a<81;++a)Vec_IntFree(cubes[a]);Vec_IntFree(esop);Vec_IntFree(out);Eso_ManStop(p);Gia_ManStop(g);
}
static void inspect_verbose(void) {
 Gia_Man_t *g=make_graph(3),*out=Eso_ManCompute(g,1,NULL);Gia_ManStop(out);Gia_ManStop(g);
}
static void inspect_print(void) {
 Gia_Man_t *g=Gia_ManStart(8);Eso_Man_t *p;Vec_Int_t *cover=Vec_IntAlloc(4);int i;
 for(i=0;i<4;++i)Gia_ManAppendCi(g);p=Eso_ManAlloc(g);
 Eso_ManCoverPrint(p,cover);Vec_IntPush(cover,p->Cube1);Vec_IntPush(cover,0);Vec_IntPush(cover,3);Eso_ManCoverPrint(p,cover);
 Vec_IntFree(cover);Eso_ManStop(p);Gia_ManStop(g);
}
int main(int argc,char **argv){
 if(argc>1 && !strcmp(argv[1],"--cubes"))inspect_cubes();
 else if(argc>1 && !strcmp(argv[1],"--verbose"))inspect_verbose();
 else if(argc>1 && !strcmp(argv[1],"--print"))inspect_print();
 else inspect_covers();
 return 0;
}

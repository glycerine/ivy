// Execute the actual original edge routines; this fixture only constructs
// valid mapped AND cones and serializes their complete assignment state.
#include "aig/gia/giaEdge.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static unsigned state;
static unsigned next(void){state=1664525u*state+1013904223u;return state>>8;}
static Gia_Man_t *graph(unsigned seed){
 Gia_Man_t *g=Gia_ManStart(128);Vec_Int_t *records=Vec_IntAlloc(128);int lits[20],n=4,i,j,k,leaf[4],root;
 state=seed;for(i=0;i<4;++i)lits[i]=Gia_ManAppendCi(g);
 for(i=0;i<16;++i){
  int size=2+next()%3;
  for(j=0;j<size;++j){int v;do{v=next()%n;for(k=0;k<j&&leaf[k]!=v;++k);}while(k<j);leaf[j]=v;}
  root=lits[leaf[0]];for(j=1;j<size;++j)root=Gia_ManAppendAnd(g,root,lits[leaf[j]]);
  Vec_IntPush(records,root>>1);Vec_IntPush(records,size);for(j=0;j<size;++j)Vec_IntPush(records,lits[leaf[j]]>>1);lits[n++]=root;
 }
 for(i=14;i<20;++i)Gia_ManAppendCo(g,lits[i]);
 g->vMapping=Vec_IntStart(Gia_ManObjNum(g));
 for(i=0;i<Vec_IntSize(records);){int id=Vec_IntEntry(records,i++),size=Vec_IntEntry(records,i++);Vec_IntWriteEntry(g->vMapping,id,Vec_IntSize(g->vMapping));Vec_IntPush(g->vMapping,size);for(j=0;j<size;++j)Vec_IntPush(g->vMapping,Vec_IntEntry(records,i++));Vec_IntPush(g->vMapping,id);}
 Vec_IntFree(records);return g;
}
static void emit(const char *name,Vec_Int_t *v){int i;printf("%s",name);for(i=0;i<Vec_IntSize(v);++i)printf(" %d",Vec_IntEntry(v,i));puts("");}
static void snapshot(Gia_Man_t *g){emit("edge1",g->vEdge1);emit("edge2",g->vEdge2);emit("delay",g->vEdgeDelay);}

static int wide(void){
 int sizes[]={5,8,16},s,mode,i;
 for(s=0;s<3;++s)for(mode=0;mode<2;++mode){
  int n=sizes[s],lits[16],root;Gia_Man_t *g=Gia_ManStart(2*n+2);
  for(i=0;i<n;++i)lits[i]=Gia_ManAppendCi(g);
  root=lits[0];for(i=1;i<n;++i)root=Gia_ManAppendAnd(g,root,lits[i]);Gia_ManAppendCo(g,root);
  if(mode){g->vMapping2=Vec_WecStart(Gia_ManObjNum(g));for(i=0;i<n;++i)Vec_WecPush(g->vMapping2,root>>1,lits[i]>>1);}
  else{g->vMapping=Vec_IntStart(Gia_ManObjNum(g));Vec_IntWriteEntry(g->vMapping,root>>1,Vec_IntSize(g->vMapping));Vec_IntPush(g->vMapping,n);for(i=0;i<n;++i)Vec_IntPush(g->vMapping,lits[i]>>1);Vec_IntPush(g->vMapping,root>>1);}
  printf("wide %d %d result %d\n",n,mode,Gia_ManComputeEdgeDelay2(g));snapshot(g);Gia_ManStop(g);
 }
 return 0;
}


static int reports(void){
 int limit;
 for(limit=0;limit<3;++limit){
  Gia_Man_t *g=graph(1);int result;
  printf("reports %d\n",limit);result=Edg_ManAssignEdgeNew(g,limit,1);
  printf("result %d\n",result);snapshot(g);Gia_ManStop(g);
 }
 return 0;
}

int main(int argc,char **argv){if(argc>1){if(!strcmp(argv[1],"reports"))return reports();return wide();}int seed,mode,two,limit;for(seed=1;seed<=32;++seed)for(mode=0;mode<2;++mode){
 Gia_Man_t *g=graph(seed);if(mode){Edg_ManToMapping(g);Vec_IntFreeP(&g->vMapping);}
 printf("case %d %d\n",seed,mode);
 for(two=0;two<2;++two){printf("forward %d %d\n",two,Gia_ManComputeEdgeDelay(g,two));snapshot(g);printf("eval %d count %d\n",Gia_ManEvalEdgeDelay(g),Gia_ManEvalEdgeCount(g));emit("evaluated",g->vEdgeDelay);}
 printf("reverse %d\n",Gia_ManComputeEdgeDelay2(g));snapshot(g);
 if(!mode)for(limit=0;limit<3;++limit){printf("opt %d %d\n",limit,Edg_ManAssignEdgeNew(g,limit,0));snapshot(g);printf("unit %d %d\n",Edg_ManEvalEdgeDelay(g),Edg_ManEvalEdgeDelayR(g));}
 Gia_ManStop(g);
 }return 0;}

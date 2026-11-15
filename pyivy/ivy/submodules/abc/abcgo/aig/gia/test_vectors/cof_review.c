#include "aig/gia/gia.h"
#include "aig/gia/giaCof.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"

static unsigned rng;
static unsigned next_random(void) { rng=rng*1664525u+1013904223u;return rng>>8; }
static int edge(int *pool,int size) {int lit=pool[next_random()%size];return lit^(next_random()&1);}
static Gia_Man_t *graph(unsigned seed) {
 Gia_Man_t *g=Gia_ManStart(256);int pool[32]={0,1},n=2,i,a,b,c,op;
 rng=seed;
 for(i=0;i<4;++i)pool[n++]=Gia_ManAppendCi(g);
 Gia_ManHashAlloc(g);
 for(i=0;i<12;++i){a=edge(pool,n);b=edge(pool,n);c=edge(pool,n);op=next_random()%3;
  pool[n++]=op==0?Gia_ManHashAnd(g,a,b):op==1?Gia_ManHashXor(g,a,b):Gia_ManHashMux(g,a,b,c);
 }
 for(i=0;i<3;++i)Gia_ManAppendCo(g,pool[n-1-i]^(seed&1));
 return g;
}
static void inspect(unsigned seed) {
 Gia_Man_t *g=graph(seed);Cof_Man_t *p=Cof_ManCreateLogicSimple(g);Cof_Obj_t *o,*a,*b,*roots[2];Vec_Ptr_t *top;
 Gia_Obj_t *obj;int i,j,k,ncand=0,candidates[256],fwd,rev,supp,removed;
 p->nLevels=Gia_ManLevelNum(g)+1;p->pLevels=ABC_CALLOC(int,p->nLevels);Cof_ManResetTravId(p);
 printf("case %u objects%d nodes%d\n",seed,p->nObjs,p->nNodes);
 Gia_ManForEachObj(g,obj,i){
  o=Cof_ManObj(p,Gia_ObjHandle(obj));printf("object %d term%d fanins%d fanouts%d mux%d",o->Id,Cof_ObjIsTerm(o),Cof_ObjFaninNum(o),Cof_ObjFanoutNum(o),o->nFanoutsM);
  for(j=0;j<Cof_ObjFaninNum(o);++j)printf(" in%d/%d",Cof_ObjFanin(o,j)->Id,o->Fanios[j].fCompl);
  for(j=0;j<Cof_ObjFanoutNum(o);++j)printf(" out%d",Cof_ObjFanout(o,j)->Id);
  puts("");if(Cof_ObjIsCi(o)||Cof_ObjIsNode(o))candidates[ncand++]=i;
 }
 for(i=0;i<ncand;++i){
  a=Cof_ManObj(p,Gia_ObjHandle(Gia_ManObj(g,candidates[i])));roots[0]=a;
  fwd=Cof_ManTfoSize(p,roots,1);rev=Cof_ManTfiSize(p,roots,1);supp=Cof_ManSuppSize(p,roots,1);
  printf("single %d %d %d %d mffc%d\n",a->Id,fwd,rev,supp,Cof_ObjMffcSize(a));
  for(j=i+1;j<ncand;++j){b=Cof_ManObj(p,Gia_ObjHandle(Gia_ManObj(g,candidates[j])));roots[1]=b;
   fwd=Cof_ManTfoSize(p,roots,2);rev=Cof_ManTfiSize(p,roots,2);supp=Cof_ManSuppSize(p,roots,2);
   printf("pair %d %d %d %d %d\n",a->Id,b->Id,fwd,rev,supp);
  }
  for(k=0;k<2;++k){removed=Cof_ManCountRemoved(p,a,k);printf("removed %d %d %d",a->Id,k,removed);
   Gia_ManForEachObj(g,obj,j){o=Cof_ManObj(p,Gia_ObjHandle(obj));if(Cof_ObjIsTravIdCurrent(p,o))printf(" %d/%d",o->Id,o->iLit);}
   puts("");
  }
 }
 for(i=1;i<=ncand+1;++i){top=Cof_ManCollectHighFanout(p,i);printf("top %d",i);for(j=0;j<Vec_PtrSize(top);++j)printf(" %d",((Cof_Obj_t*)Vec_PtrEntry(top,j))->Id);puts("");Vec_PtrFree(top);}
 Cof_ManStop(p);Gia_ManHashStop(g);Gia_ManStop(g);
}
int main(void){unsigned seed;for(seed=0;seed<32;++seed)inspect(seed);return 0;}

#include "aig/gia/gia.h"
#include "aig/gia/giaFalse.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static unsigned rng;
static unsigned next_random(void){rng=rng*1664525u+1013904223u;return rng>>8;}
static Gia_Man_t *graph(unsigned seed,int *root){
 Gia_Man_t *g=Gia_ManStart(128);int pool[32],n=0,i,a,b,ia,ib; rng=seed;
 for(i=0;i<4;++i)pool[n++]=Gia_ManAppendCi(g);
 for(i=0;i<12;++i){ia=next_random()%n;ib=next_random()%n;if(ia==ib)ib=(ib+1)%n;a=pool[ia]^(next_random()&1);b=pool[ib]^(next_random()&1);pool[n++]=Gia_ManAppendAnd(g,a,b);}
 *root=pool[n-1]>>1;Gia_ManAppendCo(g,pool[n-1]);Gia_ManAppendCo(g,pool[n-2]^1);return g;
}
int main(void){unsigned seed;int root,i,bits,packed,values[4096];Gia_Man_t *g,*out;Vec_Int_t *path;Gia_Obj_t *o;
 for(seed=0;seed<64;++seed){g=graph(seed,&root);Gia_ManLevelNum(g);path=Gia_ManCollectPath(g,Gia_ManObj(g,root));printf("path %u",seed);Vec_IntForEachEntry(path,i,bits)printf(" %d",i);puts("");Vec_IntFree(path);
 out=Gia_ManCheckOne(g,0,root,0,0,0);printf("result %u %d",seed,out?Gia_ManObjNum(out):-1);
 if(out){assert(Gia_ManObjNum(out)<4096);for(bits=0;bits<16;++bits){memset(values,0,sizeof(values));Gia_ManForEachCi(out,o,i)values[Gia_ObjId(out,o)]=(bits>>i)&1;
 Gia_ManForEachAnd(out,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
 packed=0;Gia_ManForEachCo(out,o,i)packed|=(values[Gia_ObjFaninId0p(out,o)]^Gia_ObjFaninC0(o))<<i;printf(" %d",packed);}Gia_ManStop(out);}puts("");Gia_ManStop(g);
 }return 0;
}

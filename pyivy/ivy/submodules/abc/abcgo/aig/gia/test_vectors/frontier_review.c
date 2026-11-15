#include "aig/gia/giaFront.c"
static unsigned rng;
static unsigned next_random(void){rng=rng*1664525u+1013904223u;return rng>>8;}
static Gia_Man_t *graph(unsigned seed){
 Gia_Man_t *g=Gia_ManStart(128);int pool[40],n=1,i,ia,ib,a,b;
 rng=seed;pool[0]=0;for(i=0;i<4;++i)pool[n++]=Gia_ManAppendCi(g);
 for(i=0;i<24;++i){ia=next_random()%n;ib=next_random()%n;if(ia==ib)ib=(ib+1)%n;a=pool[ia]^(next_random()&1);b=pool[ib]^(next_random()&1);pool[n++]=Gia_ManAppendAnd(g,a,b);}
 Gia_ManAppendCo(g,pool[n-1]);Gia_ManAppendCo(g,pool[n-3]^1);Gia_ManAppendCo(g,0);Gia_ManAppendCi(g);Gia_ManSetRegNum(g,1);return g;
}
static void dump(Gia_Man_t *g,const char *tag){int i;Gia_Obj_t *o;printf("%s %d %d",tag,g->nObjs,g->nFront);Gia_ManForEachObj(g,o,i)printf(" [%u %u %u %u %u %u]",o->iDiff0,o->iDiff1,o->fCompl0,o->fCompl1,o->fTerm,o->Value);puts("");}
static int outputs(Gia_Man_t *g,int bits){
 int values[64]={0},i,packed=0;Gia_Obj_t *o;
 Gia_ManForEachCi(g,o,i)values[Gia_ObjId(g,o)]=(bits>>i)&1;
 Gia_ManForEachAnd(g,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
 Gia_ManForEachCo(g,o,i)packed|=(values[Gia_ObjFaninId0p(g,o)]^Gia_ObjFaninC0(o))<<i;return packed;
}
static void buffer_counterexample(void){
 Gia_Man_t *g=Gia_ManStart(32),*q;int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),x=Gia_ManAppendBuf(g,a),z=Gia_ManAppendBuf(g,b),y,bits;
 Gia_ManAppendBuf(g,b);Gia_ManAppendBuf(g,b);Gia_ManAppendBuf(g,b);y=Gia_ManAppendAnd(g,a,z);Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,y);
 q=Gia_ManFront(g);Gia_ManFrontTransform(q);for(bits=0;bits<4;++bits)printf("buffer %d %d %d\n",bits,outputs(g,bits),outputs(q,bits));Gia_ManStop(q);Gia_ManStop(g);
}
int main(int argc,char **argv){unsigned seed;Gia_Man_t *g,*q;if(argc>1 && !strcmp(argv[1],"--buffer")){buffer_counterexample();return 0;}for(seed=0;seed<64;++seed){g=graph(seed);printf("seed %u cut %d\n",seed,Gia_ManCrossCutSimple(g));q=Gia_ManFront(g);dump(q,"front");Gia_ManFrontTransform(q);dump(q,"restored");Gia_ManStop(q);Gia_ManStop(g);}return 0;}

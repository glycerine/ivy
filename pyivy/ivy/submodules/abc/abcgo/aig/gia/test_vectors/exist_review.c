#include "aig/gia/gia.h"
#include "aig/gia/giaExist.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static int keep(void *data,int ci){return (*(int*)data>>ci)&1;}
static void inspect_quant(void){int mode,mask,phase,lit,root,i,bits,values[4096];Gia_Man_t *g;Gia_Obj_t *o;
 for(mode=0;mode<2;++mode)for(mask=0;mask<8;++mask)for(phase=0;phase<2;++phase){
  int ins[3];if(mode==0 && mask==0)continue;g=Gia_ManStart(128);Gia_ManQuantSetSuppStart(g);
  for(i=0;i<3;++i){ins[i]=Gia_ManAppendCi(g);Gia_ManQuantSetSuppZero(g);Gia_ManQuantSetSuppCi(g,Gia_ManObj(g,ins[i]>>1));}
  Gia_ManHashAlloc(g);root=Gia_ManHashMux(g,ins[0],ins[1],ins[2])^phase;
  lit=mode?Gia_ManQuantExist2(g,root,keep,&mask):Gia_ManQuantExist(g,root,keep,&mask);
  printf("%d %d %d lit%d objs%d",mode,mask,phase,lit,Gia_ManObjNum(g));
  assert(Gia_ManObjNum(g)<4096);
  for(bits=0;bits<8;++bits){memset(values,0,sizeof(values));Gia_ManForEachCi(g,o,i)values[Gia_ObjId(g,o)]=(bits>>i)&1;
   Gia_ManForEachAnd(g,o,i)values[i]=(values[Gia_ObjFaninId0(o,i)]^Gia_ObjFaninC0(o))&(values[Gia_ObjFaninId1(o,i)]^Gia_ObjFaninC1(o));
   printf(" %d",values[lit>>1]^(lit&1));}
  puts("");Gia_ManHashStop(g);Gia_ManStop(g);
 }
 return;
}

static int keep_wide(void *data,int ci){return ((ci+*(int*)data)%5)!=0;}
static void inspect_support(void){
 int sizes[]={0,1,2,63,64,65,127,128,129,191,192,193},which,mode,n,i,j,w,id,offset=2,supports;
 Gia_Man_t *g;Gia_Obj_t *o;
 for(which=0;which<12;++which)for(mode=0;mode<2;++mode){
  n=sizes[which];g=Gia_ManStart(2*n+16);Gia_ManQuantSetSuppStart(g);
  for(i=0;i<n;++i){id=Gia_ManAppendCi(g)>>1;Gia_ManQuantSetSuppZero(g);
   if(!mode || i%3!=1){if(mode)Gia_ManQuantUpdateCiSupp(g,id);else Gia_ManQuantSetSuppCi(g,Gia_ManObj(g,id));}
  }
  for(i=1;i<n;i+=7)Gia_ManAppendAnd(g,2,2*(i+1));
  printf("support %d %d objs%d words%d count%d vars",n,mode,Gia_ManObjNum(g),g->nSuppWords,g->iSuppPi);
  Vec_IntForEachEntry(&g->vSuppVars,id,i)printf(" %d",id);puts("");
  Gia_ManForEachObj(g,o,i){printf("row %d",i);for(w=0;w<g->nSuppWords;++w)printf(" %llx",(unsigned long long)Gia_ManQuantInfoId(g,i)[w]);puts("");}
  Gia_ManQuantMarkUsedCis(g,keep_wide,&offset);printf("mask");for(w=0;w<g->nSuppWords;++w)printf(" %llx",(unsigned long long)Gia_ManQuantInfoId(g,0)[w]);puts("");
  Gia_ManForEachObj(g,o,i)if(i)printf("overlap %d %d\n",i,Gia_ManQuantCheckOverlap(g,i));
  supports=g->iSuppPi;
  for(j=0;j<supports;++j){g->iSuppPi=j;Gia_ManForEachAnd(g,o,i)printf("used %d %d %d\n",i,j,Gia_ManQuantCountUsed(g,i));}
  g->iSuppPi=supports;Gia_ManStop(g);
 }
}
int main(int argc,char **argv){if(argc>1 && !strcmp(argv[1],"--support"))inspect_support();else inspect_quant();return 0;}

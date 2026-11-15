
#include "aig/gia/gia.h"
#include "base/main/main.h"
#include "base/cmd/cmd.h"
static Gia_Man_t *current;
static int batch,scenario,calls,roundCalls,rounds;
static Gia_Man_t *make_graph(int n) {
 Gia_Man_t *g=Gia_ManStart(n+8);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),i;
 for(i=0;i<n;++i)a=Gia_ManAppendAnd(g,a,b);Gia_ManAppendCo(g,a);return g;
}
static Abc_Frame_t *review_frame(void){return (Abc_Frame_t *)1;}
static Gia_Man_t *review_read(Abc_Frame_t *f){return current;}
static int review_batch(void){return batch;}
static void review_set_batch(int value){batch=value;}
static void review_update(Abc_Frame_t *f,Gia_Man_t *g){Gia_ManStop(current);current=g;roundCalls=0;++rounds;printf("reset round%d input%d\n",rounds,Gia_ManAndNum(current));}
static int review_execute(Abc_Frame_t *f,const char *command){
 int count,i=roundCalls++;int script[]={7,6,5,6,4,5,4,3,2};++calls;
 printf("cmd round%d call%d batch%d input%d %s\n",rounds,i,batch,Gia_ManAndNum(current),command);
 if(scenario==0)count=6;
 else if(scenario==1)count=5;
 else if(scenario==2)count=Abc_MaxInt(1,6-(i+1)/3);
 else count=i<9?script[i]:3;
 Gia_ManStop(current);current=make_graph(count);return 0;
}
#define Abc_FrameGetGlobalFrame review_frame
#define Abc_FrameReadGia review_read
#define Abc_FrameIsBatchMode review_batch
#define Abc_FrameSetBatchMode review_set_batch
#define Abc_FrameUpdateGia review_update
#define Cmd_CommandExecute review_execute
#include "aig/gia/giaDeep.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
int main(int argc,char **argv){
 int outer,si,two,initial,seeds[]={0,3,11},goal=argc>1?4:0;Gia_Man_t *result;
 for(outer=0;outer<2;++outer)for(scenario=0;scenario<4;++scenario)for(si=0;si<3;++si)for(two=0;two<2;++two)for(initial=0;initial<2;++initial){
  current=make_graph(6);batch=initial;calls=roundCalls=rounds=0;
  printf("case %d %d %d %d %d goal%d\n",outer,scenario,seeds[si],two,initial,goal);
  result=outer?Gia_ManDeepSyn(current,3,3,0,goal,seeds[si],two,0):Gia_ManDeepSynOne(3,0,goal,seeds[si],two,0);
  printf("result calls%d best%d current%d batch%d next%u\n",calls,Gia_ManAndNum(result),Gia_ManAndNum(current),batch,Abc_Random(0));
  Gia_ManStop(result);Gia_ManStop(current);current=NULL;
 }
 return 0;
}

#include "aig/gia/gia.h"
#include <stdio.h>
extern int Acec_DetectBoothOne(Gia_Man_t *, Gia_Obj_t *, int *);
extern int Acec_DetectBoothTwo(Gia_Man_t *, Gia_Obj_t *, int *);
int main(void) {
 for(int mode=0;mode<3;mode++) for(int mask=0;mask<32;mask++) {
  Gia_Man_t *p=Gia_ManStart(128);Gia_ManHashStart(p);
  int x[6];for(int i=0;i<6;i++)x[i]=Gia_ManAppendCi(p);
  int a=x[0]^((mask>>0)&1), b=x[1]^((mask>>1)&1), control=x[2]^((mask>>2)&1), d=x[3]^((mask>>3)&1), e=x[4]^((mask>>4)&1);
  int inner=Gia_ManHashXor(p,a,b), root;
  if(mode==2) root=Gia_ManHashXor(p,Gia_ManHashAnd(p,inner,d),control);
  else {
   int both=Gia_ManHashAnd(p,a,b),neither=Gia_ManHashAnd(p,a^1,b^1);
   int mux=Gia_ManHashMux(p,control,both,neither);
   int left=Gia_ManHashAnd(p,mux,d); int right=Gia_ManHashAnd(p,inner,e);
   int product=Gia_ManHashOr(p,left,right);
   root=Gia_ManHashXor(p,product,mode==0?control:x[5]);
  }
  int ins1[5]={-99,-99,-99,-99,-99},ins2[5]={-99,-99,-99,-99,-99};
  int one=Acec_DetectBoothOne(p,Gia_ManObj(p,root>>1),ins1);
  int two=Acec_DetectBoothTwo(p,Gia_ManObj(p,root>>1),ins2);
  printf("%d %d",one,two);
  for(int i=0;i<5;i++)printf(" %d",ins1[i]);for(int i=0;i<5;i++)printf(" %d",ins2[i]);puts("");
  Gia_ManStop(p);
 }
 return 0;
}

#include <stdio.h>
#include <stdlib.h>
#include "aig/gia/gia.h"
#include "proof/acec/acec.h"
extern void Acec_InsertFadd(Gia_Man_t *, int *, int *);
static Gia_Man_t *fixture(int width, int swap, int flip) {
 Gia_Man_t *p=Gia_ManStart(400); int a[5],b[5],cols[10][32]={{0}},sizes[10]={0};
 for(int i=0;i<width;i++)a[i]=Gia_ManAppendCi(p);
 for(int i=0;i<width;i++)b[i]=Gia_ManAppendCi(p);
 for(int i=0;i<width;i++)for(int j=0;j<width;j++){
  int x=swap?b[j]:a[i],y=swap?a[i]:b[j]; cols[i+j][sizes[i+j]++]=Gia_ManAppendAnd(p,x,y);
 }
 for(int k=0;k<2*width-1;k++) {
  while(sizes[k]>=3) {int in[3]={cols[k][0],cols[k][1],cols[k][2]},out[2];
   for(int j=3;j<sizes[k];j++)cols[k][j-3]=cols[k][j]; sizes[k]-=3;
   Acec_InsertFadd(p,in,out);cols[k][sizes[k]++]=out[0];cols[k+1][sizes[k+1]++]=out[1];
  }
  if(sizes[k]==2) {int x=cols[k][0],y=cols[k][1];cols[k][0]=Gia_ManAppendXor(p,x,y);sizes[k]=1;cols[k+1][sizes[k+1]++]=Gia_ManAppendAnd(p,x,y);}
 }
 for(int k=0;k<2*width;k++) {int out=sizes[k]?cols[k][0]:0;if(k==flip)out^=1;Gia_ManAppendCo(p,out);}
 return p;
}
int main(int argc,char **argv) {
 int flip=atoi(argv[1]), width=argc>2?atoi(argv[2]):3;
 if (width < 1 || width > 5) return 2;
 Gia_Man_t *a=fixture(width,0,-1),*b=fixture(width,1,flip);Acec_ParCec_t pars;
 Acec_ManCecSetDefaultParams(&pars); int status=Acec_Solve(a,b,&pars);printf("SOURCE_STATUS=%d\n",status);
 if(a->pCexComb) {printf("SOURCE_WITNESS=");for(int i=0;i<2*width;i++)printf("%d",Abc_InfoHasBit(a->pCexComb->pData,i));puts("");}
 fflush(stdout);return 0;
}

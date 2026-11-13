#include <stdio.h>
#include <stdlib.h>
#include "aig/gia/gia.h"
#include "proof/acec/acec.h"
extern void Acec_InsertFadd(Gia_Man_t *,int *,int *);
static Gia_Man_t *fixture(int swap,int flip){
 Gia_Man_t *p=Gia_ManStart(1000);int a[8],b[8],cols[8][32]={{0}},sizes[8]={0};
 for(int i=0;i<4;i++)a[i]=Gia_ManAppendCi(p);for(int i=0;i<4;i++)b[i]=Gia_ManAppendCi(p);for(int i=4;i<8;i++){a[i]=a[3];b[i]=b[3];}
 for(int i=0;i<8;i++)for(int j=0;i+j<8;j++){int x=swap?b[j]:a[i],y=swap?a[i]:b[j];cols[i+j][sizes[i+j]++]=Gia_ManAppendAnd(p,x,y);}
 for(int k=0;k<7;k++){
  while(sizes[k]>=3){int in[3]={cols[k][0],cols[k][1],cols[k][2]},out[2];for(int j=3;j<sizes[k];j++)cols[k][j-3]=cols[k][j];sizes[k]-=3;Acec_InsertFadd(p,in,out);cols[k][sizes[k]++]=out[0];cols[k+1][sizes[k+1]++]=out[1];}
  if(sizes[k]==2){int x=cols[k][0],y=cols[k][1];cols[k][0]=Gia_ManAppendXor(p,x,y);sizes[k]=1;cols[k+1][sizes[k+1]++]=Gia_ManAppendAnd(p,x,y);}
 }
 /* Last column must also be reduced; carry out is discarded. */
 while(sizes[7]>=3){int in[3]={cols[7][0],cols[7][1],cols[7][2]},out[2];for(int j=3;j<sizes[7];j++)cols[7][j-3]=cols[7][j];sizes[7]-=3;Acec_InsertFadd(p,in,out);cols[7][sizes[7]++]=out[0];}
 if(sizes[7]==2){cols[7][0]=Gia_ManAppendXor(p,cols[7][0],cols[7][1]);sizes[7]=1;}
 for(int k=0;k<8;k++){int out=sizes[k]?cols[k][0]:0;if(k==flip)out^=1;Gia_ManAppendCo(p,out);}return p;
}
int main(int argc,char**argv){int flip=argc>1?atoi(argv[1]):-1;Gia_Man_t*a=fixture(0,-1),*b=fixture(1,flip);Acec_ParCec_t pars;Acec_ManCecSetDefaultParams(&pars);int status=Acec_Solve(a,b,&pars);printf("SOURCE_STATUS=%d\n",status);if(a->pCexComb){printf("SOURCE_WITNESS=");for(int i=0;i<8;i++)printf("%d",Abc_InfoHasBit(a->pCexComb->pData,i));puts("");}return 0;}

#include "aig/gia/gia.h"
#include <stdio.h>
extern void Wlc_BlastAdderFast(Gia_Man_t*,int*,int*,int,int,int);
int main(void){for(int n=1;n<=4;n++)for(int sign=0;sign<2;sign++)for(int carry=0;carry<2;carry++)for(int x=0;x<(1<<n);x++)for(int y=0;y<(1<<n);y++){
 Gia_Man_t*p=Gia_ManStart(1000);int a[6]={0},b[6]={0};Gia_ManHashAlloc(p);
 for(int i=0;i<n;i++){a[i]=(x>>i)&1;b[i]=(y>>i)&1;}
 Wlc_BlastAdderFast(p,a,b,n,sign,carry);int v=0;for(int i=0;i<=n;i++)v|=a[i]<<i;
 printf("%d %d %d %d %d %d\n",n,sign,carry,x,y,v);Gia_ManStop(p);
}}

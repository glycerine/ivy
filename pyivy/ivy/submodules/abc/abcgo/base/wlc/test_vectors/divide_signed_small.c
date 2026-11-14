#include "aig/gia/gia.h"
#include "misc/vec/vec.h"
#include <stdio.h>
extern void Wlc_BlastDividerSigned(Gia_Man_t*,int*,int,int*,int,int,Vec_Int_t*);
int main(void){for(int n=1;n<=4;n++)for(int q=0;q<2;q++)for(int x=0;x<(1<<n);x++)for(int y=0;y<(1<<n);y++){
 Gia_Man_t*p=Gia_ManStart(1000);int a[4],b[4];Vec_Int_t*r=Vec_IntAlloc(n);Gia_ManHashAlloc(p);
 for(int i=0;i<n;i++){a[i]=(x>>i)&1;b[i]=(y>>i)&1;}
 Wlc_BlastDividerSigned(p,a,n,b,n,q,r);int v=0;for(int i=0;i<n;i++)v|=Vec_IntEntry(r,i)<<i;
 printf("%d %d %d %d %d\n",n,q,x,y,v);Vec_IntFree(r);Gia_ManStop(p);
}}

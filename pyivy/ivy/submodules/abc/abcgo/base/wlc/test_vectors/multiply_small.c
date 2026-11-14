#include "aig/gia/gia.h"
#include "misc/vec/vec.h"
#include <stdio.h>
extern void Wlc_BlastMultiplier(Gia_Man_t*,int*,int*,int,int,Vec_Int_t*,Vec_Int_t*,int);
int main(void) {
 for(int na=1;na<=3;na++) for(int nb=1;nb<=3;nb++) for(int sign=0;sign<=1;sign++) for(int x=0;x<(1<<na);x++) for(int y=0;y<(1<<nb);y++) {
  Gia_Man_t *p=Gia_ManStart(100);int a[3],b[3];Vec_Int_t *t=Vec_IntAlloc(0),*r=Vec_IntAlloc(0);Gia_ManHashAlloc(p);
  for(int i=0;i<na;i++)a[i]=(x>>i)&1;for(int i=0;i<nb;i++)b[i]=(y>>i)&1;
  Wlc_BlastMultiplier(p,a,b,na,nb,t,r,sign);int value=0;for(int i=0;i<na+nb;i++)value|=Vec_IntEntry(r,i)<<i;
  printf("%d %d %d %d %d %d\n",na,nb,sign,x,y,value);Vec_IntFree(t);Vec_IntFree(r);Gia_ManStop(p);
 }
}

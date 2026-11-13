#include <stdio.h>
#include "aig/gia/gia.h"
extern Vec_Int_t *Acec_OrderTreeRoots(Gia_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t *);
int main(void){Gia_Man_t *p=Gia_ManStart(8);Vec_Int_t *a=Vec_IntAlloc(6),*r=Vec_IntAlloc(3),*k=Vec_IntStartFull(3);int x[]={0,0,1,0,2,0};for(int i=0;i<6;i++)Vec_IntPush(a,x[i]);for(int i=10;i<13;i++)Vec_IntPush(r,i);Vec_IntWriteEntry(k,1,0);Vec_IntWriteEntry(k,2,1);Vec_Int_t *o=Acec_OrderTreeRoots(p,a,r,k);for(int i=0;i<Vec_IntSize(o);i++)printf("%d ",Vec_IntEntry(o,i));puts("");return 0;}

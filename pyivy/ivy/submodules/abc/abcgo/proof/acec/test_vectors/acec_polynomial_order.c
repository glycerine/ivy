#include <stdio.h>
#include "aig/gia/gia.h"
#include "proof/acec/acec.h"
extern void Acec_InsertFadd(Gia_Man_t *, int *, int *);
int main(void){Gia_Man_t *p=Gia_ManStart(128);int in[5],lo[2],hi[2],next[3];for(int i=0;i<5;i++)in[i]=Gia_ManAppendCi(p);Acec_InsertFadd(p,in,lo);next[0]=lo[1];next[1]=in[3];next[2]=in[4];Acec_InsertFadd(p,next,hi);Gia_ManAppendCo(p,lo[0]);Gia_ManAppendCo(p,hi[0]);Gia_ManAppendCo(p,hi[1]);Vec_Int_t *v=Gia_PolynReorder(p,0,0);printf("order:");for(int i=0;i<Vec_IntSize(v);i++)printf(" %d",Vec_IntEntry(v,i));puts("");return 0;}

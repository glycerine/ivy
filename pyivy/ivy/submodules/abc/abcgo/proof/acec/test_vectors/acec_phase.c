#include <stdio.h>
#include "aig/gia/gia.h"
#include "proof/acec/acec.h"
extern void Acec_InsertFadd(Gia_Man_t *,int *,int *);
extern Vec_Int_t *Gia_PolynCoreOrder(Gia_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t **,Vec_Int_t **);
extern int Pas_ManComputeCuts(Gia_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t *);
int main(void){Gia_Man_t*p=Gia_ManStart(128);int in[5],lo[2],hi[2],next[3];for(int i=0;i<5;i++)in[i]=Gia_ManAppendCi(p);Acec_InsertFadd(p,in,lo);next[0]=lo[1];next[1]=in[3];next[2]=in[4];Acec_InsertFadd(p,next,hi);Gia_ManAppendCo(p,lo[0]);Gia_ManAppendCo(p,hi[0]);Gia_ManAppendCo(p,hi[1]);Vec_Int_t*adds=Ree_ManComputeCuts(p,NULL,0),*ins,*outs;Vec_Int_t*order=Gia_PolynCoreOrder(p,adds,NULL,&ins,&outs);printf("trees=%d\n",Pas_ManComputeCuts(p,adds,order,ins,outs));return 0;}

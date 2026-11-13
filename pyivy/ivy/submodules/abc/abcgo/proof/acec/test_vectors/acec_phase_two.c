#include <stdio.h>
#include "aig/gia/gia.h"
#include "proof/acec/acec.h"
extern void Acec_InsertFadd(Gia_Man_t *,int *,int *);
extern Vec_Int_t *Gia_PolynCoreOrder(Gia_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t **,Vec_Int_t **);
extern int Pas_ManComputeCuts(Gia_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t *);
int main(void){Gia_Man_t*p=Gia_ManStart(128);int in[6],a[2],b[2];for(int i=0;i<6;i++)in[i]=Gia_ManAppendCi(p);Acec_InsertFadd(p,in,a);Acec_InsertFadd(p,in+3,b);for(int i=0;i<2;i++)Gia_ManAppendCo(p,a[i]);for(int i=0;i<2;i++)Gia_ManAppendCo(p,b[i]);Vec_Int_t*adds=Ree_ManComputeCuts(p,NULL,0),*ins,*outs;Vec_Int_t*order=Gia_PolynCoreOrder(p,adds,NULL,&ins,&outs);printf("trees=%d\n",Pas_ManComputeCuts(p,adds,order,ins,outs));return 0;}

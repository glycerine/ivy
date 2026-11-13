#include "aig/gia/gia.h"
extern void Acec_InsertFadd(Gia_Man_t *, int *, int *);
extern void *Acec_ProduceBox(Gia_Man_t *, int);
int main(void){Gia_Man_t *p=Gia_ManStart(64);int in[3],out[2];for(int i=0;i<3;i++)in[i]=Gia_ManAppendCi(p);Acec_InsertFadd(p,in,out);Gia_ManAppendCo(p,out[0]);Gia_ManAppendCo(p,out[1]);Acec_ProduceBox(p,0);return 0;}

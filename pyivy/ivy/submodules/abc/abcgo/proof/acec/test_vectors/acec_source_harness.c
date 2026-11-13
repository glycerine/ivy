#include <stdio.h>
#include "aig/gia/gia.h"
extern Vec_Bit_t * Acec_BoothFindPPG(Gia_Man_t *);
extern Vec_Int_t * Dtc_ManFindCommonCuts(Gia_Man_t *, Vec_Int_t *, Vec_Int_t *);
extern void Acec_InsertFadd(Gia_Man_t *, int *, int *);
extern Gia_Man_t *Acec_Normalize(Gia_Man_t *, int, int);
extern int Dtc_ManCompare(int *, int *);
int main(void) {
 Gia_Man_t *p=Gia_ManStart(200); int in[7];for(int i=0;i<7;i++)in[i]=Gia_ManAppendCi(p);
 Gia_ManHashStart(p);int product=Gia_ManHashAnd(p,in[0],in[1]);int digit=Gia_ManHashXor(p,in[2],in[3]);int pp=Gia_ManHashXor(p,Gia_ManHashAnd(p,digit,in[4]),in[5]);Gia_ManHashStop(p);
 int lo[2],hi[2],j0[3]={product,pp,in[6]},j1[3];Acec_InsertFadd(p,j0,lo);j1[0]=lo[1];j1[1]=in[0];j1[2]=in[5];Acec_InsertFadd(p,j1,hi);
 Gia_ManAppendCo(p,lo[0]);Gia_ManAppendCo(p,hi[0]);Gia_ManAppendCo(p,hi[1]);
 Vec_Bit_t *mask=Acec_BoothFindPPG(p);printf("mask:"); for(int i=0;i<Gia_ManObjNum(p);i++) if(Vec_BitEntry(mask,i))printf(" %d",i);printf("\n");fflush(stdout);
 Gia_Man_t *n=Acec_Normalize(p,1,0);printf("normalized: %d %d\n",Gia_ManCiNum(n),Gia_ManCoNum(n));
 int xs[]={1,2,3,10,1,2,3,11,1,2,3,12}; int ms[]={1,2,3,20,1,2,3,21};
 qsort(xs,3,4*sizeof(int),(int (*)(const void *,const void *))Dtc_ManCompare);qsort(ms,2,4*sizeof(int),(int (*)(const void *,const void *))Dtc_ManCompare);
 Vec_Int_t *x=Vec_IntAlloc(12),*m=Vec_IntAlloc(8);for(int i=0;i<12;i++)Vec_IntPush(x,xs[i]);for(int i=0;i<8;i++)Vec_IntPush(m,ms[i]);Vec_Int_t *v=Dtc_ManFindCommonCuts(p,x,m);printf("ties:");for(int i=0;i<Vec_IntSize(v);i++)printf(" %d",Vec_IntEntry(v,i));puts("");return 0;
}

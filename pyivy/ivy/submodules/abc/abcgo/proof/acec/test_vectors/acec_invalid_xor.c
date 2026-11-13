#include "aig/gia/gia.h"
extern Vec_Int_t *Acec_CollectXorTops(Gia_Man_t *);
int main(void){Gia_Man_t*p=Gia_ManStart(16);Gia_ManHashStart(p);int a=Gia_ManAppendCi(p),b=Gia_ManAppendCi(p);int x=Gia_ManHashXor(p,a,b);Gia_ManAppendCo(p,x);Gia_ManHashStop(p);Acec_CollectXorTops(p);return 0;}

#include "aig/gia/gia.h"
extern Gia_Man_t * Acec_DetectAdditional(Gia_Man_t *,int);
int main(void) {Gia_Man_t *p=Gia_ManStart(32);Gia_ManHashStart(p);int a=Gia_ManAppendCi(p),b=Gia_ManAppendCi(p),c=Gia_ManAppendCi(p);int root=Gia_ManHashXor(p,Gia_ManHashXor(p,a,b),c);Gia_ManAppendCo(p,root);Gia_ManHashStop(p);Acec_DetectAdditional(p,0);return 0;}

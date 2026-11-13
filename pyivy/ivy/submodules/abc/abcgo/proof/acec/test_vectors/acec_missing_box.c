#include "aig/gia/gia.h"
#include "proof/acec/acec.h"
int main(void){Gia_Man_t*p=Gia_ManStart(8);int a=Gia_ManAppendCi(p);Gia_ManAppendCo(p,a^1);Acec_Normalize(p,0,0);return 0;}

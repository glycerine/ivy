#include <stdio.h>
#include "aig/gia/gia.h"
#include "proof/acec/acec.h"
int main(void){Gia_Man_t*p=Gia_ManStart(32);int a=Gia_ManAppendCi(p),b=Gia_ManAppendCi(p);int x=Gia_ManAppendAnd(p,a,b);Gia_ManAppendCo(p,x);Gia_ManAppendCo(p,a^1);Vec_Int_t*order=Gia_PolynReorder(p,0,0);puts("unsigned");Gia_PolynBuild(p,order,0,1,1);puts("signed");Gia_PolynBuild(p,order,1,1,1);return 0;}

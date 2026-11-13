#include <stdio.h>
#include "aig/gia/gia.h"
#include "proof/abs/abs.h"
int main(void) {
 Gia_Man_t *g=Gia_ManStart(8);Gia_ManAppendCi(g);int q=Gia_ManAppendCi(g);Gia_ManAppendCo(g,q);Gia_ManAppendCo(g,1);Gia_ManSetRegNum(g,1);
 Abs_Par_t p;Abs_ParSetDefaults(&p);p.nFramesMax=3;p.nRatioMin=0;p.nRatioMin2=0;p.nFramesNoChangeLim=100;
 int result=Gia_ManPerformGla(g,&p);printf("SOURCE_STATUS=%d\n",result);return 0;
}

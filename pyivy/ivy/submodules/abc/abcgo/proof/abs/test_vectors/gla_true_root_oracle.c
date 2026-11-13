#include <stdio.h>
#include "aig/gia/gia.h"
#include "proof/abs/abs.h"
int main(void) {
 Gia_Man_t *g=Gia_ManStart(12); Gia_ManAppendCi(g); int q=Gia_ManAppendCi(g); Gia_ManAppendCi(g);
 int root=Gia_ManAppendAnd(g,q^1,1);
 Gia_ManAppendCo(g,root);Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,0);Gia_ManSetRegNum(g,2);
 Abs_Par_t p;Abs_ParSetDefaults(&p);p.fUseSimple=1;p.fAddLayer=1;p.nFramesMax=1;p.nRatioMin=0;p.nRatioMin2=0;p.nFramesNoChangeLim=100;
 int result=Gia_ManPerformGla(g,&p);printf("SOURCE_STATUS=%d CEX=%d\n",result,g->pCexSeq!=NULL);return 0;
}

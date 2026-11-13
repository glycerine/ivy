#include "aig/aig/aig.h"
#include "aig/gia/gia.h"
#include "proof/ssw/ssw.h"
#include <stdio.h>
extern int Ssw_RarSimulateGia(Gia_Man_t *, Ssw_RarPars_t *);
int main(void) {
 Gia_Man_t *g=Gia_ManStart(16); int x=Gia_ManAppendCi(g), q=Gia_ManAppendCi(g);
 Gia_ManAppendCo(g,Gia_ManAppendAnd(g,x,q)); Gia_ManAppendCo(g,1); Gia_ManSetRegNum(g,1);
 Ssw_RarPars_t p; Ssw_RarSetDefaultParams(&p); p.nFrames=2; p.nWords=1; p.nBinSize=1; p.nRounds=1; p.fUseFfGrouping=1; p.fSilent=1;
 int r=Ssw_RarSimulateGia(g,&p); Abc_Cex_t *c=g->pCexSeq;
 printf("result=%d original_regs=%d cex_regs=%d pis=%d frame=%d\n",r,Gia_ManRegNum(g),c?c->nRegs:-1,c?c->nPis:-1,c?c->iFrame:-1);
 Gia_ManStop(g); return 0;
}

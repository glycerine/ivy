#include "map/mpm/mpmMap.c"
#include <stdio.h>
int main(void) {
 for(int mode=0;mode<4;mode++) {
  Mpm_Man_t m={0}; Mpm_Par_t pars={0}; Mpm_Uni_t tmp={0}, a={0}, b={0}, c={0};
  m.pPars=&pars;m.nLutSize= mode==3 ? 2:6;pars.fUseDsd=mode==1;pars.fUseTruth=mode==2;
  Vec_PtrPush(&m.vFreeUnits,&tmp);a.pCut.nLeaves=2;a.pCut.pLeaves[0]=4;a.pCut.pLeaves[1]=11;a.pCut.iFunc=37;a.pCut.fCompl=1;
  b.pCut.nLeaves=2;b.pCut.pLeaves[0]=10;b.pCut.pLeaves[1]=6;
  c.pCut.nLeaves=1;c.pCut.pLeaves[0]=8;
  Mpm_Cut_t *cut=Mpm_ManMergeCuts(&m,&a.pCut,&b.pCut,mode==0?NULL:&c.pCut);
  if(!cut){puts("null");free(m.vFreeUnits.pArray);continue;}
  printf("%u %u %u %u %u %u",cut->nLeaves,cut->iFunc,cut->fCompl,m.uPermMask[1],m.uComplMask[1],m.nCutsMerged);
  for(unsigned i=0;i<cut->nLeaves;i++)printf(" %d",cut->pLeaves[i]);puts("");free(m.vFreeUnits.pArray);
 }
 return 0;
}

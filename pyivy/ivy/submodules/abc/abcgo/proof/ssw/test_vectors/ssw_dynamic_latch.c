#include "aig/aig/aig.h"
#include "aig/saig/saig.h"
#include "proof/ssw/ssw.h"
#include <stdio.h>
int main(void) {
 for(int local=0;local<2;local++) {
  Aig_Man_t *p=Aig_ManStart(16), *q; Ssw_Pars_t pars;
  Aig_Obj_t *x=Aig_ObjCreateCi(p), *a=Aig_ObjCreateCi(p), *b=Aig_ObjCreateCi(p);
  Aig_ObjCreateCo(p,Aig_And(p,a,b)); Aig_ObjCreateCo(p,x); Aig_ObjCreateCo(p,x);
  Aig_ManSetRegNum(p,2); Ssw_ManSetDefaultParamsLcorr(&pars);
  pars.fDynamic=1; pars.fLocalSim=local; pars.nSatVarMax2=1; pars.nRecycleCalls2=0;
  q=Ssw_LatchCorrespondence(p,&pars);
  printf("local=%d regs=%d pis=%d pos=%d\n",local,Aig_ManRegNum(q),Saig_ManPiNum(q),Saig_ManPoNum(q));
  Aig_ManStop(q); Aig_ManStop(p);
 }
 return 0;
}

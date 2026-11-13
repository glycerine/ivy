#include "aig/saig/saig.h"
#include "aig/aig/aig.h"
#include "proof/ssw/ssw.h"
#include <stdio.h>
int main(void) {
 Aig_Man_t *p=Aig_ManStart(16), *q;
 Aig_Obj_t *x=Aig_ObjCreateCi(p), *a=Aig_ObjCreateCi(p), *b=Aig_ObjCreateCi(p);
 Aig_ObjCreateCo(p,Aig_Exor(p,a,b));
 Aig_ObjCreateCo(p,x); Aig_ObjCreateCo(p,x);
 Aig_ManSetRegNum(p,2);
 q=Ssw_LatchCorrespondence(p,NULL);
 printf("regs=%d pis=%d pos=%d\n",Aig_ManRegNum(q),Saig_ManPiNum(q),Saig_ManPoNum(q));
 Aig_ManStop(q); Aig_ManStop(p); return 0;
}

#include "aig/aig/aig.h"
#include "proof/fra/fra.h"
#include <stdio.h>
int main(void) {
 Aig_Man_t *p=Aig_ManStart(16); Aig_Obj_t *a=Aig_ObjCreateCi(p), *b=Aig_ObjCreateCi(p);
 Aig_ObjCreateCi(p);
 Aig_ObjCreateCo(p,Aig_And(p,a,b));
 for (int flip=0;flip<2;flip++) {
  int result=Fra_FraigSat(p,0,0,0,0,0,flip,0,0,0);
  int *model=(int*)p->pData;
  printf("flip=%d result=%d model=%d,%d,%d\n",flip,result,model[0],model[1],model[2]);
  ABC_FREE(p->pData);
 }
 Aig_ManStop(p); return 0;
}

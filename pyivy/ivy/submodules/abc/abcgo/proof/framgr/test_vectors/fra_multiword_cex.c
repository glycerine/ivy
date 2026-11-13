#include "proof/fra/fra.h"
#include <stdio.h>
int main(void){
 Aig_Man_t *g=Aig_ManStart(128),*f=Aig_ManStart(128); Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g),*q=Aig_ObjCreateCi(g),*out=0;
 Aig_ObjCreateCo(g,q); Aig_ObjCreateCo(g,Aig_And(g,a,b)); Aig_ManSetRegNum(g,1);
 for(int i=0;i<40;i++) Aig_ObjCreateCi(f); for(int i=0;i<20;i++) out=Aig_ObjCreateCo(f,Aig_ManConst0(f));
 int model[41]={0};model[36]=model[37]=1;model[40]=out->Id;
 Abc_Cex_t *c=Fra_SmlCopyCounterExample(g,f,model);if(!c){puts("null");return 1;}
 printf("frame=%d po=%d bits=%d words=%08x,%08x\n",c->iFrame,c->iPo,c->nBits,c->pData[0],c->pData[1]);return 0;
}

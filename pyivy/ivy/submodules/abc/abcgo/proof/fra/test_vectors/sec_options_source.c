#include "proof/fra/fra.h"
#include "base/main/main.h"
#include <stdio.h>
#include <stdlib.h>
int main(int argc, char **argv) {
    Aig_Man_t *g, *result=NULL, *sentinel=NULL; Aig_Obj_t *x,*a,*b; Fra_Sec_t p; int status;
    Abc_Start(); g=Aig_ManStart(32); x=Aig_ObjCreateCi(g); a=Aig_ObjCreateCi(g); b=Aig_ObjCreateCi(g);
    Aig_ObjCreateCo(g,Aig_And(g,x,a)); if (argc<=9 || !atoi(argv[9])) Aig_ObjCreateCo(g,Aig_And(g,x,b));
    Aig_ObjCreateCo(g,Aig_Exor(g,x,a)); Aig_ObjCreateCo(g,Aig_Exor(g,a,b)); Aig_ManSetRegNum(g,2);
    Fra_SecSetDefaultParams(&p); p.fSilent=1; p.fRetimeFirst=0; p.fRetimeRegs=0; p.fFraiging=0;
    p.fInduction=0; p.fReachability=0; p.fUsePdr=0; p.fInterpolation=atoi(argv[1]); p.fInterSeparate=atoi(argv[2]); p.nBTLimitInter=atoi(argv[3]); p.TimeLimit=atoi(argv[4]);
    if (argc>5) p.fInduction=atoi(argv[5]);
    if (argc>6) p.nBTLimitGlobal=atoi(argv[6]);
    if (argc>7) p.nFramesMax=atoi(argv[7]);
    if (argc>8) result=sentinel=Aig_ManStart(4);
    status=Fra_FraigSec(g,&p,&result);
    printf("sentinel=%d ",sentinel && result==sentinel);
    printf("oracle status=%d model=%d output=%d frame=%d regs=%d outputs=%d\n",status,g->pSeqModel!=NULL,g->pSeqModel?g->pSeqModel->iPo:-1,g->pSeqModel?g->pSeqModel->iFrame:-1,result?Aig_ManRegNum(result):-1,result?Aig_ManCoNum(result):-1);
    if (result) Aig_ManStop(result); if (sentinel && result!=sentinel) Aig_ManStop(sentinel); Aig_ManStop(g); Abc_Stop(); return 0;
}

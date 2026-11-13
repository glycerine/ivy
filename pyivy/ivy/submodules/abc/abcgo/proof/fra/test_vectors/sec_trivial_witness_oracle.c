#include "proof/fra/fra.h"
#include "base/main/main.h"
#include "aig/saig/saig.h"
#include <stdio.h>

int main(void) {
    Aig_Man_t *g;
    Aig_Obj_t *x;
    Fra_Sec_t pars;
    int status;
    Abc_Start();
    g = Aig_ManStart(16);
    x = Aig_ObjCreateCi(g);
    Aig_ObjCreateCi(g);
    Aig_ObjCreateCo(g, x);
    Aig_ObjCreateCo(g, x);
    Aig_ManSetRegNum(g, 1);
    Fra_SecSetDefaultParams(&pars);
    pars.fSilent = 1;
    status = Fra_FraigSec(g, &pars, NULL);
    printf("oracle:%d,%u,%d\n", status, g->pSeqModel->pData[0] & 3,
           Saig_ManVerifyCex(g, g->pSeqModel));
    Aig_ManStop(g);
    Abc_Stop();
    return 0;
}

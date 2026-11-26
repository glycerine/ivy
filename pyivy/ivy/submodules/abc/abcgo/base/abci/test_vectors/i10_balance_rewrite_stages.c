/* Writes ABC's read -> balance x2 -> rewrite stages of an AIGER file.
   Regenerate the goldens from the ABC directory with a release libabc:
     cc -O2 -DLIN64 -DABC_USE_STDINT_H -Isrc -c abcgo/base/abci/test_vectors/i10_balance_rewrite_stages.c -o stages.o
     c++ stages.o libabc.a -lreadline -lpthread -lm -ldl -o stages
     ./stages i10.aig abcgo/base/abci/test_vectors/i10_c_bal2.aig abcgo/base/abci/test_vectors/i10_c_rw.aig */
#include "base/abc/abc.h"
#include "base/io/ioAbc.h"
#include "base/main/main.h"
#include <stdio.h>
extern int Abc_NtkRewrite(Abc_Ntk_t*,int,int,int,int,int);
int main(int argc, char **argv){
    Abc_Start();
    Abc_Ntk_t *n = Io_ReadAiger(argv[1], 1);
    for (int pass=0; pass<2; pass++){ Abc_Ntk_t *b = Abc_NtkBalance(n,0,0,1); Abc_NtkDelete(n); n=b; }
    Io_WriteAiger(n, argv[2], 0, 0, 0);
    printf("bal2 %d %d\n", Abc_NtkNodeNum(n), Abc_AigLevel(n));
    Abc_NtkRewrite(n,0,0,0,0,0);
    Io_WriteAiger(n, argv[3], 0, 0, 0);
    printf("rw %d %d\n", Abc_NtkNodeNum(n), Abc_AigLevel(n));
    return 0;
}

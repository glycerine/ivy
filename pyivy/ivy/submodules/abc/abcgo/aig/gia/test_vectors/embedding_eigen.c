#include "aig/gia/giaEmbed.c"

int main(void)
{
    Emb_Man_t p = {0};
    int i, j;
    p.pMatr = Emb_ManMatrAlloc(3);
    p.pEigen = Emb_ManMatrAlloc(3);
    for (i = 0; i < 3; ++i)
        for (j = 0; j < 3; ++j) p.pMatr[i][j] = 0;
    p.pMatr[0][0] = 1e9f; p.pMatr[1][1] = 1; p.pMatr[2][2] = 0.1f;
    Gia_ManRandom(1);
    for (i = 0; i < 7701; ++i) Gia_ManRandom(0);
    Emb_ManComputeEigenvectors(&p,3,2);
    for (i = 0; i < 2; ++i)
        for (j = 0; j < 3; ++j) printf("%.9g\n",p.pEigen[i][j]);
    ABC_FREE(p.pMatr); ABC_FREE(p.pEigen);
    return 0;
}

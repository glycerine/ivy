#include "aig/gia/giaEmbed.c"

int main(void)
{
    int cluster, offset, i, j;
    for (cluster = 0; cluster < 2; ++cluster)
        for (offset = 0; offset < 8; ++offset)
        {
            Gia_Man_t *g = Gia_ManStart(1);
            int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g);
            int x = Gia_ManAppendAnd(g,a,b), y = Gia_ManAppendAnd(g,x,c);
            Emb_Man_t *p;
            Gia_ManAppendCo(g,y); Gia_ManAppendCo(g,0);
            p = cluster ? Emb_ManStart(g) : Emb_ManStartSimple(g);
            Emb_ManSetValue(p);
            Gia_ManRandom(1);
            for (i = 0; i < offset; ++i) Gia_ManRandom(0);
            Emb_ManComputeDimensions(p,4);
            Emb_ManComputeCovariance(p,4);
            Emb_ManComputeEigenvectors(p,4,1);
            Emb_ManComputeSolutions(p,4,1);
            printf("%d %d %d %d",cluster,offset,p->nObjs,p->nReached);
            for (i = 0; i < 4*p->nObjs; ++i) printf(" %.9g",p->pVecs[i]);
            for (i = 0; i < 4; ++i)
                for (j = 0; j < 4; ++j) printf(" %.9g",p->pMatr[i][j]);
            for (i = 0; i < 4; ++i) printf(" %.9g",p->pEigen[0][i]);
            for (i = 0; i < p->nObjs; ++i) printf(" %.9g",p->pSols[i]);
            puts("");
            Emb_ManStop(p); Gia_ManStop(g);
        }
    return 0;
}

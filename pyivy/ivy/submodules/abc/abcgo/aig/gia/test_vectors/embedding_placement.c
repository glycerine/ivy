#include "aig/gia/giaEmbed.c"

static void record(Emb_Man_t *p)
{
    int i;
    printf(" %.0f",Emb_ManComputeHPWL(p));
    for (i = 0; i < 2*p->nObjs; ++i) printf(" %u",p->pPlacement[i]);
}

int main(void)
{
    int cluster, regs, i, iter;
    for (cluster = 0; cluster < 2; ++cluster)
        for (regs = 0; regs < 2; ++regs)
        {
            Gia_Man_t *g = Gia_ManStart(1);
            int inputs[4], root;
            Emb_Man_t *p;
            for (i = 0; i < 4; ++i) inputs[i] = Gia_ManAppendCi(g);
            root = inputs[0];
            for (i = 0; i < 12; ++i) root = Gia_ManAppendAnd(g,root,inputs[(i+1)%4]^(i&1));
            Gia_ManAppendCo(g,root); Gia_ManAppendCo(g,inputs[0]); Gia_ManAppendCo(g,inputs[3]);
            Gia_ManSetRegNum(g,regs);
            p = cluster ? Emb_ManStart(g) : Emb_ManStartSimple(g);
            Emb_ManSetValue(p);
            p->pSols = ABC_ALLOC(float,2*p->nObjs);
            for (i = 0; i < p->nObjs; ++i)
            {
                p->pSols[i] = (i*7)%17+1;
                p->pSols[p->nObjs+i] = (i*11)%19+1;
            }
            Emb_ManDerivePlacement(p,2);
            printf("%d %d %d",cluster,regs,p->nObjs); record(p);
            for (iter = 0; iter < 3; ++iter)
            {
                Emb_ManPlacementRefine(p,1,0);
                printf(" refine"); record(p);
            }
            puts("");
            Emb_ManStop(p); Gia_ManStop(g);
        }
    return 0;
}

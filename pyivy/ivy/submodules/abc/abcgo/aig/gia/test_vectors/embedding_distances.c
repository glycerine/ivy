#include "aig/gia/giaEmbed.c"

int main(void)
{
    int cluster, regs, first, second, i, h;
    for (cluster = 0; cluster < 2; ++cluster)
        for (regs = 0; regs < 2; ++regs)
        {
            Gia_Man_t *g = Gia_ManStart(1);
            int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g);
            int x = Gia_ManAppendAnd(g,a,b), y = Gia_ManAppendAnd(g,a,x);
            Emb_Man_t *p;
            Emb_Obj_t *o, *far;
            Vec_Int_t *starts = Vec_IntAlloc(2), *ids = Vec_IntAlloc(10);
            Emb_Dat_t *dist;
            Gia_ManAppendCo(g,y); Gia_ManAppendCo(g,a);
            Gia_ManSetRegNum(g,regs);
            p = cluster ? Emb_ManStart(g) : Emb_ManStartSimple(g);
            Emb_ManSetValue(p);
            Emb_ManForEachObj(p,o,h) Vec_IntPush(ids,h);
            dist = ABC_ALLOC(Emb_Dat_t,p->nObjs);
            for (first = 0; first < p->nObjs; ++first)
                for (second = first; second < p->nObjs; ++second)
                {
                    Vec_IntClear(starts);
                    Vec_IntPush(starts,Vec_IntEntry(ids,first));
                    if (second != first) Vec_IntPush(starts,Vec_IntEntry(ids,second));
                    for (i = 0; i < p->nObjs; ++i) dist[i] = -1;
                    far = Emb_ManFindDistances(p,starts,dist);
                    printf("%d %d %d %d %d %d %d",cluster,regs,first,second,p->nReached,p->nDistMax,far->Value);
                    for (i = 0; i < p->nObjs; ++i) printf(" %.0f",dist[i]);
                    puts("");
                }
            ABC_FREE(dist); Vec_IntFree(starts); Vec_IntFree(ids);
            Emb_ManStop(p); Gia_ManStop(g);
        }
    return 0;
}

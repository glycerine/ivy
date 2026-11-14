#include "aig/gia/gia.h"

int main(void)
{
    Gia_Man_t *p = Gia_ManStart(1), *q = Gia_ManStart(1);
    int a = Gia_ManAppendCi(p), b = Gia_ManAppendCi(p);
    int c = Gia_ManAppendCi(p), unused = Gia_ManAppendCi(p);
    int qc = Gia_ManAppendCi(q), qa = Gia_ManAppendCi(q), qb = Gia_ManAppendCi(q);
    int roots[7], copies[6], i, k;
    for (i = 0; i < 7; ++i)
        roots[i] = Gia_ManAppendMux(p, i == 6 ? unused : c,
            a^(i&1), b^((i>>1)&1));
    for (i = 5; i >= 0; --i)
        copies[i] = Gia_ManAppendMux(q, qc, qa^(i&1), qb^((i>>1)&1));
    for (i = 0; i < 6; ++i)
    {
        Gia_ManAppendCo(p, roots[i]);
        Gia_ManAppendCo(q, copies[i]);
    }
    p->vMapping = Vec_IntStart(p->nObjs);
    for (i = 0; i < 7; ++i)
    {
        int id = roots[i]>>1;
        Vec_IntWriteEntry(p->vMapping, id, Vec_IntSize(p->vMapping));
        Vec_IntPush(p->vMapping, 4);
        Vec_IntPush(p->vMapping, a>>1);
        Vec_IntPush(p->vMapping, unused>>1);
        Vec_IntPush(p->vMapping, b>>1);
        Vec_IntPush(p->vMapping, c>>1);
        Vec_IntPush(p->vMapping, (i&1) ? -id : id);
    }
    Gia_ManFillValue(p);
    Gia_ManConst0(p)->Value = 0;
    Gia_ManObj(p, a>>1)->Value = qa;
    Gia_ManObj(p, b>>1)->Value = qb;
    Gia_ManObj(p, c>>1)->Value = qc;
    for (i = 0; i < 6; ++i)
        Gia_ManObj(p, roots[i]>>1)->Value = copies[i]&~1;
    p->vPacking = Vec_IntAlloc(10);
    Vec_IntPush(p->vPacking, 3);
    for (i = 1, k = 0; i <= 3; ++i)
    {
        int j;
        Vec_IntPush(p->vPacking, i);
        for (j = 0; j < i; ++j)
            Vec_IntPush(p->vPacking, roots[k++]>>1);
    }
    /* Replacement must not append to, or retain, an old destination map. */
    q->vMapping = Vec_IntStart(q->nObjs);
    Gia_ManTransferMapping(q, p);
    Gia_ManTransferPacking(q, p);
    printf("mapping");
    for (i = 0; i < Vec_IntSize(q->vMapping); ++i)
        printf(" %d", Vec_IntEntry(q->vMapping, i));
    printf("\npacking");
    for (i = 0; i < Vec_IntSize(q->vPacking); ++i)
        printf(" %d", Vec_IntEntry(q->vPacking, i));
    puts("");
    Gia_ManStop(p);
    Gia_ManStop(q);
    return 0;
}

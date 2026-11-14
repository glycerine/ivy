#include "aig/gia/gia.h"
#include "aig/miniaig/miniaig.h"

extern Mini_Aig_t *Gia_ManToMiniAig(Gia_Man_t *);

int main(void)
{
    int signs, i;
    for (signs = 0; signs < 8; ++signs)
    {
        Gia_Man_t *g = Gia_ManStart(1);
        int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g);
        int x = Gia_ManAppendAnd(g, a^(signs&1), b^((signs>>1)&1));
        int y = Gia_ManAppendAnd(g, c, x^((signs>>2)&1));
        Mini_Aig_t *m;
        Gia_ManAppendCo(g, x);
        Gia_ManAppendCo(g, y^1);
        Gia_ManSetRegNum(g, 1);
        m = Gia_ManToMiniAig(g);
        printf("%d %d %d", signs, Mini_AigNodeNum(m), Mini_AigRegNum(m));
        for (i = 0; i < m->nSize; ++i) printf(" %d", m->pArray[i]);
        printf(" copies");
        for (i = 0; i < g->nObjs; ++i) printf(" %u", Gia_ManObj(g, i)->Value);
        puts("");
        Mini_AigStop(m);
        Gia_ManStop(g);
    }
    return 0;
}

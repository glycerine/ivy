/* Exercise the original list operations directly; Gia_ManFanoutStart's
   pointer-based missing-child checks are unsafe with modern GIA terminals. */
#include "aig/gia/giaFanout.c"

static void snapshot(Gia_Man_t *g, int perm, int step)
{
    int i;
    printf("perm %d step %d", perm, step);
    for (i = 0; i < 5 * g->nObjs; ++i)
        printf(" %d", g->pFanData[i]);
    puts("");
}

int main(void)
{
    int orders[6][3] = {{0,1,2},{0,2,1},{1,0,2},{1,2,0},{2,0,1},{2,1,0}};
    int perm, i;
    for (perm = 0; perm < 6; ++perm)
    {
        Gia_Man_t g = {0};
        Gia_Obj_t objects[16] = {0}, *roots[3];
        g.pObjs = objects;
        g.nObjs = 1;
        g.nObjsAlloc = 16;
        g.vCis = Vec_IntAlloc(3);
        for (i = 0; i < 3; ++i)
            Gia_ManAppendCi(&g);
        /* Same ordinary AND encodings as AppendAnd, with phases ignored by
           the structural fanout lists. No simulation/hash hooks are needed. */
        for (i = 0; i < 3; ++i)
        {
            roots[i] = Gia_ManAppendObj(&g);
            roots[i]->iDiff0 = Gia_ObjId(&g, roots[i]) - 1;
            roots[i]->iDiff1 = Gia_ObjId(&g, roots[i]) - (i == 2 ? 3 : 2);
        }
        g.nFansAlloc = 4096;
        g.pFanData = ABC_CALLOC(int, 5 * g.nFansAlloc);
        for (i = 0; i < 3; ++i)
        {
            Gia_ObjAddFanout(&g, Gia_ObjFanin0(roots[i]), roots[i]);
            Gia_ObjAddFanout(&g, Gia_ObjFanin1(roots[i]), roots[i]);
        }
        snapshot(&g, perm, 0);
        for (i = 0; i < 3; ++i)
        {
            Gia_ObjRemoveFanout(&g, Gia_ManObj(&g,1), roots[orders[perm][i]]);
            snapshot(&g, perm, i+1);
        }
        for (i = 0; i < 3; ++i)
        {
            Gia_ObjAddFanout(&g, Gia_ManObj(&g,1), roots[orders[perm][2-i]]);
            snapshot(&g, perm, i+4);
        }
        ABC_FREE(g.pFanData);
        Vec_IntFree(g.vCis);
    }
    return 0;
}

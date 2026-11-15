#include "aig/gia/giaFx.c"

int main(void)
{
    const char * covers[] = {
        " 0\n", " 1\n", "0 1\n", "1 0\n", "- 1\n",
        "11 1\n", "10 1\n01 1\n", "00 0\n11 0\n",
        "11- 1\n-11 1\n", "11- 0\n-11 0\n",
        "1--1 1\n-11- 1\n0-10 1\n", "---- 0\n"
    };
    int c, extra, mask, i, n, lit;
    for (c = 0; c < (int)(sizeof(covers)/sizeof(covers[0])); ++c)
    for (extra = 0; extra <= 2; ++extra)
    {
        n = Kit_PlaGetVarNum((char *)covers[c]);
        for (mask = 0; mask < (1 << n); ++mask)
        {
            Gia_Man_t *g = Gia_ManStart(64);
            Vec_Int_t *v = Vec_IntAlloc(n + extra);
            for (i = 0; i < n + extra; ++i)
                Vec_IntPush(v, Gia_ManAppendCi(g) ^ ((mask >> i) & 1));
            Gia_ManHashAlloc(g);
            lit = Gia_ManSopToAig(g, (char *)covers[c], v);
            printf("sop%d extra%d mask%d lit%d objs%d\n", c, extra, mask, lit, Gia_ManObjNum(g));
            Vec_IntFree(v);
            Gia_ManStop(g);
        }
    }
    return 0;
}

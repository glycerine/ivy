#include "aig/aig/aig.h"

unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

extern Vec_Int_t *Aig_ManPartitionMonolithic(Aig_Man_t *);
extern Vec_Int_t *Aig_ManPartitionLevelized(Aig_Man_t *, int);
extern Vec_Int_t *Aig_ManPartitionDfs(Aig_Man_t *, int, int);

int main(void)
{
    int cases, trial, step, a, b, mode, size, k, sizes[] = {1,2,3,7};
    assert(scanf("%d", &cases) == 1);
    for (trial = 0; trial < cases; trial++)
    {
        Aig_Man_t *p = Aig_ManStart(100);
        Aig_Obj_t *pool[16];
        pool[0] = Aig_ManConst1(p);
        for (step = 1; step < 4; step++) pool[step] = Aig_ObjCreateCi(p);
        for (step = 0; step < 12; step++)
        {
            assert(scanf("%d %d", &a, &b) == 2);
            pool[4+step] = Aig_And(p, Aig_NotCond(pool[a/2], a&1), Aig_NotCond(pool[b/2], b&1));
        }
        for (step = 0; step < 4; step++)
        {
            assert(scanf("%d", &a) == 1);
            Aig_ObjCreateCo(p, Aig_NotCond(pool[a/2], a&1));
        }
        Aig_ManCleanup(p);
        for (k = 0; k < 4; k++)
        for (mode = 0; mode < 4; mode++)
        {
            Vec_Int_t *labels;
            size = sizes[k];
            if (mode == 0) labels = Aig_ManPartitionMonolithic(p);
            else if (mode == 1) labels = Aig_ManPartitionLevelized(p, size);
            else labels = Aig_ManPartitionDfs(p, size, mode == 3);
            printf("case %d size %d mode %d:", trial, size, mode);
            for (step = 0; step < Vec_IntSize(labels); step++) printf(" %d", Vec_IntEntry(labels, step));
            puts("");
            Vec_IntFree(labels);
        }
        Aig_ManStop(p);
    }
    return 0;
}

#include "aig/gia/gia.h"
#include "aig/miniaig/minilut.h"

extern Mini_Lut_t *Gia_ManToMiniLut(Gia_Man_t *);

int main(void)
{
    int widths[] = {2, 5, 6, 7, 8, 10, 16}, w, mixed, i, k;
    for (w = 0; w < 7; ++w)
        for (mixed = 0; mixed < 2; ++mixed)
        {
            int width = widths[w], inputs[17], head, root;
            Gia_Man_t *g = Gia_ManStart(1);
            Mini_Lut_t *m;
            unsigned long long digest = 14695981039346656037ULL;
            for (i = 0; i <= width; ++i) inputs[i] = Gia_ManAppendCi(g);
            head = inputs[0];
            /* Original C receives its ordinary AND expansion of XOR. */
            for (i = 1; i < width; ++i)
                head = i%2 == 0 ? Gia_ManAppendXor(g, head, inputs[i]) : Gia_ManAppendAnd(g, head, inputs[i]^1);
            root = Gia_ManAppendAnd(g, head, inputs[width]);
            Gia_ManAppendCo(g, head^1);
            Gia_ManAppendCo(g, root^1);
            Gia_ManAppendCo(g, root^1);
            if (mixed) Gia_ManAppendCo(g, root);
            Gia_ManAppendCo(g, inputs[width]^1);
            Gia_ManAppendCo(g, inputs[width]^1);
            Gia_ManAppendCo(g, 0);
            Gia_ManAppendCo(g, 1);
            Gia_ManSetRegNum(g, 1);
            g->vMapping = Vec_IntStart(g->nObjs);
            Vec_IntWriteEntry(g->vMapping, head>>1, Vec_IntSize(g->vMapping));
            Vec_IntPush(g->vMapping, width);
            for (i = 0; i < width; ++i) Vec_IntPush(g->vMapping, inputs[i]>>1);
            Vec_IntPush(g->vMapping, head>>1);
            Vec_IntWriteEntry(g->vMapping, root>>1, Vec_IntSize(g->vMapping));
            Vec_IntPush(g->vMapping, 2);
            Vec_IntPush(g->vMapping, head>>1);
            Vec_IntPush(g->vMapping, inputs[width]>>1);
            Vec_IntPush(g->vMapping, root>>1);
            m = Gia_ManToMiniLut(g);
            for (i = 0; i < Mini_LutNodeNum(m); ++i)
            {
                for (k = 0; k < width; ++k)
                    digest = (digest ^ (unsigned)Mini_LutNodeFanin(m,i,k))*1099511628211ULL;
                for (k = 0; k < Mini_LutWordNum(width); ++k)
                    digest = (digest ^ Mini_LutNodeTruth(m,i)[k])*1099511628211ULL;
            }
            printf("%d %s %d %d %d %016llx\n", width, mixed ? "true" : "false",
                Mini_LutNodeNum(m), Mini_LutSize(m), Mini_LutRegNum(m), digest);
            Mini_LutStop(m);
            Gia_ManStop(g);
        }
    return 0;
}

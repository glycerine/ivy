#include "aig/aig/aig.h"

unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    Aig_Man_t *g;
    Aig_Obj_t *a, *b, *c, *gate, *buffer, *out;
    int mode, removed, pin, i;
    for (mode = 0; mode < 2; mode++)
    {
        g = Aig_ManStart(32);
        a = Aig_ObjCreateCi(g); b = Aig_ObjCreateCi(g);
        gate = Aig_And(g, a, b);
        buffer = Aig_ManFetchMemory(g);
        buffer->Type = AIG_OBJ_BUF; g->nObjs[AIG_OBJ_BUF]++;
        Aig_ObjConnect(g, buffer, gate, NULL);
        removed = mode ? Aig_ManSeqCleanupBasic(g) : Aig_ManSeqCleanup(g);
        printf("%s_buffer %d %d %d\n", mode ? "basic" : "full", removed,
            Aig_ManBufNum(g), Aig_ManNodeNum(g));
        Aig_ManStop(g);
    }
    g = Aig_ManStart(32);
    Aig_ManFanoutStart(g);
    a = Aig_ObjCreateCi(g); b = Aig_ObjCreateCi(g); c = Aig_ObjCreateCi(g);
    gate = Aig_And(g, a, b);
    Aig_ObjCreateCo(g, c); Aig_ObjCreateCo(g, Aig_ManConst1(g));
    Aig_ObjCreateCo(g, gate); Aig_ObjCreateCo(g, Aig_ManConst1(g));
    Aig_ManSetRegNum(g, 2);
    Aig_ManSeqCleanupBasic(g);
    printf("basic_fanouts");
    pin = Aig_ObjFanout0Int(g, 0);
    for (i = 0; i < 3; i++) { printf(" %d", pin); pin = Aig_ObjFanoutNext(g, pin); }
    puts("");
    Aig_ManStop(g);

    g = Aig_ManStart(32);
    a = Aig_ObjCreateCi(g); b = Aig_ObjCreateCi(g);
    Aig_ObjCreateCo(g, a);
    gate = Aig_And(g, a, b);
    Aig_ObjCreateCo(g, gate);
    Aig_ManSetRegNum(g, 1);
    removed = Aig_ManSeqCleanupBasic(g);
    printf("basic_patch %d %d %d\n", removed, !Aig_ObjIsNone(gate), Aig_ObjCioId(b));
    Aig_ManStop(g);

    g = Aig_ManStart(32);
    a = Aig_ObjCreateCi(g); b = Aig_ObjCreateCi(g); c = Aig_ObjCreateCi(g);
    Aig_ObjCreateCo(g, c); Aig_ObjCreateCo(g, b); out = Aig_ObjCreateCo(g, a);
    Aig_ManSetRegNum(g, 2);
    Aig_ManSeqCleanup(g);
    printf("full_ids %d %d %d\n", Aig_ManRegNum(g), Aig_ObjCioId(c), Aig_ObjCioId(out));
    Aig_ManStop(g);

    g = Aig_ManStart(32);
    a = Aig_ObjCreateCi(g); b = Aig_ObjCreateCi(g);
    buffer = Aig_ManFetchMemory(g);
    buffer->Type = AIG_OBJ_BUF; g->nObjs[AIG_OBJ_BUF]++;
    Aig_ObjConnect(g, buffer, b, NULL);
    out = Aig_ManFetchMemory(g);
    out->Type = AIG_OBJ_BUF; g->nObjs[AIG_OBJ_BUF]++;
    Aig_ObjConnect(g, out, a, NULL);
    Aig_ObjCreateCo(g, a); Aig_ObjCreateCo(g, a);
    Aig_ManSetRegNum(g, 1);
    Aig_ManSeqCleanup(g);
    printf("removed_ci_buffers %d %d\n", Aig_ManBufNum(g),
        Aig_ManObj(g, Aig_ObjFaninId0(buffer)) == NULL);
    Aig_ManStop(g);
    return 0;
}

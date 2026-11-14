#include "aig/aig/aigRet.c"

/* Standalone console environment; all graph/retiming algorithms are source. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

static Aig_Obj_t *read_lit(Aig_Obj_t **objects)
{
    int lit;
    assert(scanf("%d", &lit) == 1);
    return Aig_NotCond(objects[lit/2], lit&1);
}

static int evaluate(Aig_Obj_t *obj, unsigned char *values, unsigned char *known)
{
    int a, b;
    assert(obj->Id >= 0 && obj->Id < 4096);
    if (known[obj->Id]) return values[obj->Id];
    assert(!Aig_ObjIsCi(obj) && !Aig_ObjIsConst1(obj));
    a = evaluate(Aig_ObjFanin0(obj), values, known) ^ Aig_ObjFaninC0(obj);
    if (Aig_ObjIsNode(obj))
    {
        b = evaluate(Aig_ObjFanin1(obj), values, known) ^ Aig_ObjFaninC1(obj);
        a &= b;
    }
    else assert(Aig_ObjIsBuf(obj) || Aig_ObjIsCo(obj));
    known[obj->Id] = 1;
    return values[obj->Id] = a;
}

int main(void)
{
    Aig_Man_t *g, *out;
    Aig_Obj_t *objects[1024], *a, *b, *obj;
    unsigned char values[4096], known[4096], state[1024], next[1024];
    int npis, nregs, nnodes, npos, forward, steps, i, frame;
    while (scanf("%d%d%d%d%d%d", &npis, &nregs, &nnodes, &npos, &forward, &steps) == 6)
    {
        g = Aig_ManStart(100);
        objects[0] = Aig_ManConst1(g);
        for (i = 1; i <= npis+nregs; i++) objects[i] = Aig_ObjCreateCi(g);
        for (i = 1+npis+nregs; i <= npis+nregs+nnodes; i++)
        {
            a = read_lit(objects); b = read_lit(objects);
            objects[i] = Aig_And(g, a, b);
        }
        for (i = 0; i < npos+nregs; i++) Aig_ObjCreateCo(g, read_lit(objects));
        Aig_ManSetRegNum(g, nregs);
#ifdef RETIMING_FRONTIER
        /* Frontier uses the terminal pNext/CioId union for latch links. */
        Aig_ManCleanNext(g);
        out = Aig_ManRetimeFrontier(g, steps);
#else
        out = Rtm_ManRetime(g, forward, steps, 0);
#endif
        printf("registers %d\n", Aig_ManRegNum(out));
        memset(state, 0, sizeof(state));
        for (frame = 0; frame < 64; frame++)
        {
            memset(values, 0, sizeof(values));
            memset(known, 0, sizeof(known));
            values[0] = known[0] = 1;
            Aig_ManForEachCi(out, obj, i)
            {
                values[obj->Id] = i < npis ? ((frame*13+(frame>>2)*7) >> i)&1 : state[i-npis];
                known[obj->Id] = 1;
            }
            Aig_ManForEachCo(out, obj, i)
            {
                int value = evaluate(obj, values, known);
                if (i < npos) printf("%d", value); else next[i-npos] = value;
            }
            puts("");
            memcpy(state, next, Aig_ManRegNum(out));
        }
        if (out != g) Aig_ManStop(out);
        Aig_ManStop(g);
    }
    return 0;
}

// Compile the actual traversal routines, including their private helpers.
#include "aig/gia/gia.h"
static abctime review_clock(void) { return 0; }
#define Abc_Clock review_clock
#include "aig/gia/giaDfs.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"

static unsigned state;
static unsigned next(void)
{
    state = 1664525u * state + 1013904223u;
    return state;
}
static void emit(const char *name, Vec_Int_t *v)
{
    int i;
    printf("%s", name);
    for (i = 0; i < Vec_IntSize(v); ++i)
        printf(" %d", Vec_IntEntry(v, i));
    puts("");
}
static void marks(const char *name, Gia_Man_t *g)
{
    int i;
    printf("%s", name);
    for (i = 0; i < Gia_ManObjNum(g); ++i)
        if (Gia_ObjIsTravIdCurrentId(g, i))
            printf(" %d", i);
    puts("");
}
int main(int argc, char **argv)
{
    int seed, mask, i, j, roots[6], n, pos[4];
    if (argc > 1) {
        Gia_Man_t *g = Gia_ManStart(8);
        int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g);
        Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b));
        Gia_ManCollectTest(g);
        Gia_ManSuppSizeTest(g);
        Gia_ManCollectSeqTest(g);
        Gia_ManStop(g);
        return 0;
    }
    for (seed = 1; seed <= 32; ++seed) {
        Gia_Man_t *g = Gia_ManStart(64);
        Vec_Int_t *v = Vec_IntAlloc(32), *w;
        state = seed;
        for (i = 0; i < 5; ++i)
            Gia_ManAppendCi(g);
        for (i = 0; i < 12; ++i) {
            int a = 1 + next() % (Gia_ManObjNum(g) - 1);
            int b = 1 + next() % (Gia_ManObjNum(g) - 1);
            int pa, pb;
            if (b == a)
                b = b % (Gia_ManObjNum(g) - 1) + 1;
            pa = next() & 1;
            pb = next() & 1;
            Gia_ManAppendAnd(g, 2*a + pa, 2*b + pb);
        }
        for (i = 0; i < 6; ++i) {
            int a = 1 + next() % 17;
            Gia_ManAppendCo(g, 2*a + (next() & 1));
        }
        Gia_ManSetRegNum(g, 2);
        for (mask = 0; mask < 64; ++mask) {
            n = 0;
            for (i = 0; i < 6; ++i)
                if (mask & (1 << i))
                    roots[n++] = Gia_ObjId(g, Gia_ManCo(g, i));
            printf("case %d %d\n", seed, mask);
            Gia_ManCollectCis(g, roots, n, v);
            emit("cis", v);
            printf("supp %d\n", Gia_ManSuppSize(g, roots, n));
            marks("suppmarks", g);
            printf("cone %d\n", Gia_ManConeSize(g, roots, n));
            marks("conemarks", g);
            w = Gia_ManCollectNodesCis(g, roots, n);
            emit("nodescis", w);
            Vec_IntFree(w);
            Gia_ManIncrementTravId(g);
            Gia_ManCollectAnds(g, roots, n, v, NULL);
            emit("ands", v);
            Gia_ManCollectAnds(g, roots, n, v, NULL);
            emit("shared", v);
            w = Vec_IntAlloc(2);
            Vec_IntPush(w, 6);
            Vec_IntPush(w, 7);
            Gia_ManIncrementTravId(g);
            Gia_ManCollectAnds(g, roots, n, v, w);
            emit("leaves", v);
            Vec_IntFree(w);
        }
        for (mask = 0; mask < 16; ++mask) {
            n = 0;
            for (i = 0; i < 4; ++i)
                if (mask & (1 << i))
                    pos[n++] = i;
            printf("seq %d %d\n", seed, mask);
            w = Gia_ManCollectSeq(g, pos, n);
            emit("objects", w);
            Vec_IntFree(w);
        }
        Gia_ManStaticFanoutStart(g);
        for (i = 6; i < 18; ++i)
            for (j = 6; j < 18; ++j) {
                w = Vec_IntAlloc(2);
                Vec_IntPush(w, i);
                Vec_IntPush(w, j);
                printf("direction %d %d %d\n", seed, i, j);
                Gia_ManCollectTfi(g, w, v);
                emit("tfi", v);
                Gia_ManCollectTfo(g, w, v);
                emit("tfo", v);
                Vec_IntFree(w);
            }
        w = Gia_ManCollectAndsAll(g);
        emit("all", w);
        Vec_IntFree(w);
        for (i = 1; i < 18; ++i)
            printf("one %d %d\n", i, Gia_ManSuppSizeOne(g, Gia_ManObj(g, i)));
        {
            Vec_Vec_t *levels = Gia_ManLevelize(g);
            Vec_Ptr_t *level;
            Gia_Obj_t *obj;
            int k;
            Vec_VecForEachLevel(levels, level, i) {
                printf("level %d", i);
                Vec_PtrForEachEntry(Gia_Obj_t *, level, obj, k)
                    printf(" %d", Gia_ObjId(g, obj));
                puts("");
            }
            Vec_VecFree(levels);
        }
        w = Gia_ManOrderReverse(g);
        emit("reverse", w);
        Vec_IntFree(w);
        {
            Vec_Wec_t *levels = Gia_ManLevelizeR(g);
            Vec_Int_t *level;
            Vec_WecForEachLevel(levels, level, i) {
                printf("rlevel %d", i);
                for (j = 0; j < Vec_IntSize(level); ++j)
                    printf(" %d", Vec_IntEntry(level, j));
                puts("");
            }
            Vec_WecFree(levels);
        }
        Vec_IntFree(v);
        Gia_ManStop(g);
    }
    return 0;
}

#include "aig/gia/giaEdge.c"
#include "misc/tim/timInt.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static void emit(const char *name, Vec_Int_t *v)
{
    int i;
    printf("%s", name);
    for (i = 0; i < Vec_IntSize(v); ++i) printf(" %d", Vec_IntEntry(v, i));
    puts("");
}
static void snapshot(Gia_Man_t *g)
{
    emit("first", g->vEdge1);
    emit("second", g->vEdge2);
    emit("delay", g->vEdgeDelay);
}
static void record(Gia_Man_t *g, int id, int a, int b)
{
    Vec_IntWriteEntry(g->vMapping, id, Vec_IntSize(g->vMapping));
    Vec_IntPush(g->vMapping, 2);
    Vec_IntPushTwo(g->vMapping, a, b);
    Vec_IntPush(g->vMapping, id);
}
static void windows(void)
{
    int mode, two, replace;
    for (mode = 0; mode < 2; ++mode)
        for (two = 0; two < 2; ++two)
            for (replace = 0; replace < 2; ++replace) {
                Gia_Man_t *g = Gia_ManStart(8);
                int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g);
                int x = Gia_ManAppendAnd(g, a, b), y = Gia_ManAppendAnd(g, x, a);
                Vec_Int_t *leaves = Vec_IntAlloc(1), *nodes = Vec_IntAlloc(1), *temp = Vec_IntAlloc(2);
                Vec_Wec_t *window = Vec_WecStart(1);
                Gia_ManAppendCo(g, y);
                g->vMapping = Vec_IntStart(Gia_ManObjNum(g));
                record(g, x>>1, a>>1, b>>1);
                record(g, y>>1, x>>1, a>>1);
                Edg_ManToMapping(g);
                if (mode) Vec_IntFreeP(&g->vMapping);
                Gia_ManStaticFanoutStart(g);
                Vec_IntPush(leaves, x>>1);
                Vec_IntPush(nodes, y>>1);
                Vec_IntPush(temp, -7);
                Vec_WecPush(window, 0, replace ? a>>1 : x>>1);
                Vec_WecPush(window, 0, replace ? b>>1 : a>>1);
                Gia_ManComputeEdgeDelay(g, two);
                printf("window %d %d %d full %d\n", mode, two, replace, Gia_ManEvalWindow(g, leaves, nodes, window, temp, two));
                snapshot(g);
                emit("mapping", Gia_ObjLutFanins2(g, y>>1));
                emit("window", Vec_WecEntry(window, 0));
                emit("temp", temp);
                if (!mode) {
                    Gia_ManComputeEdgeDelay(g, two);
                    printf("inc %d\n", Gia_ManEvalWindowInc(g, leaves, nodes, window, temp, two));
                    snapshot(g);
                    emit("mapping", Gia_ObjLutFanins2(g, y>>1));
                    emit("window", Vec_WecEntry(window, 0));
                    emit("temp", temp);
                }
                Vec_IntFree(leaves); Vec_IntFree(nodes); Vec_IntFree(temp); Vec_WecFree(window);
                Gia_ManStop(g);
            }
}
static void boxes(void)
{
    int arrivals[] = {0, 6, 20}, delays[] = {0, 3, 12};
    int mode, two, ai, di;
    for (mode = 0; mode < 2; ++mode)
        for (two = 0; two < 2; ++two)
            for (ai = 0; ai < 3; ++ai)
                for (di = 0; di < 3; ++di) {
                    Gia_Man_t *g = Gia_ManStart(8);
                    int a = Gia_ManAppendCi(g), box = Gia_ManAppendCi(g), root = Gia_ManAppendAnd(g, a, box);
                    Tim_Man_t *timing = Tim_ManStart(2, 2);
                    Vec_Ptr_t *tables = Vec_PtrAlloc(1);
                    float *table = ABC_ALLOC(float, 4);
                    table[0] = 0; table[1] = table[2] = 1; table[3] = delays[di];
                    Vec_PtrPush(tables, table);
                    Tim_ManSetDelayTables(timing, tables);
                    Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0);
                    Tim_ManInitPiArrival(timing, 0, arrivals[ai]);
                    g->pManTime = timing;
                    Gia_ManAppendCo(g, a);
                    Gia_ManAppendCo(g, root);
                    g->vMapping = Vec_IntStart(Gia_ManObjNum(g));
                    record(g, root>>1, a>>1, box>>1);
                    if (mode) { Edg_ManToMapping(g); Vec_IntFreeP(&g->vMapping); }
                    printf("box %d %d %d %d compute %d\n", mode, two, ai, di, Gia_ManComputeEdgeDelay(g, two));
                    snapshot(g);
                    printf("eval %d\n", Gia_ManEvalEdgeDelay(g));
                    snapshot(g);
                    printf("times %.0f %.0f %.0f %.0f\n", Tim_ManGetCiArrival(timing, 0), Tim_ManGetCiArrival(timing, 1), Tim_ManCo(timing, 0)->timeArr, Tim_ManCo(timing, 1)->timeArr);
                    Gia_ManStop(g);
                }
}
int main(void) { windows(); boxes(); return 0; }

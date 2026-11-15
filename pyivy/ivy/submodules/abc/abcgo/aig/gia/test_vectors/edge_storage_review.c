#include "aig/gia/giaEdge.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"

static void emit(const char *name, Vec_Int_t *v)
{
    int i;
    printf("%s", name);
    for (i = 0; i < Vec_IntSize(v); ++i)
        printf(" %d", Vec_IntEntry(v, i));
    puts("");
}
static Gia_Man_t *graph(void)
{
    int i;
    Gia_Man_t *g = Gia_ManStart(8);
    for (i = 0; i < 5; ++i)
        Gia_ManAppendCi(g);
    return g;
}
static void snapshot(Gia_Man_t *g)
{
    Vec_Int_t *array = Gia_ManEdgeToArray(g);
    int i, j;
    emit("first", g->vEdge1);
    emit("second", g->vEdge2);
    emit("pairs", array);
    printf("count %d checks", Gia_ManEvalEdgeCount(g));
    for (i = 1; i <= 5; ++i)
        for (j = 1; j <= 5; ++j)
            printf(" %d", Gia_ObjCheckEdge(g, i, j));
    puts("");
    Vec_IntFree(array);
}
int main(void)
{
    int values[] = {0, 7, 9}, a, b, op, mask, i, j, bit, code;
    for (a = 0; a < 3; ++a)
        for (b = 0; b < 3; ++b)
            for (op = 0; op < 4; ++op) {
                Vec_Int_t *first, *second;
                int result = 0;
                if (op == 1 && values[a] != 7 && values[b] != 7) continue;
                if (op == 2 && values[a] != 9 && values[b] != 9) continue;
                first = Vec_IntStart(16);
                second = Vec_IntStart(16);
                Vec_IntWriteEntry(first, 5, values[a]);
                Vec_IntWriteEntry(second, 5, values[b]);
                printf("primitive %d %d %d before %d", a, b, op, Gia_ObjEdgeCount(5, first, second));
                if (op == 0) result = Gia_ObjEdgeAdd(5, 11, first, second);
                if (op == 1) Gia_ObjEdgeRemove(5, 7, first, second);
                if (op == 2) Gia_ObjEdgeRemove(5, 9, first, second);
                if (op == 3) Gia_ObjEdgeClean(5, first, second);
                printf(" result %d after %d slots %d %d\n", result, Gia_ObjEdgeCount(5, first, second), Vec_IntEntry(first, 5), Vec_IntEntry(second, 5));
                Vec_IntFree(first);
                Vec_IntFree(second);
            }
    for (mask = 0; mask < 1024; ++mask) {
        Gia_Man_t *g = graph();
        Vec_Int_t *array = Vec_IntAlloc(20);
        bit = 0;
        for (i = 1; i <= 5; ++i)
            for (j = i+1; j <= 5; ++j, ++bit)
                if (mask & (1 << bit)) Vec_IntPushTwo(array, i, j);
        printf("graph %d\n", mask);
        Gia_ManEdgeFromArray(g, array);
        snapshot(g);
        Vec_IntFree(array);
        Gia_ManStop(g);
    }
    for (code = 0; code < 81; ++code) {
        Gia_Man_t *g = graph();
        int digits = code;
        g->vPacking = Vec_IntAlloc(20);
        Vec_IntPush(g->vPacking, 4);
        for (i = 0; i < 4; ++i) {
            int size = 1 + digits % 3;
            digits /= 3;
            Vec_IntPush(g->vPacking, size);
            for (j = 0; j < size; ++j) Vec_IntPush(g->vPacking, 1 + (i+j)%5);
        }
        printf("packing %d\n", code);
        Gia_ManConvertPackingToEdges(g);
        snapshot(g);
        Gia_ManStop(g);
    }
    return 0;
}

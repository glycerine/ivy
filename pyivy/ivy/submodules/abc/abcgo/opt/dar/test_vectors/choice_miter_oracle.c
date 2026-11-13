#include "aig/gia/gia.h"
#include "aig/aig/aig.h"
#include "misc/vec/vec.h"

extern Gia_Man_t * Gia_ManChoiceMiter(Vec_Ptr_t *);
extern Aig_Man_t * Gia_ManToAigSkip(Gia_Man_t *, int);

int main(void)
{
    Gia_Man_t * g[3], * m;
    Aig_Man_t * a;
    Aig_Obj_t * obj;
    Vec_Ptr_t * snapshots = Vec_PtrAlloc(3);
    int i;
    for (i = 0; i < 3; i++) {
        int x, y, z, first, second;
        g[i] = Gia_ManStart(32);
        x = Gia_ManAppendCi(g[i]);
        y = Gia_ManAppendCi(g[i]);
        z = Gia_ManAppendCi(g[i]);
        first = Gia_ManAppendAnd(g[i], x ^ (i == 1), i == 0 ? y : z);
        second = Gia_ManAppendAnd(g[i], first, i == 2 ? y : z);
        Gia_ManAppendCo(g[i], first);
        Gia_ManAppendCo(g[i], second);
        Vec_PtrPush(snapshots, g[i]);
    }
    m = Gia_ManChoiceMiter(snapshots);
    a = Gia_ManToAigSkip(m, 3);
    printf("%d\n", Aig_ManObjNumMax(a));
    Aig_ManForEachObj(a, obj, i) {
        int kind = Aig_ObjIsConst1(obj) ? 0 : Aig_ObjIsCi(obj) ? 1 : Aig_ObjIsNode(obj) ? 2 : Aig_ObjIsCo(obj) ? 5 : -1;
        int f0 = Aig_ObjFaninId0(obj), f1 = Aig_ObjFaninId1(obj);
        printf("%d %d %d\n", kind, f0 < 0 ? -1 : 2*f0 + Aig_ObjFaninC0(obj), f1 < 0 ? -1 : 2*f1 + Aig_ObjFaninC1(obj));
    }
    return 0;
}

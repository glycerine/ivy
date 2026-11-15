#define ABC_USE_CUDD
#include "aig/gia/giaClp.c"
#include "base/main/main.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"

int main(void) {
    int widths[] = {4, 17, 33, 65}, tables[] = {6, 0x81, 0xe8, 0xca, 0x96, 0x7e};
    int wi, ti, n, table, order, phase, i, j, mask, positions[3], permutation[65];
    int root, packed, values[512];
    DdManager *dd;
    DdNode *f, *cube, *next;
    Gia_Man_t *g;
    Gia_Obj_t *obj;
    Vec_Int_t *leaves;
    Vec_Str_t *sop, *scratch;
    char *text;
    Abc_Start();
    for (wi = 0; wi < 4; ++wi) for (ti = 0; ti < 6; ++ti) for (order = 0; order < 2; ++order) {
        n = widths[wi]; table = tables[ti];
        positions[0] = 0; positions[1] = n/2; positions[2] = n-1;
        dd = Cudd_Init(n, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0);
        for (i = 0; i < n; ++i) permutation[i] = order ? n-1-i : i;
        assert(Cudd_ShuffleHeap(dd, permutation));
        Cudd_zddVarsFromBddVars(dd, 2);
        f = Cudd_Not(dd->one); Cudd_Ref(f);
        for (mask = 0; mask < 8; ++mask) if ((table >> mask) & 1) {
            cube = dd->one; Cudd_Ref(cube);
            for (i = 0; i < 3; ++i) {
                next = Cudd_bddAnd(dd, cube, Cudd_NotCond(Cudd_bddIthVar(dd, positions[i]), !((mask >> i) & 1)));
                Cudd_Ref(next); Cudd_RecursiveDeref(dd, cube); cube = next;
            }
            next = Cudd_bddOr(dd, f, cube); Cudd_Ref(next);
            Cudd_RecursiveDeref(dd, f); Cudd_RecursiveDeref(dd, cube); f = next;
        }
        for (phase = 0; phase < 8; ++phase) {
            g = Gia_ManStart(2*n+64); Gia_ManHashStart(g);
            leaves = Vec_IntAlloc(n); sop = Vec_StrAlloc(0); scratch = Vec_StrAlloc(0);
            for (i = 0; i < n; ++i) {
                root = Gia_ManAppendCi(g);
                for (j = 0; j < 3; ++j) if (i == positions[j]) root ^= (phase >> j) & 1;
                Vec_IntPush(leaves, root);
            }
            root = Gia_ManRebuildIsop(dd, f, g, leaves, sop, scratch);
            printf("case %d %x %d %d %d sop", n, table, order, phase, root);
            text = Vec_StrArray(sop);
            for (i = 0; text[i]; ++i) printf("%02x", (unsigned char)text[i]);
            printf(" nodes");
            Gia_ManForEachAnd(g, obj, i) printf(" %d:%d", Gia_ObjFaninLit0p(g, obj), Gia_ObjFaninLit1p(g, obj));
            packed = 0;
            for (mask = 0; mask < 8; ++mask) {
                memset(values, 0, sizeof(values));
                for (i = 0; i < 3; ++i) values[Gia_ObjId(g, Gia_ManCi(g, positions[i]))] = (mask >> i) & 1;
                Gia_ManForEachAnd(g, obj, i) values[i] = (values[Gia_ObjFaninId0(obj, i)] ^ Gia_ObjFaninC0(obj)) & (values[Gia_ObjFaninId1(obj, i)] ^ Gia_ObjFaninC1(obj));
                packed |= (values[root >> 1] ^ (root & 1)) << mask;
            }
            printf(" truth %x\n", packed);
            Gia_ManStop(g); Vec_IntFree(leaves); Vec_StrFree(sop); Vec_StrFree(scratch);
        }
        Cudd_RecursiveDeref(dd, f); Cudd_Quit(dd);
    }
    Abc_Stop();
    return 0;
}

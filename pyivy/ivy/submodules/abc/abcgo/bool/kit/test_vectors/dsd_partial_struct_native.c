// Structured functions and valid (whole-component) bound sets for
// Kit_DsdTruthPartialTwo and Kit_DsdTruthPartial; produces
// dsd_partial_struct_native.txt (first 5000 lines kept).
#include "misc/util/abc_global.h"
#include "bool/kit/kit.h"
#include <stdio.h>
static unsigned long long s = 12345;
static unsigned rnd(void) { s = s * 6364136223846793005ULL + 1442695040888963407ULL; return (unsigned)(s >> 33); }
#define W 8
static void gen(unsigned *t, int nVars, int depth) {
    int nW = Kit_TruthWordNum(nVars), k, op = rnd() % 6;
    unsigned a[W], b[W], c[W];
    if (depth == 0 || op == 0) { Kit_TruthIthVar(t, nVars, rnd() % nVars); if (rnd() & 1) Kit_TruthNot(t, t, nVars); return; }
    gen(a, nVars, depth - 1); gen(b, nVars, depth - 1);
    for (k = 0; k < nW; k++) {
        if (op == 1 || op == 2) t[k] = a[k] & b[k];
        else if (op == 3) t[k] = a[k] ^ b[k];
        else { if (k == 0) gen(c, nVars, depth - 1); t[k] = op == 4 ? (a[k] & c[k]) | (b[k] & ~c[k]) : (a[k] & b[k]) | (a[k] & c[k]) | (b[k] & c[k]); }
    }
    if (rnd() & 1) Kit_TruthNot(t, t, nVars);
}
int main(void) {
    Kit_DsdMan_t *m = Kit_DsdManAlloc(8, 64);
    unsigned t[8], co[8], dec[8];
    for (int it = 0; it < 200000; it++) {
        int nVars = 4 + rnd() % 5, nW = Kit_TruthWordNum(nVars);
        // structured-ish random function: OR/AND/XOR of random small terms
        gen(t, nVars, 2 + rnd() % 4);
        Kit_DsdNtk_t *n = Kit_DsdDecomposeExpand(t, nVars);
        Kit_DsdGetSupports(n);
        unsigned bound = 0;
        if (n->nNodes > 0) {
            Kit_DsdObj_t *o = n->pNodes[rnd() % n->nNodes];
            unsigned pick = rnd(), i, lit;
            if (o->Type == KIT_DSD_PRIME) { lit = o->pFans[pick % o->nFans]; bound = Kit_DsdLitSupport(n, lit); }
            else if (o->Type == KIT_DSD_AND || o->Type == KIT_DSD_XOR)
                for (i = 0; i < o->nFans; i++) if ((pick >> i) & 1) bound |= Kit_DsdLitSupport(n, o->pFans[i]);
        }
        if (Kit_WordCountOnes(bound) < 2) { Kit_DsdNtkFree(n); continue; }
        int iVar = Kit_WordFindFirstBit(bound);
        Kit_DsdTruthPartialTwo(m, n, bound, iVar, co, dec);
        Kit_DsdNtkFree(n);
        unsigned part[8];
        n = Kit_DsdDecompose(t, nVars);
        Kit_DsdTruthPartial(m, n, part, bound);
        Kit_DsdNtkFree(n);
        printf("%d %d %x", it, nVars, bound);
        for (int w = 0; w < nW; w++) printf(" %08x", t[w]);
        printf(" |");
        for (int w = 0; w < nW; w++) printf(" %08x", co[w]);
        printf(" |");
        for (int w = 0; w < nW; w++) printf(" %08x", dec[w]);
        printf(" |");
        for (int w = 0; w < nW; w++) printf(" %08x", part[w]);
        printf("\n");
    }
    return 0;
}

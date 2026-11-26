// Random structured (decomposable) functions: prints the truth table, then
// Kit_DsdPrint of Kit_DsdDecompose, Kit_DsdDecomposeExpand and
// Kit_DsdDecomposeMux(.., 8)+Kit_DsdExpand. Produces dsd_struct_native.txt.
#include "misc/util/abc_global.h"
#include "bool/kit/kit.h"
#include <stdio.h>
static unsigned long long s = 4242;
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
    unsigned t[W];
    for (int it = 0; it < 3000; it++) {
        int nVars = 3 + rnd() % 6, nW = Kit_TruthWordNum(nVars);
        gen(t, nVars, 2 + rnd() % 4);
        printf("%d", nVars);
        for (int w = 0; w < nW; w++) printf(" %08x", t[w]);
        Kit_DsdNtk_t *n = Kit_DsdDecompose(t, nVars), *e;
        printf(" | "); Kit_DsdPrint(stdout, n); Kit_DsdNtkFree(n);
        n = Kit_DsdDecomposeExpand(t, nVars);
        printf(" | "); Kit_DsdPrint(stdout, n); Kit_DsdNtkFree(n);
        n = Kit_DsdDecomposeMux(t, nVars, 8); e = Kit_DsdExpand(n);
        printf(" | "); Kit_DsdPrint(stdout, e); Kit_DsdNtkFree(e); Kit_DsdNtkFree(n);
        printf("\n");
    }
    return 0;
}

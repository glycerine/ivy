#include "misc/util/abc_global.h"
#include "bool/kit/kit.h"
#include <stdio.h>
static unsigned long long s = 12345;
static unsigned rnd(void) { s = s * 6364136223846793005ULL + 1442695040888963407ULL; return (unsigned)(s >> 33); }
int main(void) {
    Kit_DsdMan_t *m = Kit_DsdManAlloc(8, 64);
    unsigned t[8], co[8], dec[8];
    for (int it = 0; it < 200000; it++) {
        int nVars = 4 + rnd() % 5, nW = Kit_TruthWordNum(nVars);
        // structured-ish random function: OR/AND/XOR of random small terms
        for (int w = 0; w < nW; w++) t[w] = rnd() & rnd() | (rnd() & rnd() & rnd());
        if (nVars < 5) t[0] = (t[0] & 0xFFFF) | ((t[0] & 0xFFFF) << 16);
        unsigned bound = rnd() & ((1u << nVars) - 1);
        if (Kit_WordCountOnes(bound) < 2) continue;
        Kit_DsdNtk_t *n = Kit_DsdDecomposeExpand(t, nVars);
        int iVar = Kit_WordFindFirstBit(bound);
        Kit_DsdTruthPartialTwo(m, n, bound, iVar, co, dec);
        Kit_DsdNtkFree(n);
        printf("%d %d %x", it, nVars, bound);
        for (int w = 0; w < nW; w++) printf(" %08x", t[w]);
        printf(" |");
        for (int w = 0; w < nW; w++) printf(" %08x", co[w]);
        printf(" |");
        for (int w = 0; w < nW; w++) printf(" %08x", dec[w]);
        printf("\n");
    }
    return 0;
}

// Prints Kit_DsdPrintFromTruth and Kit_DsdWriteFromTruth for random
// functions; produces dsd_print_native.txt.
#include "misc/util/abc_global.h"
#include "bool/kit/kit.h"
#include <stdio.h>
static unsigned long long s = 777;
static unsigned rnd(void) { s = s * 6364136223846793005ULL + 1442695040888963407ULL; return (unsigned)(s >> 33); }
int main(void) {
    unsigned t[8];
    char buf[4096];
    for (int it = 0; it < 400; it++) {
        int nVars = 2 + rnd() % 7, nW = Kit_TruthWordNum(nVars);
        for (int w = 0; w < nW; w++) t[w] = (it & 1) ? (rnd() & rnd()) | (rnd() & rnd() & rnd()) : rnd() & rnd() & rnd();
        if (nVars < 5) { unsigned m = (1u << (1 << nVars)) - 1; t[0] &= m; for (int k = 1 << nVars; k < 32; k <<= 1) t[0] |= t[0] << k; }
        printf("%d", nVars);
        for (int w = 0; w < nW; w++) printf(" %08x", t[w]);
        printf(" | ");
        Kit_DsdPrintFromTruth(t, nVars);
        Kit_DsdWriteFromTruth(buf, t, nVars);
        printf(" | %s\n", buf);
    }
    return 0;
}

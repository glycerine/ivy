#include "aig/aig/aigCanon.c"

/* Console hooks only; canonicalization and Kit are compiled unchanged. */
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    int n, ret, i, j, words;
    while (scanf("%d %d", &n, &ret) == 2)
    {
        unsigned in[RMAX_MAXWORD], out[RMAX_MAXWORD], phase, *result;
        Aig_VSig_t sigs[2*RMAN_MAXVARS];
        char perm[RMAN_MAXVARS];
        assert(n >= 0 && n <= RMAN_MAXVARS);
        words = Kit_TruthWordNum(n);
        for (i = 0; i < words; i++) assert(scanf("%u", in+i) == 1);
        for (i = 0; i < n; i++) perm[i] = i;
        phase = Aig_RManSemiCanonicize(out, in, n, perm, sigs, ret);
        result = ret ? in : out;
        printf("%u %d", phase, Aig_RManVarsAreUnique(sigs, n));
        for (i = 0; i < n; i++) printf(" %u", (unsigned char)perm[i]);
        for (i = 0; i < words; i++) printf(" %u", result[i]);
        for (i = 0; i < 2*n; i++)
        {
            printf(" %d", sigs[i].nOnes);
            for (j = 0; j < n; j++) printf(" %d", sigs[i].nCofOnes[j]);
        }
        puts("");
    }
    return 0;
}

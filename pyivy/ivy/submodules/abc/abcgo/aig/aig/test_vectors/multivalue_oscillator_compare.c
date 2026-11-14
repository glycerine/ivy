#include "aig/saig/saigSimMv.c"

unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }

int main(void)
{
    int n, i, value;
    while (scanf("%d", &n) == 1)
    {
        Saig_MvMan_t p;
        unsigned values[64];
        memset(&p, 0, sizeof(p));
        assert(n > 0 && n <= 64);
        p.nFlops = 1; p.vStates = Vec_PtrAlloc(n+1); Vec_PtrPush(p.vStates, NULL);
        for (i = 0; i < n; i++)
        {
            unsigned *state = ABC_CALLOC(unsigned, 2);
            assert(scanf("%d", &value) == 1); state[1] = value;
            values[i] = value;
            Vec_PtrPush(p.vStates, state);
        }
        printf("%d %d\n", Saig_MvManCheckOscilator(&p, 0), Saig_MvSimHash(values, n, 503));
        Vec_PtrFreeFree(p.vStates);
    }
    return 0;
}

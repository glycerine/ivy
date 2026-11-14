#include "aig/aig/aigTsim.c"

int main(void)
{
    int i, binary, ternary;
    Aig_ManRandom(1);
    for (i = 0; i < 32; i++)
    {
        /* Separate draws: C does not specify argument evaluation order. */
        binary = Aig_XsimRand2(); ternary = Aig_XsimRand3();
        printf("%d %d\n", binary, ternary);
    }
    printf("next %u\n", Aig_ManRandom(0));
    return 0;
}

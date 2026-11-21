#include "aig/ioa/ioa.h"
#include <time.h>

/* Encode the already-loaded circuit, including release of the output buffer. */
int main(int argc, char **argv)
{
    int rounds = argc >= 3 ? atoi(argv[2]) : 0;
    if (rounds <= 0 || argc > 4)
        return 1;
    Aig_Man_t *aig = Ioa_ReadAiger(argv[1], 1);
    if (!aig)
        return 1;
    int size = 0;
    struct timespec begin, end;
    clock_gettime(CLOCK_MONOTONIC, &begin);
    for (int i = 0; i < rounds; ++i) {
        Vec_Str_t *bytes = Ioa_WriteAigerIntoMemoryStr(aig);
        size = Vec_StrSize(bytes);
        Vec_StrFree(bytes);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    if (argc == 4) {
        Vec_Str_t *bytes = Ioa_WriteAigerIntoMemoryStr(aig);
        FILE *out = fopen(argv[3], "wb");
        if (!out || fwrite(Vec_StrArray(bytes), 1, size, out) != (size_t)size || fclose(out))
            return 1;
        Vec_StrFree(bytes);
    }
    printf("%.3f us/op  %d bytes  %d ANDs\n",
        ((end.tv_sec - begin.tv_sec) * 1e6 + (end.tv_nsec - begin.tv_nsec) * .001) / rounds,
        size, Aig_ManNodeNum(aig));
    Aig_ManStop(aig);
    return 0;
}

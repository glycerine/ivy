#include "misc/extra/extra.h"
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

/* Construct and release a caller-owned copy of all four NPN tables. */
int main(int argc, char **argv)
{
    int count = argc == 2 ? atoi(argv[1]) : 0;
    if (count <= 0)
        return 1;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < count; i++) {
        unsigned short *canon;
        char *phase, *perm;
        unsigned char *classes;
        Extra_Truth4VarNPN(&canon, &phase, &perm, &classes);
        if (canon[0xfffe] != 1 || phase[0xffff] != 16 || perm[1] || classes[1] != 1)
            return 2;
        ABC_FREE(canon);
        ABC_FREE(phase);
        ABC_FREE(perm);
        ABC_FREE(classes);
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double ms = 1e3 * (end.tv_sec - start.tv_sec) + 1e-6 * (end.tv_nsec - start.tv_nsec);
    printf("%d NPN tables, %.3f ms/op\n", count, ms / count);
    return 0;
}

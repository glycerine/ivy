#include "bool/kit/kit.h"
#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static uint64_t mix(uint64_t h, unsigned v)
{
    return (h ^ v) * UINT64_C(1099511628211);
}

int main(int argc, char **argv)
{
    int count = argc == 2 ? atoi(argv[1]) : 0;
    if (count <= 0)
        return 1;
    Vec_Int_t *cover = Vec_IntAlloc(256);
    uint64_t digest = 0;
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int batch = 0; batch < count; ++batch) {
        uint32_t state = UINT32_C(0x12345678);
        uint64_t h = 1;
        for (int i = 0; i < 256; ++i) {
            state = UINT32_C(1664525) * state + UINT32_C(1013904223);
            unsigned truth = state;
            int phase = Kit_TruthIsop(&truth, 5, cover, i & 1);
            if (phase < 0)
                return 2;
            h = mix(h, phase);
            h = mix(h, Vec_IntSize(cover));
            for (int j = 0; j < Vec_IntSize(cover); ++j)
                h = mix(h, Vec_IntEntry(cover, j));
        }
        if (batch && h != digest)
            return 3;
        digest = h;
    }
    clock_gettime(CLOCK_MONOTONIC, &end);
    double us = 1e6 * (end.tv_sec - start.tv_sec) + 1e-3 * (end.tv_nsec - start.tv_nsec);
    printf("256 five-variable covers, digest %016" PRIx64 ", %.2f us/batch\n", digest, us / count);
    Vec_IntFree(cover);
    return 0;
}

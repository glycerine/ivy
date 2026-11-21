// Build from the ABC root with -O3 -DABC_USE_STDINT_H -Isrc.
#include "misc/vec/vec.h"
#include <inttypes.h>
#include <time.h>

enum { Records = 128, Width = 17 };
static word data[Records][Width];

static uint64_t digest(int width, int hint)
{
    Vec_Mem_t *mem = Vec_MemAlloc(width, 4);
    Vec_MemHashAlloc(mem, hint);
    uint64_t hash = Vec_IntSize(mem->vTable);
    for (int i = 0; i < Records; ++i) {
        unsigned key = Vec_MemHashKey(mem, data[i]);
        int id = Vec_MemHashInsert(mem, data[i]);
        int prior = Vec_MemHashInsert(mem, data[i / 3]);
        hash = (hash ^ key) * UINT64_C(1099511628211);
        hash = (hash ^ id) * UINT64_C(1099511628211);
        hash = (hash ^ prior) * UINT64_C(1099511628211);
    }
    Vec_MemHashFree(mem);
    Vec_MemFree(mem);
    return hash;
}

int main(int argc, char **argv)
{
    int runs = argc > 1 ? atoi(argv[1]) : 1;
    if (runs < 1)
        return 1;
    word state = UINT64_C(0x18532dc021);
    for (int i = 0; i < Records; ++i)
        for (int k = 0; k < Width; ++k) {
            state ^= state << 13;
            state ^= state >> 7;
            state ^= state << 17;
            data[i][k] = state;
        }
    int widths[] = {1, 2, 4, 5, 9, 17}, hints[] = {0, 1, 2, 3, 4, 17, 32};
    volatile uint64_t sum = 0;
    struct timespec start, stop;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int r = 0; r < runs; ++r) {
        uint64_t hash = UINT64_C(14695981039346656037);
        for (int w = 0; w < 6; ++w)
            for (int h = 0; h < 7; ++h)
                hash = (hash ^ digest(widths[w], hints[h])) * UINT64_C(1099511628211);
        sum = hash;
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    double ns = (stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec;
    printf("%016" PRIx64 " %.2f process ns/record\n", sum, ns/(runs*6.0*7*Records));
    return sum != UINT64_C(0xfce2fa148b5a01ac);
}

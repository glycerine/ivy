// Build from the ABC root with -O3 -DABC_USE_STDINT_H -Isrc.
#include "misc/util/abc_global.h"
#include "misc/util/utilTruth.h"
#include <inttypes.h>
#include <time.h>

static uint64_t digest(unsigned size)
{
    const word repeat = UINT64_C(0x0001000100010001);
    uint32_t state = 0x27159385;
    uint64_t hash = UINT64_C(14695981039346656037);
    for (unsigned i = 0; i < size; ++i) {
        state ^= state << 13;
        state ^= state >> 17;
        state ^= state << 5;
        uint16_t exact = (uint16_t)(i * 40503u + 7919u);
        for (int mode = 0; mode < 2; ++mode) {
            uint16_t allowed = mode ? (uint16_t)state : exact;
            word on = (exact & allowed) * repeat;
            int count = 7;
            word result = Abc_Tt6Isop(on, allowed * repeat, 4, &count);
            hash = (hash ^ result) * UINT64_C(1099511628211);
            hash = (hash ^ count) * UINT64_C(1099511628211);
        }
    }
    return hash;
}

int main(int argc, char **argv)
{
    int runs = argc > 1 ? atoi(argv[1]) : 1;
    unsigned size = argc > 2 ? (unsigned)atoi(argv[2]) : 65536;
    if (runs < 1 || !size || size > 65536)
        return 1;
    struct timespec start, stop;
    volatile uint64_t hash = 0;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < runs; ++i)
        hash = digest(size);
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    double ns = (stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec;
    printf("%016" PRIx64 " %.2f process ns/pair\n", hash, ns/(runs*size*2.0));
    if (size == 65536)
        return hash != UINT64_C(0x8e445a19aae238cd);
    if (size == 128)
        return hash != UINT64_C(0x3b6cdca3e521626d);
    return 0;
}

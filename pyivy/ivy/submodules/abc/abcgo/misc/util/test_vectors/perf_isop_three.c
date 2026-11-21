// Build from the ABC root with -O3 -DABC_USE_STDINT_H -Isrc.
#include "misc/util/abc_global.h"
#include "misc/util/utilTruth.h"
#include <inttypes.h>
#include <time.h>

static uint64_t isop_digest(void)
{
    const word repeat = UINT64_C(0x0101010101010101);
    uint64_t hash = UINT64_C(14695981039346656037);
    for (unsigned allowed = 0; allowed < 256; ++allowed) {
        for (unsigned on = allowed; ; on = (on - 1) & allowed) {
            int count = 5;
            word truth = Abc_Tt6Isop(on * repeat, allowed * repeat, 3, &count);
            hash = (hash ^ truth) * UINT64_C(1099511628211);
            hash = (hash ^ count) * UINT64_C(1099511628211);
            if (!on)
                break;
        }
    }
    return hash;
}

int main(int argc, char **argv)
{
    int runs = argc > 1 ? atoi(argv[1]) : 1;
    if (runs < 1)
        return 1;
    struct timespec start, stop;
    volatile uint64_t digest = 0;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < runs; ++i)
        digest = isop_digest();
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    double ns = (stop.tv_sec - start.tv_sec) * 1e9 + stop.tv_nsec - start.tv_nsec;
    printf("%016" PRIx64 " %.2f process ns/pair\n", digest, ns / (6561.0 * runs));
    return digest != UINT64_C(0xa40a7d1f15a0d833);
}

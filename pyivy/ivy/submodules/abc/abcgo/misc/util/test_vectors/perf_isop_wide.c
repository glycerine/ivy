// Build from the ABC root with -O3 -DABC_USE_STDINT_H -Isrc.
#include "misc/util/abc_global.h"
#include "misc/util/utilTruth.h"
#include <inttypes.h>
#include <time.h>

static word next_word(word *state)
{
    *state ^= *state << 13;
    *state ^= *state >> 7;
    *state ^= *state << 17;
    return *state;
}

static uint64_t mix(uint64_t hash, word value)
{
    return (hash ^ value) * UINT64_C(1099511628211);
}

static uint64_t wide_digest(void)
{
    uint64_t hash = UINT64_C(14695981039346656037);
    word state = UINT64_C(0xd51d28de17);
    for (int trial = 0; trial < 96; ++trial) {
        int n = 6 + trial % 3;
        word on[4], dc[4], out[4], covered[4];
        for (int k = 0; k < 4; ++k) {
            dc[k] = next_word(&state);
            on[k] = dc[k] & next_word(&state);
        }
        if (trial % 4 == 0)
            memcpy(dc, on, sizeof(on));
        else if (trial % 4 == 1) {
            for (int k = 2; k < 4; ++k) {
                on[k] = on[k-2];
                dc[k] = dc[k-2];
            }
        } else if (trial % 4 == 2) {
            on[1] = on[0];
            dc[1] = dc[0];
        }
        int codes[260] = {7, 11}, count = 2;
        int size = Abc_Tt8Isop(on, dc, n, out);
        Abc_Tt8IsopCover(on, dc, n, covered, codes, &count);
        hash = mix(mix(hash, size), count);
        for (int k = 0; k < 4; ++k)
            hash = mix(mix(hash, out[k]), covered[k]);
        for (int k = 0; k < count; ++k)
            hash = mix(hash, codes[k]);
    }
    return hash;
}

static uint64_t word_count_digest(void)
{
    uint64_t hash = UINT64_C(14695981039346656037);
    word state = UINT64_C(0x143d5aa9cb);
    for (unsigned i = 0; i < 8192; ++i) {
        int n = 5 + (i & 1);
        word dc = next_word(&state), on = dc & next_word(&state);
        if (i % 3 == 0)
            dc = on;
        if (i % 5 == 0) {
            on = (uint16_t)on * UINT64_C(0x0001000100010001);
            dc = (uint16_t)dc * UINT64_C(0x0001000100010001);
        }
        if (n == 5 || i % 7 == 0) {
            on = (uint32_t)on * UINT64_C(0x0000000100000001);
            dc = (uint32_t)dc * UINT64_C(0x0000000100000001);
        }
        int count = 3;
        word result = Abc_Tt6Isop(on, dc, n, &count);
        hash = mix(mix(hash, result), (word)count);
    }
    return hash;
}

int main(int argc, char **argv)
{
    int runs = argc > 1 ? atoi(argv[1]) : 1;
    int counts = argc > 2 && strcmp(argv[2], "count") == 0;
    if (runs < 1)
        return 1;
    struct timespec start, stop;
    volatile uint64_t digest = 0;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &start);
    for (int i = 0; i < runs; ++i)
        digest = counts ? word_count_digest() : wide_digest();
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &stop);
    double ns = (stop.tv_sec-start.tv_sec)*1e9 + stop.tv_nsec-start.tv_nsec;
    printf("%016" PRIx64 " %.2f process ns/table\n", digest, ns / ((counts ? 8192.0 : 96.0)*runs));
    if (counts) return digest != UINT64_C(0x211d762f71704346);
    return digest != UINT64_C(0xc5fa775abb4222c8);
}

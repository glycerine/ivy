// Reference workloads for the Go GIA benchmarks. See PERF_IMPROVE.md.
#include "aig/gia/gia.h"
#include "sat/satoko/satoko.h"
#include <time.h>

extern satoko_t *Gia_ManSatokoCreate(Gia_Man_t *, satoko_opts_t *);

static double now(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts.tv_sec + 1e-9 * ts.tv_nsec;
}

static int load_i10(char *path)
{
    Gia_Man_t *p = Gia_AigerRead(path, 0, 0, 0);
    if (!p)
        return 0;
    int ok = Gia_ManCiNum(p) == 257 && Gia_ManCoNum(p) == 224 && Gia_ManAndNum(p) == 2675;
    Gia_ManStop(p);
    return ok;
}

static int hash_chain(void)
{
    Gia_Man_t *p = Gia_ManStart(64);
    int a = Gia_ManAppendCi(p);
    int x = Gia_ManAppendCi(p);
    Gia_ManHashAlloc(p);
    int ok = 1;
    for (int i = 0; i < 4096; ++i) {
        int fanin = x ^ (i & 1);
        int y = Gia_ManHashAnd(p, a, fanin);
        if (Gia_ManHashAnd(p, fanin, a) != y) {
            ok = 0;
            break;
        }
        a = y;
    }
    Gia_ManHashStop(p);
    Gia_ManStop(p);
    return ok;
}

int main(int argc, char **argv)
{
    if (argc != 2 && argc != 3) {
        fprintf(stderr, "usage: %s cone | hash | (load|cones|sat) i10.aig\n", argv[0]);
        return 2;
    }
    int runs;
    double start;
    if (argc == 2 && !strcmp(argv[1], "cone")) {
        Gia_Man_t *p = Gia_ManStart(2048);
        int a = Gia_ManAppendCi(p);
        int other = Gia_ManAppendCi(p);
        for (int i = 0; i < 1024; ++i)
            a = Gia_ManAppendAnd(p, a, other);
        Gia_ManAppendCo(p, a);
        int root = Gia_ManObjNum(p) - 1;
        runs = 200000;
        start = now();
        for (int i = 0; i < runs; ++i)
            if (Gia_ManConeSize(p, &root, 1) != 1024)
                return 1;
        printf("cone %.0f ns/op\n", (now() - start) * 1e9 / runs);
        Gia_ManStop(p);
        return 0;
    }
    if (argc == 3 && !strcmp(argv[1], "cones")) {
        Gia_Man_t *p = Gia_AigerRead(argv[2], 0, 0, 0);
        if (!p)
            return 1;
        runs = 4000;
        long long total = 0;
        start = now();
        for (int i = 0; i < runs; ++i) {
            for (int j = 0; j < Gia_ManCoNum(p); ++j) {
                int root = Gia_ObjId(p, Gia_ManCo(p, j));
                total += Gia_ManConeSize(p, &root, 1);
            }
        }
        printf("cones %.0f ns/op, %lld nodes/sweep\n", (now() - start) * 1e9 / runs, total / runs);
        Gia_ManStop(p);
        return 0;
    }
    if (argc == 3 && !strcmp(argv[1], "sat")) {
        satoko_opts_t opts;
        satoko_default_opts(&opts);
        runs = 30;
        start = now();
        long conflicts = 0;
        for (int i = 0; i < runs; ++i) {
            Gia_Man_t *p = Gia_AigerRead(argv[2], 0, 0, 0);
            if (!p)
                return 1;
            satoko_t *s = Gia_ManSatokoCreate(p, &opts);
            if (!s || satoko_solve(s) != SATOKO_SAT)
                return 1;
            conflicts += satoko_stats(s)->n_conflicts;
            satoko_destroy(s);
            Gia_ManStop(p);
        }
        printf("sat %.0f ns/op, %ld conflicts/op\n", (now() - start) * 1e9 / runs, conflicts / runs);
        return 0;
    }
    if (argc == 2 && !strcmp(argv[1], "hash")) {
        runs = 4000;
        start = now();
        for (int i = 0; i < runs; ++i)
            if (!hash_chain())
                return 1;
    } else if (argc == 3 && !strcmp(argv[1], "load")) {
        runs = 4000;
        start = now();
        for (int i = 0; i < runs; ++i)
            if (!load_i10(argv[2]))
                return 1;
    } else {
        return 2;
    }
    printf("%s %.0f ns/op\n", argv[1], (now() - start) * 1e9 / runs);
    return 0;
}

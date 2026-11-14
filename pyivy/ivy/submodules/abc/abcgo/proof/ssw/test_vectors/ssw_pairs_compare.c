#include "proof/ssw/sswPairs.c"
#include <sys/prctl.h>

int main(int argc, char **argv)
{
    Vec_Int_t *v, **c;
    int i, j;
    /* The invalid-pair oracle intentionally aborts. Avoid host crash-report
       handlers (including piped handlers, which can ignore RLIMIT_CORE). */
    if (prctl(PR_SET_DUMPABLE, 0, 0, 0, 0) != 0)
    {
        perror("disable oracle core dumps");
        return 2;
    }
    if (argc == 2 && strcmp(argv[1], "--dumpable") == 0)
    {
        printf("%d\n", prctl(PR_GET_DUMPABLE, 0, 0, 0, 0));
        return 0;
    }
    v = Vec_IntAlloc(argc);
    for (i = 1; i < argc; i++)
        Vec_IntPush(v, atoi(argv[i]));
    c = Ssw_TransformPairsIntoTempClasses(v, 12);
    for (i = 0; i < 12; i++)
        if (c[i])
        {
            printf("%d:", i);
            for (j = 0; j < Vec_IntSize(c[i]); j++)
                printf("%s%d", j ? "," : "", Vec_IntEntry(c[i], j));
            printf(";");
        }
    puts("");
    return 0;
}

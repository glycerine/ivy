#include "proof/fraig/fraigInt.h"
#include <stdio.h>
#include <stdlib.h>

int main(void)
{
    int sizes[] = {8, 64, 1000, 9000};
    int counts[] = {1024, 1024, 65, 8};
    for (int k = 0; k < 4; k++) {
        Fraig_MemFixed_t *pool = Fraig_MemFixedStart(sizes[k]);
        char **entries = malloc((counts[k] + 1) * sizeof(*entries));
        for (int i = 0; i <= counts[k]; i++)
            entries[i] = Fraig_MemFixedEntryFetch(pool);
        int before = Fraig_MemFixedReadMemUsage(pool);
        Fraig_MemFixedEntryRecycle(pool, entries[1]);
        Fraig_MemFixedEntryRecycle(pool, entries[2]);
        int first = Fraig_MemFixedEntryFetch(pool) == entries[2];
        int second = Fraig_MemFixedEntryFetch(pool) == entries[1];
        Fraig_MemFixedRestart(pool);
        int after = Fraig_MemFixedReadMemUsage(pool);
        int restart = Fraig_MemFixedEntryFetch(pool) == entries[0];
        printf("%d %d %d %d %d %d\n", sizes[k], before, after, first, second, restart);
        free(entries);
        Fraig_MemFixedStop(pool, 0);
    }
    return 0;
}

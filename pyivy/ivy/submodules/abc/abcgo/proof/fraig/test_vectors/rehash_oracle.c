#include "proof/fraig/fraigInt.h"
#include <stdio.h>

int main(void)
{
    Fraig_Man_t manager = {0};
    Fraig_HashTable_t table = {0};
    Fraig_Node_t a = {0}, b = {0}, c = {0};

    table.nBins = 3;
    table.nEntries = 3;
    table.pBins = calloc(3, sizeof(*table.pBins));
    manager.pTableF0 = &table;
    a.uHashD = 6;
    b.uHashD = 6;
    c.uHashD = 0;
    a.pNextF = &b;
    b.pNextF = &c;
    table.pBins[0] = &a;

    for (int i = 0; i < 3; i++) {
        b.uHashD = 6;
        int linked = Fraig_TableRehashF0(&manager, 1);
        int grouped = a.pNextD == &b || b.pNextD == &a;
        b.uHashD = 5;
        Fraig_TableRehashF0(&manager, 0);
        int separate = a.pNextD == NULL && b.pNextD == NULL && c.pNextD == NULL;
        printf("%d %d %d %d\n", linked, grouped, separate, table.nEntries);
    }

    free(table.pBins);
    return 0;
}

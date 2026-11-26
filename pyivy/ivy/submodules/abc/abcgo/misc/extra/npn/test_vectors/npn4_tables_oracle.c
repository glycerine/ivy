// Prints ABC's Extra_Truth4VarNPN tables, one truth per line:
// truth canon phase perm class
#include "misc/extra/extra.h"
#include <stdio.h>

// Stubs for framework hooks that extraUtilMisc.c references but this
// standalone oracle never calls.
unsigned enable_dbg_outs = 0;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { return 0; }
word Abc_RandomW(int fReset) { return 0; }

int main(void)
{
    unsigned short *canons;
    char *phases, *perms;
    unsigned char *map;
    Extra_Truth4VarNPN(&canons, &phases, &perms, &map);
    for (int t = 0; t < (1 << 16); t++)
        printf("%d %d %d %d %d\n", t, canons[t], phases[t] & 0xff, perms[t] & 0xff, map[t]);
    return 0;
}

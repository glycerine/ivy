#include <stdio.h>
#include <stdlib.h>
extern char *Wln_ConstFromBits(int *, int);
extern char *Wln_ConstFromStr(char *, int);
int main(void) {
    int bits[] = {1, 1};
    char *a = Wln_ConstFromBits(bits, 33);
    char *b = Wln_ConstFromStr("10xzextra", 4);
    puts(a); puts(b);
    free(a); free(b);
    return 0;
}

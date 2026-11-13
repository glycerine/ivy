#include "aig/gia/gia.h"
extern void *Acec_ProduceBox(Gia_Man_t *, int);
int main(void) {
    Gia_Man_t *p = Gia_ManStart(8);
    int a = Gia_ManAppendCi(p);
    Gia_ManAppendCo(p, a);
    Acec_ProduceBox(p, 0);
    return 0;
}

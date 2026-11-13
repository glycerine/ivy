#include <stdio.h>
#include "base/wln/wln.h"
int main(void) {
    Wln_Ntk_t *p = Wln_NtkAlloc(NULL, 16);
    int obj = Wln_ObjAlloc(p, ABC_OPER_NONE, 0, 0, 0);
    Wln_ObjSetFanin(p, obj, 1, 7);
    printf("inline_metadata=%d\n", Wln_ObjFanin(p, obj, 1));
    for (int i = 0; i < 10; i++) {
        Wln_ObjAddFanin(p, obj, 10 + i);
        printf("size=%d cap=%d values=", Wln_ObjFaninNum(p, obj), p->vFanins[obj].nCap);
        for (int j = 0; j <= i; j++) printf("%s%d", j ? "," : "", Wln_ObjFanin(p, obj, j));
        puts("");
    }
    Wln_NtkFree(p);
    return 0;
}

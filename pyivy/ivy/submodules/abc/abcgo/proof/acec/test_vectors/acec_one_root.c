#include "aig/gia/gia.h"
extern Vec_Int_t *Acec_OrderTreeRoots(Gia_Man_t *,Vec_Int_t *,Vec_Int_t *,Vec_Int_t *);
int main(void){Gia_Man_t *p=Gia_ManStart(8);Vec_Int_t *adds=Vec_IntAlloc(0),*roots=Vec_IntAlloc(1),*ranks=Vec_IntAlloc(0);Vec_IntPush(roots,1);Acec_OrderTreeRoots(p,adds,roots,ranks);return 0;}

#include <stdio.h>
#include "base/cba/cba.h"
int main(void) {
 for(int i=-1;i<=CBA_BOX_LAST;i++) printf("%d %d %d %d %d\n",i,Cba_TypeIsBox(i),Cba_TypeIsSeq(i),Cba_TypeIsUnary(i),Cba_TypeIsMux(i));
 return 0;
}

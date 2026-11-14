#include <stdio.h>
#include "base/cba/cbaReadBlif.c"
int main(void){for(int i=0;i<256;i++){char c=(char)i;printf("%d %d %d\n",Prs_CharIsSpace(c),Prs_CharIsStop(c),Prs_CharIsLit(c));}return 0;}

#include <stdio.h>
#include "base/cba/cbaReadVer.c"
int main(void) {
 for(int i=0;i<256;i++) {
  char c=(char)i;
  printf("%d %d %d %d %d %d %d\n",Prs_CharIsSpace(c),Prs_CharIsDigit(c),Prs_CharIsDigitB(c),Prs_CharIsDigitH(c),Prs_CharIsChar(c),Prs_CharIsSymb1(c),Prs_CharIsSymb2(c));
 }
 return 0;
}

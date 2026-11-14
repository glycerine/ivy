#include <stdio.h>
#include "base/cba/cbaWriteVer.c"
int main(void) {
 char *names[]={"wire","wire","_x9$","9wire","a-b","\\escaped name","a\0-b"};
 int ids[]={12,13,13,13,13,1,13};
 for(int i=0;i<7;i++) printf("%d %d\n",Cba_NameIsLegalInVerilog(names[i],ids[i]),Prs_NameIsLegalInVerilog(names[i],ids[i]));
 return 0;
}

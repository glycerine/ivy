#include <stdio.h>
#include "base/cba/cba.h"
#include "base/cba/cbaPrs.h"
int main(void) {
 char *s[]={" 0\n"," 1\n","1 1\n","0 1\n","11 1\n","00 1\n","00 0\n","-1 1\n1- 1\n","1- 1\n-1 1\n","01 1\n10 1\n","10 1\n01 1\n","11 1\n00 1\n","00 1\n11 1\n","10 1\n","01 1\n"};
 for(int i=0;i<15;i++) printf("%d %s\n",Ptr_SopToType(s[i]),Ptr_SopToTypeName(s[i]));
 return 0;
}

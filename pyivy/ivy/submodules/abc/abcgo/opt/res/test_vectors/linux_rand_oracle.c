#include <stdio.h>
#include <stdlib.h>
int main(void) { srand(0xabc); for (int i=0;i<16;i++) printf("%d\n",rand()); return 0; }

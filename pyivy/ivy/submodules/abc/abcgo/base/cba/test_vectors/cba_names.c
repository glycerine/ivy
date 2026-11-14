#include <stdio.h>
extern int Cba_StrCmp(char **, char **);
int main(void) {
 char *v[][2]={{"wire2","wire10"},{"wire19","wire3"},{"a_1_x","a_2_x"},{"wire","wire1"},{"wire1","wire"},{"node3a","node3b"},{"a#2","a$1"},{"x1000000000","x999999999"}};
 for(unsigned i=0;i<sizeof(v)/sizeof(v[0]);i++) printf("%d\n",Cba_StrCmp(&v[i][0],&v[i][1]));
 return 0;
}

#include <stdio.h>
int BacManReadBacNameAndNums(char *, int *, int *, int *, int *);
int main(void){
 char a[][32]={"design 2 ","net 1 2 4 3 ","net 1","net -1 2","net 0x 7"};
 for(int i=0;i<5;i++){int w,x,y,z;int ok=BacManReadBacNameAndNums(a[i],&w,&x,&y,&z);printf("%s|%d|%d,%d,%d,%d\n",a[i],ok,w,x,y,z);}
}

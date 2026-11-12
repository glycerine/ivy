#include <stdio.h>
#include <stdlib.h>
#define LIN64
#define BPI 32
#define ABC_NAMESPACE_HEADER_START
#define ABC_NAMESPACE_HEADER_END
#define ABC_NAMESPACE_IMPL_START
#define ABC_NAMESPACE_IMPL_END
#define ALLOC(type,num) ((type*)malloc(sizeof(type)*(num)))
#define REALLOC(type,p,num) ((type*)realloc((p),sizeof(type)*(num)))
#define FREE(p) free(p)
#define ABS(x) ((x)<0?-(x):(x))
#include "../../../../src/misc/espresso/set.c"
int bit_count[256];
void fatal(char *s){fprintf(stderr,"%s\n",s);exit(1);}
int main(void){int i;pset a,b,r;pset_family f;
for(i=0;i<256;i++){int v=i;while(v){bit_count[i]+=v&1;v>>=1;}}
a=set_new(65);b=set_new(65);r=set_new(65);
set_insert(a,0);set_insert(a,31);set_insert(a,32);set_insert(a,64);
set_insert(b,1);set_insert(b,31);set_insert(b,33);set_insert(b,64);
SET(r,PRIME);set_and(r,a,b);printf("and ");set_write(stdout,r);
set_or(r,a,b);printf("or ");set_write(stdout,r);
set_diff(r,a,b);printf("diff ");set_write(stdout,r);
set_xor(r,a,b);printf("xor ");set_write(stdout,r);
printf("counts %d %d %d\n",set_ord(a),set_dist(a,b),bit_index(0x80000000));
f=sf_new(2,65);sf_addset(f,a);sf_addset(f,b);sf_write(stdout,f);
f=sf_addcol(f,31,2);printf("inserted %s\n",ps1(GETSET(f,0)));
f=sf_delcol(f,31,2);printf("deleted %s\n",ps1(GETSET(f,0)));
f=sf_transpose(f);printf("transpose %d %d %s\n",f->count,f->sf_size,ps1(GETSET(f,31)));
return 0;}

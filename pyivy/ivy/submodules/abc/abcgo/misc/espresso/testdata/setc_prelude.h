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

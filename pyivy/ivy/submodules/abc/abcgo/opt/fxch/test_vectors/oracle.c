// Driver for unchanged ABC FXCH original implementation.
#include "opt/fxch/Fxch.h"
// Only application logging hooks are supplied; extraction is unchanged.
unsigned enable_dbg_outs=0;
int Abc_FrameIsBridgeMode(void){return 0;}
char *vnsprintf(const char *format,va_list args){va_list copy;va_copy(copy,args);int n=vsnprintf(NULL,0,format,copy);va_end(copy);char *s=malloc(n+1);vsnprintf(s,n+1,format,args);return s;}
int main(int argc,char **argv){int mode=argc>1?atoi(argv[1]):0,i,j;Vec_Wec_t *v=Vec_WecAlloc(10);int a[][5]={{8,0,2,4,-1},{8,0,2,6,-1},{8,0,2,8,-1}};for(i=0;i<3;i++)for(j=0;j<4;j++)Vec_WecPush(v,i,a[i][j]);if(mode){Vec_WecClear(v);int b[][4]={{8,0,4,-1},{8,2,4,-1},{9,0,6,-1},{9,2,6,-1}};for(i=0;i<4;i++)for(j=0;j<3;j++)Vec_WecPush(v,i,b[i][j]);}if(mode==2){Vec_WecClear(v);int m[][5]={{8,0,2,-1,-1},{8,1,4,-1,-1},{9,0,2,6,-1},{9,1,4,6,-1}};for(i=0;i<4;i++)for(j=0;j<5&&m[i][j]>=0;j++)Vec_WecPush(v,i,m[i][j]);}Fxch_FastExtract(v,10,1,0,0);for(i=0;i<Vec_WecSize(v);i++){Vec_Int_t *c=Vec_WecEntry(v,i);for(j=0;j<Vec_IntSize(c);j++)printf("%s%d",j?" ":"",Vec_IntEntry(c,j));printf("\n");}return 0;}

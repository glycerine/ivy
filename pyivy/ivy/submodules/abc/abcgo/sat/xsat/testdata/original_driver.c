#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include "sat/xsat/xsat.h"
// The standalone oracle uses ABC's ordinary non-bridge diagnostics mode.
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void){return 0;}
int Gia_ManToBridgeText(FILE *file,int size,unsigned char *buffer){return (int)fwrite(buffer,1,(size_t)size,file);}
char *vnsprintf(const char *format,va_list args){va_list copy;va_copy(copy,args);int n=vsnprintf(NULL,0,format,copy);va_end(copy);if(n<0)return NULL;char *s=malloc((size_t)n+1);if(s)vsnprintf(s,(size_t)n+1,format,args);return s;}
int main(int argc,char **argv){
 if(argc!=2)return 2;
 FILE *f=fopen(argv[1],"r");if(!f)return 2;
 xSAT_Solver_t *s=NULL;
 int parsed=xSAT_SolverParseDimacs(f,&s);fclose(f);
 if(!parsed){puts("RESULT=-1");return 0;}
 int result=xSAT_SolverSolve(s);
 printf("RESULT=%d\n",result);
 xSAT_SolverDestroy(s);return 0;
}

#include <stdio.h>
typedef struct node {int type;char *name;void *object;struct node *left,*right;} node;
extern node *parseFormulaCreateAST(char *);
extern node *readLtlFormula(char *);
extern void resetGlobalVar(void);
extern int startOfSuffixString;
int main(void) {
 char a[]=">signal trailing",b[]="alpha beta";
 node *x=parseFormulaCreateAST(a),*y=parseFormulaCreateAST(b);
 printf("%d:%s\n%d:%s\n",x->type,x->name,y->type,y->name);
 resetGlobalVar();x=readLtlFormula(b);printf("%s:%d\n",x->name,startOfSuffixString);
 y=readLtlFormula(b);printf("%s:%d\n",y->name,startOfSuffixString);
 resetGlobalVar();char prefix[]="skip", formula[]="skip:G :F signal trailing";
 readLtlFormula(prefix);x=parseFormulaCreateAST(formula);
 printf("primed:%d:%d:%s:%d\n",x->type,x->left->type,x->left->left->name,startOfSuffixString);
 return 0;
}

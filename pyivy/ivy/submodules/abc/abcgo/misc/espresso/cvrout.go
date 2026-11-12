package espresso

import "fmt"

import "os"

/*
 * Revision Control Information
 *
 * $Source$
 * $Author$
 * $Revision$
 * $Date$
 *
 */
/*
    module: cvrout.c
    purpose: cube and cover output routines
*/

//#include "espresso.h"

//ABC_NAMESPACE_IMPL_START


void fprint_pla(fp, PLA, output_type)
INOUT FILE *fp;
IN pPLA PLA;
IN int output_type;
{
    int num;
    register pcube last, p;

    if (output_type & CONSTRAINTS_type) != 0 {
    output_symbolic_constraints(fp, PLA, 0);
    output_type &= ~ CONSTRAINTS_type;
    if output_type == 0 {
        return;
    }
    }

    if (output_type & SYMBOLIC_CONSTRAINTS_type) != 0 {
    output_symbolic_constraints(fp, PLA, 1);
    output_type &= ~ SYMBOLIC_CONSTRAINTS_type;
    if output_type == 0 {
        return;
    }
    }

    if output_type == PLEASURE_type {
    pls_output(PLA);
    } else if output_type == EQNTOTT_type {
    eqn_output(PLA);
    } else if output_type == KISS_type {
    kiss_output(fp, PLA);
    } else {
    fpr_header(fp, PLA, output_type);

    num = 0;
    if (output_type & F_type) num += (PLA.F).count;
    if (output_type & D_type) num += (PLA.D).count;
    if (output_type & R_type) num += (PLA.R).count;
    (void) fmt.Fprintf(fp, ".p %d\n", num);

    /* quick patch 01/17/85 to support TPLA ! */
    if output_type == F_type {
        // inlining of macro foreach_set
        for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
        print_cube(fp, p, "01");
        }
        (void) fmt.Fprintf(fp, ".e\n");
    } else {
        if output_type & F_type {
        // inlining of macro foreach_set
        for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
            print_cube(fp, p, "~1");
        }
        }
        if output_type & D_type {
        // inlining of macro foreach_set
        for p=PLA.D.data,last=p+PLA.D.count*PLA.D.wsize;p<last;p+=PLA.D.wsize {
            print_cube(fp, p, "~2");
        }
        }
        if output_type & R_type {
        // inlining of macro foreach_set
        for p=PLA.R.data,last=p+PLA.R.count*PLA.R.wsize;p<last;p+=PLA.R.wsize {
            print_cube(fp, p, "~0");
        }
        }
        (void) fmt.Fprintf(fp, ".end\n");
    }
    }
}

void fpr_header(fp, PLA, output_type)
var fp *os.File;
var PLA pPLA;
var output_type int32;
{
    register int i, var;
    int first, last;

    /* .type keyword gives logical type */
    if output_type != F_type {
    (void) fmt.Fprintf(fp, ".type ");
    if (output_type & F_type) putc('f', fp);
    if (output_type & D_type) putc('d', fp);
    if (output_type & R_type) putc('r', fp);
    putc('\n', fp);
    }

    /* Check for binary or multiple-valued labels */
    if cube.num_mv_vars <= 1 {
    (void) fmt.Fprintf(fp, ".i %d\n", cube.num_binary_vars);
    if (cube.output != -1)
        (void) fmt.Fprintf(fp, ".o %d\n", cube.part_size[cube.output]);
    } else {
    (void) fmt.Fprintf(fp, ".mv %d %d", cube.num_vars, cube.num_binary_vars);
    for(var = cube.num_binary_vars; var < cube.num_vars; var++)
        (void) fmt.Fprintf(fp, " %d", cube.part_size[var]);
    (void) fmt.Fprintf(fp, "\n");
    }

    /* binary valued labels */
    if PLA.label != NIL(char *) && PLA.label[1] != NILbyte(
        && cube.num_binary_vars) > 0 {
    (void) fmt.Fprintf(fp, ".ilb");
    for(var = 0; var < cube.num_binary_vars; var++)
      /* see (NIL) OUTLABELS comment below */
      // inlining of macro INLABEL
      if PLA.label[cube.first_part[var] + 1] == NIL(char){
        (void) fmt.Fprintf(fp, " (null)");
      }
      else{
        // inlining of macro INLABEL
        (void) fmt.Fprintf(fp, " %s", PLA.label[cube.first_part[var] + 1]);
      }
    putc('\n', fp);
    }

    /* output-part (last multiple-valued variable) labels */
    if PLA.label != NIL(char *) &&
        PLA.label[cube.first_part[cube.output]] != NILbyte(
        && cube.output) != -1 {
    (void) fmt.Fprintf(fp, ".ob");
    for(i = 0; i < cube.part_size[cube.output]; i++)
      /* (NIL) OUTLABELS caused espresso to segfault under solaris */
      // inlining of macro OUTLABEL
      if PLA.label[cube.first_part[cube.output] + i] == NIL(char){
        (void) fmt.Fprintf(fp, " (null)");
      }
      else{
        // inlining of macro OUTLABEL
        (void) fmt.Fprintf(fp, " %s", PLA.label[cube.first_part[cube.output] + i]);
      }
    putc('\n', fp);
    }

    /* multiple-valued labels */
    for var = cube.num_binary_vars; var < cube.num_vars-1; var++ {
    first = cube.first_part[var];
    last = cube.last_part[var];
    if PLA.label != NULL && PLA.label[first] != NULL {
        (void) fmt.Fprintf(fp, ".label var=%d", var);
        for i = first; i <= last; i++ {
        (void) fmt.Fprintf(fp, " %s", PLA.label[i]);
        }
        putc('\n', fp);
    }
    }

    if PLA.phase != (pcube) NULL {
    first = cube.first_part[cube.output];
    last = cube.last_part[cube.output];
    (void) fmt.Fprintf(fp, "#.phase ");
    for(i = first; i <= last; i++)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        putc((PLA.phase[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) ? '1' : '0', fp);
    (void) fmt.Fprintf(fp, "\n");
    }
}

void pls_output(PLA)
IN pPLA PLA;
{
    register pcube last, p;

    (void) fmt.Printf(".option unmerged\n");
    makeup_labels(PLA);
    pls_label(PLA, os.Stdout);
    pls_group(PLA, os.Stdout);
    (void) fmt.Printf(".p %d\n", PLA.F.count);
    // inlining of macro foreach_set
    for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
    print_expanded_cube(os.Stdout, p, PLA.phase);
    }
    (void) fmt.Printf(".end\n");
}


void pls_group(PLA, fp)
var PLA pPLA;
var fp *os.File;
{
    int var, i, col, len;

    (void) fmt.Fprintf(fp, "\n.group");
    col = 6;
    for var = 0; var < cube.num_vars-1; var++ {
    (void) fmt.Fprintf(fp, " ("), col += 2;
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        len = len(PLA.label[i]);
        if (col + len > 75)
        (void) fmt.Fprintf(fp, " \\\n"), col = 0;
        else if (i != 0)
        putc(' ', fp), col += 1;
        (void) fmt.Fprintf(fp, "%s", PLA.label[i]), col += len;
    }
    (void) fmt.Fprintf(fp, ")"), col += 1;
    }
    (void) fmt.Fprintf(fp, "\n");
}


void pls_label(PLA, fp)
var PLA pPLA;
var fp *os.File;
{
    int var, i, col, len;

    (void) fmt.Fprintf(fp, ".label");
    col = 6;
    for(var = 0; var < cube.num_vars; var++)
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        len = len(PLA.label[i]);
        if (col + len > 75)
        (void) fmt.Fprintf(fp, " \\\n"), col = 0;
        else
        putc(' ', fp), col += 1;
        (void) fmt.Fprintf(fp, "%s", PLA.label[i]), col += len;
    }
}



/*
    eqntott output mode -- output algebraic equations
*/
void eqn_output(PLA)
var PLA pPLA;
{
    register pcube p, last;
    register int i, var, col, len;
    int x;
    bool firstand, firstor;

    if (cube.output == -1)
    fatal("Cannot have no-output function for EQNTOTT output mode");
    if (cube.num_mv_vars != 1)
    fatal("Must have binary-valued function for EQNTOTT output mode");
    makeup_labels(PLA);

    /* Write a single equation for each output */
    for i = 0; i < cube.part_size[cube.output]; i++ {
    // inlining of macro OUTLABEL
    (void) fmt.Printf("%s = ", PLA.label[cube.first_part[cube.output] + i]);
    // inlining of macro OUTLABEL
    col = len(PLA.label[cube.first_part[cube.output] + i]) + 3;
    firstor = TRUE;

    /* Write product terms for each cube in this output */
    // inlining of macro foreach_set
    for(p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize)
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((i + cube.first_part[cube.output]) >> LOGBPI) + 1)] & (1 << ((i + cube.first_part[cube.output]) & (BPI-1))) {
        if (firstor)
            (void) fmt.Printf("("), col += 1;
        else
            (void) fmt.Printf(" | (");
col += 4;
        firstor = FALSE;
        firstand = TRUE;

        /* print out a product term */
        for(var = 0; var < cube.num_binary_vars; var++)
            // inlining of macro GETINPUT
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            if (x=((p[(((2*var) >> LOGBPI) + 1)] >> ((2*var) & (BPI-1))) & 3)) != DASH {
            // inlining of macro INLABEL
            len = len(PLA.label[cube.first_part[var] + 1]);
            if (col+len > 72)
                (void) fmt.Printf("\n    "), col = 4;
            if (! firstand)
                (void) fmt.Printf("&"), col += 1;
            firstand = FALSE;
            if (x == ZERO)
                (void) fmt.Printf("!"), col += 1;
            // inlining of macro INLABEL
            (void) fmt.Printf("%s", PLA.label[cube.first_part[var] + 1]), col += len;
            }
        (void) fmt.Printf(")"), col += 1;
        }
    (void) fmt.Printf(";\n\n");
    }
}


char *fmt_cube(c, out_map, s)
register pcube c;
register char *out_map, *s;
{
    register int i, var, last, len = 0;

    for var = 0; var < cube.num_binary_vars; var++ {
    // inlining of macro GETINPUT
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    s[len++] = "?01-" [((c[(((2*var) >> LOGBPI) + 1)] >> ((2*var) & (BPI-1))) & 3)];
    }
    for var = cube.num_binary_vars; var < cube.num_vars - 1; var++ {
    s[len++] = ' ';
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        s[len++] = "01" [(c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
    }
    }
    if cube.output != -1 {
    last = cube.last_part[cube.output];
    s[len++] = ' ';
    for i = cube.first_part[cube.output]; i <= last; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        s[len++] = out_map [(c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
    }
    }
    s[len] = '\0';
    return s;
}


void print_cube(fp, c, out_map)
register FILE *fp;
register pcube c;
register char *out_map;
{
    register int i, var, ch;
    int last;

    for var = 0; var < cube.num_binary_vars; var++ {
    // inlining of macro GETINPUT
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    ch = "?01-" [((c[(((2*var) >> LOGBPI) + 1)] >> ((2*var) & (BPI-1))) & 3)];
    putc(ch, fp);
    }
    for var = cube.num_binary_vars; var < cube.num_vars - 1; var++ {
    putc(' ', fp);
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        ch = "01" [(c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
        putc(ch, fp);
    }
    }
    if cube.output != -1 {
    last = cube.last_part[cube.output];
    putc(' ', fp);
    for i = cube.first_part[cube.output]; i <= last; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        ch = out_map [(c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
        putc(ch, fp);
    }
    }
    putc('\n', fp);
}


void print_expanded_cube(fp, c, phase)
register FILE *fp;
register pcube c;
pcube phase;
{
    register int i, var, ch;
    char *out_map;

    for var = 0; var < cube.num_binary_vars; var++ {
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        ch = "~1" [(c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
        putc(ch, fp);
    }
    }
    for var = cube.num_binary_vars; var < cube.num_vars - 1; var++ {
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        ch = "1~" [(c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
        putc(ch, fp);
    }
    }
    if cube.output != -1 {
    var = cube.num_vars - 1;
    putc(' ', fp);
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if phase == (pcube) NULL || (phase[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) {
        out_map = "~1";
        } else {
        out_map = "~0";
        }
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        ch = out_map[(c[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
        putc(ch, fp);
    }
    }
    putc('\n', fp);
}


char *pc1(c) pcube c;
{static char s1[256];return fmt_cube(c, "01", s1);}
char *pc2(c) pcube c;
{static char s2[256];return fmt_cube(c, "01", s2);}


void debug_print(T, name, level)
pcube *T;
var name string;
var level int32;
{
    register pcube *T1, p, temp;
    register int cnt;

    // inlining of macro CUBELISTSIZE
    cnt = (((pcube *) T[1] - T) - 3);
    // inlining of macro new_cube
    // inlining of macro set_new
    temp = set_clear(ALLOC(unsigned int, SET_SIZE(cube.size)), cube.size);
    if (verbose_debug && level == 0)
    (void) fmt.Printf("\n");
    (void) fmt.Printf("%s[%d]: ord(T)=%d\n", name, level, cnt);
    if verbose_debug {
    (void) fmt.Printf("cofactor=%s\n", pc1(T[0]));
    for(T1 = T+2, cnt = 1; (p = *T1++) != (pcube) NULL; cnt++)
        (void) fmt.Printf("%4d. %s\n", cnt, pc1(set_or(temp, p, T[0])));
    }
    // inlining of macro free_cube
    // inlining of macro set_free
    FREE(temp);
}


void debug1_print(T, name, num)
pcover T;
var name string;
var num int32;
{
    register int cnt = 1;
    register pcube p, last;

    if (verbose_debug && num == 0)
    (void) fmt.Printf("\n");
    (void) fmt.Printf("%s[%d]: ord(T)=%d\n", name, num, T.count);
    if (verbose_debug)
    // inlining of macro foreach_set
    for(p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize)
        (void) fmt.Printf("%4d. %s\n", cnt++, pc1(p));
}


void cprint(T)
pcover T;
{
    register pcube p, last;

    // inlining of macro foreach_set
    for(p=T.data,last=p+T.count*T.wsize;p<last;p+=T.wsize)
    (void) fmt.Printf("%s\n", pc1(p));
}


void makeup_labels(PLA)
var PLA pPLA;
{
    int var, i, ind;

    if (PLA.label == (char **) NULL)
    PLA_labels(PLA);

    for(var = 0; var < cube.num_vars; var++)
    for i = 0; i < cube.part_size[var]; i++ {
        ind = cube.first_part[var] + i;
        if PLA.label[ind] == (char *) NULL {
        PLA.label[ind] = ALLOC(char, 15);
        if (var < cube.num_binary_vars)
            if ((i % 2) == 0)
            (void) sprintf(PLA.label[ind], "v%d.bar", var);
            else
            (void) sprintf(PLA.label[ind], "v%d", var);
        else
            (void) sprintf(PLA.label[ind], "v%d.%d", var, i);
        }
    }
}


void kiss_output(fp, PLA)
var fp *os.File;
var PLA pPLA;
{
    register pset last, p;

    // inlining of macro foreach_set
    for p=PLA.F.data,last=p+PLA.F.count*PLA.F.wsize;p<last;p+=PLA.F.wsize {
    kiss_print_cube(fp, PLA, p, "~1");
    }
    // inlining of macro foreach_set
    for p=PLA.D.data,last=p+PLA.D.count*PLA.D.wsize;p<last;p+=PLA.D.wsize {
    kiss_print_cube(fp, PLA, p, "~2");
    }
}


void kiss_print_cube(fp, PLA, p, out_string)
var fp *os.File;
var PLA pPLA;
pcube p;
var out_string string;
{
    register int i, var;
    int part, x;

    for var = 0; var < cube.num_binary_vars; var++ {
    // inlining of macro GETINPUT
    // inlining of macro WHICH_WORD
    // inlining of macro WHICH_BIT
    x = "?01-" [((p[(((2*var) >> LOGBPI) + 1)] >> ((2*var) & (BPI-1))) & 3)];
    putc(x, fp);
    }

    for var = cube.num_binary_vars; var < cube.num_vars - 1; var++ {
    putc(' ', fp);
    if setp_implies(cube.var_mask[var], p) {
        putc('-', fp);
    } else {
        part = -1;
        for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        if p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1))) {
            if part != -1 {
            fatal("more than 1 part in a symbolic variable\n");
            }
            part = i;
        }
        }
        if part == -1 {
        putc('~', fp);    /* no parts, hope its an output ... */
        } else {
        (void) fputs(PLA.label[part], fp);
        }
    }
    }

    if (var = cube.output) != -1 {
    putc(' ', fp);
    for i = cube.first_part[var]; i <= cube.last_part[var]; i++ {
        // inlining of macro is_in_set
        // inlining of macro WHICH_WORD
        // inlining of macro WHICH_BIT
        x = out_string [(p[(((i) >> LOGBPI) + 1)] & (1 << ((i) & (BPI-1)))) != 0];
        putc(x, fp);
    }
    }

    putc('\n', fp);
}

void output_symbolic_constraints(fp, PLA, output_symbolic)
var fp *os.File;
var PLA pPLA;
var output_symbolic int32;
{
    pset_family A;
    register int i, j;
    int size, var, npermute, *permute, *weight, noweight;

    if (cube.num_vars - cube.num_binary_vars) <= 1 {
    return;
    }
    makeup_labels(PLA);

    for var=cube.num_binary_vars; var < cube.num_vars-1; var++ {

    /* pull out the columns for variable "var" */
    npermute = cube.part_size[var];
    permute = ALLOC(int, npermute);
    for i=0; i < npermute; i++ {
        permute[i] = cube.first_part[var] + i;
    }
    A = sf_permute(sf_save(PLA.F), permute, npermute);
    FREE(permute);


    /* Delete the singletons and the full sets */
    noweight = 0;
    for i = 0; i < A.count; i++ {
        // inlining of macro GETSET
        size = set_ord(((A).data + (A).wsize * (i)));
        if size == 1 || size == A.sf_size {
        sf_delset(A, i--);
        noweight++;
        }
    }


    /* Count how many times each is duplicated */
    weight = ALLOC(int, A.count);
    for i = 0; i < A.count; i++ {
        // inlining of macro RESET
        // inlining of macro GETSET
        (((A).data + (A).wsize * (i))[0] &= ~ (COVERED));
    }
    for i = 0; i < A.count; i++ {
        weight[i] = 0;
        // inlining of macro TESTP
        // inlining of macro GETSET
        if ! (((A).data + (A).wsize * (i))[0] & (COVERED)) {
        weight[i] = 1;
        for j = i+1; j < A.count; j++ {
            // inlining of macro GETSET
            if setp_equal(((A).data + (A).wsize * (i)), ((A).data + (A).wsize * (j))) {
            weight[i]++;
            // inlining of macro SET
            // inlining of macro GETSET
            (((A).data + (A).wsize * (j))[0] |= (COVERED));
            }
        }
        }
    }


    /* Print out the constraints */
    if ! (output_symbolic != 0) {
        (void) fmt.Fprintf(fp, "# Symbolic constraints for variable %d (Numeric form)\n", var);
        (void) fmt.Fprintf(fp, "# unconstrained weight = %d\n", noweight);
        (void) fmt.Fprintf(fp, "num_codes=%d\n", cube.part_size[var]);
        for i = 0; i < A.count; i++ {
        if weight[i] > 0 {
            (void) fmt.Fprintf(fp, "weight=%d: ", weight[i]);
            for j = 0; j < A.sf_size; j++ {
            // inlining of macro is_in_set
            // inlining of macro GETSET
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            if ((A).data + (A).wsize * (i))[(((j) >> LOGBPI) + 1)] & (1 << ((j) & (BPI-1))) {
                (void) fmt.Fprintf(fp, " %d", j);
            }
            }
            (void) fmt.Fprintf(fp, "\n");
        }
        }
    } else {
        (void) fmt.Fprintf(fp, "# Symbolic constraints for variable %d (Symbolic form)\n", var);
        for i = 0; i < A.count; i++ {
        if weight[i] > 0 {
            (void) fmt.Fprintf(fp, "#   w=%d: (", weight[i]);
            for j = 0; j < A.sf_size; j++ {
            // inlining of macro is_in_set
            // inlining of macro GETSET
            // inlining of macro WHICH_WORD
            // inlining of macro WHICH_BIT
            if ((A).data + (A).wsize * (i))[(((j) >> LOGBPI) + 1)] & (1 << ((j) & (BPI-1))) {
                (void) fmt.Fprintf(fp, " %s", PLA.label[cube.first_part[var]+j]);
            }
            }
            (void) fmt.Fprintf(fp, " )\n");
        }
        }
        FREE(weight);
    }
    }
}
//ABC_NAMESPACE_IMPL_END


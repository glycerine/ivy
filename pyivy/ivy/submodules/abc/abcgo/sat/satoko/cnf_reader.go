package satoko

import "fmt"

import "os"

//===--- cnf_reader.h -------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===
//#include <assert.h>
//#include <ctype.h>
//#include <stdlib.h>
//#include <stdio.h>
//#include <string.h>

//#include "satoko.h"
//#include "solver.h"
//#include "utils/mem.h"
//#include "utils/vec/vec_uint.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_IMPL_START

/** Read the file into an internal buffer.
 *
 * This function will receive a file name. The return data is a string ended
 * with '\0'.
 *
 */
func file_open(fname string) string {
    var file *os.File  = os.Open(fname);
    var buffer string;
    var sz_file int32;
    var ret int32;

    if file == nil {
        fmt.Printf("Couldn't open file: %s\n", fname);
        return nil;
    }
    fseek(file, 0, SEEK_END);
    sz_file = ftell(file);
    rewind(file);
    // inlining of macro satoko_alloc
    buffer = ((char *) malloc((sz_file + 3) * sizeof(char)));
    ret = fread(buffer, sz_file, 1, file);
    buffer[sz_file + 0] = '\n';
    buffer[sz_file + 1] = '\0';
    return buffer;
}

func skip_spaces(pos *string) {
    assert(pos != nil, "pos != nil");
    for ; isspace(**pos); (*pos)++ {}
}

func skip_line(pos *string) {
    assert(pos != nil, "pos != nil");
    for ; **pos != '\n' && **pos != '\r' && **pos != -1; (*pos)++ {}
    if **pos != -1 {
        (*pos)++; }
    return;
}

func read_int(token *string) int32 {
    var value int32;
    var neg int32;

    skip_spaces(token);
    if **token == '-' {
        neg = 1;
        (*token)++;
    } else if **token == '+' {
        (*token)++; }

    if !isdigit(**token) {
        fmt.Printf("Parsing error. Unexpected char: %c.\n", **token);
        os.Exit(int(1));
    }
    for isdigit(**token) {
        value = (value * 10) + (**token - '0');
        (*token)++;
    }
    if neg != 0 {
        return -value;
    } else {
        return value;
    }
}

func read_clause(token *string, lits *vec_uint_t) {
    var var int32;
    var sign uint32;

    vec_uint_clear(lits);
    for {
        var = read_int(token);
        if var == 0 {
            break; }
        sign = (var > 0);
        var = abs(var) - 1;
        vec_uint_push_back(lits, var2lit(uint32( var), byte(!sign)));
    }
}

/** Start the solver and reads the DIMAC file.
 *
 * Returns false upon immediate conflict.
 */
func satoko_parse_dimacs(fname string, solver **satoko_t) int32 {
    var p *satoko_t  = nil;
    var lits *vec_uint_t  = nil;
    var n_var int32;
    var n_clause int32;
    var buffer string  = file_open(fname);
    var token string;

    if buffer == nil {
        return -1; }

    token = buffer;
    for {
        skip_spaces(&token);
        if *token == 0 {
            break;
        } else if *token == 'c' {
            skip_line(&token);
        } else if *token == 'p' {
            token++;
            skip_spaces(&token);
            for ; !isspace(*token); token++ {} /* skip 'cnf' */

            n_var = read_int(&token);
            n_clause = read_int(&token);
            skip_line(&token);
            lits = vec_uint_alloc(uint32( n_var));
            p = satoko_create();
        } else {
            if lits == nil {
                fmt.Printf("There is no parameter line.\n");
                // inlining of macro satoko_free
                for { free(buffer); buffer = nil; 
if !(false) { break }
} 
                return -1;
            }
            read_clause(&token, lits);
            if !satoko_add_clause(p, (int*)vec_uint_data(lits), vec_uint_size(lits)) {
                fmt.Printf("Unable to add clause: ");
                vec_uint_print(lits);
                return SATOKO_ERR;
            }
            }
    }
    vec_uint_free(lits);
    // inlining of macro satoko_free
    for { free(buffer); buffer = nil; 
if !(false) { break }
} 
    *solver = p;
    return SATOKO_OK;
}

//ABC_NAMESPACE_IMPL_END

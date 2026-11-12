package bsat2

import (
	"fmt"
	"os"
	"strings"
)

/***************************************************************************************[Options.h]
Copyright (c) 2008-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/




//#include <stdlib.h>
//#include <stdio.h>
//#include <math.h>
//#include <string.h>

//#include "IntTypes.h"
//#include "Vec.h"
//#include "ParseUtils.h"

namespace Minisat {

//==================================================================================================
// Top-level option parse/help functions:


extern int parseOptions     (int& argc, char** argv, bool strict = false);
extern int printUsageAndExit(int  argc, char** argv, bool verbose = false);
extern void setUsageHelp     (char* str);
extern void setHelpPrefixStr (char* str);


//==================================================================================================
// Options is an abstract class that gives the interface for all types options:


class Option
{
 public:
    char* name;
    char* description;
    char* category;
    char* type_name;

    static vec<Option*>& getOptionList () { static vec<Option*> options; return options; }
    static char*&  getUsageString() { static char* usage_str; return usage_str; }
    static char*&  getHelpPrefixString() { static char* help_prefix_str = ""; return help_prefix_str; }

    struct OptionLt {
        bool operator()(Option* x, Option* y) {
            int test1 = int32(strings.Compare(x.category, y.category));
            return test1 < 0 || test1 == 0 && int32(strings.Compare(x.type_name, y.type_name)) < 0;
        }
    };

    Option(char* name_, 
           char* desc_,
           char* cate_,
           char* type_) : 
      name       (name_)
    , description(desc_)
    , category   (cate_)
    , type_name  (type_)
    { 
        getOptionList().push(this); 
    }

 public:
    virtual ~Option() {}

    virtual bool parse             (char* str)      = 0;
    virtual void help              (bool verbose = false) = 0;

    friend  int parseOptions      (int& argc, char** argv, bool strict);
    friend  int printUsageAndExit (int  argc, char** argv, bool verbose);
    friend  void setUsageHelp      (char* str);
    friend  void setHelpPrefixStr  (char* str);
};


//==================================================================================================
// Range classes with specialization for floating types:


struct IntRange {
    int begin;
    int end;
    IntRange(int b, int e) : begin(b), end(e) {}
};

struct Int64Range {
    int64_t begin;
    int64_t end;
    Int64Range(int64_t b, int64_t e) : begin(b), end(e) {}
};

struct DoubleRange {
    double begin;
    double end;
    bool  begin_inclusive;
    bool  end_inclusive;
    DoubleRange(double b, bool binc, double e, bool einc) : begin(b), end(e), begin_inclusive(binc), end_inclusive(einc) {}
};


//==================================================================================================
// Double options:


class DoubleOption : public Option
{
 protected:
    DoubleRange range;
    double      value;

 public:
    DoubleOption(char* c, char* n, char* d, double def = double(), DoubleRange r = DoubleRange(-HUGE_VAL, false, HUGE_VAL, false))
        : Option(n, d, c, "<double>"), range(r), value(def) {
        // FIXME: set LC_NUMERIC to "C" to make sure that strtof/strtod parses decimal point correctly.
    }

    operator      double   (void) { return value; }
    operator      double&  (void)       { return value; }
    DoubleOption& operator=(double x)   { value = x; return *this; }

    virtual bool parse(char* str){
        char* span = str; 

        if (!match(span, "-") || !match(span, name) || !match(span, "="))
            return false;

        char*  end;
        double tmp = strtod(span, &end);

        if (end == NULL) 
            return false;
        else if tmp >= range.end && (!range.end_inclusive || tmp != range.end){
            fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too large for option \"%s\".\n", span, name);
            os.Exit(int(1));
        }else if tmp <= range.begin && (!range.begin_inclusive || tmp != range.begin){
            fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too small for option \"%s\".\n", span, name);
            os.Exit(int(1)); }

        value = tmp;
        // fprintf(stderr, "READ VALUE: %g\n", value);

        return true;
    }

    virtual void help (bool verbose = false){
        fmt.Fprintf(os.Stderr, "  -%-12s = %-8s %c%4.2g .. %4.2g%c (default: %g)\n", name, type_name, range.begin_inclusive ? '[' : '(', range.begin, range.end, range.end_inclusive ? ']' : ')', value);
        if verbose{
            fmt.Fprintf(os.Stderr, "\n        %s\n", description);
            fmt.Fprintf(os.Stderr, "\n");
        }
    }
};


//==================================================================================================
// Int options:


class IntOption : public Option
{
 protected:
    IntRange range;
    int32_t  value;

 public:
    IntOption(char* c, char* n, char* d, int32_t def = int32_t(), IntRange r = IntRange(INT32_MIN, INT32_MAX))
        : Option(n, d, c, "<int32>"), range(r), value(def) {}
 
    operator   int32_t   (void) { return value; }
    operator   int32_t&  (void)       { return value; }
    IntOption& operator= (int32_t x)  { value = x; return *this; }

    virtual bool parse(char* str){
        char* span = str; 

        if (!match(span, "-") || !match(span, name) || !match(span, "="))
            return false;

        char*   end;
        int32_t tmp = strtol(span, &end, 10);

        if (end == NULL) 
            return false;
        else if tmp > range.end{
            fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too large for option \"%s\".\n", span, name);
            os.Exit(int(1));
        }else if tmp < range.begin{
            fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too small for option \"%s\".\n", span, name);
            os.Exit(int(1)); }

        value = tmp;

        return true;
    }

    virtual void help (bool verbose = false){
        fmt.Fprintf(os.Stderr, "  -%-12s = %-8s [", name, type_name);
        if (range.begin == INT32_MIN)
            fmt.Fprintf(os.Stderr, "imin");
        else
            fmt.Fprintf(os.Stderr, "%4d", range.begin);

        fmt.Fprintf(os.Stderr, " .. ");
        if (range.end == INT32_MAX)
            fmt.Fprintf(os.Stderr, "imax");
        else
            fmt.Fprintf(os.Stderr, "%4d", range.end);

        fmt.Fprintf(os.Stderr, "] (default: %d)\n", value);
        if verbose{
            fmt.Fprintf(os.Stderr, "\n        %s\n", description);
            fmt.Fprintf(os.Stderr, "\n");
        }
    }
};

/*
// Leave this out for visual C++ until Microsoft implements C99 and gets support for strtoll.
#ifndef _MSC_VER

class Int64Option : public Option
{
 protected:
    Int64Range range;
    int64_t  value;

 public:
    Int64Option(const char* c, const char* n, const char* d, int64_t def = int64_t(), Int64Range r = Int64Range(INT64_MIN, INT64_MAX))
        : Option(n, d, c, "<int64>"), range(r), value(def) {}
 
    operator     int64_t   (void) const { return value; }
    operator     int64_t&  (void)       { return value; }
    Int64Option& operator= (int64_t x)  { value = x; return *this; }

    virtual bool parse(const char* str){
        const char* span = str; 

        if (!match(span, "-") || !match(span, name) || !match(span, "="))
            return false;

        char*   end;
        int64_t tmp = strtoll(span, &end, 10);

        if (end == NULL) 
            return false;
        else if (tmp > range.end){
            fprintf(stderr, "ERROR! value <%s> is too large for option \"%s\".\n", span, name);
            exit(1);
        }else if (tmp < range.begin){
            fprintf(stderr, "ERROR! value <%s> is too small for option \"%s\".\n", span, name);
            exit(1); }

        value = tmp;

        return true;
    }

    virtual void help (bool verbose = false){
        fprintf(stderr, "  -%-12s = %-8s [", name, type_name);
        if (range.begin == INT64_MIN)
            fprintf(stderr, "imin");
        else
            fprintf(stderr, "%4"PRIi64, range.begin);

        fprintf(stderr, " .. ");
        if (range.end == INT64_MAX)
            fprintf(stderr, "imax");
        else
            fprintf(stderr, "%4"PRIi64, range.end);

        fprintf(stderr, "] (default: %"PRIi64")\n", value);
        if (verbose){
            fprintf(stderr, "\n        %s\n", description);
            fprintf(stderr, "\n");
        }
    }
};
#endif
*/

//==================================================================================================
// String option:


class StringOption : public Option
{
    char* value;
 public:
    StringOption(char* c, char* n, char* d, char* def = NULL) 
        : Option(n, d, c, "<string>"), value(def) {}

    operator      char*  (void) { return value; }
    operator      char*& (void)           { return value; }
    StringOption& operator=    (char* x)  { value = x; return *this; }

    virtual bool parse(char* str){
        char* span = str; 

        if (!match(span, "-") || !match(span, name) || !match(span, "="))
            return false;

        value = span;
        return true;
    }

    virtual void help (bool verbose = false){
        fmt.Fprintf(os.Stderr, "  -%-10s = %8s\n", name, type_name);
        if verbose{
            fmt.Fprintf(os.Stderr, "\n        %s\n", description);
            fmt.Fprintf(os.Stderr, "\n");
        }
    }    
};


//==================================================================================================
// Bool option:


class BoolOption : public Option
{
    bool value;

 public:
    BoolOption(char* c, char* n, char* d, bool v) 
        : Option(n, d, c, "<bool>"), value(v) {}

    operator    bool     (void) { return value; }
    operator    bool&    (void)       { return value; }
    BoolOption& operator=(bool b)     { value = b; return *this; }

    virtual bool parse(char* str){
        char* span = str; 
        
        if match(span, "-"){
            bool b = !match(span, "no-");

            if int32(strings.Compare(span, name)) == 0{
                value = b;
                return true; }
        }

        return false;
    }

    virtual void help (bool verbose = false){

        fmt.Fprintf(os.Stderr, "  -%s, -no-%s", name, name);

        for (uint32_t i = 0; i < 32 - len(name)*2; i++)
            fmt.Fprintf(os.Stderr, " ");

        fmt.Fprintf(os.Stderr, " ");
        fmt.Fprintf(os.Stderr, "(default: %s)\n", value ? "on" : "off");
        if verbose{
            fmt.Fprintf(os.Stderr, "\n        %s\n", description);
            fmt.Fprintf(os.Stderr, "\n");
        }
    }
};

//=================================================================================================
}



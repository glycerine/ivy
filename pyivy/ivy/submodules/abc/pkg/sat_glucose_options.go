package abc

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Port of src/sat/glucose/Options.h and Options.cpp (namespace Gluco).
// See sat_bsat2_vec.go for the C++ -> Go conventions.

// GlucoOptionI is the virtual interface of class Option: parse and help are
// pure virtual; Base gives access to the common Option part of the object so that
// vec<Option*> (Option::getOptionList) can hold any option subclass.
type GlucoOptionI interface {
	Base() *GlucoOption
	Parse(str string) bool
	Help(verbose bool)
}

// GlucoOption is Gluco::Option, the abstract base of all command-line options.
type GlucoOption struct {
	Name        string // const char* name
	Description string // const char* description
	Category    string // const char* category
	Type_name   string // const char* type_name
}

var (
	glucoOptionList       GlucoVec[GlucoOptionI] // function-local static of getOptionList
	glucoOptionUsage      string                 // function-local static of getUsageString (NULL == "")
	glucoOptionUsageSet   bool
	glucoOptionHelpPrefix string // function-local static of getHelpPrefixString, initially ""
)

// GlucoOptionGetOptionList is Option::getOptionList (static; holds a function-local static vec<Option*>).
// C: static vec<Option*>& getOptionList()
func GlucoOptionGetOptionList() *GlucoVec[GlucoOptionI] { return &glucoOptionList }

// GlucoOptionGetUsageString is Option::getUsageString (static; function-local static const char*).
// C: static const char*& getUsageString()
func GlucoOptionGetUsageString() *string { return &glucoOptionUsage }

// GlucoOptionGetHelpPrefixString is Option::getHelpPrefixString (static; function-local static, initially "").
// C: static const char*& getHelpPrefixString()
func GlucoOptionGetHelpPrefixString() *string { return &glucoOptionHelpPrefix }

// GlucoOptionLt is Option::OptionLt: orders options by category then type_name.
type GlucoOptionLt struct{}

// Call is OptionLt::operator().
// C: bool operator()(const Option* x, const Option* y)
func (l GlucoOptionLt) Call(x GlucoOptionI, y GlucoOptionI) bool {
	test1 := strings.Compare(x.Base().Category, y.Base().Category)
	return test1 < 0 || test1 == 0 && strings.Compare(x.Base().Type_name, y.Base().Type_name) < 0
}

// NewGlucoOption is the protected constructor Option(name_, desc_, cate_, type_).
// The C++ constructor also registers `this` in getOptionList(); the Go
// constructors of the concrete option classes do that registration, because only
// they hold the derived object (see registerGlucoOption).
// C: Option(const char* name_, const char* desc_, const char* cate_, const char* type_)
func NewGlucoOption(name_ string, desc_ string, cate_ string, type_ string) *GlucoOption {
	return &GlucoOption{Name: name_, Description: desc_, Category: cate_, Type_name: type_}
}

// registerGlucoOption is the `getOptionList().push(this)` of the Option constructor.
func registerGlucoOption(o GlucoOptionI) { glucoOptionList.PushElem(o) }

// Destroy is Option's virtual destructor.
// C: virtual ~Option()
func (o *GlucoOption) Destroy() {}

// GlucoIntRange is Gluco::IntRange.
type GlucoIntRange struct {
	Begin int
	End   int
}

// NewGlucoIntRange is IntRange(int b, int e).
// C: IntRange(int b, int e)
func NewGlucoIntRange(b int, e int) GlucoIntRange { return GlucoIntRange{Begin: b, End: e} }

// GlucoInt64Range is Gluco::Int64Range.
type GlucoInt64Range struct {
	Begin int64
	End   int64
}

// NewGlucoInt64Range is Int64Range(int64_t b, int64_t e).
// C: Int64Range(int64_t b, int64_t e)
func NewGlucoInt64Range(b int64, e int64) GlucoInt64Range {
	return GlucoInt64Range{Begin: b, End: e}
}

// GlucoDoubleRange is Gluco::DoubleRange.
type GlucoDoubleRange struct {
	Begin           float64
	End             float64
	Begin_inclusive bool
	End_inclusive   bool
}

// NewGlucoDoubleRange is DoubleRange(double b, bool binc, double e, bool einc).
// C: DoubleRange(double b, bool binc, double e, bool einc)
func NewGlucoDoubleRange(b float64, binc bool, e float64, einc bool) GlucoDoubleRange {
	return GlucoDoubleRange{Begin: b, End: e, Begin_inclusive: binc, End_inclusive: einc}
}

// glucoStrtod is C strtod on the prefix of s that forms a number (0 if none).
func glucoStrtod(s string) float64 {
	i := 0
	for i < len(s) && (s[i] == ' ' || (s[i] >= 9 && s[i] <= 13)) {
		i++
	}
	start := i
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	lower := strings.ToLower(s[i:])
	if strings.HasPrefix(lower, "infinity") {
		return glucoSignedInf(s[start:i])
	}
	if strings.HasPrefix(lower, "inf") {
		return glucoSignedInf(s[start:i])
	}
	if strings.HasPrefix(lower, "nan") {
		return math.NaN()
	}
	digits := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		digits++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
			digits++
		}
	}
	if digits == 0 {
		return 0
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		if j < len(s) && s[j] >= '0' && s[j] <= '9' {
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			i = j
		}
	}
	v, err := strconv.ParseFloat(s[start:i], 64)
	if err != nil && !math.IsInf(v, 0) {
		return 0
	}
	return v
}

func glucoSignedInf(sign string) float64 {
	if sign == "-" {
		return math.Inf(-1)
	}
	return math.Inf(1)
}

// glucoStrtol is C strtol(s, &end, base 10) on a 64-bit long: the prefix of s
// that forms a number, saturating at the long limits (0 if none).
func glucoStrtol(s string) int64 {
	i := 0
	for i < len(s) && (s[i] == ' ' || (s[i] >= 9 && s[i] <= 13)) {
		i++
	}
	start := i
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	d := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == d {
		return 0
	}
	v, err := strconv.ParseInt(s[start:i], 10, 64)
	if err != nil {
		if strings.HasPrefix(s[start:], "-") {
			return math.MinInt64
		}
		return math.MaxInt64
	}
	return v
}

// glucoHeaderMatch is the common start of every parse(): match "-", name, "=".
func glucoHeaderMatch(str string, name string) (*GlucoCharPtr, bool) {
	span := NewGlucoCharPtr(str)
	if !glucoMatch(span, "-") || !glucoMatch(span, name) || !glucoMatch(span, "=") {
		return span, false
	}
	return span, true
}

func glucoHelpVerbose(o *GlucoOption, verbose bool) {
	if verbose {
		fmt.Fprintf(os.Stderr, "\n        %s\n", o.Description)
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// GlucoDoubleOption is Gluco::DoubleOption : public Option.
type GlucoDoubleOption struct {
	GlucoOption
	range_ GlucoDoubleRange // protected: range
	value  float64          // protected
}

// NewGlucoDoubleOption is the constructor (def defaults to double(), r to DoubleRange(-HUGE_VAL, false, HUGE_VAL, false)).
// C: DoubleOption(const char* c, const char* n, const char* d, double def = double(), DoubleRange r = DoubleRange(-HUGE_VAL, false, HUGE_VAL, false))
func NewGlucoDoubleOption(c string, n string, d string, def float64, r GlucoDoubleRange) *GlucoDoubleOption {
	o := &GlucoDoubleOption{GlucoOption: *NewGlucoOption(n, d, c, "<double>"), range_: r, value: def}
	registerGlucoOption(o)
	return o
}

// Value is `operator double() const`.
// C: operator double(void) const
func (o *GlucoDoubleOption) Value() float64 { return o.value }

// ValueRef is `operator double&()`.
// C: operator double&(void)
func (o *GlucoDoubleOption) ValueRef() *float64 { return &o.value }

// Assign is operator=(double).
// C: DoubleOption& operator=(double x)
func (o *GlucoDoubleOption) Assign(x float64) *GlucoDoubleOption {
	o.value = x
	return o
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoDoubleOption) Base() *GlucoOption { return &o.GlucoOption }

// Parse is DoubleOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoDoubleOption) Parse(str string) bool {
	span, ok := glucoHeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := glucoStrtod(span.Rest())
	if tmp >= o.range_.End && (!o.range_.End_inclusive || tmp != o.range_.End) {
		fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too large for option \"%s\".\n", span.Rest(), o.Name)
		os.Exit(1)
	} else if tmp <= o.range_.Begin && (!o.range_.Begin_inclusive || tmp != o.range_.Begin) {
		fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too small for option \"%s\".\n", span.Rest(), o.Name)
		os.Exit(1)
	}
	o.value = tmp
	return true
}

// Help is DoubleOption::help (virtual; verbose defaults to false).
// C: virtual void help(bool verbose = false)
func (o *GlucoDoubleOption) Help(verbose bool) {
	bl, el := ')', ')'
	bl = '('
	if o.range_.Begin_inclusive {
		bl = '['
	}
	if o.range_.End_inclusive {
		el = ']'
	}
	fmt.Fprintf(os.Stderr, "  -%-12s = %-8s %c%4.2g .. %4.2g%c (default: %g)\n",
		o.Name, o.Type_name, bl, o.range_.Begin, o.range_.End, el, o.value)
	glucoHelpVerbose(&o.GlucoOption, verbose)
}

// GlucoIntOption is Gluco::IntOption : public Option.
type GlucoIntOption struct {
	GlucoOption
	range_ GlucoIntRange // protected: range
	value  int32         // protected
}

// NewGlucoIntOption is the constructor (def defaults to int32_t(), r to IntRange(INT32_MIN, INT32_MAX)).
// C: IntOption(const char* c, const char* n, const char* d, int32_t def = int32_t(), IntRange r = IntRange(INT32_MIN, INT32_MAX))
func NewGlucoIntOption(c string, n string, d string, def int32, r GlucoIntRange) *GlucoIntOption {
	o := &GlucoIntOption{GlucoOption: *NewGlucoOption(n, d, c, "<int32>"), range_: r, value: def}
	registerGlucoOption(o)
	return o
}

// Value is `operator int32_t() const`.
// C: operator int32_t(void) const
func (o *GlucoIntOption) Value() int32 { return o.value }

// ValueRef is `operator int32_t&()`.
// C: operator int32_t&(void)
func (o *GlucoIntOption) ValueRef() *int32 { return &o.value }

// Assign is operator=(int32_t).
// C: IntOption& operator=(int32_t x)
func (o *GlucoIntOption) Assign(x int32) *GlucoIntOption {
	o.value = x
	return o
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoIntOption) Base() *GlucoOption { return &o.GlucoOption }

// Parse is IntOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoIntOption) Parse(str string) bool {
	span, ok := glucoHeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := int32(glucoStrtol(span.Rest())) // int32_t tmp = strtol(...) truncates the long
	if int(tmp) > o.range_.End {
		fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too large for option \"%s\".\n", span.Rest(), o.Name)
		os.Exit(1)
	} else if int(tmp) < o.range_.Begin {
		fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too small for option \"%s\".\n", span.Rest(), o.Name)
		os.Exit(1)
	}
	o.value = tmp
	return true
}

// Help is IntOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *GlucoIntOption) Help(verbose bool) {
	fmt.Fprintf(os.Stderr, "  -%-12s = %-8s [", o.Name, o.Type_name)
	if o.range_.Begin == math.MinInt32 {
		fmt.Fprintf(os.Stderr, "imin")
	} else {
		fmt.Fprintf(os.Stderr, "%4d", o.range_.Begin)
	}
	fmt.Fprintf(os.Stderr, " .. ")
	if o.range_.End == math.MaxInt32 {
		fmt.Fprintf(os.Stderr, "imax")
	} else {
		fmt.Fprintf(os.Stderr, "%4d", o.range_.End)
	}
	fmt.Fprintf(os.Stderr, "] (default: %d)\n", o.value)
	glucoHelpVerbose(&o.GlucoOption, verbose)
}

// GlucoStringOption is Gluco::StringOption : public Option.
type GlucoStringOption struct {
	GlucoOption
	value string // const char* value (NULL default is the empty string)
}

// NewGlucoStringOption is the constructor (def defaults to NULL).
// C: StringOption(const char* c, const char* n, const char* d, const char* def = NULL)
func NewGlucoStringOption(c string, n string, d string, def string) *GlucoStringOption {
	o := &GlucoStringOption{GlucoOption: *NewGlucoOption(n, d, c, "<string>"), value: def}
	registerGlucoOption(o)
	return o
}

// Value is `operator const char*() const`.
// C: operator const char*(void) const
func (o *GlucoStringOption) Value() string { return o.value }

// ValueRef is `operator const char*&()`.
// C: operator const char*&(void)
func (o *GlucoStringOption) ValueRef() *string { return &o.value }

// Assign is operator=(const char*).
// C: StringOption& operator=(const char* x)
func (o *GlucoStringOption) Assign(x string) *GlucoStringOption {
	o.value = x
	return o
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoStringOption) Base() *GlucoOption { return &o.GlucoOption }

// Parse is StringOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoStringOption) Parse(str string) bool {
	span, ok := glucoHeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	o.value = span.Rest()
	return true
}

// Help is StringOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *GlucoStringOption) Help(verbose bool) {
	fmt.Fprintf(os.Stderr, "  -%-10s = %8s\n", o.Name, o.Type_name)
	glucoHelpVerbose(&o.GlucoOption, verbose)
}

// GlucoBoolOption is Gluco::BoolOption : public Option.
type GlucoBoolOption struct {
	GlucoOption
	value bool
}

// NewGlucoBoolOption is the constructor.
// C: BoolOption(const char* c, const char* n, const char* d, bool v)
func NewGlucoBoolOption(c string, n string, d string, v bool) *GlucoBoolOption {
	o := &GlucoBoolOption{GlucoOption: *NewGlucoOption(n, d, c, "<bool>"), value: v}
	registerGlucoOption(o)
	return o
}

// Value is `operator bool() const`.
// C: operator bool(void) const
func (o *GlucoBoolOption) Value() bool { return o.value }

// ValueRef is `operator bool&()`.
// C: operator bool&(void)
func (o *GlucoBoolOption) ValueRef() *bool { return &o.value }

// Assign is operator=(bool).
// C: BoolOption& operator=(bool b)
func (o *GlucoBoolOption) Assign(b bool) *GlucoBoolOption {
	o.value = b
	return o
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoBoolOption) Base() *GlucoOption { return &o.GlucoOption }

// Parse is BoolOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoBoolOption) Parse(str string) bool {
	span := NewGlucoCharPtr(str)
	if glucoMatch(span, "-") {
		b := !glucoMatch(span, "no-")
		if span.Rest() == o.Name {
			o.value = b
			return true
		}
	}
	return false
}

// Help is BoolOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *GlucoBoolOption) Help(verbose bool) {
	fmt.Fprintf(os.Stderr, "  -%s, -no-%s", o.Name, o.Name)
	for i := uint32(0); i < 32-uint32(len(o.Name))*2; i++ {
		fmt.Fprintf(os.Stderr, " ")
	}
	fmt.Fprintf(os.Stderr, " ")
	d := "off"
	if o.value {
		d = "on"
	}
	fmt.Fprintf(os.Stderr, "(default: %s)\n", d)
	glucoHelpVerbose(&o.GlucoOption, verbose)
}

// GlucoParseOptions is Gluco::parseOptions (returns void in this version); consumed
// options are removed from argv and argc. An unknown flag in strict mode exits with status 1.
// C: void Gluco::parseOptions(int& argc, char** argv, bool strict)
func GlucoParseOptions(argc *int, argv []string, strict bool) {
	var i, j int
	for i, j = 1, 1; i < *argc; i++ {
		str := NewGlucoCharPtr(argv[i])
		if glucoMatch(str, "--") && glucoMatch(str, glucoOptionHelpPrefix) && glucoMatch(str, "help") {
			if str.Deref() == 0 {
				GlucoPrintUsageAndExit(*argc, argv, false)
			} else if glucoMatch(str, "-verb") {
				GlucoPrintUsageAndExit(*argc, argv, true)
			}
		} else {
			parsedOk := false
			for k := 0; !parsedOk && k < glucoOptionList.Size(); k++ {
				parsedOk = (*glucoOptionList.At(k)).Parse(argv[i])
			}
			if !parsedOk {
				cur := NewGlucoCharPtr(argv[i])
				if strict && glucoMatch(cur, "-") {
					// match() advanced argv[i] past the '-', which is what the message prints.
					fmt.Fprintf(os.Stderr, "ERROR! Unknown flag \"%s\". Use '--%shelp' for help.\n", cur.Rest(), glucoOptionHelpPrefix)
					os.Exit(1)
				} else {
					argv[j] = argv[i]
					j++
				}
			}
		}
	}
	*argc -= i - j
}

// GlucoSetUsageHelp is Gluco::setUsageHelp.
// C: void Gluco::setUsageHelp(const char* str)
func GlucoSetUsageHelp(str string) {
	glucoOptionUsage = str
	glucoOptionUsageSet = true
}

// GlucoSetHelpPrefixStr is Gluco::setHelpPrefixStr.
// C: void Gluco::setHelpPrefixStr(const char* str)
func GlucoSetHelpPrefixStr(str string) { glucoOptionHelpPrefix = str }

// GlucoPrintUsageAndExit is Gluco::printUsageAndExit (verbose defaults to false); prints the
// usage and exits with status 0.
// C: void Gluco::printUsageAndExit(int argc, char** argv, bool verbose)
func GlucoPrintUsageAndExit(argc int, argv []string, verbose bool) {
	if glucoOptionUsageSet {
		fmt.Fprintf(os.Stderr, strings.Replace(glucoOptionUsage, "%s", "%[1]s", -1), argv[0])
	}
	GlucoSortVec(&glucoOptionList, GlucoOptionLt{})
	prevCat := ""
	prevType := ""
	first := true
	for i := 0; i < glucoOptionList.Size(); i++ {
		o := (*glucoOptionList.At(i)).Base()
		// The C code compares the const char* pointers: equal strings of one
		// category share a pointer because they come from the same literal.
		if first || o.Category != prevCat {
			fmt.Fprintf(os.Stderr, "\n%s OPTIONS:\n\n", o.Category)
		} else if o.Type_name != prevType {
			fmt.Fprintf(os.Stderr, "\n")
		}
		(*glucoOptionList.At(i)).Help(verbose)
		prevCat = o.Category
		prevType = o.Type_name
		first = false
	}
	fmt.Fprintf(os.Stderr, "\nHELP OPTIONS:\n\n")
	fmt.Fprintf(os.Stderr, "  --%shelp        Print help message.\n", glucoOptionHelpPrefix)
	fmt.Fprintf(os.Stderr, "  --%shelp-verb   Print verbose help message.\n", glucoOptionHelpPrefix)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(0)
}

// GlucoInt64Option is Gluco::Int64Option : public Option (inside `#ifndef _MSC_VER`).
type GlucoInt64Option struct {
	GlucoOption
	range_ GlucoInt64Range // protected: range
	value  int64           // protected
}

// NewGlucoInt64Option is the constructor (def defaults to int64_t(), r to Int64Range(INT64_MIN, INT64_MAX)).
// C: Int64Option(const char* c, const char* n, const char* d, int64_t def = int64_t(), Int64Range r = Int64Range(INT64_MIN, INT64_MAX))
func NewGlucoInt64Option(c string, n string, d string, def int64, r GlucoInt64Range) *GlucoInt64Option {
	o := &GlucoInt64Option{GlucoOption: *NewGlucoOption(n, d, c, "<int64>"), range_: r, value: def}
	registerGlucoOption(o)
	return o
}

// Value is `operator int64_t() const`.
// C: operator int64_t(void) const
func (o *GlucoInt64Option) Value() int64 { return o.value }

// ValueRef is `operator int64_t&()`.
// C: operator int64_t&(void)
func (o *GlucoInt64Option) ValueRef() *int64 { return &o.value }

// Assign is operator=(int64_t).
// C: Int64Option& operator=(int64_t x)
func (o *GlucoInt64Option) Assign(x int64) *GlucoInt64Option {
	o.value = x
	return o
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoInt64Option) Base() *GlucoOption { return &o.GlucoOption }

// Parse is Int64Option::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoInt64Option) Parse(str string) bool {
	span, ok := glucoHeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := glucoStrtol(span.Rest()) // strtoll
	if tmp > o.range_.End {
		fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too large for option \"%s\".\n", span.Rest(), o.Name)
		os.Exit(1)
	} else if tmp < o.range_.Begin {
		fmt.Fprintf(os.Stderr, "ERROR! value <%s> is too small for option \"%s\".\n", span.Rest(), o.Name)
		os.Exit(1)
	}
	o.value = tmp
	return true
}

// Help is Int64Option::help (virtual); the range ends are printed as (int).
// C: virtual void help(bool verbose = false)
func (o *GlucoInt64Option) Help(verbose bool) {
	fmt.Fprintf(os.Stderr, "  -%-12s = %-8s [", o.Name, o.Type_name)
	if o.range_.Begin == math.MinInt64 {
		fmt.Fprintf(os.Stderr, "imin")
	} else {
		fmt.Fprintf(os.Stderr, "%4d", int32(o.range_.Begin))
	}
	fmt.Fprintf(os.Stderr, " .. ")
	if o.range_.End == math.MaxInt64 {
		fmt.Fprintf(os.Stderr, "imax")
	} else {
		fmt.Fprintf(os.Stderr, "%4d", int32(o.range_.End))
	}
	fmt.Fprintf(os.Stderr, "] (default: %d)\n", int32(o.value))
	glucoHelpVerbose(&o.GlucoOption, verbose)
}

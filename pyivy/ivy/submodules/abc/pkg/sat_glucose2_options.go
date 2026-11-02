package abc

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Port of src/sat/glucose2/Options.h and Options.cpp (namespace Gluco2).
// See sat_bsat2_vec.go for the C++ -> Go conventions.

// Gluco2OptionI is the virtual interface of class Option: parse and help are
// pure virtual; Base gives access to the common Option part of the object so that
// vec<Option*> (Option::getOptionList) can hold any option subclass.
type Gluco2OptionI interface {
	Base() *Gluco2Option
	Parse(str string) bool
	Help(verbose bool)
}

// Gluco2Option is Gluco2::Option, the abstract base of all command-line options.
type Gluco2Option struct {
	Name        string // const char* name
	Description string // const char* description
	Category    string // const char* category
	Type_name   string // const char* type_name
}

var (
	gluco2OptionList       Gluco2Vec[Gluco2OptionI] // function-local static of getOptionList
	gluco2OptionUsage      string                   // function-local static of getUsageString (NULL == "")
	gluco2OptionUsageSet   bool
	gluco2OptionHelpPrefix string // function-local static of getHelpPrefixString, initially ""
)

// Gluco2OptionGetOptionList is Option::getOptionList (static; holds a function-local static vec<Option*>).
// C: static vec<Option*>& getOptionList()
func Gluco2OptionGetOptionList() *Gluco2Vec[Gluco2OptionI] { return &gluco2OptionList }

// Gluco2OptionGetUsageString is Option::getUsageString (static; function-local static const char*).
// C: static const char*& getUsageString()
func Gluco2OptionGetUsageString() *string { return &gluco2OptionUsage }

// Gluco2OptionGetHelpPrefixString is Option::getHelpPrefixString (static; function-local static, initially "").
// C: static const char*& getHelpPrefixString()
func Gluco2OptionGetHelpPrefixString() *string { return &gluco2OptionHelpPrefix }

// Gluco2OptionLt is Option::OptionLt: orders options by category then type_name.
type Gluco2OptionLt struct{}

// Call is OptionLt::operator().
// C: bool operator()(const Option* x, const Option* y)
func (l Gluco2OptionLt) Call(x Gluco2OptionI, y Gluco2OptionI) bool {
	test1 := strings.Compare(x.Base().Category, y.Base().Category)
	return test1 < 0 || test1 == 0 && strings.Compare(x.Base().Type_name, y.Base().Type_name) < 0
}

// NewGluco2Option is the protected constructor Option(name_, desc_, cate_, type_).
// The C++ constructor also registers `this` in getOptionList(); the Go
// constructors of the concrete option classes do that registration, because only
// they hold the derived object (see registerGluco2Option).
// C: Option(const char* name_, const char* desc_, const char* cate_, const char* type_)
func NewGluco2Option(name_ string, desc_ string, cate_ string, type_ string) *Gluco2Option {
	return &Gluco2Option{Name: name_, Description: desc_, Category: cate_, Type_name: type_}
}

// registerGluco2Option is the `getOptionList().push(this)` of the Option constructor.
func registerGluco2Option(o Gluco2OptionI) { gluco2OptionList.PushElem(o) }

// Destroy is Option's virtual destructor.
// C: virtual ~Option()
func (o *Gluco2Option) Destroy() {}

// Gluco2IntRange is Gluco2::IntRange.
type Gluco2IntRange struct {
	Begin int
	End   int
}

// NewGluco2IntRange is IntRange(int b, int e).
// C: IntRange(int b, int e)
func NewGluco2IntRange(b int, e int) Gluco2IntRange { return Gluco2IntRange{Begin: b, End: e} }

// Gluco2Int64Range is Gluco2::Int64Range.
type Gluco2Int64Range struct {
	Begin int64
	End   int64
}

// NewGluco2Int64Range is Int64Range(int64_t b, int64_t e).
// C: Int64Range(int64_t b, int64_t e)
func NewGluco2Int64Range(b int64, e int64) Gluco2Int64Range {
	return Gluco2Int64Range{Begin: b, End: e}
}

// Gluco2DoubleRange is Gluco2::DoubleRange.
type Gluco2DoubleRange struct {
	Begin           float64
	End             float64
	Begin_inclusive bool
	End_inclusive   bool
}

// NewGluco2DoubleRange is DoubleRange(double b, bool binc, double e, bool einc).
// C: DoubleRange(double b, bool binc, double e, bool einc)
func NewGluco2DoubleRange(b float64, binc bool, e float64, einc bool) Gluco2DoubleRange {
	return Gluco2DoubleRange{Begin: b, End: e, Begin_inclusive: binc, End_inclusive: einc}
}

// gluco2Strtod is C strtod on the prefix of s that forms a number (0 if none).
func gluco2Strtod(s string) float64 {
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
		return gluco2SignedInf(s[start:i])
	}
	if strings.HasPrefix(lower, "inf") {
		return gluco2SignedInf(s[start:i])
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

func gluco2SignedInf(sign string) float64 {
	if sign == "-" {
		return math.Inf(-1)
	}
	return math.Inf(1)
}

// gluco2Strtol is C strtol(s, &end, base 10) on a 64-bit long: the prefix of s
// that forms a number, saturating at the long limits (0 if none).
func gluco2Strtol(s string) int64 {
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

// gluco2HeaderMatch is the common start of every parse(): match "-", name, "=".
func gluco2HeaderMatch(str string, name string) (*Gluco2CharPtr, bool) {
	span := NewGluco2CharPtr(str)
	if !gluco2Match(span, "-") || !gluco2Match(span, name) || !gluco2Match(span, "=") {
		return span, false
	}
	return span, true
}

func gluco2HelpVerbose(o *Gluco2Option, verbose bool) {
	if verbose {
		fmt.Fprintf(os.Stderr, "\n        %s\n", o.Description)
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// Gluco2DoubleOption is Gluco2::DoubleOption : public Option.
type Gluco2DoubleOption struct {
	Gluco2Option
	range_ Gluco2DoubleRange // protected: range
	value  float64           // protected
}

// NewGluco2DoubleOption is the constructor (def defaults to double(), r to DoubleRange(-HUGE_VAL, false, HUGE_VAL, false)).
// C: DoubleOption(const char* c, const char* n, const char* d, double def = double(), DoubleRange r = DoubleRange(-HUGE_VAL, false, HUGE_VAL, false))
func NewGluco2DoubleOption(c string, n string, d string, def float64, r Gluco2DoubleRange) *Gluco2DoubleOption {
	o := &Gluco2DoubleOption{Gluco2Option: *NewGluco2Option(n, d, c, "<double>"), range_: r, value: def}
	registerGluco2Option(o)
	return o
}

// Value is `operator double() const`.
// C: operator double(void) const
func (o *Gluco2DoubleOption) Value() float64 { return o.value }

// ValueRef is `operator double&()`.
// C: operator double&(void)
func (o *Gluco2DoubleOption) ValueRef() *float64 { return &o.value }

// Assign is operator=(double).
// C: DoubleOption& operator=(double x)
func (o *Gluco2DoubleOption) Assign(x float64) *Gluco2DoubleOption {
	o.value = x
	return o
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2DoubleOption) Base() *Gluco2Option { return &o.Gluco2Option }

// Parse is DoubleOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2DoubleOption) Parse(str string) bool {
	span, ok := gluco2HeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := gluco2Strtod(span.Rest())
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
func (o *Gluco2DoubleOption) Help(verbose bool) {
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
	gluco2HelpVerbose(&o.Gluco2Option, verbose)
}

// Gluco2IntOption is Gluco2::IntOption : public Option.
type Gluco2IntOption struct {
	Gluco2Option
	range_ Gluco2IntRange // protected: range
	value  int32          // protected
}

// NewGluco2IntOption is the constructor (def defaults to int32_t(), r to IntRange(INT32_MIN, INT32_MAX)).
// C: IntOption(const char* c, const char* n, const char* d, int32_t def = int32_t(), IntRange r = IntRange(INT32_MIN, INT32_MAX))
func NewGluco2IntOption(c string, n string, d string, def int32, r Gluco2IntRange) *Gluco2IntOption {
	o := &Gluco2IntOption{Gluco2Option: *NewGluco2Option(n, d, c, "<int32>"), range_: r, value: def}
	registerGluco2Option(o)
	return o
}

// Value is `operator int32_t() const`.
// C: operator int32_t(void) const
func (o *Gluco2IntOption) Value() int32 { return o.value }

// ValueRef is `operator int32_t&()`.
// C: operator int32_t&(void)
func (o *Gluco2IntOption) ValueRef() *int32 { return &o.value }

// Assign is operator=(int32_t).
// C: IntOption& operator=(int32_t x)
func (o *Gluco2IntOption) Assign(x int32) *Gluco2IntOption {
	o.value = x
	return o
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2IntOption) Base() *Gluco2Option { return &o.Gluco2Option }

// Parse is IntOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2IntOption) Parse(str string) bool {
	span, ok := gluco2HeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := int32(gluco2Strtol(span.Rest())) // int32_t tmp = strtol(...) truncates the long
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
func (o *Gluco2IntOption) Help(verbose bool) {
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
	gluco2HelpVerbose(&o.Gluco2Option, verbose)
}

// Gluco2StringOption is Gluco2::StringOption : public Option.
type Gluco2StringOption struct {
	Gluco2Option
	value string // const char* value (NULL default is the empty string)
}

// NewGluco2StringOption is the constructor (def defaults to NULL).
// C: StringOption(const char* c, const char* n, const char* d, const char* def = NULL)
func NewGluco2StringOption(c string, n string, d string, def string) *Gluco2StringOption {
	o := &Gluco2StringOption{Gluco2Option: *NewGluco2Option(n, d, c, "<string>"), value: def}
	registerGluco2Option(o)
	return o
}

// Value is `operator const char*() const`.
// C: operator const char*(void) const
func (o *Gluco2StringOption) Value() string { return o.value }

// ValueRef is `operator const char*&()`.
// C: operator const char*&(void)
func (o *Gluco2StringOption) ValueRef() *string { return &o.value }

// Assign is operator=(const char*).
// C: StringOption& operator=(const char* x)
func (o *Gluco2StringOption) Assign(x string) *Gluco2StringOption {
	o.value = x
	return o
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2StringOption) Base() *Gluco2Option { return &o.Gluco2Option }

// Parse is StringOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2StringOption) Parse(str string) bool {
	span, ok := gluco2HeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	o.value = span.Rest()
	return true
}

// Help is StringOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *Gluco2StringOption) Help(verbose bool) {
	fmt.Fprintf(os.Stderr, "  -%-10s = %8s\n", o.Name, o.Type_name)
	gluco2HelpVerbose(&o.Gluco2Option, verbose)
}

// Gluco2BoolOption is Gluco2::BoolOption : public Option.
type Gluco2BoolOption struct {
	Gluco2Option
	value bool
}

// NewGluco2BoolOption is the constructor.
// C: BoolOption(const char* c, const char* n, const char* d, bool v)
func NewGluco2BoolOption(c string, n string, d string, v bool) *Gluco2BoolOption {
	o := &Gluco2BoolOption{Gluco2Option: *NewGluco2Option(n, d, c, "<bool>"), value: v}
	registerGluco2Option(o)
	return o
}

// Value is `operator bool() const`.
// C: operator bool(void) const
func (o *Gluco2BoolOption) Value() bool { return o.value }

// ValueRef is `operator bool&()`.
// C: operator bool&(void)
func (o *Gluco2BoolOption) ValueRef() *bool { return &o.value }

// Assign is operator=(bool).
// C: BoolOption& operator=(bool b)
func (o *Gluco2BoolOption) Assign(b bool) *Gluco2BoolOption {
	o.value = b
	return o
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2BoolOption) Base() *Gluco2Option { return &o.Gluco2Option }

// Parse is BoolOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2BoolOption) Parse(str string) bool {
	span := NewGluco2CharPtr(str)
	if gluco2Match(span, "-") {
		b := !gluco2Match(span, "no-")
		if span.Rest() == o.Name {
			o.value = b
			return true
		}
	}
	return false
}

// Help is BoolOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *Gluco2BoolOption) Help(verbose bool) {
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
	gluco2HelpVerbose(&o.Gluco2Option, verbose)
}

// Gluco2ParseOptions is Gluco2::parseOptions (returns void in this version); consumed
// options are removed from argv and argc. An unknown flag in strict mode exits with status 1.
// C: void Gluco2::parseOptions(int& argc, char** argv, bool strict)
func Gluco2ParseOptions(argc *int, argv []string, strict bool) {
	var i, j int
	for i, j = 1, 1; i < *argc; i++ {
		str := NewGluco2CharPtr(argv[i])
		if gluco2Match(str, "--") && gluco2Match(str, gluco2OptionHelpPrefix) && gluco2Match(str, "help") {
			if str.Deref() == 0 {
				Gluco2PrintUsageAndExit(*argc, argv, false)
			} else if gluco2Match(str, "-verb") {
				Gluco2PrintUsageAndExit(*argc, argv, true)
			}
		} else {
			parsedOk := false
			for k := 0; !parsedOk && k < gluco2OptionList.Size(); k++ {
				parsedOk = (*gluco2OptionList.At(k)).Parse(argv[i])
			}
			if !parsedOk {
				cur := NewGluco2CharPtr(argv[i])
				if strict && gluco2Match(cur, "-") {
					// match() advanced argv[i] past the '-', which is what the message prints.
					fmt.Fprintf(os.Stderr, "ERROR! Unknown flag \"%s\". Use '--%shelp' for help.\n", cur.Rest(), gluco2OptionHelpPrefix)
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

// Gluco2SetUsageHelp is Gluco2::setUsageHelp.
// C: void Gluco2::setUsageHelp(const char* str)
func Gluco2SetUsageHelp(str string) {
	gluco2OptionUsage = str
	gluco2OptionUsageSet = true
}

// Gluco2SetHelpPrefixStr is Gluco2::setHelpPrefixStr.
// C: void Gluco2::setHelpPrefixStr(const char* str)
func Gluco2SetHelpPrefixStr(str string) { gluco2OptionHelpPrefix = str }

// Gluco2PrintUsageAndExit is Gluco2::printUsageAndExit (verbose defaults to false); prints the
// usage and exits with status 0.
// C: void Gluco2::printUsageAndExit(int argc, char** argv, bool verbose)
func Gluco2PrintUsageAndExit(argc int, argv []string, verbose bool) {
	if gluco2OptionUsageSet {
		fmt.Fprintf(os.Stderr, strings.Replace(gluco2OptionUsage, "%s", "%[1]s", -1), argv[0])
	}
	Gluco2SortVec(&gluco2OptionList, Gluco2OptionLt{})
	prevCat := ""
	prevType := ""
	first := true
	for i := 0; i < gluco2OptionList.Size(); i++ {
		o := (*gluco2OptionList.At(i)).Base()
		// The C code compares the const char* pointers: equal strings of one
		// category share a pointer because they come from the same literal.
		if first || o.Category != prevCat {
			fmt.Fprintf(os.Stderr, "\n%s OPTIONS:\n\n", o.Category)
		} else if o.Type_name != prevType {
			fmt.Fprintf(os.Stderr, "\n")
		}
		(*gluco2OptionList.At(i)).Help(verbose)
		prevCat = o.Category
		prevType = o.Type_name
		first = false
	}
	fmt.Fprintf(os.Stderr, "\nHELP OPTIONS:\n\n")
	fmt.Fprintf(os.Stderr, "  --%shelp        Print help message.\n", gluco2OptionHelpPrefix)
	fmt.Fprintf(os.Stderr, "  --%shelp-verb   Print verbose help message.\n", gluco2OptionHelpPrefix)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(0)
}

// Gluco2Int64Option is Gluco2::Int64Option : public Option (inside `#ifndef _MSC_VER`).
type Gluco2Int64Option struct {
	Gluco2Option
	range_ Gluco2Int64Range // protected: range
	value  int64            // protected
}

// NewGluco2Int64Option is the constructor (def defaults to int64_t(), r to Int64Range(INT64_MIN, INT64_MAX)).
// C: Int64Option(const char* c, const char* n, const char* d, int64_t def = int64_t(), Int64Range r = Int64Range(INT64_MIN, INT64_MAX))
func NewGluco2Int64Option(c string, n string, d string, def int64, r Gluco2Int64Range) *Gluco2Int64Option {
	o := &Gluco2Int64Option{Gluco2Option: *NewGluco2Option(n, d, c, "<int64>"), range_: r, value: def}
	registerGluco2Option(o)
	return o
}

// Value is `operator int64_t() const`.
// C: operator int64_t(void) const
func (o *Gluco2Int64Option) Value() int64 { return o.value }

// ValueRef is `operator int64_t&()`.
// C: operator int64_t&(void)
func (o *Gluco2Int64Option) ValueRef() *int64 { return &o.value }

// Assign is operator=(int64_t).
// C: Int64Option& operator=(int64_t x)
func (o *Gluco2Int64Option) Assign(x int64) *Gluco2Int64Option {
	o.value = x
	return o
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2Int64Option) Base() *Gluco2Option { return &o.Gluco2Option }

// Parse is Int64Option::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2Int64Option) Parse(str string) bool {
	span, ok := gluco2HeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := gluco2Strtol(span.Rest()) // strtoll
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
func (o *Gluco2Int64Option) Help(verbose bool) {
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
	gluco2HelpVerbose(&o.Gluco2Option, verbose)
}

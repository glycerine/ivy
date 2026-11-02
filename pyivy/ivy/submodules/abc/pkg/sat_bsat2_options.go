package abc

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Port of src/sat/bsat2/Options.h and Options.cpp (namespace Minisat).
// See sat_bsat2_vec.go for the C++ -> Go conventions.

// MinisatOptionI is the virtual interface of class Option: parse and help are
// pure virtual; Base gives access to the common Option part of the object so that
// vec<Option*> (Option::getOptionList) can hold any option subclass.
type MinisatOptionI interface {
	Base() *MinisatOption
	Parse(str string) bool
	Help(verbose bool)
}

// MinisatOption is Minisat::Option, the abstract base of all command-line options.
type MinisatOption struct {
	Name        string // const char* name
	Description string // const char* description
	Category    string // const char* category
	Type_name   string // const char* type_name
}

var (
	minisatOptionList       MinisatVec[MinisatOptionI] // function-local static of getOptionList
	minisatOptionUsage      string                     // function-local static of getUsageString (NULL == "")
	minisatOptionUsageSet   bool
	minisatOptionHelpPrefix string // function-local static of getHelpPrefixString, initially ""
)

// MinisatOptionGetOptionList is Option::getOptionList (static; holds a function-local static vec<Option*>).
// C: static vec<Option*>& getOptionList()
func MinisatOptionGetOptionList() *MinisatVec[MinisatOptionI] { return &minisatOptionList }

// MinisatOptionGetUsageString is Option::getUsageString (static; function-local static const char*).
// C: static const char*& getUsageString()
func MinisatOptionGetUsageString() *string { return &minisatOptionUsage }

// MinisatOptionGetHelpPrefixString is Option::getHelpPrefixString (static; function-local static, initially "").
// C: static const char*& getHelpPrefixString()
func MinisatOptionGetHelpPrefixString() *string { return &minisatOptionHelpPrefix }

// MinisatOptionLt is Option::OptionLt: orders options by category then type_name.
type MinisatOptionLt struct{}

// Call is OptionLt::operator().
// C: bool operator()(const Option* x, const Option* y)
func (l MinisatOptionLt) Call(x MinisatOptionI, y MinisatOptionI) bool {
	test1 := strings.Compare(x.Base().Category, y.Base().Category)
	return test1 < 0 || test1 == 0 && strings.Compare(x.Base().Type_name, y.Base().Type_name) < 0
}

// NewMinisatOption is the protected constructor Option(name_, desc_, cate_, type_).
// The C++ constructor also registers `this` in getOptionList(); the Go
// constructors of the concrete option classes do that registration, because only
// they hold the derived object (see registerMinisatOption).
// C: Option(const char* name_, const char* desc_, const char* cate_, const char* type_)
func NewMinisatOption(name_ string, desc_ string, cate_ string, type_ string) *MinisatOption {
	return &MinisatOption{Name: name_, Description: desc_, Category: cate_, Type_name: type_}
}

// registerMinisatOption is the `getOptionList().push(this)` of the Option constructor.
func registerMinisatOption(o MinisatOptionI) { minisatOptionList.PushElem(o) }

// Destroy is Option's virtual destructor.
// C: virtual ~Option()
func (o *MinisatOption) Destroy() {}

// MinisatIntRange is Minisat::IntRange.
type MinisatIntRange struct {
	Begin int
	End   int
}

// NewMinisatIntRange is IntRange(int b, int e).
// C: IntRange(int b, int e)
func NewMinisatIntRange(b int, e int) MinisatIntRange { return MinisatIntRange{Begin: b, End: e} }

// MinisatInt64Range is Minisat::Int64Range.
type MinisatInt64Range struct {
	Begin int64
	End   int64
}

// NewMinisatInt64Range is Int64Range(int64_t b, int64_t e).
// C: Int64Range(int64_t b, int64_t e)
func NewMinisatInt64Range(b int64, e int64) MinisatInt64Range {
	return MinisatInt64Range{Begin: b, End: e}
}

// MinisatDoubleRange is Minisat::DoubleRange.
type MinisatDoubleRange struct {
	Begin           float64
	End             float64
	Begin_inclusive bool
	End_inclusive   bool
}

// NewMinisatDoubleRange is DoubleRange(double b, bool binc, double e, bool einc).
// C: DoubleRange(double b, bool binc, double e, bool einc)
func NewMinisatDoubleRange(b float64, binc bool, e float64, einc bool) MinisatDoubleRange {
	return MinisatDoubleRange{Begin: b, End: e, Begin_inclusive: binc, End_inclusive: einc}
}

// minisatStrtod is C strtod on the prefix of s that forms a number (0 if none).
func minisatStrtod(s string) float64 {
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
		return minisatSignedInf(s[start:i])
	}
	if strings.HasPrefix(lower, "inf") {
		return minisatSignedInf(s[start:i])
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

func minisatSignedInf(sign string) float64 {
	if sign == "-" {
		return math.Inf(-1)
	}
	return math.Inf(1)
}

// minisatStrtol is C strtol(s, &end, base 10) on a 64-bit long: the prefix of s
// that forms a number, saturating at the long limits (0 if none).
func minisatStrtol(s string) int64 {
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

// minisatHeaderMatch is the common start of every parse(): match "-", name, "=".
func minisatHeaderMatch(str string, name string) (*MinisatCharPtr, bool) {
	span := NewMinisatCharPtr(str)
	if !minisatMatch(span, "-") || !minisatMatch(span, name) || !minisatMatch(span, "=") {
		return span, false
	}
	return span, true
}

func minisatHelpVerbose(o *MinisatOption, verbose bool) {
	if verbose {
		fmt.Fprintf(os.Stderr, "\n        %s\n", o.Description)
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// MinisatDoubleOption is Minisat::DoubleOption : public Option.
type MinisatDoubleOption struct {
	MinisatOption
	range_ MinisatDoubleRange // protected: range
	value  float64            // protected
}

// NewMinisatDoubleOption is the constructor (def defaults to double(), r to DoubleRange(-HUGE_VAL, false, HUGE_VAL, false)).
// C: DoubleOption(const char* c, const char* n, const char* d, double def = double(), DoubleRange r = DoubleRange(-HUGE_VAL, false, HUGE_VAL, false))
func NewMinisatDoubleOption(c string, n string, d string, def float64, r MinisatDoubleRange) *MinisatDoubleOption {
	o := &MinisatDoubleOption{MinisatOption: *NewMinisatOption(n, d, c, "<double>"), range_: r, value: def}
	registerMinisatOption(o)
	return o
}

// Value is `operator double() const`.
// C: operator double(void) const
func (o *MinisatDoubleOption) Value() float64 { return o.value }

// ValueRef is `operator double&()`.
// C: operator double&(void)
func (o *MinisatDoubleOption) ValueRef() *float64 { return &o.value }

// Assign is operator=(double).
// C: DoubleOption& operator=(double x)
func (o *MinisatDoubleOption) Assign(x float64) *MinisatDoubleOption {
	o.value = x
	return o
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatDoubleOption) Base() *MinisatOption { return &o.MinisatOption }

// Parse is DoubleOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatDoubleOption) Parse(str string) bool {
	span, ok := minisatHeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := minisatStrtod(span.Rest())
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
func (o *MinisatDoubleOption) Help(verbose bool) {
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
	minisatHelpVerbose(&o.MinisatOption, verbose)
}

// MinisatIntOption is Minisat::IntOption : public Option.
type MinisatIntOption struct {
	MinisatOption
	range_ MinisatIntRange // protected: range
	value  int32           // protected
}

// NewMinisatIntOption is the constructor (def defaults to int32_t(), r to IntRange(INT32_MIN, INT32_MAX)).
// C: IntOption(const char* c, const char* n, const char* d, int32_t def = int32_t(), IntRange r = IntRange(INT32_MIN, INT32_MAX))
func NewMinisatIntOption(c string, n string, d string, def int32, r MinisatIntRange) *MinisatIntOption {
	o := &MinisatIntOption{MinisatOption: *NewMinisatOption(n, d, c, "<int32>"), range_: r, value: def}
	registerMinisatOption(o)
	return o
}

// Value is `operator int32_t() const`.
// C: operator int32_t(void) const
func (o *MinisatIntOption) Value() int32 { return o.value }

// ValueRef is `operator int32_t&()`.
// C: operator int32_t&(void)
func (o *MinisatIntOption) ValueRef() *int32 { return &o.value }

// Assign is operator=(int32_t).
// C: IntOption& operator=(int32_t x)
func (o *MinisatIntOption) Assign(x int32) *MinisatIntOption {
	o.value = x
	return o
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatIntOption) Base() *MinisatOption { return &o.MinisatOption }

// Parse is IntOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatIntOption) Parse(str string) bool {
	span, ok := minisatHeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	tmp := int32(minisatStrtol(span.Rest())) // int32_t tmp = strtol(...) truncates the long
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
func (o *MinisatIntOption) Help(verbose bool) {
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
	minisatHelpVerbose(&o.MinisatOption, verbose)
}

// MinisatStringOption is Minisat::StringOption : public Option.
type MinisatStringOption struct {
	MinisatOption
	value string // const char* value (NULL default is the empty string)
}

// NewMinisatStringOption is the constructor (def defaults to NULL).
// C: StringOption(const char* c, const char* n, const char* d, const char* def = NULL)
func NewMinisatStringOption(c string, n string, d string, def string) *MinisatStringOption {
	o := &MinisatStringOption{MinisatOption: *NewMinisatOption(n, d, c, "<string>"), value: def}
	registerMinisatOption(o)
	return o
}

// Value is `operator const char*() const`.
// C: operator const char*(void) const
func (o *MinisatStringOption) Value() string { return o.value }

// ValueRef is `operator const char*&()`.
// C: operator const char*&(void)
func (o *MinisatStringOption) ValueRef() *string { return &o.value }

// Assign is operator=(const char*).
// C: StringOption& operator=(const char* x)
func (o *MinisatStringOption) Assign(x string) *MinisatStringOption {
	o.value = x
	return o
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatStringOption) Base() *MinisatOption { return &o.MinisatOption }

// Parse is StringOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatStringOption) Parse(str string) bool {
	span, ok := minisatHeaderMatch(str, o.Name)
	if !ok {
		return false
	}
	o.value = span.Rest()
	return true
}

// Help is StringOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *MinisatStringOption) Help(verbose bool) {
	fmt.Fprintf(os.Stderr, "  -%-10s = %8s\n", o.Name, o.Type_name)
	minisatHelpVerbose(&o.MinisatOption, verbose)
}

// MinisatBoolOption is Minisat::BoolOption : public Option.
type MinisatBoolOption struct {
	MinisatOption
	value bool
}

// NewMinisatBoolOption is the constructor.
// C: BoolOption(const char* c, const char* n, const char* d, bool v)
func NewMinisatBoolOption(c string, n string, d string, v bool) *MinisatBoolOption {
	o := &MinisatBoolOption{MinisatOption: *NewMinisatOption(n, d, c, "<bool>"), value: v}
	registerMinisatOption(o)
	return o
}

// Value is `operator bool() const`.
// C: operator bool(void) const
func (o *MinisatBoolOption) Value() bool { return o.value }

// ValueRef is `operator bool&()`.
// C: operator bool&(void)
func (o *MinisatBoolOption) ValueRef() *bool { return &o.value }

// Assign is operator=(bool).
// C: BoolOption& operator=(bool b)
func (o *MinisatBoolOption) Assign(b bool) *MinisatBoolOption {
	o.value = b
	return o
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatBoolOption) Base() *MinisatOption { return &o.MinisatOption }

// Parse is BoolOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatBoolOption) Parse(str string) bool {
	span := NewMinisatCharPtr(str)
	if minisatMatch(span, "-") {
		b := !minisatMatch(span, "no-")
		if span.Rest() == o.Name {
			o.value = b
			return true
		}
	}
	return false
}

// Help is BoolOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *MinisatBoolOption) Help(verbose bool) {
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
	minisatHelpVerbose(&o.MinisatOption, verbose)
}

// MinisatParseOptions is Minisat::parseOptions; consumed options are removed from argv and argc.
// Returns 0 on an unknown flag in strict mode (or after printing the help), else 1.
// C: int Minisat::parseOptions(int& argc, char** argv, bool strict)
func MinisatParseOptions(argc *int, argv []string, strict bool) int {
	var i, j int
	for i, j = 1, 1; i < *argc; i++ {
		str := NewMinisatCharPtr(argv[i])
		if minisatMatch(str, "--") && minisatMatch(str, minisatOptionHelpPrefix) && minisatMatch(str, "help") {
			if str.Deref() == 0 {
				return MinisatPrintUsageAndExit(*argc, argv, false)
			} else if minisatMatch(str, "-verb") {
				return MinisatPrintUsageAndExit(*argc, argv, true)
			}
		} else {
			parsedOk := false
			for k := 0; !parsedOk && k < minisatOptionList.Size(); k++ {
				parsedOk = (*minisatOptionList.At(k)).Parse(argv[i])
			}
			if !parsedOk {
				cur := NewMinisatCharPtr(argv[i])
				if strict && minisatMatch(cur, "-") {
					// match() advanced argv[i] past the '-', which is what the message prints.
					fmt.Fprintf(os.Stderr, "ERROR! Unknown flag \"%s\". Use '--%shelp' for help.\n", cur.Rest(), minisatOptionHelpPrefix)
					return 0
				}
				argv[j] = argv[i]
				j++
			}
		}
	}
	*argc -= i - j
	return 1
}

// MinisatSetUsageHelp is Minisat::setUsageHelp.
// C: void Minisat::setUsageHelp(const char* str)
func MinisatSetUsageHelp(str string) {
	minisatOptionUsage = str
	minisatOptionUsageSet = true
}

// MinisatSetHelpPrefixStr is Minisat::setHelpPrefixStr.
// C: void Minisat::setHelpPrefixStr(const char* str)
func MinisatSetHelpPrefixStr(str string) { minisatOptionHelpPrefix = str }

// MinisatPrintUsageAndExit is Minisat::printUsageAndExit (verbose defaults to false);
// despite the name it only prints and returns 0 (the exit(0) is commented out in C).
// C: int Minisat::printUsageAndExit(int argc, char** argv, bool verbose)
func MinisatPrintUsageAndExit(argc int, argv []string, verbose bool) int {
	if minisatOptionUsageSet {
		fmt.Fprintf(os.Stderr, strings.Replace(minisatOptionUsage, "%s", "%[1]s", -1), argv[0])
	}
	MinisatSortVec(&minisatOptionList, MinisatOptionLt{})
	prevCat := ""
	prevType := ""
	first := true
	for i := 0; i < minisatOptionList.Size(); i++ {
		o := (*minisatOptionList.At(i)).Base()
		// The C code compares the const char* pointers: equal strings of one
		// category share a pointer because they come from the same literal.
		if first || o.Category != prevCat {
			fmt.Fprintf(os.Stderr, "\n%s OPTIONS:\n\n", o.Category)
		} else if o.Type_name != prevType {
			fmt.Fprintf(os.Stderr, "\n")
		}
		(*minisatOptionList.At(i)).Help(verbose)
		prevCat = o.Category
		prevType = o.Type_name
		first = false
	}
	fmt.Fprintf(os.Stderr, "\nHELP OPTIONS:\n\n")
	fmt.Fprintf(os.Stderr, "  --%shelp        Print help message.\n", minisatOptionHelpPrefix)
	fmt.Fprintf(os.Stderr, "  --%shelp-verb   Print verbose help message.\n", minisatOptionHelpPrefix)
	fmt.Fprintf(os.Stderr, "\n")
	return 0
}

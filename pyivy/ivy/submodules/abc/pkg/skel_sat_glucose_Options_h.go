package abc

// Hand-checked skeleton of src/sat/glucose/Options.h (namespace Gluco).
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

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

// GlucoOptionGetOptionList is Option::getOptionList (static; holds a function-local static vec<Option*>).
// C: static vec<Option*>& getOptionList()
func GlucoOptionGetOptionList() *GlucoVec[GlucoOptionI] {
	panic("TODO, stubbed")
}

// GlucoOptionGetUsageString is Option::getUsageString (static; function-local static const char*).
// C: static const char*& getUsageString()
func GlucoOptionGetUsageString() *string {
	panic("TODO, stubbed")
}

// GlucoOptionGetHelpPrefixString is Option::getHelpPrefixString (static; function-local static, initially "").
// C: static const char*& getHelpPrefixString()
func GlucoOptionGetHelpPrefixString() *string {
	panic("TODO, stubbed")
}

// GlucoOptionLt is Option::OptionLt: orders options by category then type_name.
type GlucoOptionLt struct{}

// Call is OptionLt::operator().
// C: bool operator()(const Option* x, const Option* y)
func (l GlucoOptionLt) Call(x GlucoOptionI, y GlucoOptionI) bool {
	panic("TODO, stubbed")
}

// NewGlucoOption is the protected constructor Option(name_, desc_, cate_, type_); it registers the option in getOptionList().
// C: Option(const char* name_, const char* desc_, const char* cate_, const char* type_)
func NewGlucoOption(name_ string, desc_ string, cate_ string, type_ string) *GlucoOption {
	panic("TODO, stubbed")
}

// Destroy is Option's virtual destructor.
// C: virtual ~Option()
func (o *GlucoOption) Destroy() {
	panic("TODO, stubbed")
}

// GlucoIntRange is Gluco::IntRange.
type GlucoIntRange struct {
	Begin int
	End   int
}

// NewGlucoIntRange is IntRange(int b, int e).
// C: IntRange(int b, int e)
func NewGlucoIntRange(b int, e int) GlucoIntRange {
	panic("TODO, stubbed")
}

// GlucoInt64Range is Gluco::Int64Range.
type GlucoInt64Range struct {
	Begin int64
	End   int64
}

// NewGlucoInt64Range is Int64Range(int64_t b, int64_t e).
// C: Int64Range(int64_t b, int64_t e)
func NewGlucoInt64Range(b int64, e int64) GlucoInt64Range {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator double() const`.
// C: operator double(void) const
func (o *GlucoDoubleOption) Value() float64 {
	panic("TODO, stubbed")
}

// ValueRef is `operator double&()`.
// C: operator double&(void)
func (o *GlucoDoubleOption) ValueRef() *float64 {
	panic("TODO, stubbed")
}

// Assign is operator=(double).
// C: DoubleOption& operator=(double x)
func (o *GlucoDoubleOption) Assign(x float64) *GlucoDoubleOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoDoubleOption) Base() *GlucoOption {
	panic("TODO, stubbed")
}

// Parse is DoubleOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoDoubleOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is DoubleOption::help (virtual; verbose defaults to false).
// C: virtual void help(bool verbose = false)
func (o *GlucoDoubleOption) Help(verbose bool) {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator int32_t() const`.
// C: operator int32_t(void) const
func (o *GlucoIntOption) Value() int32 {
	panic("TODO, stubbed")
}

// ValueRef is `operator int32_t&()`.
// C: operator int32_t&(void)
func (o *GlucoIntOption) ValueRef() *int32 {
	panic("TODO, stubbed")
}

// Assign is operator=(int32_t).
// C: IntOption& operator=(int32_t x)
func (o *GlucoIntOption) Assign(x int32) *GlucoIntOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoIntOption) Base() *GlucoOption {
	panic("TODO, stubbed")
}

// Parse is IntOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoIntOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is IntOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *GlucoIntOption) Help(verbose bool) {
	panic("TODO, stubbed")
}

// GlucoStringOption is Gluco::StringOption : public Option.
type GlucoStringOption struct {
	GlucoOption
	value string // const char* value (NULL default is the empty string)
}

// NewGlucoStringOption is the constructor (def defaults to NULL).
// C: StringOption(const char* c, const char* n, const char* d, const char* def = NULL)
func NewGlucoStringOption(c string, n string, d string, def string) *GlucoStringOption {
	panic("TODO, stubbed")
}

// Value is `operator const char*() const`.
// C: operator const char*(void) const
func (o *GlucoStringOption) Value() string {
	panic("TODO, stubbed")
}

// ValueRef is `operator const char*&()`.
// C: operator const char*&(void)
func (o *GlucoStringOption) ValueRef() *string {
	panic("TODO, stubbed")
}

// Assign is operator=(const char*).
// C: StringOption& operator=(const char* x)
func (o *GlucoStringOption) Assign(x string) *GlucoStringOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoStringOption) Base() *GlucoOption {
	panic("TODO, stubbed")
}

// Parse is StringOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoStringOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is StringOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *GlucoStringOption) Help(verbose bool) {
	panic("TODO, stubbed")
}

// GlucoBoolOption is Gluco::BoolOption : public Option.
type GlucoBoolOption struct {
	GlucoOption
	value bool
}

// NewGlucoBoolOption is the constructor.
// C: BoolOption(const char* c, const char* n, const char* d, bool v)
func NewGlucoBoolOption(c string, n string, d string, v bool) *GlucoBoolOption {
	panic("TODO, stubbed")
}

// Value is `operator bool() const`.
// C: operator bool(void) const
func (o *GlucoBoolOption) Value() bool {
	panic("TODO, stubbed")
}

// ValueRef is `operator bool&()`.
// C: operator bool&(void)
func (o *GlucoBoolOption) ValueRef() *bool {
	panic("TODO, stubbed")
}

// Assign is operator=(bool).
// C: BoolOption& operator=(bool b)
func (o *GlucoBoolOption) Assign(b bool) *GlucoBoolOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoBoolOption) Base() *GlucoOption {
	panic("TODO, stubbed")
}

// Parse is BoolOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoBoolOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is BoolOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *GlucoBoolOption) Help(verbose bool) {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator int64_t() const`.
// C: operator int64_t(void) const
func (o *GlucoInt64Option) Value() int64 {
	panic("TODO, stubbed")
}

// ValueRef is `operator int64_t&()`.
// C: operator int64_t&(void)
func (o *GlucoInt64Option) ValueRef() *int64 {
	panic("TODO, stubbed")
}

// Assign is operator=(int64_t).
// C: Int64Option& operator=(int64_t x)
func (o *GlucoInt64Option) Assign(x int64) *GlucoInt64Option {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements GlucoOptionI).
func (o *GlucoInt64Option) Base() *GlucoOption {
	panic("TODO, stubbed")
}

// Parse is Int64Option::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *GlucoInt64Option) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is Int64Option::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *GlucoInt64Option) Help(verbose bool) {
	panic("TODO, stubbed")
}

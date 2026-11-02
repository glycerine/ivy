package abc

// Hand-checked skeleton of src/sat/glucose2/Options.h (namespace Gluco2).
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

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

// Gluco2OptionGetOptionList is Option::getOptionList (static; holds a function-local static vec<Option*>).
// C: static vec<Option*>& getOptionList()
func Gluco2OptionGetOptionList() *Gluco2Vec[Gluco2OptionI] {
	panic("TODO, stubbed")
}

// Gluco2OptionGetUsageString is Option::getUsageString (static; function-local static const char*).
// C: static const char*& getUsageString()
func Gluco2OptionGetUsageString() *string {
	panic("TODO, stubbed")
}

// Gluco2OptionGetHelpPrefixString is Option::getHelpPrefixString (static; function-local static, initially "").
// C: static const char*& getHelpPrefixString()
func Gluco2OptionGetHelpPrefixString() *string {
	panic("TODO, stubbed")
}

// Gluco2OptionLt is Option::OptionLt: orders options by category then type_name.
type Gluco2OptionLt struct{}

// Call is OptionLt::operator().
// C: bool operator()(const Option* x, const Option* y)
func (l Gluco2OptionLt) Call(x Gluco2OptionI, y Gluco2OptionI) bool {
	panic("TODO, stubbed")
}

// NewGluco2Option is the protected constructor Option(name_, desc_, cate_, type_); it registers the option in getOptionList().
// C: Option(const char* name_, const char* desc_, const char* cate_, const char* type_)
func NewGluco2Option(name_ string, desc_ string, cate_ string, type_ string) *Gluco2Option {
	panic("TODO, stubbed")
}

// Destroy is Option's virtual destructor.
// C: virtual ~Option()
func (o *Gluco2Option) Destroy() {
	panic("TODO, stubbed")
}

// Gluco2IntRange is Gluco2::IntRange.
type Gluco2IntRange struct {
	Begin int
	End   int
}

// NewGluco2IntRange is IntRange(int b, int e).
// C: IntRange(int b, int e)
func NewGluco2IntRange(b int, e int) Gluco2IntRange {
	panic("TODO, stubbed")
}

// Gluco2Int64Range is Gluco2::Int64Range.
type Gluco2Int64Range struct {
	Begin int64
	End   int64
}

// NewGluco2Int64Range is Int64Range(int64_t b, int64_t e).
// C: Int64Range(int64_t b, int64_t e)
func NewGluco2Int64Range(b int64, e int64) Gluco2Int64Range {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator double() const`.
// C: operator double(void) const
func (o *Gluco2DoubleOption) Value() float64 {
	panic("TODO, stubbed")
}

// ValueRef is `operator double&()`.
// C: operator double&(void)
func (o *Gluco2DoubleOption) ValueRef() *float64 {
	panic("TODO, stubbed")
}

// Assign is operator=(double).
// C: DoubleOption& operator=(double x)
func (o *Gluco2DoubleOption) Assign(x float64) *Gluco2DoubleOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2DoubleOption) Base() *Gluco2Option {
	panic("TODO, stubbed")
}

// Parse is DoubleOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2DoubleOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is DoubleOption::help (virtual; verbose defaults to false).
// C: virtual void help(bool verbose = false)
func (o *Gluco2DoubleOption) Help(verbose bool) {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator int32_t() const`.
// C: operator int32_t(void) const
func (o *Gluco2IntOption) Value() int32 {
	panic("TODO, stubbed")
}

// ValueRef is `operator int32_t&()`.
// C: operator int32_t&(void)
func (o *Gluco2IntOption) ValueRef() *int32 {
	panic("TODO, stubbed")
}

// Assign is operator=(int32_t).
// C: IntOption& operator=(int32_t x)
func (o *Gluco2IntOption) Assign(x int32) *Gluco2IntOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2IntOption) Base() *Gluco2Option {
	panic("TODO, stubbed")
}

// Parse is IntOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2IntOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is IntOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *Gluco2IntOption) Help(verbose bool) {
	panic("TODO, stubbed")
}

// Gluco2StringOption is Gluco2::StringOption : public Option.
type Gluco2StringOption struct {
	Gluco2Option
	value string // const char* value (NULL default is the empty string)
}

// NewGluco2StringOption is the constructor (def defaults to NULL).
// C: StringOption(const char* c, const char* n, const char* d, const char* def = NULL)
func NewGluco2StringOption(c string, n string, d string, def string) *Gluco2StringOption {
	panic("TODO, stubbed")
}

// Value is `operator const char*() const`.
// C: operator const char*(void) const
func (o *Gluco2StringOption) Value() string {
	panic("TODO, stubbed")
}

// ValueRef is `operator const char*&()`.
// C: operator const char*&(void)
func (o *Gluco2StringOption) ValueRef() *string {
	panic("TODO, stubbed")
}

// Assign is operator=(const char*).
// C: StringOption& operator=(const char* x)
func (o *Gluco2StringOption) Assign(x string) *Gluco2StringOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2StringOption) Base() *Gluco2Option {
	panic("TODO, stubbed")
}

// Parse is StringOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2StringOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is StringOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *Gluco2StringOption) Help(verbose bool) {
	panic("TODO, stubbed")
}

// Gluco2BoolOption is Gluco2::BoolOption : public Option.
type Gluco2BoolOption struct {
	Gluco2Option
	value bool
}

// NewGluco2BoolOption is the constructor.
// C: BoolOption(const char* c, const char* n, const char* d, bool v)
func NewGluco2BoolOption(c string, n string, d string, v bool) *Gluco2BoolOption {
	panic("TODO, stubbed")
}

// Value is `operator bool() const`.
// C: operator bool(void) const
func (o *Gluco2BoolOption) Value() bool {
	panic("TODO, stubbed")
}

// ValueRef is `operator bool&()`.
// C: operator bool&(void)
func (o *Gluco2BoolOption) ValueRef() *bool {
	panic("TODO, stubbed")
}

// Assign is operator=(bool).
// C: BoolOption& operator=(bool b)
func (o *Gluco2BoolOption) Assign(b bool) *Gluco2BoolOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2BoolOption) Base() *Gluco2Option {
	panic("TODO, stubbed")
}

// Parse is BoolOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2BoolOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is BoolOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *Gluco2BoolOption) Help(verbose bool) {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator int64_t() const`.
// C: operator int64_t(void) const
func (o *Gluco2Int64Option) Value() int64 {
	panic("TODO, stubbed")
}

// ValueRef is `operator int64_t&()`.
// C: operator int64_t&(void)
func (o *Gluco2Int64Option) ValueRef() *int64 {
	panic("TODO, stubbed")
}

// Assign is operator=(int64_t).
// C: Int64Option& operator=(int64_t x)
func (o *Gluco2Int64Option) Assign(x int64) *Gluco2Int64Option {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements Gluco2OptionI).
func (o *Gluco2Int64Option) Base() *Gluco2Option {
	panic("TODO, stubbed")
}

// Parse is Int64Option::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *Gluco2Int64Option) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is Int64Option::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *Gluco2Int64Option) Help(verbose bool) {
	panic("TODO, stubbed")
}

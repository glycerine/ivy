package abc

// Hand-checked skeleton of src/sat/bsat2/Options.h (namespace Minisat).
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

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

// MinisatOptionGetOptionList is Option::getOptionList (static; holds a function-local static vec<Option*>).
// C: static vec<Option*>& getOptionList()
func MinisatOptionGetOptionList() *MinisatVec[MinisatOptionI] {
	panic("TODO, stubbed")
}

// MinisatOptionGetUsageString is Option::getUsageString (static; function-local static const char*).
// C: static const char*& getUsageString()
func MinisatOptionGetUsageString() *string {
	panic("TODO, stubbed")
}

// MinisatOptionGetHelpPrefixString is Option::getHelpPrefixString (static; function-local static, initially "").
// C: static const char*& getHelpPrefixString()
func MinisatOptionGetHelpPrefixString() *string {
	panic("TODO, stubbed")
}

// MinisatOptionLt is Option::OptionLt: orders options by category then type_name.
type MinisatOptionLt struct{}

// Call is OptionLt::operator().
// C: bool operator()(const Option* x, const Option* y)
func (l MinisatOptionLt) Call(x MinisatOptionI, y MinisatOptionI) bool {
	panic("TODO, stubbed")
}

// NewMinisatOption is the protected constructor Option(name_, desc_, cate_, type_); it registers the option in getOptionList().
// C: Option(const char* name_, const char* desc_, const char* cate_, const char* type_)
func NewMinisatOption(name_ string, desc_ string, cate_ string, type_ string) *MinisatOption {
	panic("TODO, stubbed")
}

// Destroy is Option's virtual destructor.
// C: virtual ~Option()
func (o *MinisatOption) Destroy() {
	panic("TODO, stubbed")
}

// MinisatIntRange is Minisat::IntRange.
type MinisatIntRange struct {
	Begin int
	End   int
}

// NewMinisatIntRange is IntRange(int b, int e).
// C: IntRange(int b, int e)
func NewMinisatIntRange(b int, e int) MinisatIntRange {
	panic("TODO, stubbed")
}

// MinisatInt64Range is Minisat::Int64Range.
type MinisatInt64Range struct {
	Begin int64
	End   int64
}

// NewMinisatInt64Range is Int64Range(int64_t b, int64_t e).
// C: Int64Range(int64_t b, int64_t e)
func NewMinisatInt64Range(b int64, e int64) MinisatInt64Range {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator double() const`.
// C: operator double(void) const
func (o *MinisatDoubleOption) Value() float64 {
	panic("TODO, stubbed")
}

// ValueRef is `operator double&()`.
// C: operator double&(void)
func (o *MinisatDoubleOption) ValueRef() *float64 {
	panic("TODO, stubbed")
}

// Assign is operator=(double).
// C: DoubleOption& operator=(double x)
func (o *MinisatDoubleOption) Assign(x float64) *MinisatDoubleOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatDoubleOption) Base() *MinisatOption {
	panic("TODO, stubbed")
}

// Parse is DoubleOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatDoubleOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is DoubleOption::help (virtual; verbose defaults to false).
// C: virtual void help(bool verbose = false)
func (o *MinisatDoubleOption) Help(verbose bool) {
	panic("TODO, stubbed")
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
	panic("TODO, stubbed")
}

// Value is `operator int32_t() const`.
// C: operator int32_t(void) const
func (o *MinisatIntOption) Value() int32 {
	panic("TODO, stubbed")
}

// ValueRef is `operator int32_t&()`.
// C: operator int32_t&(void)
func (o *MinisatIntOption) ValueRef() *int32 {
	panic("TODO, stubbed")
}

// Assign is operator=(int32_t).
// C: IntOption& operator=(int32_t x)
func (o *MinisatIntOption) Assign(x int32) *MinisatIntOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatIntOption) Base() *MinisatOption {
	panic("TODO, stubbed")
}

// Parse is IntOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatIntOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is IntOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *MinisatIntOption) Help(verbose bool) {
	panic("TODO, stubbed")
}

// MinisatStringOption is Minisat::StringOption : public Option.
type MinisatStringOption struct {
	MinisatOption
	value string // const char* value (NULL default is the empty string)
}

// NewMinisatStringOption is the constructor (def defaults to NULL).
// C: StringOption(const char* c, const char* n, const char* d, const char* def = NULL)
func NewMinisatStringOption(c string, n string, d string, def string) *MinisatStringOption {
	panic("TODO, stubbed")
}

// Value is `operator const char*() const`.
// C: operator const char*(void) const
func (o *MinisatStringOption) Value() string {
	panic("TODO, stubbed")
}

// ValueRef is `operator const char*&()`.
// C: operator const char*&(void)
func (o *MinisatStringOption) ValueRef() *string {
	panic("TODO, stubbed")
}

// Assign is operator=(const char*).
// C: StringOption& operator=(const char* x)
func (o *MinisatStringOption) Assign(x string) *MinisatStringOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatStringOption) Base() *MinisatOption {
	panic("TODO, stubbed")
}

// Parse is StringOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatStringOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is StringOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *MinisatStringOption) Help(verbose bool) {
	panic("TODO, stubbed")
}

// MinisatBoolOption is Minisat::BoolOption : public Option.
type MinisatBoolOption struct {
	MinisatOption
	value bool
}

// NewMinisatBoolOption is the constructor.
// C: BoolOption(const char* c, const char* n, const char* d, bool v)
func NewMinisatBoolOption(c string, n string, d string, v bool) *MinisatBoolOption {
	panic("TODO, stubbed")
}

// Value is `operator bool() const`.
// C: operator bool(void) const
func (o *MinisatBoolOption) Value() bool {
	panic("TODO, stubbed")
}

// ValueRef is `operator bool&()`.
// C: operator bool&(void)
func (o *MinisatBoolOption) ValueRef() *bool {
	panic("TODO, stubbed")
}

// Assign is operator=(bool).
// C: BoolOption& operator=(bool b)
func (o *MinisatBoolOption) Assign(b bool) *MinisatBoolOption {
	panic("TODO, stubbed")
}

// Base returns the Option part (implements MinisatOptionI).
func (o *MinisatBoolOption) Base() *MinisatOption {
	panic("TODO, stubbed")
}

// Parse is BoolOption::parse (virtual).
// C: virtual bool parse(const char* str)
func (o *MinisatBoolOption) Parse(str string) bool {
	panic("TODO, stubbed")
}

// Help is BoolOption::help (virtual).
// C: virtual void help(bool verbose = false)
func (o *MinisatBoolOption) Help(verbose bool) {
	panic("TODO, stubbed")
}

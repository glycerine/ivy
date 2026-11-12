# Extended precision doubles

`EpDouble` holds a floating significand in `Value` and a machine-integer
`Exponent`. Binary operations interpret this as `Value * 2^Exponent`; decimal
operations interpret it as `Value * 10^Exponent`, as in the original EPD API.
Do not interchange these representations without conversion.

The package uses only the Go standard library. The arithmetic names follow C's
in-place/two-input conventions. `EpdGetString` returns its text rather than
writing a caller-owned C buffer. `EpdFree` is unnecessary under Go garbage
collection and retained as a compatibility no-op. `EpdCmp` compares representations
and returns zero for equality (it is not a numerical ordering comparator).

Decimal power calculation preserves the C implementation's recursive multiplication
order, and decimal exponent extraction uses its six-digit scientific rounding.
A differential test compiles the repository's original C implementation in a
temporary directory when a C compiler is available, then compares decimal powers
and exact formatted output over binary exponents -1022 through 100000. Production
code does not invoke the C compiler or depend on that reference implementation.

Intentional corrections to legacy numerical defects:

- Binary normalization preserves zero and correctly handles subnormal doubles.
  The original implementation overwrites their exponent bits, turning zero into
  a nonzero extended value and misnormalizing subnormals.
- Negative binary/decimal powers beyond the hardware double range retain their
  extended exponent instead of first underflowing through a hardware double.
- Multiplication of zero and infinity produces NaN, following IEEE floating-point
  arithmetic; the original infinity shortcut returned infinity.

The C tests deliberately compare only the reference's valid normalization domain;
separate Go tests cover the corrected boundary behavior. Neither implementation
provides arbitrary significand precision: normal floating rounding still applies.

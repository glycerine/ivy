# EPD review

## Design/readability pass

Reviewed all production entry points separately from correctness validation.
The representation is a normalized significand plus an explicit integer exponent;
there are no redundant union/bitfield views, unsafe casts, dependencies or C
runtime adapters. In-place operations share three-operand implementations;
copying input operands before writing results preserves aliasing. Decimal powers
use the original recursive multiplication order instead of logarithm-based
approximation. Allocation/free/copy and string-return API differences are stated
in README. No further structural changes were needed after this pass.

## Bug-review pass

Reviewed zero/subnormal/special handling, aliasing, normalization, exponent
conversion, decimal formatting and extreme powers. Found and fixed an additional
subnormal decimal normalization defect: `math.Pow10(-324)` underflows to zero,
so direct division produced infinity. Scaling is now split into representable
factors; TestDecimalNormalizeSubnormal was observed failing before the fix.

Documented intentional fixes to original C zero/subnormal normalization,
negative-extreme power underflow and zero-times-infinity handling. These are not
claimed to be defect-compatible with C. Original-C differential tests compare
valid-domain mantissas/exponents and exact formatted output for 13 binary powers
from -1022 to 100000; the C compiler is test-only.

Validation: `go -C abcgo test ./bdd/epd -count=1` passes, including the compiled C
oracle, arithmetic/aliasing tests and extreme/subnormal/formatting regressions.
No pending review finding remains in the implemented package; floating rounding
and machine-integer exponent limits remain documented representation limits.

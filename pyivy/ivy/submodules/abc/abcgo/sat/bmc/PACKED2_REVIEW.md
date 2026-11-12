# BMC two-bit field helper review

`Abc_InfoGet2Bits`, `Abc_InfoSet2Bits`, and `Abc_InfoAdd2Bits` store sixteen
two-bit object fields in each 32-bit word. The port uses explicit `uint32`
shifts and masks, then reads/writes the backing `Vec_Int_t` word through its
public entry API. This removes the C pointer cast and pointer arithmetic while
preserving the original packed layout and the distinction between replacement
(`Set`) and bitwise accumulation (`Add`).

The helper tests cover all sixteen offsets, replacement isolation, and the
source's OR semantics. Verified independently with:

```
go test -race packed2.go packed2_test.go -count=1
go vet packed2.go packed2_test.go
```

The containing BMC package is still not buildable: other generated files have
untranslated C assignment expressions in loop/conditional headers. Therefore
these checks verify this helper component only, not BMC integration.

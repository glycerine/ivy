# AU hierarchy original-C oracle

Generated with the existing unmodified original ABC executable:

```
./abc -c '&read_cblif abcgo/base/abc/test_vectors/hierarchy_au.cblif; &w /tmp/hierarchy_au_oracle.aig'
```

`hierarchy_au_oracle.aig` is that actual binary output. The native test reads it
with the independent IOA codec and compares all eight assignments and both
outputs against the AU CBLIF parser/flattening pipeline. The fixture exercises a
forward model reference, hierarchical AND, XOR,
and complementary outputs. Original source reports two nodes, one instance,
eight ports, one AND and one XOR; native hierarchical count checks match.

The separate complemented-actual regression intentionally differs from the
original AU implementation: the C box binding copies the underlying object
value without the literal complement bit. The native port preserves that bit,
matching CBLIF literal semantics and the ordinary hierarchy converter.

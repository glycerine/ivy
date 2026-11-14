`ashen_unreachable.blif` reproduces an unsound Ashenhurst rewrite in the original
C implementation. With the repository's native `abc` binary:

```
abc -c 'read_blif abcgo/base/acb/test_vectors/ashen_unreachable.blif; mfse -d -v; write_blif /tmp/ashen-out.blif; cec abcgo/base/acb/test_vectors/ashen_unreachable.blif /tmp/ashen-out.blif'
```

The C code chooses intermediate truth `0x4C` and output truth `0xB`, then CEC
reports NOT EQUIVALENT for `n5=1 n6=1 n2=0 n1=0 n4=0`. The Go regression uses
this same network and exhaustively checks all 64 input assignments. Candidate
validation rejects the bad rewrite while retaining the source support-search
and decomposition algorithms.

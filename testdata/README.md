# Frozen acceptance fixtures

These are independent expected-data inputs, not generated from this library's
production map or conversion implementation. They originate from nb1 commit
`112e224db3bcb7af823e4f403eec01f59a19d418`, `internal/colorinput/testdata/`.

- `css-named-2026.tsv`: 148 opaque CSS names, hex and decimal RGB channels.
  W3C CSS Color 4 source: https://www.w3.org/TR/2026/CRD-css-color-4-20261103/#named-colors
  SHA256: `521686f82ac6c649de3f22589783ac86e2bfeb0c1858fd953f59efc4e1d2fd09`.
- `hsv-golden.json`: 37 fixed expected hexadecimal results, including sector,
  decimal precision and half-up boundaries.
  SHA256: `de9d72d991ea51364aaabf680da983f9ea3754249a0dfee635983e2103c19c17`.

`internal/testvectors/vectors.go` carries the same literal data for the internal
WASM acceptance executable, where no filesystem fixture is read. Tests compare
that data back to these hashed source fixtures. There is no code generation
step, third-party dependency, JSON wire contract or hardware requirement.

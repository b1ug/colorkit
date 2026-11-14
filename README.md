# colorkit

Opaque RGB and HSV values, CSS color names, and Hex conversion for Go.

The module targets **Go 1.13 source compatibility**, uses only the standard
library, and performs no device, file, environment, or network access. Use a
maintained Go toolchain for development and delivery. Release tags identify
versioned module snapshots.

## Values and conversions

```go
import "github.com/b1ug/colorkit"

c, err := colorkit.Parse("#abc")
if err != nil {
    return err
}
fmt.Println(c.Hex()) // #aabbcc

h := colorkit.HSV{H: 210, S: 75, V: 90}
rgb, err := h.RGB8()
if err != nil {
    return err
}
fmt.Println(rgb.Hex()) // #398fe6
```

The fragment runs in a caller that imports `fmt` and returns an error.

`RGB8` contains `R`, `G`, and `B` as `uint8`. Its zero value is black. It
implements `image/color.Color`, with 16-bit RGBA channels and full opacity.

`HSV` contains `H`, `S`, and `V` as `float64`: hue is in degrees `[0,360]`,
saturation/value in percent `[0,100]`. Hue 360 equals 0. `HSV.Validate` and
`HSV.RGB8` reject NaN, infinity, and out-of-range channels. HSV converts to RGB8
before entering Go's error-free `color.Color` interface.

`RGB8.HSV()` describes the current quantized RGB value. It does not recover
the HSV parameters originally used to create the color.

## Text and names

- `Parse`: one complete name or `#rgb` / `#rrggbb` value.
- `ParseHex`: only those hexadecimal forms.
- `LookupName`: a name lookup returning `(RGB8, bool)`.
- `Names`: a sorted copy of all supported names.
- `RGB8.Hex`: canonical lowercase `#rrggbb` output.

Text inputs allow surrounding ASCII whitespace and case-insensitive names/hex
digits. The original input is limited to 64 ASCII bytes. The fixed CSS table
contains 148 opaque names and 139 distinct RGB values, including aliases and
`rebeccapurple`. `green` is `#008000`; `lime` is `#00ff00`.

The accepted grammar excludes alpha, `transparent`, `currentColor`, CSS
functions, device commands, and local color names such as `@mint`. Hardware
gamma, calibration, random colors, and artistic palettes belong to callers.

## Decimal precision

`HSV.RGB8()` first validates the input, then normalizes each channel with
`strconv.FormatFloat(x, 'f', 6, 64)`. Negative zero becomes zero. The resulting
decimals enter a shared exact rational conversion; RGB channels are rounded
half up once at the final 8-bit boundary. Values are never clamped.

Adapters that retain original decimal tokens can call
`FromHSVDecimal(hDeg, sPct, vPct string)`. Each token permits at most three
integer digits and six fractional digits, without whitespace, signs, exponents,
or leading zeros. Out-of-range values return an error. This entry does not pass
the decimal source through `float64`.

## Standard Go interfaces and errors

`RGB8` implements `encoding.TextMarshaler` and `encoding.TextUnmarshaler`.
`encoding/json` therefore encodes it as a Hex string. Invalid text leaves the
receiver unchanged. A caller needing an RGB object uses its own wire struct.

`FromColor` accepts a fully opaque standard `color.Color`. It rejects nil,
non-opaque alpha, and channels outside the 16-bit interface contract. Each valid
channel is rounded to the nearest RGB8 value using `(channel + 128) / 257`.

Failures return a zero RGB8 and a `*colorkit.Error` carrying `Op`, `Field`, and
an underlying `ErrSyntax`, `ErrRange`, or `ErrAlpha`. Inspect them with
`errors.Is` and `errors.As`; error text is not a machine contract.

## Development

```sh
go test ./...
go vet ./...
```

CI checks real Go 1.13.15 source compatibility, current Go 1.26/1.27 on Linux,
macOS and Windows, race tests, and actual Node execution of Go/WASM vectors.
Go 1.13 is a compatibility floor, not the recommended development toolchain.
`tools/check-module.py` verifies the minimal go.mod and the standard-library-only
dependency graph. The package has no CGO requirement.

The color package is independent of blink(1), blink1-go, and nb1. Downstream
adoption follows the package's implementation, acceptance checks, and version
tag. SDK and application dependency updates are validated separately.

See [NOTICE](NOTICE) for frozen data provenance and [releasing](docs/releasing.md)
for the release sequence.

## License

MIT. See [LICENSE](LICENSE).

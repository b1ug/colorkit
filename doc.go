// Package colorkit provides opaque sRGB values, named colors, hexadecimal
// literals, and HSV conversion using only the Go standard library.
//
// RGB8 implements image/color.Color and the standard text encoding interfaces.
// Text input accepts complete CSS named colors or three- and six-digit
// hexadecimal literals. It is limited to 64 bytes before trimming, permits
// ASCII surrounding whitespace, and matches names without case sensitivity.
// The fixed table contains 148 opaque CSS names, including aliases; it does
// not include transparent, currentColor, or application-defined names.
//
// HSV uses degrees for hue and percentages for saturation and value. Its
// conversion rejects non-finite and out-of-range channels. Float input is
// normalized with strconv.FormatFloat(x, 'f', 6, 64) before the shared exact
// decimal conversion. FromHSVDecimal instead preserves original decimal
// tokens with up to six fractional digits. Final RGB8 channels are rounded
// half up once. HSV does not implement color.Color because invalid values
// must be reported as errors rather than silently corrected.
//
// The package performs no I/O, device access, gamma correction, compositing,
// or application configuration lookup. It is not a complete CSS parser.
package colorkit

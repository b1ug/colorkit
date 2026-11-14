package colorkit

import (
	"encoding"
	"encoding/hex"
	"image/color"
	"reflect"
	"sort"
	"strings"
)

// RGB8 is an opaque sRGB color with eight bits per channel.
// Its zero value is black.
type RGB8 struct {
	R uint8
	G uint8
	B uint8
}

var (
	_ color.Color              = RGB8{}
	_ encoding.TextMarshaler   = RGB8{}
	_ encoding.TextUnmarshaler = (*RGB8)(nil)
)

// RGBA implements color.Color. It expands each channel to 16 bits and returns
// an alpha value of 65535, making the returned channels fully opaque.
func (c RGB8) RGBA() (r, g, b, a uint32) {
	return uint32(c.R) * 257, uint32(c.G) * 257, uint32(c.B) * 257, 65535
}

// Hex returns the canonical lowercase #rrggbb representation.
func (c RGB8) Hex() string {
	const digits = "0123456789abcdef"
	return string([]byte{
		'#', digits[c.R>>4], digits[c.R&15], digits[c.G>>4], digits[c.G&15], digits[c.B>>4], digits[c.B&15],
	})
}

// MarshalText implements encoding.TextMarshaler using canonical Hex output.
func (c RGB8) MarshalText() ([]byte, error) { return []byte(c.Hex()), nil }

// UnmarshalText implements encoding.TextUnmarshaler using Parse's syntax.
// On failure, the receiver is unchanged. A nil receiver returns ErrSyntax.
func (c *RGB8) UnmarshalText(text []byte) error {
	const op = "RGB8.UnmarshalText"
	if c == nil {
		return invalid(op, "receiver", ErrSyntax)
	}
	next, err := parse(op, string(text), false)
	if err != nil {
		return err
	}
	*c = next
	return nil
}

// Parse accepts one complete CSS named color or #rgb/#rrggbb literal.
// It permits ASCII surrounding whitespace and case-insensitive names and hex
// digits. The original input must be ASCII and no longer than 64 bytes.
func Parse(text string) (RGB8, error) { return parse("Parse", text, false) }

// ParseHex accepts a complete #rgb or #rrggbb literal, using the same ASCII
// whitespace and length rules as Parse. It does not accept color names.
func ParseHex(text string) (RGB8, error) { return parse("ParseHex", text, true) }

func parse(op, text string, hexOnly bool) (RGB8, error) {
	s, ok := normalizeText(text)
	if !ok {
		return RGB8{}, invalid(op, "text", ErrSyntax)
	}
	if len(s) > 0 && s[0] == '#' {
		if len(s) != 4 && len(s) != 7 {
			return RGB8{}, invalid(op, "text", ErrSyntax)
		}
		digits := s[1:]
		if len(digits) == 3 {
			digits = string([]byte{digits[0], digits[0], digits[1], digits[1], digits[2], digits[2]})
		}
		data, err := hex.DecodeString(digits)
		if err != nil {
			return RGB8{}, invalid(op, "text", ErrSyntax)
		}
		return RGB8{data[0], data[1], data[2]}, nil
	}
	if !hexOnly {
		if c, ok := namedColors[s]; ok {
			return c, nil
		}
	}
	return RGB8{}, invalid(op, "text", ErrSyntax)
}

// LookupName looks up a fixed CSS name without accepting hexadecimal input.
// It uses the same ASCII, case, whitespace, and length rules as Parse.
func LookupName(name string) (RGB8, bool) {
	s, ok := normalizeText(name)
	if !ok {
		return RGB8{}, false
	}
	c, ok := namedColors[s]
	return c, ok
}

// Names returns a sorted copy of all supported names, including aliases.
// Modifying the returned slice does not affect future calls or lookups.
func Names() []string {
	names := make([]string, 0, len(namedColors))
	for name := range namedColors {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalizeText(text string) (string, bool) {
	if len(text) > 64 {
		return "", false
	}
	for i := 0; i < len(text); i++ {
		if text[i] >= 128 {
			return "", false
		}
	}
	return strings.ToLower(strings.Trim(text, " \t\r\n\v\f")), true
}

// FromColor converts a fully opaque color.Color to RGB8. Nil colors, including
// typed nil values, return ErrSyntax; channels outside the 16-bit color.Color
// contract return ErrRange; non-opaque alpha returns ErrAlpha. Each valid
// 16-bit channel is rounded to the nearest 8-bit value using (v+128)/257.
func FromColor(c color.Color) (RGB8, error) {
	const op = "FromColor"
	if c == nil || isNilColor(c) {
		return RGB8{}, invalid(op, "color", ErrSyntax)
	}
	r, g, b, a := c.RGBA()
	for _, channel := range []struct {
		name  string
		value uint32
	}{{"r", r}, {"g", g}, {"b", b}, {"a", a}} {
		if channel.value > 65535 {
			return RGB8{}, invalid(op, channel.name, ErrRange)
		}
	}
	if a != 65535 {
		return RGB8{}, invalid(op, "a", ErrAlpha)
	}
	return RGB8{uint8((r + 128) / 257), uint8((g + 128) / 257), uint8((b + 128) / 257)}, nil
}

func isNilColor(c color.Color) bool {
	v := reflect.ValueOf(c)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

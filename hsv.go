package colorkit

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// HSV describes hue, saturation, and value (also called brightness).
// Its zero value is black. HSV does not implement color.Color; use RGB8 to
// validate and convert it before passing it to a standard color interface.
type HSV struct {
	H float64 // Hue in degrees, [0, 360]; 360 is equivalent to 0.
	S float64 // Saturation in percent, [0, 100].
	V float64 // Value in percent, [0, 100].
}

// Validate rejects non-finite and out-of-range channels without correcting
// them. Negative zero is valid and is normalized to zero during conversion.
func (h HSV) Validate() error { return h.validate("HSV.Validate") }

func (h HSV) validate(op string) error {
	for _, channel := range []struct {
		name       string
		value, max float64
	}{{"h", h.H, 360}, {"s", h.S, 100}, {"v", h.V, 100}} {
		if math.IsNaN(channel.value) || math.IsInf(channel.value, 0) || channel.value < 0 || channel.value > channel.max {
			return invalid(op, channel.name, ErrRange)
		}
	}
	return nil
}

// RGB8 validates h, then normalizes each channel using
// strconv.FormatFloat(x, 'f', 6, 64). The normalized decimal channels use the
// same exact conversion as FromHSVDecimal, with one final half-up rounding
// to 8-bit RGB. Channels are never clamped or reduced modulo their range.
func (h HSV) RGB8() (RGB8, error) {
	const op = "HSV.RGB8"
	if err := h.validate(op); err != nil {
		return RGB8{}, err
	}
	format := func(value float64) string {
		if value == 0 {
			return "0"
		}
		return strconv.FormatFloat(value, 'f', 6, 64)
	}
	return fromHSVDecimal(op, format(h.H), format(h.S), format(h.V))
}

// HSV converts c to hue in degrees and saturation/value in percent.
// Achromatic colors have hue zero. The result describes the quantized RGB8
// value and does not restore the HSV originally used to create it.
func (c RGB8) HSV() HSV {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min
	h := HSV{V: max * 100}
	if delta == 0 {
		return h
	}
	h.S = delta / max * 100
	switch max {
	case r:
		h.H = math.Mod((g-b)/delta, 6) * 60
	case g:
		h.H = ((b-r)/delta + 2) * 60
	default:
		h.H = ((r-g)/delta + 4) * 60
	}
	if h.H < 0 {
		h.H += 360
	}
	return h
}

// FromHSVDecimal converts original decimal tokens without first converting
// them to floating point. H is in degrees [0,360], S and V are percentages
// [0,100], and 360 degrees is equivalent to zero. Tokens accept at most six
// fractional digits, with no signs, exponent, whitespace, or leading zeros.
// Each token has at most three integer digits. Final RGB8 channels are rounded
// half up using exact rational arithmetic.
func FromHSVDecimal(hDeg, sPct, vPct string) (RGB8, error) {
	return fromHSVDecimal("FromHSVDecimal", hDeg, sPct, vPct)
}

func fromHSVDecimal(op, hDeg, sPct, vPct string) (RGB8, error) {
	h, err := decimal(op, "h", hDeg, 360)
	if err != nil {
		return RGB8{}, err
	}
	s, err := decimal(op, "s", sPct, 100)
	if err != nil {
		return RGB8{}, err
	}
	v, err := decimal(op, "v", vPct, 100)
	if err != nil {
		return RGB8{}, err
	}
	s.Quo(s, big.NewRat(100, 1))
	v.Quo(v, big.NewRat(100, 1))
	h.Quo(h, big.NewRat(60, 1))
	if h.Cmp(big.NewRat(6, 1)) == 0 {
		h.SetInt64(0)
	}
	sector := new(big.Int).Quo(h.Num(), h.Denom()).Int64()
	f := new(big.Rat).Sub(h, big.NewRat(sector, 1))
	one := big.NewRat(1, 1)
	p := new(big.Rat).Mul(v, new(big.Rat).Sub(one, s))
	q := new(big.Rat).Mul(v, new(big.Rat).Sub(one, new(big.Rat).Mul(s, f)))
	t := new(big.Rat).Mul(v, new(big.Rat).Sub(one, new(big.Rat).Mul(s, new(big.Rat).Sub(one, f))))
	var channels [3]*big.Rat
	switch sector {
	case 0:
		channels = [3]*big.Rat{v, t, p}
	case 1:
		channels = [3]*big.Rat{q, v, p}
	case 2:
		channels = [3]*big.Rat{p, v, t}
	case 3:
		channels = [3]*big.Rat{p, q, v}
	case 4:
		channels = [3]*big.Rat{t, p, v}
	default: // The validated hue guarantees sector 5 here.
		channels = [3]*big.Rat{v, p, q}
	}
	return RGB8{quantize(channels[0]), quantize(channels[1]), quantize(channels[2])}, nil
}

func decimal(op, field, text string, max int64) (*big.Rat, error) {
	if !decimalSyntax(text) {
		return nil, invalid(op, field, ErrSyntax)
	}
	r, ok := new(big.Rat).SetString(text)
	if !ok || r.Cmp(big.NewRat(max, 1)) > 0 {
		return nil, invalid(op, field, ErrRange)
	}
	return r, nil
}

func decimalSyntax(text string) bool {
	// At most three integer digits, one point, and six fractional digits.
	if len(text) == 0 || len(text) > 10 {
		return false
	}
	whole := text
	if point := strings.IndexByte(text, '.'); point >= 0 {
		fraction := text[point+1:]
		if len(fraction) < 1 || len(fraction) > 6 {
			return false
		}
		whole = text[:point]
		for i := 0; i < len(fraction); i++ {
			if fraction[i] < '0' || fraction[i] > '9' {
				return false
			}
		}
	}
	if len(whole) == 0 || len(whole) > 3 || len(whole) > 1 && whole[0] == '0' {
		return false
	}
	for i := 0; i < len(whole); i++ {
		if whole[i] < '0' || whole[i] > '9' {
			return false
		}
	}
	return true
}

func quantize(channel *big.Rat) uint8 {
	rounded := new(big.Rat).Add(new(big.Rat).Mul(channel, big.NewRat(255, 1)), big.NewRat(1, 2))
	return uint8(new(big.Int).Quo(rounded.Num(), rounded.Denom()).Int64())
}

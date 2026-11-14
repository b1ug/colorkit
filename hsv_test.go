package colorkit_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/b1ug/colorkit"
	"github.com/b1ug/colorkit/internal/testvectors"
)

func TestFrozenHSVGoldenVectors(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/hsv-golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "de9d72d991ea51364aaabf680da983f9ea3754249a0dfee635983e2103c19c17" {
		t.Fatal("fixture hash changed", got)
	}
	var fixtures []struct{ Name, Input, Want string }
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) != 37 || len(testvectors.HSV) != 37 {
		t.Fatal("incomplete golden set")
	}
	for i, row := range fixtures {
		var input struct {
			H json.Number `json:"h_deg"`
			S json.Number `json:"s_pct"`
			V json.Number `json:"v_pct"`
		}
		if err = json.Unmarshal([]byte(row.Input), &input); err != nil {
			t.Fatal(err)
		}
		compiled := testvectors.HSV[i]
		if compiled.Name != row.Name || compiled.H != input.H.String() || compiled.S != input.S.String() || compiled.V != input.V.String() || compiled.Hex != row.Want {
			t.Fatal("compiled fixture differs", i)
		}
		t.Run(row.Name, func(t *testing.T) {
			got, e := colorkit.FromHSVDecimal(input.H.String(), input.S.String(), input.V.String())
			if e != nil || got.Hex() != row.Want {
				t.Fatal(got, e, row.Want)
			}
			h, _ := strconv.ParseFloat(input.H.String(), 64)
			s, _ := strconv.ParseFloat(input.S.String(), 64)
			v, _ := strconv.ParseFloat(input.V.String(), 64)
			numeric, e := (colorkit.HSV{H: h, S: s, V: v}).RGB8()
			if e != nil || numeric != got {
				t.Fatal("numeric/decimal differ", numeric, got, e)
			}
		})
	}
}

func TestHSVRejectsEveryInvalidField(t *testing.T) {
	fields := []string{"h", "s", "v"}
	for i, field := range fields {
		max := 100.0
		if i == 0 {
			max = 360
		}
		bad := []float64{math.NaN(), math.Inf(1), math.Inf(-1), -math.SmallestNonzeroFloat64, math.Nextafter(max, math.Inf(1))}
		for j, value := range bad {
			h := colorkit.HSV{H: 10, S: 20, V: 30}
			switch i {
			case 0:
				h.H = value
			case 1:
				h.S = value
			case 2:
				h.V = value
			}
			t.Run(fmt.Sprintf("%s/%d", field, j), func(t *testing.T) {
				requireColorError(t, h.Validate(), colorkit.ErrRange, field)
				got, e := h.RGB8()
				if got != (colorkit.RGB8{}) {
					t.Fatal("partial invalid result", got)
				}
				requireColorError(t, e, colorkit.ErrRange, field)
			})
		}
	}
	for i, field := range fields {
		for _, token := range []string{"", " ", "-0", "-1", "+1", ".5", "1.", "01", "1e2", "NaN", "Inf", "0.0000001", "０", "0\x00", strings.Repeat("9", 1024)} {
			parts := []string{"10", "20", "30"}
			parts[i] = token
			got, e := colorkit.FromHSVDecimal(parts[0], parts[1], parts[2])
			if got != (colorkit.RGB8{}) {
				t.Fatal("partial decimal result", field, token, got)
			}
			requireColorError(t, e, colorkit.ErrSyntax, field)
		}
		parts := []string{"10", "20", "30"}
		if i == 0 {
			parts[i] = "360.000001"
		} else {
			parts[i] = "100.000001"
		}
		got, e := colorkit.FromHSVDecimal(parts[0], parts[1], parts[2])
		if got != (colorkit.RGB8{}) {
			t.Fatal(got)
		}
		requireColorError(t, e, colorkit.ErrRange, field)
	}
}

func TestHSVPrecisionNormalizationAndNegativeZero(t *testing.T) {
	cases := []struct {
		name string
		h    colorkit.HSV
		want string
	}{
		{"hue_rounds_down", colorkit.HSV{H: 29.99999949, S: 100, V: 100}, "#ff7f00"},
		{"hue_rounds_up", colorkit.HSV{H: 29.99999951, S: 100, V: 100}, "#ff8000"},
		{"gray_rounds_down", colorkit.HSV{V: 49.99999949}, "#7f7f7f"},
		{"gray_rounds_up", colorkit.HSV{V: 49.99999951}, "#808080"},
		{"wrap_normalized_360", colorkit.HSV{H: 359.9999999, S: 100, V: 100}, "#ff0000"},
		{"degree_360", colorkit.HSV{H: 360, S: 100, V: 100}, "#ff0000"},
	}
	for _, x := range cases {
		t.Run(x.name, func(t *testing.T) {
			got, e := x.h.RGB8()
			if e != nil || got.Hex() != x.want {
				t.Fatal(got, e, x.want)
			}
		})
	}
	negativeZero := math.Copysign(0, -1)
	for _, h := range []colorkit.HSV{{H: negativeZero}, {S: negativeZero}, {V: negativeZero}, {H: negativeZero, S: negativeZero, V: negativeZero}, {}} {
		got, e := h.RGB8()
		if e != nil || got != (colorkit.RGB8{}) || h.Validate() != nil {
			t.Fatal(h, got, e)
		}
	}
}

func TestRGBHSVReverseAndRoundTrip(t *testing.T) {
	for _, x := range []struct {
		rgb colorkit.RGB8
		hsv colorkit.HSV
	}{{colorkit.RGB8{R: 255}, colorkit.HSV{S: 100, V: 100}}, {colorkit.RGB8{G: 255}, colorkit.HSV{H: 120, S: 100, V: 100}}, {colorkit.RGB8{B: 255}, colorkit.HSV{H: 240, S: 100, V: 100}}, {colorkit.RGB8{}, colorkit.HSV{}}, {colorkit.RGB8{R: 255, G: 255, B: 255}, colorkit.HSV{V: 100}}} {
		if got := x.rgb.HSV(); got != x.hsv {
			t.Fatal(x.rgb, got, x.hsv)
		}
	}
	for v := 0; v < 256; v++ {
		h := (colorkit.RGB8{R: uint8(v), G: uint8(v), B: uint8(v)}).HSV()
		if h.H != 0 || h.S != 0 || math.Abs(h.V-float64(v)*100/255) > 1e-12 {
			t.Fatal("achromatic inverse", v, h)
		}
	}
	check := func(c colorkit.RGB8) {
		h := c.HSV()
		if err := h.Validate(); err != nil || h.H < 0 || h.H >= 360 {
			t.Fatal("invalid reverse", c, h, err)
		}
		back, e := h.RGB8()
		if e != nil || back != c {
			t.Fatal("RGB roundtrip", c, h, back, e)
		}
	}
	grid := []uint8{0, 1, 17, 63, 64, 127, 128, 191, 192, 254, 255}
	for _, r := range grid {
		for _, g := range grid {
			for _, b := range grid {
				check(colorkit.RGB8{R: r, G: g, B: b})
			}
		}
	}
	// Explicit fixed-seed integer samples, independent of math/rand version.
	state := uint32(0x1bad5eed)
	next := func() uint8 { state = 1664525*state + 1013904223; return uint8(state >> 24) }
	for i := 0; i < 2048; i++ {
		check(colorkit.RGB8{R: next(), G: next(), B: next()})
	}
}

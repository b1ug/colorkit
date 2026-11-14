// wasmvectors is an internal acceptance executable, not a user-facing CLI.
// It validates fixed expectations before writing deterministic stdout. It has
// no filesystem, device, network, clock or random input.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/b1ug/colorkit"
	"github.com/b1ug/colorkit/internal/testvectors"
)

func ensure(ok bool, context string) {
	if !ok {
		panic(context)
	}
}

type fixedColor struct{ r, g, b, a uint32 }

func (c fixedColor) RGBA() (uint32, uint32, uint32, uint32) { return c.r, c.g, c.b, c.a }

func main() {
	names := colorkit.Names()
	ensure(len(names) == len(testvectors.Named), "name count")
	for i, x := range testvectors.Named {
		c, ok := colorkit.LookupName(x.Name)
		want := colorkit.RGB8{R: x.R, G: x.G, B: x.B}
		ensure(ok && c == want && c.Hex() == x.Hex && names[i] == x.Name, "name "+x.Name)
		fmt.Println("name", x.Name, c.Hex())
	}
	for _, x := range testvectors.HSV {
		c, err := colorkit.FromHSVDecimal(x.H, x.S, x.V)
		ensure(err == nil && c.Hex() == x.Hex, "decimal "+x.Name)
		h, _ := strconv.ParseFloat(x.H, 64)
		s, _ := strconv.ParseFloat(x.S, 64)
		v, _ := strconv.ParseFloat(x.V, 64)
		numeric, err := (colorkit.HSV{H: h, S: s, V: v}).RGB8()
		ensure(err == nil && numeric == c, "numeric "+x.Name)
		fmt.Println("hsv", x.Name, c.Hex())
	}
	for i, h := range []colorkit.HSV{{H: math.NaN()}, {S: math.Inf(1)}, {V: math.Inf(-1)}, {H: 361}, {S: -1}, {V: 101}} {
		c, err := h.RGB8()
		var detail *colorkit.Error
		ensure(c == (colorkit.RGB8{}) && errors.Is(err, colorkit.ErrRange) && errors.As(err, &detail), "invalid HSV")
		fmt.Println("invalid-hsv", i, detail.Field)
	}
	for _, x := range []struct{ input, want string }{{"#abc", "#aabbcc"}, {" #AbCdEf\t", "#abcdef"}, {" ReD ", "#ff0000"}} {
		c, err := colorkit.Parse(x.input)
		ensure(err == nil && c.Hex() == x.want, "literal")
		fmt.Println("literal", c.Hex())
	}
	for _, x := range []struct {
		in   fixedColor
		want colorkit.RGB8
	}{{fixedColor{r: 129, g: 32896, b: 65406, a: 65535}, colorkit.RGB8{R: 1, G: 128, B: 254}}, {fixedColor{r: 128, g: 65407, a: 65535}, colorkit.RGB8{G: 255}}} {
		c, err := colorkit.FromColor(x.in)
		ensure(err == nil && c == x.want, "FromColor")
		fmt.Println("from-color", c.Hex())
	}
	for _, c := range []colorkit.RGB8{{}, {R: 255}, {G: 255}, {B: 255}, {R: 17, G: 63, B: 192}, {R: 128, G: 128, B: 128}} {
		back, err := c.HSV().RGB8()
		ensure(err == nil && back == c, "reverse")
		fmt.Println("roundtrip", back.Hex())
	}
	c := colorkit.RGB8{R: 170, G: 187, B: 204}
	raw, err := json.Marshal(c)
	ensure(err == nil && string(raw) == `"#aabbcc"`, "standard JSON")
	var decoded colorkit.RGB8
	err = json.Unmarshal(raw, &decoded)
	ensure(err == nil && decoded == c, "standard JSON decode")
	before := decoded
	err = decoded.UnmarshalText([]byte("invalid"))
	ensure(errors.Is(err, colorkit.ErrSyntax) && decoded == before, "atomic text")
	fmt.Println("json", string(raw))
	fmt.Println("atomic-text", decoded.Hex())
}

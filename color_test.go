package colorkit_test

import (
	"bufio"
	"crypto/sha256"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/b1ug/colorkit"
	"github.com/b1ug/colorkit/internal/testvectors"
)

var _ color.Color = colorkit.RGB8{}
var _ encoding.TextMarshaler = colorkit.RGB8{}
var _ encoding.TextUnmarshaler = (*colorkit.RGB8)(nil)

func TestFrozenCSSFixture(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/css-named-2026.tsv")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "521686f82ac6c649de3f22589783ac86e2bfeb0c1858fd953f59efc4e1d2fd09" {
		t.Fatalf("fixture source hash changed: %s", got)
	}
	var names []string
	unique := map[colorkit.RGB8]bool{}
	scan := bufio.NewScanner(strings.NewReader(string(raw)))
	index := 0
	for scan.Scan() {
		f := strings.Fields(scan.Text())
		if len(f) != 5 {
			t.Fatalf("invalid fixture row %q", scan.Text())
		}
		var channels [3]uint8
		for i := range channels {
			v, e := strconv.Atoi(f[i+2])
			if e != nil || v < 0 || v > 255 {
				t.Fatal(f)
			}
			channels[i] = uint8(v)
		}
		want := colorkit.RGB8{R: channels[0], G: channels[1], B: channels[2]}
		got, ok := colorkit.LookupName(f[0])
		if !ok || got != want || got.Hex() != f[1] {
			t.Fatalf("%s: got %v/%t, want %v/%s", f[0], got, ok, want, f[1])
		}
		parsed, e := colorkit.Parse(f[0])
		if e != nil || parsed != want {
			t.Fatal(f[0], parsed, e)
		}
		if index >= len(testvectors.Named) {
			t.Fatal("compiled fixture too short")
		}
		v := testvectors.Named[index]
		if v.Name != f[0] || v.Hex != f[1] || v.R != channels[0] || v.G != channels[1] || v.B != channels[2] {
			t.Fatal("compiled fixture differs", index)
		}
		names = append(names, f[0])
		unique[got] = true
		index++
	}
	if err = scan.Err(); err != nil {
		t.Fatal(err)
	}
	if len(names) != 148 || len(unique) != 139 || len(testvectors.Named) != 148 {
		t.Fatal(len(names), len(unique))
	}
	if !reflect.DeepEqual(names, colorkit.Names()) {
		t.Fatal("Names differs from frozen complete sorted list")
	}
}

func TestParseAndHexBoundaries(t *testing.T) {
	for input, want := range map[string]string{"#abc": "#aabbcc", "#AbC": "#aabbcc", "#0012fF": "#0012ff", " \t\r\n\v\f#AbC \t": "#aabbcc", " ReD ": "#ff0000", strings.Repeat(" ", 61) + "red": "#ff0000"} {
		got, err := colorkit.Parse(input)
		if err != nil || got.Hex() != want {
			t.Fatalf("Parse(%q)=%v,%v", input, got, err)
		}
	}
	for input, want := range map[string]string{"#abc": "#aabbcc", " #ABC\t": "#aabbcc", "#0012fF": "#0012ff", strings.Repeat(" ", 60) + "#abc": "#aabbcc"} {
		got, err := colorkit.ParseHex(input)
		if err != nil || got.Hex() != want {
			t.Fatal(input, got, err)
		}
	}
	invalid := []string{"", " ", "#", "#ab", "#abcd", "#abcde", "#aabbccdd", "#ggg", "0xff0000", "rgb(255,0,0)", "set red", "red blue", "red\x00", "r ed", "\u00a0red", "red\u2003", "réd", string([]byte{0xff}), strings.Repeat(" ", 62) + "red", strings.Repeat(" ", 61) + "#abc"}
	for _, input := range invalid {
		got, err := colorkit.Parse(input)
		if got != (colorkit.RGB8{}) {
			t.Fatal("partial result", input, got)
		}
		requireColorError(t, err, colorkit.ErrSyntax, "")
	}
	for _, input := range []string{"red", "abc", "#ab", "#xxxxxx", "\u00a0#abc", strings.Repeat(" ", 61) + "#abc"} {
		got, err := colorkit.ParseHex(input)
		if got != (colorkit.RGB8{}) {
			t.Fatal(got)
		}
		requireColorError(t, err, colorkit.ErrSyntax, "")
	}
	for _, input := range []string{"#abc", "transparent", "currentColor", "mint", "apricot", "red\x00", "\u00a0red", strings.Repeat(" ", 62) + "red"} {
		if got, ok := colorkit.LookupName(input); ok || got != (colorkit.RGB8{}) {
			t.Fatal(input, got, ok)
		}
	}
	if got, ok := colorkit.LookupName(" \tReD\r\n"); !ok || got != (colorkit.RGB8{R: 255}) {
		t.Fatal(got, ok)
	}
}

func TestAliasesNamesCopiesAndConcurrentReads(t *testing.T) {
	aliases := [][2]string{{"aqua", "cyan"}, {"fuchsia", "magenta"}, {"gray", "grey"}, {"darkgray", "darkgrey"}, {"darkslategray", "darkslategrey"}, {"dimgray", "dimgrey"}, {"lightgray", "lightgrey"}, {"lightslategray", "lightslategrey"}, {"slategray", "slategrey"}}
	for _, pair := range aliases {
		a, ok := colorkit.LookupName(pair[0])
		b, ok2 := colorkit.LookupName(pair[1])
		if !ok || !ok2 || a != b {
			t.Fatal(pair)
		}
	}
	names := colorkit.Names()
	if !sort.StringsAreSorted(names) {
		t.Fatal("unsorted")
	}
	names[0] = "mutated"
	if colorkit.Names()[0] != "aliceblue" {
		t.Fatal("Names leaked writable storage")
	}
	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 40; j++ {
				n := colorkit.Names()
				if len(n) != 148 || n[0] != "aliceblue" {
					errs <- "Names"
					return
				}
				n[0] = "local"
				c, ok := colorkit.LookupName("red")
				if !ok || c != (colorkit.RGB8{R: 255}) {
					errs <- "LookupName"
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

func TestStandardInterfacesAndAtomicText(t *testing.T) {
	if _, ok := interface{}(colorkit.HSV{}).(color.Color); ok {
		t.Fatal("HSV must not silently implement non-error Color")
	}
	for n := 0; n < 256; n++ {
		c := colorkit.RGB8{R: uint8(n), G: uint8(255 - n), B: uint8(n ^ 85)}
		r, g, b, a := c.RGBA()
		if r != uint32(c.R)*257 || g != uint32(c.G)*257 || b != uint32(c.B)*257 || a != 65535 {
			t.Fatal(c, r, g, b, a)
		}
		text, e := c.MarshalText()
		if e != nil || string(text) != c.Hex() {
			t.Fatal(c, text, e)
		}
		var decoded colorkit.RGB8
		if e = decoded.UnmarshalText(text); e != nil || decoded != c {
			t.Fatal(c, decoded, e)
		}
	}
	c := colorkit.RGB8{R: 170, G: 187, B: 204}
	encoded, e := json.Marshal(c)
	if e != nil || string(encoded) != `"#aabbcc"` {
		t.Fatal(string(encoded), e)
	}
	for _, text := range []string{`"#abc"`, `"#AaBbCc"`} {
		var out colorkit.RGB8
		if e = json.Unmarshal([]byte(text), &out); e != nil || out != c {
			t.Fatal(text, out, e)
		}
	}
	for _, bad := range []string{"oops", "#xyz", "#ab", strings.Repeat(" ", 65)} {
		before := c
		if e = c.UnmarshalText([]byte(bad)); e == nil || c != before {
			t.Fatal("failed text mutation", bad, c, e)
		}
		requireColorError(t, e, colorkit.ErrSyntax, "")
	}
	before := c
	for _, bad := range []string{`"oops"`, `123`, `{"r":1}`, `[1,2,3]`} {
		if e = json.Unmarshal([]byte(bad), &c); e == nil || c != before {
			t.Fatal("failed standard JSON mutation", bad, c, e)
		}
	}
	var nilReceiver *colorkit.RGB8
	requireColorError(t, nilReceiver.UnmarshalText([]byte("red")), colorkit.ErrSyntax, "receiver")
	if (colorkit.RGB8{}).Hex() != "#000000" {
		t.Fatal("zero RGB8 must be black")
	}
}

type fixedColor struct{ r, g, b, a uint32 }

func (c fixedColor) RGBA() (uint32, uint32, uint32, uint32) { return c.r, c.g, c.b, c.a }

type nilColorMap map[string]int

func (c nilColorMap) RGBA() (uint32, uint32, uint32, uint32) { panic("nil map must not be called") }

func TestFromColorNilAlphaAndRanges(t *testing.T) {
	var typedNil *color.NRGBA
	var mapNil nilColorMap
	for _, input := range []color.Color{nil, typedNil, mapNil} {
		got, e := colorkit.FromColor(input)
		if got != (colorkit.RGB8{}) {
			t.Fatal(got)
		}
		requireColorError(t, e, colorkit.ErrSyntax, "color")
	}
	for a := uint32(0); a < 65535; a++ {
		got, e := colorkit.FromColor(fixedColor{a: a})
		if got != (colorkit.RGB8{}) || !errors.Is(e, colorkit.ErrAlpha) {
			t.Fatal("accepted alpha", a, got, e)
		}
		if a == 0 || a == 1 || a == 65534 {
			requireColorError(t, e, colorkit.ErrAlpha, "a")
		}
	}
	for i, input := range []fixedColor{{r: 65536, a: 65535}, {g: 65536, a: 65535}, {b: 65536, a: 65535}, {a: 65536}, {r: 0xffffffff, a: 0}} {
		got, e := colorkit.FromColor(input)
		if got != (colorkit.RGB8{}) {
			t.Fatal(got)
		}
		field := []string{"r", "g", "b", "a", "r"}[i]
		requireColorError(t, e, colorkit.ErrRange, field)
	}
	for _, input := range []color.Color{color.RGBA{R: 255, G: 128, A: 255}, color.NRGBA{R: 255, G: 128, A: 255}} {
		got, e := colorkit.FromColor(input)
		if e != nil || got != (colorkit.RGB8{R: 255, G: 128}) {
			t.Fatal(got, e)
		}
	}
}

func TestFromColorNearest16BitQuantization(t *testing.T) {
	// A nearest 8-bit result is uniquely within 128 of a 16-bit sample because
	// adjacent 8-bit expansions are 257 apart. This is an independent property.
	for v := uint32(0); v <= 65535; v++ {
		c, e := colorkit.FromColor(fixedColor{r: v, g: v, b: v, a: 65535})
		if e != nil || c.R != c.G || c.R != c.B {
			t.Fatal(v, c, e)
		}
		distance := int(v) - int(c.R)*257
		if distance < 0 {
			distance = -distance
		}
		if distance > 128 {
			t.Fatal("not nearest", v, c, distance)
		}
	}
	for _, x := range []struct {
		value uint32
		want  uint8
	}{{0, 0}, {128, 0}, {129, 1}, {32767, 127}, {32768, 128}, {65406, 254}, {65407, 255}, {65535, 255}} {
		c, e := colorkit.FromColor(color.Gray16{Y: uint16(x.value)})
		if e != nil || c != (colorkit.RGB8{R: x.want, G: x.want, B: x.want}) {
			t.Fatal(x, c, e)
		}
	}
	c, e := colorkit.FromColor(color.NRGBA64{R: 129, G: 32896, B: 65406, A: 65535})
	if e != nil || c != (colorkit.RGB8{R: 1, G: 128, B: 254}) {
		t.Fatal(c, e)
	}
}

package svg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/wzshiming/democtl/pkg/color"
	"github.com/wzshiming/democtl/pkg/minify"
	"github.com/wzshiming/democtl/pkg/renderer"
)

type canvas struct {
	svg      *svg.SVG
	output   io.WriteCloser
	noWindow bool
	getColor func(i int) string

	width, height int

	offsets []time.Duration

	stylesIndex   map[string]string
	stylesContent map[string]string

	defsIndex   map[string]string
	defsContent map[string]func()

	bold      bool
	italic    bool
	underline bool
	blink     bool
}

const (
	rowHeight = 30
	colWidth  = 12
	padding   = 20
)

func NewCanvas(output io.Writer, noWindow bool) renderer.Renderer {
	return &canvas{
		output:   minify.SVGWithWriter(output),
		noWindow: noWindow,
		getColor: color.DefaultColors().GetColorForSVG,
	}
}

func (c *canvas) Initialize(ctx context.Context, x, y int, width, height int) error {
	c.width = width
	c.height = height

	c.svg = svg.New(c.output)

	c.svg.Start(c.paddingRight(), c.paddingBottom())
	c.createWindow()

	c.svg.Group(fmt.Sprintf(`class="main"`))
	return nil
}

func (c *canvas) Finish(ctx context.Context) error {
	err := c.addStyles()
	if err != nil {
		return err
	}
	err = c.addDefs()
	if err != nil {
		return err
	}

	c.svg.Gend()
	c.svg.End()

	err = c.output.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *canvas) Frame(ctx context.Context, index int, offset time.Duration) (renderer.Frame, error) {
	c.offsets = append(c.offsets, offset)
	c.svg.Gtransform(fmt.Sprintf("translate(%d)", c.paddingRight()*index))
	return &frame{
		canvas:    c,
		heightOff: c.paddingTop(),
		widthOff:  c.paddingLeft(),
		finish: func() error {
			c.svg.Gend()
			return nil
		},
	}, nil
}

func (c *canvas) getFG(fg string) string {
	if c.stylesIndex == nil {
		c.stylesIndex = map[string]string{}
		c.stylesContent = map[string]string{}
	}
	id, ok := c.stylesIndex[fg]
	if ok {
		return id
	}
	id = encodeIndex(uint64(len(c.stylesContent)))

	c.stylesIndex[fg] = id

	c.stylesContent[id] = fmt.Sprintf(`
.%s {
  fill: %s;
}
	`, id, fg)

	return id
}

func (c *canvas) getBG(bg string) string {
	if c.defsIndex == nil {
		c.defsIndex = map[string]string{}
		c.defsContent = map[string]func(){}
	}

	id, ok := c.defsIndex[bg]
	if ok {
		return id
	}
	id = encodeIndex(uint64(len(c.defsContent)))

	c.defsIndex[bg] = id

	c.defsContent[id] = func() {
		c.svg.Filter(id)
		defer c.svg.Fend()

		c.svg.FeFlood(svg.Filterspec{Result: "bg"}, bg, 1.0)
		c.svg.FeMerge([]string{`bg`, `SourceGraphic`})
	}

	return id
}

func (c *canvas) paddingLeft() int {
	return padding
}

func (c *canvas) paddingRight() int {
	return (c.width+2)*colWidth + padding
}

func (c *canvas) paddingTop() int {
	if c.noWindow {
		return padding
	}
	return padding + padding/2
}

func (c *canvas) paddingBottom() int {
	return (c.height)*rowHeight + c.paddingTop()
}

func (c *canvas) createWindow() {
	if c.noWindow {
		c.svg.Rect(0, 0, c.paddingRight(), c.paddingBottom(), "fill:"+c.getColor(int(renderer.DefaultBG)))
		return
	}
	windowRadius := 5
	buttonRadius := 7
	buttonColors := [3]string{"#ff5f58", "#ffbd2e", "#18c132"}
	c.svg.Roundrect(0, 0, c.paddingRight(), c.paddingBottom(), windowRadius, windowRadius, "fill:"+c.getColor(int(renderer.DefaultBG)))
	for i := range buttonColors {
		c.svg.Circle((i*(padding+buttonRadius/2))+padding, padding, buttonRadius, fmt.Sprintf("fill:%s", buttonColors[i]))
	}
}

func (c *canvas) addStyles() error {
	styles := []string{}

	keys := make([]string, 0, len(c.stylesContent))
	for k := range c.stylesContent {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		styles = append(styles, c.stylesContent[key])
	}

	if c.bold {
		styles = append(styles, `
.bold {
  font-weight: bold;
}
`)
	}
	if c.italic {
		styles = append(styles, `
.italic {
  font-style: italic;
}
`)
	}
	if c.underline {
		styles = append(styles, `
.underline {
  text-decoration: underline;
}
`)
	}

	if c.blink {
		styles = append(styles, `
.blink {
  animation: blink-animation 1s steps(2, start) infinite;
}
@keyframes blink-animation {
  to {
    visibility: hidden;
  }
}
`)
	}

	styles = append(styles,
		fmt.Sprintf(`
text {
  fill: %s;
}
`, c.getColor(int(renderer.DefaultFG))),
	)

	styles = append(styles,
		fmt.Sprintf(`
.cursor {
  fill: %s;
  opacity: 0.8;
}
`, c.getColor(int(renderer.DefaultCursor))),
	)

	styles = append(styles,
		fmt.Sprintf(`
.main {
  animation-duration: %.2fs;
  animation-iteration-count: infinite;
  animation-name: k;
  animation-timing-function: steps(1,end);
  font-family: Monaco,Consolas,Menlo,'Bitstream Vera Sans Mono','Powerline Symbols',monospace;
  font-size: 20px;
}
`, float64(c.offsets[len(c.offsets)-1])/float64(time.Second)),
	)

	styles = append(styles, generateKeyframes(c.offsets, int32(c.paddingRight())))

	allCss, err := minify.CSSWithString(strings.Join(styles, "\n"))
	if err != nil {
		return err
	}
	c.svg.Style("text/css", allCss)
	return nil
}

func (c *canvas) addDefs() error {
	c.svg.Def()
	defer c.svg.DefEnd()

	keys := make([]string, 0, len(c.defsContent))
	for k := range c.defsContent {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		c.defsContent[key]()
	}
	return nil
}

func (c *canvas) toGlyph(fg, bg renderer.Color, mode int16) []string {
	classes := []string{}
	filters := []string{}

	if fg != renderer.DefaultFG {
		classes = append(classes, c.getFG(c.getColor(int(fg))))
	}

	if bg != renderer.DefaultBG {
		filters = append(filters, fmt.Sprintf(`url(#%s)`, c.getBG(c.getColor(int(bg)))))
	}

	if mode&0b00000010 != 0 {
		classes = append(classes, `underline`)
		c.underline = true
	}
	if mode&0b00000100 != 0 {
		classes = append(classes, `bold`)
		c.bold = true
	}
	if mode&0b00010000 != 0 {
		classes = append(classes, `italic`)
		c.italic = true
	}
	if mode&0b00100000 != 0 {
		classes = append(classes, `blink`)
		c.blink = true
	}

	out := []string{}
	if len(classes) != 0 {
		out = append(out, fmt.Sprintf(`class=%q`, strings.Join(classes, " ")))
	}
	if len(filters) != 0 {
		out = append(out, fmt.Sprintf(`filter=%q`, strings.Join(filters, " ")))
	}

	return out
}

func generateKeyframes(offsets []time.Duration, width int32) string {
	dur := offsets[len(offsets)-1]
	buf := bytes.NewBuffer(nil)
	buf.WriteString("@keyframes k {")
	for i, offset := range offsets {
		fmt.Fprintf(buf, "%.3f%%{transform:translateX(-%dpx)}", float32(offset)*100/float32(dur), width*int32(i))
	}
	buf.WriteString("}")
	return buf.String()
}

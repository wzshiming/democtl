package svg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/wzshiming/democtl/pkg/renderer"
	"github.com/wzshiming/vt10x"
)

type canvas struct {
	output   io.Writer
	noWindow bool
	getColor func(i vt10x.Color) string

	width, height int

	offsets []time.Duration

	stylesIndex   map[string]string
	stylesContent []string

	defsIndex   map[string]string
	defsContent []string
}

const (
	rowHeight = 30
	colWidth  = 12
	padding   = 20
)

func NewCanvas(output io.Writer, noWindow bool, getColor func(i vt10x.Color) string) renderer.Renderer {
	return &canvas{
		output:   newMinifyWriter(output),
		noWindow: noWindow,
		getColor: getColor,
	}
}

func (c *canvas) Initialize(ctx context.Context, x, y int, width, height int) error {
	c.width = width
	c.height = height

	fmt.Fprintf(c.output, `<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, c.paddingRight(), c.paddingBottom())

	c.createWindow()

	fmt.Fprintf(c.output, `<g>`)

	return nil
}

func (c *canvas) Finish(ctx context.Context) error {

	fmt.Fprintf(c.output, `</g>`)

	err := c.addDefs()
	if err != nil {
		return err
	}

	err = c.addStyles()
	if err != nil {
		return err
	}

	fmt.Fprintf(c.output, `</svg>`)
	return nil
}

func (c *canvas) Frame(ctx context.Context, index int, offset time.Duration) (renderer.Frame, error) {
	c.offsets = append(c.offsets, offset)
	fmt.Fprintf(c.output, `<svg x="%d">`, c.paddingRight()*index)
	return &frame{
		canvas:    c,
		heightOff: c.paddingTop(),
		widthOff:  c.paddingLeft(),
		finish: func() error {
			fmt.Fprintf(c.output, `</svg>`)
			return nil
		},
	}, nil
}

func (c *canvas) getFG(fg string) string {
	return c.getStyles(fg, func(id string) string {
		return fmt.Sprintf(`
.%s {
  fill: %s;
}
`, id, fg)
	})
}

func (c *canvas) getBG(bg string) string {
	return c.getDefs(bg, func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<filter id="%s">
<feFlood result="bg" flood-color="%s"/>
<feMerge>
<feMergeNode in="bg"/>
<feMergeNode in="SourceGraphic"/>
</feMerge>
</filter>
`, id, bg)
		return buf.String()
	})
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
		fmt.Fprintf(c.output, `<rect width="%d" height="%d" style="fill:%s"/>`,
			c.paddingRight(), c.paddingBottom(), c.getColor(vt10x.DefaultBG))
		return
	}
	windowRadius := 5
	buttonRadius := 7
	buttonColors := [3]string{"#ff5f58", "#ffbd2e", "#18c132"}

	fmt.Fprintf(c.output, `<rect width="%d" height="%d" rx="%d" ry="%d" style="fill:%s"/>`,
		c.paddingRight(), c.paddingBottom(), windowRadius, windowRadius, c.getColor(vt10x.DefaultBG))
	for i := range buttonColors {
		fmt.Fprintf(c.output, `<circle cx="%d" cy="%d" r="%d" style="fill:%s"/>`,
			(i*(padding+buttonRadius/2))+padding, padding, buttonRadius, buttonColors[i],
		)
	}
}

func (c *canvas) addStyles() error {
	styles := []string{}
	for _, content := range c.stylesContent {
		styles = append(styles, content)
	}

	styles = append(styles,
		fmt.Sprintf(`
symbol {
  overflow: visible;
}
`),
	)

	styles = append(styles,
		fmt.Sprintf(`
text {
  font-family: Monaco,Consolas,Menlo,monospace;
  font-size: 20px;
  dominant-baseline: hanging;
  text-anchor: start;
  fill: %s;
}
`, c.getColor(vt10x.DefaultFG)),
	)

	styles = append(styles,
		fmt.Sprintf(`
g {
  animation-duration: %.2fs;
  animation-iteration-count: infinite;
  animation-name: k;
  animation-timing-function: steps(1,end);
}
`, float64(c.offsets[len(c.offsets)-1])/float64(time.Second)),
	)

	styles = append(styles, generateKeyframes(c.offsets, int32(c.paddingRight())))

	fmt.Fprintf(c.output, `<style type="text/css">`)
	defer fmt.Fprintf(c.output, `</style>`)

	s, err := minifyCSS(strings.Join(styles, ""))
	if err != nil {
		return err
	}
	c.output.Write([]byte(s))
	return nil
}

func (c *canvas) addDefs() error {
	fmt.Fprintf(c.output, `<defs>`)
	defer fmt.Fprintf(c.output, `</defs>`)

	for _, d := range c.defsContent {
		c.output.Write([]byte(d))
	}
	return nil
}

func (c *canvas) useDef(id string, x, y int) {
	fmt.Fprintf(c.output, `<use href="#%s" x="%d" y="%d"/>`, id, x, y)
}

func (c *canvas) getDefs(unique string, f func(id string) string) string {
	if c.defsIndex == nil {
		c.defsIndex = map[string]string{}
	}

	id, ok := c.defsIndex[unique]
	if ok {
		return id
	}
	id = encodeIndex(uint64(len(c.defsContent)))

	c.defsIndex[unique] = id

	c.defsContent = append(c.defsContent, f(id))

	return id
}

func (c *canvas) getStyles(unique string, f func(id string) string) string {
	if c.stylesIndex == nil {
		c.stylesIndex = map[string]string{}
	}
	id, ok := c.stylesIndex[unique]
	if ok {
		return id
	}
	id = encodeIndex(uint64(len(c.stylesContent)))

	c.stylesIndex[unique] = id

	c.stylesContent = append(c.stylesContent, f(id))

	return id
}

func (c *canvas) toGlyph(fg, bg vt10x.Color, mode vt10x.AttrFlag) []string {
	classes := []string{}
	filters := []string{}

	if mode&vt10x.AttrReverse != 0 {
		fg, bg = bg, fg
	}

	if fg != vt10x.DefaultFG {
		classes = append(classes, c.getFG(c.getColor(fg)))
	}

	if bg != vt10x.DefaultBG {
		filters = append(filters, fmt.Sprintf(`url(#%s)`, c.getBG(c.getColor(bg))))
	}

	if mode&vt10x.AttrUnderline != 0 {
		if mode&vt10x.AttrStrike != 0 {
			id := c.getStyles("underline-strike", func(id string) string {
				return fmt.Sprintf(`
.%s {
  text-decoration: underline line-through;
}
`, id)
			})
			classes = append(classes, id)
		} else {
			id := c.getStyles("underline", func(id string) string {
				return fmt.Sprintf(`
.%s {
  text-decoration: underline;
}
`, id)
			})
			classes = append(classes, id)
		}
	} else {
		if mode&vt10x.AttrStrike != 0 {
			id := c.getStyles("strike", func(id string) string {
				return fmt.Sprintf(`
.%s {
  text-decoration: line-through;
}
`, id)
			})
			classes = append(classes, id)
		}
	}

	if mode&vt10x.AttrDim != 0 {
		id := c.getDefs("dim", func(id string) string {
			buf := bytes.NewBuffer(nil)
			fmt.Fprintf(buf, `
<filter id="%s">
<feComponentTransfer>
  <feFuncR type="linear" slope=".5" intercept="0"/>
  <feFuncG type="linear" slope=".5" intercept="0"/>
  <feFuncB type="linear" slope=".5" intercept="0"/>
</feComponentTransfer>
</filter>
`, id)
			return buf.String()
		})
		filters = append(filters, fmt.Sprintf(`url(#%s)`, id))
	}
	if mode&vt10x.AttrBold != 0 {
		id := c.getStyles("bold", func(id string) string {
			return fmt.Sprintf(`
.%s {
  font-weight: bold;
}
`, id)
		})
		classes = append(classes, id)
	}
	if mode&vt10x.AttrItalic != 0 {
		id := c.getStyles("italic", func(id string) string {
			return fmt.Sprintf(`
.%s {
  font-style: italic;
}
`, id)
		})
		classes = append(classes, id)
	}
	if mode&vt10x.AttrBlink != 0 {
		id := c.getStyles("blink", func(id string) string {
			return fmt.Sprintf(`
.%s {
  animation: b 1s steps(2, start) infinite;
}
@keyframes b {
  to {
    visibility: hidden;
  }
}
`, id)
		})
		classes = append(classes, id)
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

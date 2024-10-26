package svg

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	svg "github.com/ajstarks/svgo"
	"github.com/hinshun/vt10x"
	"github.com/wzshiming/democtl/pkg/cast"
	"github.com/wzshiming/democtl/pkg/color"
	"github.com/wzshiming/democtl/pkg/minify"
)

type Canvas struct {
	svg      *svg.SVG
	header   cast.Header
	events   []cast.Event
	getColor func(i int) string
	noWindow bool
}

const (
	rowHeight = 30
	colWidth  = 12
	padding   = 20
)

func NewCanvas() *Canvas {
	return &Canvas{
		getColor: color.DefaultColors().GetColorForSVG,
	}
}

func (c *Canvas) Run(input io.Reader, output io.Writer, noWindow bool) error {
	decoder := cast.NewDecoder(input)

	header, err := decoder.DecodeHeader()
	if err != nil {
		return err
	}

	var events []cast.Event
	for {
		event, err := decoder.DecodeEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		events = append(events, event)
	}
	c.noWindow = noWindow
	c.header = header
	c.events = events
	c.svg = svg.New(output)

	c.svg.Start(c.paddingRight(), c.paddingBottom())
	defer c.svg.End()

	err = c.addDefine()
	if err != nil {
		return err
	}

	c.createWindow()

	err = c.frames()
	if err != nil {
		return err
	}
	return nil
}

func (c *Canvas) paddingLeft() int {
	return padding
}

func (c *Canvas) paddingRight() int {
	return (c.header.Width+2)*colWidth + padding
}

func (c *Canvas) paddingTop() int {
	if c.noWindow {
		return padding
	}
	return padding + padding/2
}

func (c *Canvas) paddingBottom() int {
	return (c.header.Height)*rowHeight + c.paddingTop()
}

func (c *Canvas) createWindow() {
	bg := c.getColor(int(vt10x.DefaultBG))
	if c.noWindow {
		c.svg.Rect(0, 0, c.paddingRight(), c.paddingBottom(), "fill:"+bg)
		return
	}
	windowRadius := 5
	buttonRadius := 7
	buttonColors := [3]string{"#ff5f58", "#ffbd2e", "#18c132"}
	c.svg.Roundrect(0, 0, c.paddingRight(), c.paddingBottom(), windowRadius, windowRadius, "fill:"+bg)
	for i := range buttonColors {
		c.svg.Circle((i*(padding+buttonRadius/2))+padding, padding, buttonRadius, fmt.Sprintf("fill:%s", buttonColors[i]))
	}
}

func (c *Canvas) addDefine() error {
	styles := []string{}
	styles = append(styles, generateKeyframes(c.events, int32(c.paddingRight())))

	for i := 0; i != 16; i++ {
		styles = append(styles,
			fmt.Sprintf(`
.%s {
  fill: %s;
}
`, colorID(i), c.getColor(i)),
		)
	}

	styles = append(styles, `
.bold {
  font-weight: bold;
}
.italic {
  font-style: italic;
}
.underline {
  text-decoration: underline;
}
.blink {
  animation: blink-animation 1s steps(2, start) infinite;
}
@keyframes blink-animation {
  to {
    visibility: hidden;
  }
}
`)

	styles = append(styles,
		fmt.Sprintf(`
text {
  fill: %s;
}
`, c.getColor(int(vt10x.DefaultFG)),
		),
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
`, c.events[len(c.events)-1].Time),
	)

	styles = append(styles,
		fmt.Sprintf(`
.cursor {
  fill: %s;
  opacity: 0.8;
}
`, c.getColor(int(vt10x.DefaultCursor)),
		),
	)

	allCss, err := minify.CSSWithString(strings.Join(styles, "\n"))
	if err != nil {
		return err
	}
	c.svg.Style("text/css", allCss)

	c.svg.Def()
	defer c.svg.DefEnd()
	for i := 0; i != 16; i++ {
		c.svg.Filter(colorID(i))
		c.svg.FeFlood(svg.Filterspec{Result: "bg"}, c.getColor(i), 1.0)
		c.svg.FeMerge([]string{`bg`, `SourceGraphic`})
		c.svg.Fend()
	}
	return nil
}

func colorID(i int) string {
	return fmt.Sprintf("c%d", i)
}

func toGlyph(fg, bg vt10x.Color, mode int16) []string {
	classes := []string{}
	filters := []string{}

	if fg != vt10x.DefaultFG {
		classes = append(classes, colorID(int(fg)))
	}

	if bg != vt10x.DefaultBG {
		filters = append(filters, fmt.Sprintf(`url(#%s)`, colorID(int(bg))))
	}

	if mode&0b00000010 != 0 {
		classes = append(classes, `underline`)
	}
	if mode&0b00000100 != 0 {
		classes = append(classes, `bold`)
	}
	if mode&0b00010000 != 0 {
		classes = append(classes, `italic`)
	}
	if mode&0b00100000 != 0 {
		classes = append(classes, `blink`)
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

func (c *Canvas) frames() error {
	heightOff := c.paddingTop()
	widthOff := c.paddingLeft()
	c.svg.Group(fmt.Sprintf(`class="main"`))
	defer c.svg.Gend()

	term := vt10x.New(vt10x.WithSize(c.header.Width, c.header.Height))
	for i, event := range c.events {
		_, err := term.Write([]byte(event.Data))
		if err != nil {
			return err
		}

		err = c.frame(term, i, heightOff, widthOff)
		if err != nil {
			return err
		}
	}
	return nil
}

func clearGlyph(c vt10x.Glyph) vt10x.Glyph {
	if c.Char == ' ' {
		c.Char = '\u2009'
	}
	return c
}

func (c *Canvas) frame(term vt10x.Terminal, i int, heightOff, widthOff int) error {
	c.svg.Gtransform(fmt.Sprintf("translate(%d)", c.paddingRight()*i))
	defer c.svg.Gend()
	show := false

	for row := 1; row < c.header.Height; row++ {
		frame := ""
		lastCell := clearGlyph(term.Cell(0, row))
		lastColorFG := lastCell.FG
		lastColorBG := lastCell.BG
		lastMode := lastCell.Mode
		lastColumn := 0

		for col := 0; col < c.header.Width; col++ {
			cell := clearGlyph(term.Cell(col, row))

			if cell.FG != lastColorFG ||
				cell.BG != lastColorBG ||
				cell.Mode != lastMode {
				if frame != "" {
					show = true
					c.svg.Textspan(
						widthOff+lastColumn*colWidth,
						heightOff+row*rowHeight,
						"",
						toGlyph(lastColorFG, lastColorBG, lastMode)...,
					)

					c.svg.Writer.Write([]byte(frame))
					c.svg.TextEnd()
					frame = ""
				}

				lastColorFG = cell.FG
				lastColorBG = cell.BG
				lastMode = cell.Mode
				lastColumn = col
			}

			frame += string(cell.Char)

		}

		if strings.TrimSpace(frame) != "" {
			show = true
			c.svg.Textspan(
				widthOff+lastColumn*colWidth,
				heightOff+row*rowHeight,
				"",
				toGlyph(lastColorFG, lastColorBG, lastMode)...,
			)

			c.svg.Writer.Write([]byte(frame))
			c.svg.TextEnd()
		}
	}

	if show && term.CursorVisible() {
		cursor := term.Cursor()
		if cursor.Y != 0 {
			c.svg.Rect(
				widthOff+cursor.X*colWidth,
				heightOff-3-padding+cursor.Y*rowHeight,
				colWidth,
				rowHeight,
				`class="cursor"`)
		}
	}

	return nil
}

func generateKeyframes(events []cast.Event, width int32) string {
	dur := events[len(events)-1].Time
	buf := bytes.NewBuffer(nil)
	buf.WriteString("@keyframes k {")
	for i, frame := range events {
		fmt.Fprintf(buf, "%.3f%%{transform:translateX(-%dpx)}", float32(frame.Time*100/dur), width*int32(i))
	}
	buf.WriteString("}")
	return buf.String()
}

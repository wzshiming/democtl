package video

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
	"github.com/hinshun/vt10x"
	"github.com/wzshiming/democtl/pkg/cast"
	dc "github.com/wzshiming/democtl/pkg/color"
	"github.com/wzshiming/democtl/pkg/fonts"
	"golang.org/x/image/font"
)

type Canvas struct {
	header   cast.Header
	events   []cast.Event
	getColor func(i int) string
	noWindow bool

	regular       font.Face
	regularItalic font.Face
	bold          font.Face
	boldItalic    font.Face
}

const (
	rowHeight = 30
	colWidth  = 12
	padding   = 20
)

func NewCanvas() (*Canvas, error) {
	opt := fonts.Options{
		Size: 20,
		DPI:  72,
	}
	regular, err := fonts.LoadFontFace(fonts.Regular, opt)
	if err != nil {
		return nil, err
	}
	regularItalic, err := fonts.LoadFontFace(fonts.RegularItalic, opt)
	if err != nil {
		return nil, err
	}
	bold, err := fonts.LoadFontFace(fonts.Bold, opt)
	if err != nil {
		return nil, err
	}
	boldItalic, err := fonts.LoadFontFace(fonts.BoldItalic, opt)
	if err != nil {
		return nil, err
	}
	return &Canvas{
		getColor:      dc.DefaultColors().GetColorForSVG,
		regular:       regular,
		regularItalic: regularItalic,
		bold:          bold,
		boldItalic:    boldItalic,
	}, nil
}

func (c *Canvas) Run(input io.Reader, outputDir string, noWindow bool) error {
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

	err = c.frames(outputDir)
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

func (c *Canvas) createWindow(dc *gg.Context) {
	bg := c.getColor(int(vt10x.DefaultBG))
	if c.noWindow {
		dc.SetHexColor(bg)
		dc.Clear()
		return
	}

	windowRadius := 5.0
	buttonRadius := 7.0
	buttonColors := [3]string{"#ff5f58", "#ffbd2e", "#18c132"}

	dc.SetHexColor(bg)
	dc.DrawRoundedRectangle(0, 0, float64(c.paddingRight()), float64(c.paddingBottom()), windowRadius)
	dc.Fill()

	for i, color := range buttonColors {
		x := float64(i*(padding+int(buttonRadius/2))) + padding
		y := float64(padding)
		dc.SetHexColor(color)
		dc.DrawCircle(x, y, buttonRadius)
		dc.Fill()
	}
}

func (c *Canvas) frames(outputDir string) error {
	heightOff := c.paddingTop()
	widthOff := c.paddingLeft()

	term := vt10x.New(vt10x.WithSize(c.header.Width, c.header.Height))

	frames, err := os.OpenFile(filepath.Join(outputDir, "frames.txt"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer frames.Close()

	lastTime := 0.
	for i, event := range c.events {
		_, err := term.Write([]byte(event.Data))
		if err != nil {
			return err
		}
		var delay = 0.
		if i != len(c.events)-1 {
			delay = c.events[i+1].Time - lastTime
			if delay < 0.001 {
				continue
			}
		}
		lastTime = event.Time

		img, err := c.frame(term, heightOff, widthOff)
		if err != nil {
			return err
		}

		imgName := fmt.Sprintf("frame%d.png", i)
		_, err = fmt.Fprintf(frames, "file '%s'\n", imgName)
		if err != nil {
			return err
		}

		if i != len(c.events)-1 {
			_, err = fmt.Fprintf(frames, "duration %f\n", delay)
			if err != nil {
				return err
			}
		}

		frame, err := os.OpenFile(filepath.Join(outputDir, imgName), os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		err = png.Encode(frame, img)
		if err != nil {
			return err
		}
		frame.Close()
	}

	return nil
}

func (c *Canvas) frame(term vt10x.Terminal, heightOff, widthOff int) (image.Image, error) {
	width := c.paddingRight()
	height := c.paddingBottom()
	dc := gg.NewContext(width, height)

	c.createWindow(dc)

	for row := 1; row < c.header.Height; row++ {
		frame := ""
		lastCell := term.Cell(0, row)
		lastColorFG := lastCell.FG
		lastColorBG := lastCell.BG
		lastMode := lastCell.Mode
		lastColumn := 0

		for col := 0; col < c.header.Width; col++ {
			cell := term.Cell(col, row)
			if cell.FG != lastColorFG ||
				cell.BG != lastColorBG ||
				cell.Mode != lastMode {
				if frame != "" {
					c.drawText(dc,
						frame,
						widthOff+lastColumn*colWidth,
						heightOff+row*rowHeight,
						lastColorFG,
						lastColorBG,
						lastMode,
					)
					frame = ""
				}
				lastColorFG = cell.FG
				lastColorBG = cell.BG
				lastMode = cell.Mode
				lastColumn = col
			}
			frame += string(cell.Char)
		}

		if frame != "" {
			c.drawText(dc,
				frame,
				widthOff+lastColumn*colWidth,
				heightOff+row*rowHeight,
				lastColorFG,
				lastColorBG,
				lastMode,
			)
		}
	}

	if term.CursorVisible() {
		cursor := term.Cursor()
		if cursor.Y != 0 {
			c.drawCursor(dc,
				widthOff+cursor.X*colWidth,
				heightOff-3-padding+cursor.Y*rowHeight,
			)
		}
	}

	return dc.Image(), nil
}

func (c *Canvas) drawText(dc *gg.Context, text string, x, y int, fg, bg vt10x.Color, mode int16) {

	if bg != vt10x.DefaultBG {
		bgColorStr := c.getColor(int(bg))
		dc.SetHexColor(bgColorStr)
		dc.DrawRectangle(float64(x), float64(y-rowHeight)+7, float64(len(text)*colWidth), rowHeight)
		dc.Fill()
	}

	colorStr := c.getColor(int(fg))

	if mode&0b00000100 != 0 {
		if mode&0b00010000 != 0 {
			dc.SetFontFace(c.boldItalic)
		} else {
			dc.SetFontFace(c.bold)
		}
	} else {
		if mode&0b00010000 != 0 {
			dc.SetFontFace(c.regularItalic)
		} else {
			dc.SetFontFace(c.regular)
		}
	}

	// TODO: blink
	//if mode&0b00100000 != 0 {
	//}

	dc.SetHexColor(colorStr)
	dc.DrawString(text, float64(x), float64(y))

	if mode&0b00000010 != 0 {
		width, _ := dc.MeasureString(text)
		dc.DrawLine(float64(x), float64(y)+5, float64(x)+width, float64(y)+5)
		dc.Stroke()
	}
}

func (c *Canvas) drawCursor(dc *gg.Context, x, y int) {
	dc.SetHexColor(c.getColor(int(vt10x.DefaultCursor)) + "aa")
	dc.DrawRectangle(float64(x)+3, float64(y), colWidth, rowHeight)
	dc.Fill()
}

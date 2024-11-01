package video

import (
	"context"
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fogleman/gg"
	"github.com/wzshiming/democtl/pkg/renderer"
	"github.com/wzshiming/vt10x"
	"golang.org/x/image/font"
)

type canvas struct {
	getColor func(i vt10x.Color) string
	noWindow bool

	regular       font.Face
	regularItalic font.Face
	bold          font.Face
	boldItalic    font.Face

	width, height int

	frames io.WriteCloser
	output string

	lastOffset time.Duration
}

const (
	rowHeight = 30
	colWidth  = 12
	padding   = 20
)

func NewCanvas(output string, noWindow bool, getColor func(i vt10x.Color) string) renderer.Renderer {
	return &canvas{
		output:   output,
		noWindow: noWindow,
		getColor: getColor,
	}
}

func (c *canvas) Initialize(ctx context.Context, x, y int, width, height int) error {
	c.width = width
	c.height = height
	frames, err := os.OpenFile(filepath.Join(c.output, "frames.txt"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	c.frames = frames
	return nil
}

func (c *canvas) Finish(ctx context.Context) error {
	return c.frames.Close()
}

func (c *canvas) Frame(ctx context.Context, index int, offset time.Duration) (renderer.Frame, error) {
	width := c.paddingRight()
	height := c.paddingBottom()
	dc := gg.NewContext(width, height)
	c.createWindow(dc)

	return &frame{
		canvas:    c,
		dc:        dc,
		offset:    offset,
		heightOff: c.paddingTop(),
		widthOff:  c.paddingLeft(),
		finish: func() error {
			imgName := fmt.Sprintf("frame%d.png", index)

			if index != 0 {
				_, err := fmt.Fprintf(c.frames, "duration %f\n", float64(offset-c.lastOffset)/float64(time.Second))
				if err != nil {
					return err
				}
			}

			_, err := fmt.Fprintf(c.frames, "file '%s'\n", imgName)
			if err != nil {
				return err
			}

			c.lastOffset = offset

			frame, err := os.OpenFile(filepath.Join(c.output, imgName), os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			err = png.Encode(frame, dc.Image())
			if err != nil {
				return err
			}
			frame.Close()

			return nil
		},
	}, nil
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

func (c *canvas) createWindow(dc *gg.Context) {
	bg := c.getColor(vt10x.DefaultBG)
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

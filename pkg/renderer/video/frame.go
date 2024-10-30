package video

import (
	"context"

	"github.com/fogleman/gg"
	"github.com/hinshun/vt10x"
	"github.com/wzshiming/democtl/pkg/fonts"
	"github.com/wzshiming/democtl/pkg/renderer"
)

type frame struct {
	*canvas

	dc *gg.Context

	heightOff, widthOff int

	finish func() error
}

var opt = fonts.Options{
	Size: 20,
	DPI:  72,
}

func (f *frame) offsetX(x int) float64 {
	return float64(f.widthOff + x*colWidth)
}

func (f *frame) offsetY(y int) float64 {
	return float64(f.heightOff + y*rowHeight)
}

func (f *frame) DrawText(ctx context.Context, x, y int, text string, fg, bg renderer.Color, mode int16) error {
	offsetX := f.offsetX(x)
	offsetY := f.offsetY(y)
	width, _ := f.dc.MeasureString(text)

	if bg != vt10x.DefaultBG {
		bgColorStr := f.getColor(int(bg))
		f.dc.SetHexColor(bgColorStr)
		f.dc.DrawRectangle(offsetX, offsetY-rowHeight+7, width*colWidth, rowHeight)
		f.dc.Fill()
	}

	for i, r := range text {
		err := f.drawText(ctx, x+i, y, r, fg, bg, mode)
		if err != nil {
			return err
		}
	}

	if mode&0b00000010 != 0 {
		f.dc.DrawLine(offsetX, offsetY+5, offsetX+width, offsetY+5)
		f.dc.Stroke()
	}
	return nil
}

func (f *frame) drawText(ctx context.Context, x, y int, text rune, fg, bg renderer.Color, mode int16) error {
	offsetX := f.offsetX(x)
	offsetY := f.offsetY(y)

	colorStr := f.getColor(int(fg))

	var err error
	if mode&0b00000100 != 0 {
		if mode&0b00010000 != 0 {
			if f.boldItalic == nil {
				f.boldItalic, err = fonts.LoadFontFace(fonts.BoldItalic, opt)
				if err != nil {
					return err
				}
			}
			f.dc.SetFontFace(f.boldItalic)
		} else {
			if f.bold == nil {
				f.bold, err = fonts.LoadFontFace(fonts.Bold, opt)
				if err != nil {
					return err
				}
			}
			f.dc.SetFontFace(f.bold)
		}
	} else {
		if mode&0b00010000 != 0 {
			if f.regularItalic == nil {
				f.regularItalic, err = fonts.LoadFontFace(fonts.RegularItalic, opt)
				if err != nil {
					return err
				}
			}
			f.dc.SetFontFace(f.regularItalic)
		} else {
			if f.regular == nil {
				f.regular, err = fonts.LoadFontFace(fonts.Regular, opt)
				if err != nil {
					return err
				}
			}
			f.dc.SetFontFace(f.regular)
		}
	}

	// TODO: blink
	//if mode&0b00100000 != 0 {
	//}

	f.dc.SetHexColor(colorStr)
	f.dc.DrawStringAnchored(string(text), offsetX, offsetY, 0, 0)

	return nil
}

func (f *frame) DrawCursor(ctx context.Context, x, y int) error {
	offsetX := f.offsetX(x)
	offsetY := f.offsetY(y)

	f.dc.SetHexColor(f.getColor(int(vt10x.DefaultCursor)) + "aa")
	f.dc.DrawRectangle(
		offsetX+3, offsetY-20,
		colWidth, rowHeight)
	f.dc.Fill()
	return nil
}

func (f *frame) Finish(ctx context.Context) error {
	if f.finish == nil {
		return nil
	}
	err := f.finish()
	if err != nil {
		return err
	}
	return nil
}

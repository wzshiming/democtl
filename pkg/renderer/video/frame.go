package video

import (
	"context"
	"time"

	"github.com/fogleman/gg"
	"github.com/wzshiming/democtl/pkg/color"
	"github.com/wzshiming/democtl/pkg/fonts"
	"github.com/wzshiming/democtl/pkg/utils"
	"github.com/wzshiming/vt10x"
)

type frame struct {
	*canvas

	dc *gg.Context

	offset time.Duration

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

func (f *frame) setFont(mode vt10x.AttrFlag) error {
	var err error
	if mode&vt10x.AttrBold != 0 {
		if mode&vt10x.AttrItalic != 0 {
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
		if mode&vt10x.AttrItalic != 0 {
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
	return nil
}

func (f *frame) DrawText(ctx context.Context, x, y int, text string, fg, bg vt10x.Color, mode vt10x.AttrFlag) error {
	if mode&vt10x.AttrReverse != 0 {
		fg, bg = bg, fg
	}

	offsetX := f.offsetX(x)
	offsetY := f.offsetY(y)
	width := float64(utils.StrLen(text))

	bgColorStr := f.getColor(bg)
	if bg != vt10x.DefaultBG {
		f.dc.SetHexColor(bgColorStr)
		f.dc.DrawRectangle(offsetX, offsetY-rowHeight+7, width*colWidth, rowHeight)
		f.dc.Fill()
	}

	colorStr := f.getColor(fg)
	if mode&vt10x.AttrDim != 0 {
		r, g, b := color.ParseHexColor(colorStr)
		colorStr = color.FormatHexColor(r/2, g/2, b/2)
	}

	if colorStr == bgColorStr {
		return nil
	}

	err := f.setFont(mode)
	if err != nil {
		return err
	}

	f.dc.SetHexColor(colorStr)
	for i, r := range text {
		err := f.drawText(ctx, x+i, y, r, fg, bg, mode)
		if err != nil {
			return err
		}
	}

	if mode&vt10x.AttrUnderline != 0 {
		f.dc.DrawLine(offsetX, offsetY+5, offsetX+width*colWidth, offsetY+5)
		f.dc.Stroke()
	}
	if mode&vt10x.AttrStrike != 0 {
		f.dc.DrawLine(offsetX, offsetY-5, offsetX+width*colWidth, offsetY-5)
		f.dc.Stroke()
	}
	return nil
}

func (f *frame) drawText(ctx context.Context, x, y int, text rune, fg, bg vt10x.Color, mode vt10x.AttrFlag) error {
	offsetX := f.offsetX(x)
	offsetY := f.offsetY(y)

	f.dc.DrawStringAnchored(string(text), offsetX, offsetY, 0, 0)
	return nil
}

func (f *frame) DrawCursor(ctx context.Context, x, y int) error {
	offsetX := f.offsetX(x)
	offsetY := f.offsetY(y)

	f.dc.SetHexColor(f.getColor(vt10x.DefaultCursor) + "aa")
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

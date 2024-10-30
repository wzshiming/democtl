package svg

import (
	"context"
	"strings"

	"github.com/wzshiming/democtl/pkg/renderer"
)

type frame struct {
	*canvas

	heightOff, widthOff int

	finish func() error
}

func (f *frame) offsetX(x int) int {
	return f.widthOff + x*colWidth
}

func (f *frame) offsetY(y int) int {
	return f.heightOff + y*rowHeight
}

func (f *frame) DrawText(ctx context.Context, x, y int, text string, fg, bg renderer.Color, mode int16) error {
	f.svg.Textspan(
		f.offsetX(x),
		f.offsetY(y),
		"",
		f.toGlyph(fg, bg, mode)...,
	)
	f.svg.Writer.Write([]byte(strings.ReplaceAll(text, " ", "\u2009")))

	f.svg.TextEnd()
	return nil
}

func (f *frame) DrawCursor(ctx context.Context, x, y int) error {
	f.svg.Rect(
		f.offsetX(x),
		f.offsetY(y)-3-padding,
		colWidth,
		rowHeight,
		`class="cursor"`)
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

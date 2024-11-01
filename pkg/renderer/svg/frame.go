package svg

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/wzshiming/vt10x"
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

func (f *frame) DrawText(ctx context.Context, x, y int, text string, fg, bg vt10x.Color, mode vt10x.AttrFlag) error {
	key := fmt.Sprintf("%d,%d,%d,%s", fg, bg, mode, text)
	attrs := f.toGlyph(fg, bg, mode)
	if strings.HasPrefix(text, " ") ||
		strings.HasSuffix(text, " ") ||
		strings.Contains(text, "  ") {
		attrs = append(attrs, `xml:space="preserve"`)
	}

	id := f.getDefs(key, func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<symbol id="%s">
<text %s>%s</text>
</symbol>
`, id, strings.Join(attrs, " "), escapeText(text))
		return buf.String()
	})
	f.useDef(id, f.offsetX(x), f.offsetY(y)-17)
	return nil
}

func (f *frame) DrawCursor(ctx context.Context, x, y int) error {
	id := f.getDefs("cursor", func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<symbol id="%s">
<rect width="%d" height="%d" style="fill:%s;opacity:0.8"/>
</symbol>
`, id, colWidth, rowHeight, f.getColor(vt10x.DefaultCursor))
		return buf.String()
	})

	f.useDef(id, f.offsetX(x), f.offsetY(y)-23)
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

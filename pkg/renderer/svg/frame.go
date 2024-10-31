package svg

import (
	"bytes"
	"context"
	"encoding/xml"
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
	text = strings.ReplaceAll(text, " ", "\u2009")
	key := fmt.Sprintf("%d,%d,%d,%s", fg, bg, mode, text)

	attrs := strings.Join(f.toGlyph(fg, bg, mode), " ")
	id := f.getDefs(key, func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<symbol id="%s">
<text %s x="10" y="25">
`, id, attrs)
		xml.Escape(buf, []byte(text))
		fmt.Fprintf(buf, `
</text>
</symbol>
`)
		return buf.String()
	})

	f.useDef(id, f.offsetX(x)-10, f.offsetY(y)-25)
	return nil
}

func (f *frame) DrawCursor(ctx context.Context, x, y int) error {
	sid := f.getStyles("cursor", func(id string) string {
		return fmt.Sprintf(`
.%s {
  fill: %s;
  opacity: 0.8;
}
`, id, f.getColor(vt10x.DefaultCursor))
	})

	id := f.getDefs("cursor", func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<symbol id="%s">
<rect width="%d" height="%d" class="%s" />
</symbol>
`, id, colWidth, rowHeight, sid)
		return buf.String()
	})

	f.useDef(id, f.offsetX(x), f.offsetY(y)-3-padding)
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

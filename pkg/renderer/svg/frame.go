package svg

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/wzshiming/democtl/pkg/utils"
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
	if mode&vt10x.AttrReverse != 0 {
		fg, bg = bg, fg
		mode &^= vt10x.AttrReverse
	}

	attrs := f.toGlyph(fg, bg, mode)
	if strings.HasPrefix(text, " ") ||
		strings.HasSuffix(text, " ") ||
		strings.Contains(text, "  ") {
		attrs = append(attrs, `xml:space="preserve"`)
	}

	if bg != vt10x.DefaultBG {
		bid := f.drawRect(utils.StrLen(text), bg)
		f.useDef(bid, f.offsetX(x), f.offsetY(y)-23)
	}

	id := f.getDefs(fmt.Sprintf("%d,%d,%s", fg, mode, text), func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<text id="%s" %s>%s</text>
`, id, strings.Join(attrs, " "), escapeText(text))
		return buf.String()
	})

	f.useDef(id, f.offsetX(x), f.offsetY(y)-17)
	return nil
}

func (f *frame) drawRect(width int, bg vt10x.Color) string {
	return f.getDefs(fmt.Sprintf("%d,%d", width, bg), func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<rect id="%s" width="%d" height="%d" style="fill:%s"/>
`, id, colWidth*width, rowHeight, f.getColor(bg))
		return buf.String()
	})
}

func (f *frame) DrawCursor(ctx context.Context, x, y int) error {
	id := f.getDefs("cursor", func(id string) string {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, `
<rect id="%s" width="%d" height="%d" style="fill:%s;opacity:0.8"/>
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

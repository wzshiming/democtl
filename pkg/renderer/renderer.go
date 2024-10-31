package renderer

import (
	"context"
	"io"
	"time"
	"strings"

	"github.com/wzshiming/democtl/pkg/cast"
	"github.com/wzshiming/vt10x"
)

type Renderer interface {
	Initialize(ctx context.Context, x, y int, width, height int) error
	Frame(ctx context.Context, index int, offset time.Duration) (Frame, error)
	Finish(ctx context.Context) error
}

type Frame interface {
	DrawText(ctx context.Context, x, y int, text string, fg, bg vt10x.Color, mode vt10x.AttrFlag) error
	DrawCursor(ctx context.Context, x, y int) error
	Finish(ctx context.Context) error
}

type renderContent struct {
	ctx context.Context

	header cast.Header
	events []cast.Event

	renderer Renderer
}

func Render(ctx context.Context, renderer Renderer, input io.Reader) error {
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

	c := &renderContent{
		ctx:      ctx,
		renderer: renderer,
		header:   header,
		events:   compress(events, 60),
	}
	err = frames(c)
	if err != nil {
		return err
	}

	return nil
}

func frames(c *renderContent) (err error) {
	term := vt10x.New(vt10x.WithSize(c.header.Width, c.header.Height))

	err = c.renderer.Initialize(c.ctx, 0, 0,
		c.header.Width, c.header.Height,
	)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = c.renderer.Finish(c.ctx)
		}
	}()

	for i, event := range c.events {
		_, err = term.Write([]byte(event.Data))
		if err != nil {
			return err
		}

		f, err := c.renderer.Frame(c.ctx, i, time.Duration(event.Time*float64(time.Second)))
		if err != nil {
			return err
		}
		err = frame(c, term, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func isEmpty(text string, bg vt10x.Color, mode vt10x.AttrFlag) bool {
	if mode&vt10x.AttrHidden != 0 {
		return true
	}
	empty := bg == vt10x.DefaultBG && !strings.ContainsFunc(text, func(r rune) bool {
		return r != ' '
	})
	if empty {
		return true
	}

	return false
}

func frame(c *renderContent, term vt10x.Terminal, frame Frame) (err error) {
	defer func() {
		if err == nil {
			err = frame.Finish(c.ctx)
		}
	}()

	for row := 1; row < c.header.Height; row++ {
		f := ""
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
				if f != "" {
					if !isEmpty(f, lastColorBG, lastMode) {
						err = frame.DrawText(c.ctx,
							lastColumn,
							row,
							f,
							lastColorFG,
							lastColorBG,
							lastMode,
						)
						if err != nil {
							return err
						}
					}
					f = ""
				}
				lastColorFG = cell.FG
				lastColorBG = cell.BG
				lastMode = cell.Mode
				lastColumn = col
			}
			f += string(cell.Char)
		}

		if f != "" {
			if !isEmpty(f, lastColorBG, lastMode) {
				err = frame.DrawText(c.ctx,
					lastColumn,
					row,
					f,
					lastColorFG,
					lastColorBG,
					lastMode,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	if term.CursorVisible() {
		cursor := term.Cursor()
		if cursor.Y != 0 {
			err := frame.DrawCursor(c.ctx, cursor.X, cursor.Y)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func compress(events []cast.Event, fps int) []cast.Event {
	var out []cast.Event
	var minInterval = 1.0 / float64(fps)
	for i, event := range events {
		if i == 0 {
			out = append(out, event)
			continue
		}
		lastEvent := &out[len(out)-1]
		if event.Time-lastEvent.Time < minInterval {
			lastEvent.Data += event.Data
			lastEvent.Time = event.Time
		} else {
			out = append(out, event)
		}
	}
	return out
}

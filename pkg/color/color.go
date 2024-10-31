package color

import (
	"github.com/wzshiming/vt10x"
)

type Colors struct {
	Color0  string // Black
	Color1  string // Red
	Color2  string // Green
	Color3  string // Yellow
	Color4  string // Blue
	Color5  string // Magenta
	Color6  string // Cyan
	Color7  string // Grey
	Color8  string // Dark Grey
	Color9  string // Light Red
	Color10 string // Light Green
	Color11 string // Light Yellow
	Color12 string // Light Blue
	Color13 string // Light Magenta
	Color14 string // Light Cyan
	Color15 string // White

	Foreground  string
	Background  string
	CursorColor string
}

func (c Colors) GetColorForHex(i vt10x.Color) string {
	switch i {
	case vt10x.DefaultBG:
		return c.Background
	case vt10x.DefaultFG:
		return c.Foreground
	case vt10x.DefaultCursor:
		return c.CursorColor
	case 0:
		return c.Color0
	case 1:
		return c.Color1
	case 2:
		return c.Color2
	case 3:
		return c.Color3
	case 4:
		return c.Color4
	case 5:
		return c.Color5
	case 6:
		return c.Color6
	case 7:
		return c.Color7
	case 8:
		return c.Color8
	case 9:
		return c.Color9
	case 10:
		return c.Color10
	case 11:
		return c.Color11
	case 12:
		return c.Color12
	case 13:
		return c.Color13
	case 14:
		return c.Color14
	case 15:
		return c.Color15
	}
	return ""
}

func DefaultColors() *Colors {
	return &Colors{
		Color0:  colors[0],
		Color1:  colors[1],
		Color2:  colors[2],
		Color3:  colors[3],
		Color4:  colors[4],
		Color5:  colors[5],
		Color6:  colors[6],
		Color7:  colors[7],
		Color8:  colors[8],
		Color9:  colors[9],
		Color10: colors[10],
		Color11: colors[11],
		Color12: colors[12],
		Color13: colors[13],
		Color14: colors[14],
		Color15: colors[15],

		Foreground:  colors[15],
		Background:  "#222324",
		CursorColor: "#bbbbbb",
	}
}

var colors = [...]string{
	0x00: "#000000", // Black
	0x01: "#cd0000", // Red
	0x02: "#00cd00", // Green
	0x03: "#cdcd00", // Yellow
	0x04: "#0000ee", // Blue
	0x05: "#cd00cd", // Magenta
	0x06: "#00cdcd", // Cyan
	0x07: "#e5e5e5", // Grey
	0x08: "#7f7f7f", // Dark Grey
	0x09: "#ff0000", // Light Red
	0x0a: "#00ff00", // Light Green
	0x0b: "#ffff00", // Light Yellow
	0x0c: "#5c5cff", // Light Blue
	0x0d: "#ff00ff", // Light Magenta
	0x0e: "#00ffff", // Light Cyan
	0x0f: "#ffffff", // White
}

package styles

import (
	"os"

	"github.com/wzshiming/vt10x"
	"gopkg.in/yaml.v3"
)

type Styles struct {
	Color0  string `yaml:"color0,omitempty"`  // Black
	Color1  string `yaml:"color1,omitempty"`  // Red
	Color2  string `yaml:"color2,omitempty"`  // Green
	Color3  string `yaml:"color3,omitempty"`  // Yellow
	Color4  string `yaml:"color4,omitempty"`  // Blue
	Color5  string `yaml:"color5,omitempty"`  // Magenta
	Color6  string `yaml:"color6,omitempty"`  // Cyan
	Color7  string `yaml:"color7,omitempty"`  // Grey
	Color8  string `yaml:"color8,omitempty"`  // Dark Grey
	Color9  string `yaml:"color9,omitempty"`  // Light Red
	Color10 string `yaml:"color10,omitempty"` // Light Green
	Color11 string `yaml:"color11,omitempty"` // Light Yellow
	Color12 string `yaml:"color12,omitempty"` // Light Blue
	Color13 string `yaml:"color13,omitempty"` // Light Magenta
	Color14 string `yaml:"color14,omitempty"` // Light Cyan
	Color15 string `yaml:"color15,omitempty"` // White

	Foreground  string `yaml:"foreground,omitempty"`
	Background  string `yaml:"background,omitempty"`
	CursorColor string `yaml:"cursorColor,omitempty"`

	NoWindows bool `yaml:"noWindows,omitempty"`
}

func NewStylesFromFile(path string) (*Styles, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &Styles{}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s Styles) GetColorForHex(i vt10x.Color) string {
	switch i {
	case vt10x.DefaultBG:
		return s.Background
	case vt10x.DefaultFG:
		return s.Foreground
	case vt10x.DefaultCursor:
		return s.CursorColor
	case 0:
		return s.Color0
	case 1:
		return s.Color1
	case 2:
		return s.Color2
	case 3:
		return s.Color3
	case 4:
		return s.Color4
	case 5:
		return s.Color5
	case 6:
		return s.Color6
	case 7:
		return s.Color7
	case 8:
		return s.Color8
	case 9:
		return s.Color9
	case 10:
		return s.Color10
	case 11:
		return s.Color11
	case 12:
		return s.Color12
	case 13:
		return s.Color13
	case 14:
		return s.Color14
	case 15:
		return s.Color15
	}
	if !i.ANSI() {
		r, g, b, ok := i.RGB()
		if ok {
			return FormatHexColor(r, g, b)
		}
	}
	return ""
}

func Default() *Styles {
	return &Styles{
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

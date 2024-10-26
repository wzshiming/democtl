package fonts

import (
	_ "embed"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	//go:embed SF-Mono-Regular.otf
	Regular []byte
	//go:embed SF-Mono-RegularItalic.otf
	RegularItalic []byte

	//go:embed SF-Mono-Heavy.otf
	Bold []byte
	//go:embed SF-Mono-HeavyItalic.otf
	BoldItalic []byte
)

type Options = opentype.FaceOptions

func LoadFontFace(data []byte, opt Options) (font.Face, error) {
	f, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(f, &opt)
}

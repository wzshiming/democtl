package minify

import (
	"io"

	"github.com/tdewolff/minify/v2"
	mcss "github.com/tdewolff/minify/v2/css"
	msvg "github.com/tdewolff/minify/v2/svg"
)

func SVGWithWriter(w io.Writer) io.WriteCloser {
	m := minify.New()
	m.AddFunc("image/svg+xml", msvg.Minify)
	return m.Writer("image/svg+xml", w)
}

func CSSWithString(data string) (string, error) {
	m := minify.New()
	m.AddFunc("text/css", mcss.Minify)
	return m.String("text/css", data)
}

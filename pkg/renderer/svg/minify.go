package svg

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify/v2"
	mcss "github.com/tdewolff/minify/v2/css"
)

func minifyCSS(data string) (string, error) {
	m := minify.New()
	m.AddFunc("text/css", mcss.Minify)
	return m.String("text/css", data)
}

type minifyWriter struct {
	w io.Writer
}

func newMinifyWriter(w io.Writer) *minifyWriter {
	return &minifyWriter{w: w}
}

func (m *minifyWriter) Write(p []byte) (n int, err error) {
	for len(p) != 0 {
		index := bytes.IndexByte(p, '\n')
		if index == -1 {
			return m.w.Write(p)
		}

		m, err := m.w.Write(p[:index])
		if err != nil {
			return n, err
		}
		n += m

		p = p[index+1:]
	}
	return n, nil
}

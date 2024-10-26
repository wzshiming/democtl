package cast

import (
	"encoding/json"
	"io"
)

type Encoder struct {
	w *json.Encoder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: json.NewEncoder(w),
	}
}

func (e *Encoder) EncodeHeader(h Header) error {
	h.Version = 2
	return e.w.Encode(h)
}

func (e *Encoder) EncodeEvent(v Event) error {
	return e.w.Encode(v)
}

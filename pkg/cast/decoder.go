package cast

import (
	"encoding/json"
	"io"
)

type Decoder struct {
	r *json.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: json.NewDecoder(r)}
}

func (d *Decoder) DecodeHeader() (Header, error) {
	var h Header
	if err := d.r.Decode(&h); err != nil {
		return Header{}, err
	}

	return h, nil
}

func (d *Decoder) DecodeEvent() (Event, error) {
	var e Event
	if err := d.r.Decode(&e); err != nil {
		return Event{}, err
	}
	return e, nil
}

package cast

import (
	"encoding/json"
	"io"
)

type Decoder struct {
	r *json.Decoder

	stdout     []Event
	index      int
	timeOffset float64
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: json.NewDecoder(r)}
}

func (d *Decoder) DecodeHeader() (Header, error) {
	var h Header
	if err := d.r.Decode(&h); err != nil {
		return Header{}, err
	}

	if len(h.Stdout) != 0 {
		d.stdout = h.Stdout
		h.Stdout = nil
		h.Version = 2
	}

	return h, nil
}

func (d *Decoder) DecodeEvent() (Event, error) {
	if d.stdout != nil {
		if d.index == len(d.stdout) {
			return Event{}, io.EOF
		}
		e := d.stdout[d.index]
		t := e.Time
		e.Time += d.timeOffset
		d.timeOffset += t
		d.index++

		return e, nil
	}

	var e Event
	if err := d.r.Decode(&e); err != nil {
		return Event{}, err
	}
	return e, nil
}

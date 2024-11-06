package cast

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Time float64
	Data string
}

// UnmarshalJSON reads json list as Event fields.
func (e *Event) UnmarshalJSON(data []byte) error {
	var v []json.RawMessage
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if len(v) == 2 {
		err = json.Unmarshal(v[0], &e.Time)
		if err != nil {
			return err
		}

		err = json.Unmarshal(v[1], &e.Data)
		if err != nil {
			return err
		}
		return nil
	}

	if len(v) != 3 {
		return fmt.Errorf("wrong event length (%d): expected 3 elements", len(v))
	}

	var t string
	err = json.Unmarshal(v[1], &t)
	if err != nil {
		return err
	}
	if t != "o" {
		return fmt.Errorf("wrong event type (%s): expected o", t)
	}

	err = json.Unmarshal(v[0], &e.Time)
	if err != nil {
		return err
	}

	err = json.Unmarshal(v[2], &e.Data)
	if err != nil {
		return err
	}

	return nil
}

// MarshalJSON reads json list as Event fields.
func (e Event) MarshalJSON() ([]byte, error) {
	data := [...]any{e.Time, "o", e.Data}
	return json.Marshal(data)
}

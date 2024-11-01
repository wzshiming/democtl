package svg

import (
	"bytes"
	"encoding/xml"
)

func escapeText(s string) string {
	buf := bytes.NewBuffer(nil)
	err := xml.EscapeText(buf, []byte(s))
	if err != nil {
		return s
	}
	return buf.String()
}

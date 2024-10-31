package color

import (
	"fmt"
	"strings"
)

func ParseHexColor(x string) (r, g, b int) {
	x = strings.TrimPrefix(x, "#")
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	return
}

func FormatHexColor(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

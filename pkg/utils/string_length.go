package utils

import (
	"unicode/utf8"
)

func runeWidth(r rune) int {
	switch {
	case r == utf8.RuneError || r < '\x20':
		return 0

	case '\x20' <= r && r < '\u2000':
		return 1

	case '\u2000' <= r && r < '\uFF61':
		return 2

	case '\uFF61' <= r && r < '\uFFA0':
		return 1

	case '\uFFA0' <= r:
		return 2
	}

	return 0
}

func StrLen(str string) int {
	i := 0
	for _, v := range str {
		i += runeWidth(v)
	}
	return i
}

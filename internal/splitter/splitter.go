package splitter

import (
	"unicode/utf8"

	"go.osspkg.com/unic/internal/dict"
)

func Func(data []byte, atEOF bool) (int, []byte, error) {
	start := 0
	raw := false

	for width := 0; start < len(data); start += width {
		var char rune
		char, width = utf8.DecodeRune(data[start:])

		switch true {
		case dict.IsRawChar(char) && !raw:
			raw = true
			start += width
		case dict.IsSkipChar(char):
			continue
		case dict.IsStopChar(char):
			return start + width, data[start : start+width], nil
		default:
		}

		break
	}

	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		switch true {
		case dict.IsRawChar(r) && raw:
			return i + width, data[start:i], nil
		case raw:
			continue
		case dict.IsStopChar(r):
			return i, data[start:i], nil
		case dict.IsSkipChar(r):
			return i + width, data[start:i], nil
		default:
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	return start, nil, nil
}

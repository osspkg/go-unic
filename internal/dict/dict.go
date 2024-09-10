package dict

import "unicode/utf8"

const (
	Comment    = "#"
	NewLine    = "\n"
	Space      = " "
	KeyEnd     = ";"
	BlockOpen  = "{"
	BlockClose = "}"
	Apostrophe = "`"
)

func IsMultiline(data []byte) bool {
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if IsSkipChar(r) {
			return true
		}
	}
	return false
}

func IsRawChar(r rune) bool {
	switch r {
	case '`':
		return true
	}
	return false
}

func IsStopChar(r rune) bool {
	switch r {
	case '{', '}', '#', ';', '\n':
		return true
	}
	return false
}

func IsSkipChar(r rune) bool {
	if r <= '\u00FF' {
		switch r {
		case ' ', '\t', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}

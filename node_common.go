package coma

import "unicode/utf8"

const (
	wordComment    = "#"
	wordNewLine    = "\n"
	wordSpace      = " "
	wordDataEnd    = ";"
	wordBlockOpen  = "{"
	wordBlockClose = "}"
)

func isMultiline(data []byte) bool {
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if isSkipSplitWord(r) {
			return true
		}
	}
	return false
}

func splitWordsFunc(data []byte, atEOF bool) (int, []byte, error) {
	start := 0
	raw := false

	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])

		switch true {
		case isRawSplitWord(r) && !raw:
			raw = true
			start += width
			break
		case isSkipSplitWord(r):
			continue
		case isStopSplitWord(r):
			return start + width, data[start : start+width], nil
		default:
		}

		break
	}

	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		switch true {
		case isRawSplitWord(r) && raw:
			return i + width, data[start:i], nil
		case raw:
			continue
		case isStopSplitWord(r):
			return i, data[start:i], nil
		case isSkipSplitWord(r):
			return i + width, data[start:i], nil
		default:
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	return start, nil, nil
}

func isRawSplitWord(r rune) bool {
	switch r {
	case '`':
		return true
	}
	return false
}

func isStopSplitWord(r rune) bool {
	switch r {
	case '{', '}', '#', ';', '\n':
		return true
	}
	return false
}

func isSkipSplitWord(r rune) bool {
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

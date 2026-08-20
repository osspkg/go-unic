/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type fieldTag struct {
	name       string
	hasDefault bool
	defaultVal string
	omitempty  bool
	attr       int
	desc       string
}

func parseUnicTag(tag string) (fieldTag, bool, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" || tag == "-" {
		return fieldTag{}, false, nil
	}

	name, rest, err := splitTagName(tag)
	if err != nil {
		return fieldTag{}, false, err
	}
	ft := fieldTag{name: name}
	for rest != "" {
		var opt string
		opt, rest, err = splitTagOption(rest)
		if err != nil {
			return fieldTag{}, false, err
		}
		if opt == "" {
			continue
		}
		key, val, hasVal := splitOpt(opt)
		switch key {
		case "omitempty":
			ft.omitempty = true
		case "default":
			ft.hasDefault = true
			ft.defaultVal = val
		case "attr":
			if !hasVal {
				return fieldTag{}, false, fmt.Errorf("unic: attr requires a positive index")
			}
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				return fieldTag{}, false, fmt.Errorf("unic: attr must be > 0")
			}
			ft.attr = n
		case "desc":
			ft.desc = val
		default:
			// unknown options are ignored so new flags do not break decode
		}
	}
	if ft.name == "" {
		return fieldTag{}, false, nil
	}
	return ft, true, nil
}

//nolint:unparam
func splitTagName(tag string) (string, string, error) {
	i := 0
	for i < len(tag) {
		r, size := utf8.DecodeRuneInString(tag[i:])
		if r == ',' {
			break
		}
		i += size
	}
	return strings.TrimSpace(tag[:i]), strings.TrimSpace(trimComma(tag[i:])), nil
}

func splitTagOption(s string) (opt, rest string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", nil
	}
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == '\'' || r == '"' {
			end, qerr := skipQuoted(s, i)
			if qerr != nil {
				return "", "", qerr
			}
			i = end
			continue
		}
		if r == ',' {
			return strings.TrimSpace(s[:i]), strings.TrimSpace(trimComma(s[i:])), nil
		}
		i += size
	}
	return strings.TrimSpace(s), "", nil
}

func skipQuoted(s string, i int) (int, error) {
	quote, size := utf8.DecodeRuneInString(s[i:])
	i += size
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r == quote {
			return i, nil
		}
	}
	return 0, fmt.Errorf("unic: unterminated quote in struct tag")
}

func splitOpt(opt string) (key, val string, hasVal bool) {
	key, val, hasVal = strings.Cut(opt, "=")
	key = strings.TrimSpace(key)
	val = strings.TrimSpace(val)
	if hasVal {
		val = unquoteTagValue(val)
	}
	return key, val, hasVal
}

func unquoteTagValue(v string) string {
	if len(v) >= 2 {
		if (v[0] == '\'' && v[len(v)-1] == '\'') || (v[0] == '"' && v[len(v)-1] == '"') {
			return v[1 : len(v)-1]
		}
	}
	return v
}

func trimComma(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, ",") {
		return strings.TrimSpace(s[1:])
	}
	return s
}

func splitDefaultList(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

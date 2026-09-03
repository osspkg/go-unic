/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import (
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"go.osspkg.com/bb"

	"go.osspkg.com/unic/internal/pool"
)

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokIdent
	tokString
	tokLBrace
	tokRBrace
	tokLBrack
	tokRBrack
	tokLParen
	tokRParen
	tokSemi
	tokComma
)

type token struct {
	kind      tokenKind
	val       string
	line, col int
}

type scanner struct {
	buf               *bb.Buffer
	line, col         int
	prevLine, prevCol int
	peeked            token
	hasPeek           bool
}

func newScanner(buf *bb.Buffer) *scanner {
	return &scanner{buf: buf, line: 1, col: 1, prevLine: 1, prevCol: 1}
}

func (s *scanner) next() (token, error) {
	if s.hasPeek {
		s.hasPeek = false
		return s.peeked, nil
	}
	return s.scan()
}

func (s *scanner) peek() (token, error) {
	if s.hasPeek {
		return s.peeked, nil
	}
	t, err := s.scan()
	if err != nil {
		return token{}, err
	}
	s.peeked = t
	s.hasPeek = true
	return t, nil
}

func (s *scanner) scan() (token, error) {
	for {
		r, err := s.peekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return token{kind: tokEOF, line: s.line, col: s.col}, nil
			}
			return token{}, s.wrap(err)
		}
		if unicode.IsSpace(r) {
			if _, err = s.readRune(); err != nil {
				return token{}, s.wrap(err)
			}
			continue
		}
		if r == '#' {
			if err = s.skipComment(); err != nil {
				return token{}, err
			}
			continue
		}
		break
	}

	line, col := s.line, s.col
	r, err := s.peekRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return token{kind: tokEOF, line: line, col: col}, nil
		}
		return token{}, s.wrap(err)
	}

	switch r {
	case '{':
		_, _ = s.readRune()
		return token{kind: tokLBrace, val: "{", line: line, col: col}, nil
	case '}':
		_, _ = s.readRune()
		return token{kind: tokRBrace, val: "}", line: line, col: col}, nil
	case '[':
		_, _ = s.readRune()
		return token{kind: tokLBrack, val: "[", line: line, col: col}, nil
	case ']':
		_, _ = s.readRune()
		return token{kind: tokRBrack, val: "]", line: line, col: col}, nil
	case '(':
		_, _ = s.readRune()
		return token{kind: tokLParen, val: "(", line: line, col: col}, nil
	case ')':
		_, _ = s.readRune()
		return token{kind: tokRParen, val: ")", line: line, col: col}, nil
	case ';':
		_, _ = s.readRune()
		return token{kind: tokSemi, val: ";", line: line, col: col}, nil
	case ',':
		_, _ = s.readRune()
		return token{kind: tokComma, val: ",", line: line, col: col}, nil
	case '\'', '"':
		return s.scanQuoted(r)
	case '`':
		return s.scanRaw()
	default:
		if isDelim(r) {
			return token{}, fmt.Errorf("unic:%d:%d: unexpected %q", line, col, r)
		}
		return s.scanIdent()
	}
}

func (s *scanner) scanIdent() (token, error) {
	line, col := s.line, s.col

	b := pool.Bytes.Get()
	defer pool.Bytes.Put(b)

	for {
		r, err := s.peekRune()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return token{}, s.wrap(err)
		}
		if unicode.IsSpace(r) || isDelim(r) {
			break
		}
		r, err = s.readRune()
		if err != nil {
			return token{}, s.wrap(err)
		}
		b.B = utf8.AppendRune(b.B, r)
	}
	if len(b.B) == 0 {
		return token{}, fmt.Errorf("unic:%d:%d: empty token", line, col)
	}
	return token{kind: tokIdent, val: string(b.B), line: line, col: col}, nil
}

func (s *scanner) scanQuoted(quote rune) (token, error) {
	line, col := s.line, s.col
	if _, err := s.readRune(); err != nil {
		return token{}, s.wrap(err)
	}

	b := pool.Bytes.Get()
	defer pool.Bytes.Put(b)

	for {
		r, err := s.readRune()
		if errors.Is(err, io.EOF) {
			return token{}, fmt.Errorf("unic:%d:%d: unterminated string", line, col)
		}
		if err != nil {
			return token{}, s.wrap(err)
		}
		if r == quote {
			return token{kind: tokString, val: string(b.B), line: line, col: col}, nil
		}
		b.B = utf8.AppendRune(b.B, r)
	}
}

func (s *scanner) scanRaw() (token, error) {
	line, col := s.line, s.col
	for i := 0; i < 3; i++ {
		r, err := s.readRune()
		if err != nil {
			return token{}, fmt.Errorf("unic:%d:%d: unterminated raw string", line, col)
		}
		if r != '`' {
			return token{}, fmt.Errorf("unic:%d:%d: raw strings must start with ```", line, col)
		}
	}

	b := pool.Bytes.Get()
	defer pool.Bytes.Put(b)

	ticks := 0
	for {
		r, err := s.readRune()
		if errors.Is(err, io.EOF) {
			return token{}, fmt.Errorf("unic:%d:%d: unterminated raw string", line, col)
		}
		if err != nil {
			return token{}, s.wrap(err)
		}
		if r == '`' {
			ticks++
			if ticks == 3 {
				return token{kind: tokString, val: string(b.B), line: line, col: col}, nil
			}
			continue
		}
		for ticks > 0 {
			b.B = append(b.B, '`')
			ticks--
		}
		b.B = utf8.AppendRune(b.B, r)
	}
}

func (s *scanner) skipComment() error {
	for {
		r, err := s.readRune()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return s.wrap(err)
		}
		if r == '\n' {
			return nil
		}
	}
}

func (s *scanner) peekRune() (rune, error) {
	r, size, err := s.buf.ReadRune()
	if err != nil {
		return 0, err
	}
	if r == utf8.RuneError && size == 1 {
		return 0, fmt.Errorf("invalid UTF-8 sequence")
	}
	if err = s.buf.UnreadRune(); err != nil {
		return 0, err
	}
	return r, nil
}

func (s *scanner) readRune() (rune, error) {
	r, size, err := s.buf.ReadRune()
	if err != nil {
		return 0, err
	}
	if r == utf8.RuneError && size == 1 {
		return 0, fmt.Errorf("invalid UTF-8 sequence")
	}
	s.prevLine, s.prevCol = s.line, s.col
	if r == '\n' {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	return r, nil
}

func (s *scanner) wrap(err error) error {
	return fmt.Errorf("unic:%d:%d: %w", s.line, s.col, err)
}

func isDelim(r rune) bool {
	switch r {
	case '{', '}', '[', ']', '(', ')', '#', ';', ',', '\'', '"', '`':
		return true
	default:
		return false
	}
}

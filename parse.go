/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import (
	"fmt"

	"go.osspkg.com/bb"
)

type nodeKind int

const (
	nodeEmpty nodeKind = iota
	nodeScalar
	nodeList
	nodeMap
	nodeBlock
)

type pair struct {
	key, val *node
}

type field struct {
	key string
	val *node
}

type node struct {
	kind      nodeKind
	scalar    string
	list      []*node
	pairs     []*pair
	attrs     []*node
	fields    []*field
	line, col int
}

type parser struct {
	s *scanner
}

func parseDocument(buf *bb.Buffer) (*node, error) {
	p := &parser{s: newScanner(buf)}
	fields, err := p.parseFields(true)
	if err != nil {
		return nil, err
	}
	t, err := p.s.peek()
	if err != nil {
		return nil, err
	}
	if t.kind != tokEOF {
		return nil, fmt.Errorf("unic:%d:%d: unexpected %s", t.line, t.col, t.val)
	}
	return &node{kind: nodeBlock, fields: fields, line: 1, col: 1}, nil
}

func (p *parser) parseFields(top bool) ([]*field, error) {
	fields := make([]*field, 0, 10)
	for {
		t, err := p.s.peek()
		if err != nil {
			return nil, err
		}
		if t.kind == tokEOF {
			if top {
				return fields, nil
			}
			return nil, fmt.Errorf("unic:%d:%d: unclosed block", t.line, t.col)
		}
		if t.kind == tokRBrace {
			if top {
				return nil, fmt.Errorf("unic:%d:%d: unexpected '}'", t.line, t.col)
			}
			return fields, nil
		}
		f, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
}

func (p *parser) parseField() (*field, error) {
	key, err := p.s.next()
	if err != nil {
		return nil, err
	}
	if key.kind != tokIdent && key.kind != tokString {
		return nil, fmt.Errorf("unic:%d:%d: expected field name, got %s", key.line, key.col, tokenName(key))
	}

	t, err := p.s.peek()
	if err != nil {
		return nil, err
	}

	switch t.kind {
	case tokLBrack:
		list, err := p.parseList()
		if err != nil {
			return nil, err
		}
		if err = p.expect(tokSemi); err != nil {
			return nil, err
		}
		return &field{key: key.val, val: list}, nil
	case tokLParen:
		m, err := p.parseMap()
		if err != nil {
			return nil, err
		}
		if err = p.expect(tokSemi); err != nil {
			return nil, err
		}
		return &field{key: key.val, val: m}, nil
	case tokLBrace:
		block, err := p.parseBlock(nil)
		if err != nil {
			return nil, err
		}
		return &field{key: key.val, val: block}, nil
	case tokSemi:
		_, _ = p.s.next()
		return &field{key: key.val, val: &node{kind: nodeScalar, line: t.line, col: t.col}}, nil
	case tokEOF, tokRBrace:
		return nil, fmt.Errorf("unic:%d:%d: expected value or ';' after %s", key.line, key.col, key.val)
	default:
	}

	vals := make([]*node, 0, 10)
	for {
		t, err = p.s.peek()
		if err != nil {
			return nil, err
		}
		switch t.kind {
		case tokSemi:
			_, _ = p.s.next()
			return &field{key: key.val, val: valuesNode(vals, key.line, key.col)}, nil
		case tokLBrace:
			block, err := p.parseBlock(vals)
			if err != nil {
				return nil, err
			}
			return &field{key: key.val, val: block}, nil
		case tokIdent, tokString:
			tok, err := p.s.next()
			if err != nil {
				return nil, err
			}
			vals = append(vals, &node{kind: nodeScalar, scalar: tok.val, line: tok.line, col: tok.col})
		default:
			return nil, fmt.Errorf("unic:%d:%d: unexpected %s in field %s", t.line, t.col, tokenName(t), key.val)
		}
	}
}

func valuesNode(vals []*node, line, col int) *node {
	switch len(vals) {
	case 0:
		return &node{kind: nodeScalar, line: line, col: col}
	case 1:
		return vals[0]
	default:
		return &node{kind: nodeList, list: vals, line: line, col: col}
	}
}

func (p *parser) parseBlock(attrs []*node) (*node, error) {
	open, err := p.s.next()
	if err != nil {
		return nil, err
	}
	if open.kind != tokLBrace {
		return nil, fmt.Errorf("unic:%d:%d: expected '{'", open.line, open.col)
	}
	fields, err := p.parseFields(false)
	if err != nil {
		return nil, err
	}
	if err = p.expect(tokRBrace); err != nil {
		return nil, err
	}
	return &node{kind: nodeBlock, attrs: attrs, fields: fields, line: open.line, col: open.col}, nil
}

func (p *parser) parseList() (*node, error) {
	open, err := p.s.next()
	if err != nil {
		return nil, err
	}
	n := &node{kind: nodeList, line: open.line, col: open.col, list: make([]*node, 0, 4)}
	for {
		t, err := p.s.peek()
		if err != nil {
			return nil, err
		}
		if t.kind == tokRBrack {
			_, _ = p.s.next()
			return n, nil
		}
		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		n.list = append(n.list, item)
		t, err = p.s.peek()
		if err != nil {
			return nil, err
		}
		switch t.kind {
		case tokComma:
			_, _ = p.s.next()
		case tokRBrack:
			continue
		default:
			return nil, fmt.Errorf("unic:%d:%d: expected ',' or ']' in list", t.line, t.col)
		}
	}
}

func (p *parser) parseMap() (*node, error) {
	open, err := p.s.next()
	if err != nil {
		return nil, err
	}
	n := &node{kind: nodeMap, line: open.line, col: open.col, pairs: make([]*pair, 0, 4)}
	items := make([]*node, 0, 10)
	for {
		t, err := p.s.peek()
		if err != nil {
			return nil, err
		}
		if t.kind == tokRParen {
			_, _ = p.s.next()
			break
		}
		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		t, err = p.s.peek()
		if err != nil {
			return nil, err
		}
		switch t.kind {
		case tokComma:
			_, _ = p.s.next()
		case tokRParen:
			continue
		default:
			return nil, fmt.Errorf("unic:%d:%d: expected ',' or ')' in map", t.line, t.col)
		}
	}
	if len(items)%2 != 0 {
		return nil, fmt.Errorf("unic:%d:%d: map must contain an even number of values", n.line, n.col)
	}
	for i := 0; i < len(items); i += 2 {
		n.pairs = append(n.pairs, &pair{key: items[i], val: items[i+1]})
	}
	return n, nil
}

func (p *parser) parseValue() (*node, error) {
	t, err := p.s.peek()
	if err != nil {
		return nil, err
	}
	switch t.kind {
	case tokLBrack:
		return p.parseList()
	case tokLParen:
		return p.parseMap()
	case tokLBrace:
		return p.parseBlock(nil)
	case tokIdent, tokString:
		tok, err := p.s.next()
		if err != nil {
			return nil, err
		}
		return &node{kind: nodeScalar, scalar: tok.val, line: tok.line, col: tok.col}, nil
	default:
		return nil, fmt.Errorf("unic:%d:%d: expected value, got %s", t.line, t.col, tokenName(t))
	}
}

func (p *parser) expect(k tokenKind) error {
	t, err := p.s.next()
	if err != nil {
		return err
	}
	if t.kind != k {
		return fmt.Errorf("unic:%d:%d: expected %s, got %s", t.line, t.col, kindName(k), tokenName(t))
	}
	return nil
}

func tokenName(t token) string {
	if t.kind == tokEOF {
		return "EOF"
	}
	if t.val != "" {
		return t.val
	}
	return kindName(t.kind)
}

func kindName(k tokenKind) string {
	switch k {
	case tokEOF:
		return "EOF"
	case tokIdent:
		return "identifier"
	case tokString:
		return "string"
	case tokLBrace:
		return "'{'"
	case tokRBrace:
		return "'}'"
	case tokLBrack:
		return "'['"
	case tokRBrack:
		return "']'"
	case tokLParen:
		return "'('"
	case tokRParen:
		return "')'"
	case tokSemi:
		return "';'"
	case tokComma:
		return "','"
	default:
		return "token"
	}
}

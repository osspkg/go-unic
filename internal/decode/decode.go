package decode

import (
	"bufio"
	"fmt"
	"io"

	"go.osspkg.com/unic/internal/dict"
	"go.osspkg.com/unic/internal/node"
	"go.osspkg.com/unic/internal/splitter"
)

const (
	caseOpen = iota + 1
	caseValue
	caseAttach
	caseDeattach
	caseClose
)

type Decoder struct {
	scanner *bufio.Scanner
	block   *node.Block
	comment bool
}

func New(r io.Reader) *Decoder {
	scanner := bufio.NewScanner(r)
	scanner.Split(splitter.Func)
	return &Decoder{
		scanner: scanner,
		block:   node.NewBlock(),
		comment: false,
	}
}

func (v *Decoder) GetBlock() *node.Block {
	return v.block
}

func (v *Decoder) Decode() error {
	next := caseOpen

	for v.scanner.Scan() {
		data := v.scanner.Text()

		if v.ignore(data) {
			continue
		}

		switch data {
		case dict.KeyEnd:
			next = caseClose
		case dict.BlockOpen:
			next = caseAttach
		case dict.BlockClose:
			next = caseDeattach
		}

		switch next {
		case caseOpen:
			v.block = v.block.Next()
			v.block.Key().Set(data)
			next = caseValue
		case caseClose:
			v.block = v.block.Previous()
			next = caseOpen
		case caseValue:
			v.block.Key().Set(data)
		case caseAttach:
			next = caseOpen
		case caseDeattach:
			v.block = v.block.Previous()
			next = caseOpen
		}
	}

	if !v.block.IsRoot() {
		return fmt.Errorf("not all closing brackets of the block were found")
	}

	return nil
}

func (v *Decoder) ignore(data string) bool {
	if data == dict.NewLine && v.comment {
		v.comment = false
		return true
	}
	if data == dict.Comment && !v.comment {
		v.comment = true
		return true
	}
	if v.comment || data == dict.NewLine {
		return true
	}

	return false
}

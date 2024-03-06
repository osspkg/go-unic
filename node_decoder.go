package coma

import (
	"bufio"
	"fmt"
	"io"
)

const (
	stepOpen = iota + 1
	stepNext
)

type nodeDecoder struct {
	scanner *bufio.Scanner
	block   *block
	kv      *nodeData
	comment bool
}

func newNodeDecoder(r io.Reader) *nodeDecoder {
	scanner := bufio.NewScanner(r)
	scanner.Split(splitWordsFunc)
	return &nodeDecoder{
		scanner: scanner,
		block:   newNodeBlock(),
		kv:      newNodeData(),
		comment: false,
	}
}

func (v *nodeDecoder) RootNode() *block {
	for {
		if v.block.Parent == nil {
			break
		}
		v.block = v.block.Parent
	}
	return v.block
}

func (v *nodeDecoder) Decode() error {
	step := stepOpen

	for v.scanner.Scan() {
		data := v.scanner.Text()

		if v.needSkipWord(data) {
			continue
		}

		if step == stepOpen {
			switch data {
			case wordBlockClose:
				if err := v.backParent(); err != nil {
					return err
				}
			default:
				v.kv.Key = data
				step = stepNext
			}
			continue
		}

		switch data {
		case wordDataEnd:
			v.nextNodeData()
			step = stepOpen
		case wordBlockOpen:
			v.nextNodeBlock()
			step = stepOpen
		case wordBlockClose:
			if err := v.backParent(); err != nil {
				return err
			}
		default:
			v.kv.Values = append(v.kv.Values, data)
		}
	}

	return nil
}

func (v *nodeDecoder) nextNodeBlock() {
	nb := newNodeBlock()
	v.block.Child = append(v.block.Child, nb)
	nb.Parent = v.block
	nb.Data = v.kv
	v.kv = newNodeData()
	v.block = nb
}

func (v *nodeDecoder) nextNodeData() {
	v.block.SubData = append(v.block.SubData, v.kv)
	v.kv = newNodeData()
}

func (v *nodeDecoder) backParent() error {
	if v.block.Parent != nil {
		v.block = v.block.Parent
		return nil
	}
	return fmt.Errorf("node parse fail")
}

func (v *nodeDecoder) needSkipWord(data string) bool {
	if data == wordNewLine && v.comment {
		v.comment = false
		return true
	}
	if data == wordComment && !v.comment {
		v.comment = true
		return true
	}
	if v.comment || data == wordNewLine {
		return true
	}

	return false
}

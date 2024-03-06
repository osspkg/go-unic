package coma

import (
	"strings"
)

type nodeData struct {
	Key    string
	Values []string
}

func newNodeData() *nodeData {
	return &nodeData{
		Key:    "",
		Values: make([]string, 0, 2),
	}
}

func (v *nodeData) toString() string {
	if v == nil {
		return ""
	}
	values := make([]string, 0, len(v.Values)+1)
	values = append(values, v.Key)
	for _, value := range v.Values {
		if isMultiline([]byte(value)) {
			value = "`" + value + "`"
		}
		values = append(values, value)
	}
	return strings.Join(values, wordSpace)
}

// ---------------------------------------------------------------------------------------------------------------------

type block struct {
	Data    *nodeData
	SubData []*nodeData
	Child   []*block
	Parent  *block
}

func newNodeBlock() *block {
	return &block{
		Data:    nil,
		SubData: nil,
		Child:   nil,
		Parent:  nil,
	}
}

func (v *block) toString(c int) string {
	result := ""
	if v.Data != nil {
		result += wordNewLine + indentSpace(c) + v.Data.toString() + wordSpace + wordBlockOpen + wordNewLine
		c++
	}
	for _, datum := range v.SubData {
		result += indentSpace(c) + datum.toString() + wordDataEnd + wordNewLine
	}
	for _, datum := range v.Child {
		result += datum.toString(c)
	}
	if v.Data != nil {
		c--
		result += indentSpace(c) + wordBlockClose + wordNewLine
	}
	return result
}

func indentSpace(c int) string {
	return strings.Repeat(wordSpace, c*2)
}

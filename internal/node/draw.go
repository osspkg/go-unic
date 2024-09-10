package node

import (
	"io"
	"strings"

	"go.osspkg.com/unic/internal/dict"
)

func DrawKey(w io.StringWriter, k *Key) {
	w.WriteString(k.key) //nolint:errcheck
	for _, value := range k.values {
		if dict.IsMultiline([]byte(value)) {
			w.WriteString(dict.Space)      //nolint:errcheck
			w.WriteString(dict.Apostrophe) //nolint:errcheck
			w.WriteString(value)           //nolint:errcheck
			w.WriteString(dict.Apostrophe) //nolint:errcheck
			continue
		}
		w.WriteString(dict.Space) //nolint:errcheck
		w.WriteString(value)      //nolint:errcheck
	}
}

func DrawBlock(w io.StringWriter, level int, b *Block) {
	w.WriteString(indentSpace(level)) //nolint:errcheck
	DrawKey(w, b.Key())
	if !b.HasChild() {
		w.WriteString(dict.KeyEnd)  //nolint:errcheck
		w.WriteString(dict.NewLine) //nolint:errcheck
		return
	}

	w.WriteString(dict.Space)     //nolint:errcheck
	w.WriteString(dict.BlockOpen) //nolint:errcheck
	w.WriteString(dict.NewLine)   //nolint:errcheck

	level++
	for _, c := range b.Child() {
		DrawBlock(w, level, c)
	}
	level--

	w.WriteString(indentSpace(level)) //nolint:errcheck
	w.WriteString(dict.BlockClose)    //nolint:errcheck
	w.WriteString(dict.NewLine)       //nolint:errcheck

}

func indentSpace(c int) string {
	return strings.Repeat(dict.Space, c*4)
}

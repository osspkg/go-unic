package node

import (
	"fmt"
	"io"
	"strings"

	"go.osspkg.com/unic/internal/dict"
)

func DrawKey(w io.Writer, k *Key) {
	fmt.Fprint(w, k.key) //nolint:errcheck
	for _, value := range k.values {
		if dict.IsMultiline([]byte(value)) {
			fmt.Fprint(w, dict.Space)      //nolint:errcheck
			fmt.Fprint(w, dict.Apostrophe) //nolint:errcheck
			fmt.Fprint(w, value)           //nolint:errcheck
			fmt.Fprint(w, dict.Apostrophe) //nolint:errcheck
			continue
		}
		fmt.Fprint(w, dict.Space) //nolint:errcheck
		fmt.Fprint(w, value)      //nolint:errcheck
	}
}

func DrawBlock(w io.Writer, level int, b *Block) {
	fmt.Fprint(w, indentSpace(level)) //nolint:errcheck
	DrawKey(w, b.Key())
	if !b.HasChild() {
		fmt.Fprint(w, dict.KeyEnd)  //nolint:errcheck
		fmt.Fprint(w, dict.NewLine) //nolint:errcheck
		return
	}

	fmt.Fprint(w, dict.Space)     //nolint:errcheck
	fmt.Fprint(w, dict.BlockOpen) //nolint:errcheck
	fmt.Fprint(w, dict.NewLine)   //nolint:errcheck

	level++
	for _, c := range b.Child() {
		DrawBlock(w, level, c)
	}
	level--

	fmt.Fprint(w, indentSpace(level)) //nolint:errcheck
	fmt.Fprint(w, dict.BlockClose)    //nolint:errcheck
	fmt.Fprint(w, dict.NewLine)       //nolint:errcheck

}

func indentSpace(c int) string {
	return strings.Repeat(dict.Space, c*4)
}

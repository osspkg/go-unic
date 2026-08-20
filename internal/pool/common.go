/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package pool

import (
	"go.osspkg.com/bb"
)

type TransBytes struct {
	B []byte
}

func (b *TransBytes) Reset() {
	b.B = b.B[:0]
}

var Bytes = New[*TransBytes](func() *TransBytes {
	return &TransBytes{B: make([]byte, 0, 512)}
})

var Buffer = New[*bb.Buffer](func() *bb.Buffer {
	return bb.New(512)
})

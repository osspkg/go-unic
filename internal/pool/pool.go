/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package pool

import "sync"

type TPool interface {
	Reset()
}

type Pool[T TPool] struct {
	pool *sync.Pool
}

func New[T TPool](callNew func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{New: func() any { return callNew() }},
	}
}

func (v *Pool[T]) Get() T {
	return v.pool.Get().(T)
}

func (v *Pool[T]) Put(t T) {
	t.Reset()
	v.pool.Put(t)
}

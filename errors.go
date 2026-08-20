/*
 *  Copyright (c) 2024-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unic

import "errors"

//nolint:unused
var (
	errUnknownNode       = errors.New("unknown node")
	errDuplicateField    = errors.New("duplicate field")
	errArrayManyValues   = errors.New("too many values for array")
	errMapKeyString      = errors.New("map key must be string")
	errCanSetEmbeddedPtr = errors.New("cannot set embedded pointer")
	errNotStruct         = errors.New("not a struct")
	errNotMaps           = errors.New("not a maps")
	errCantAllocatePtr   = errors.New("cannot allocate pointer")
	errNestedStructBlock = errors.New("nested struct must be a block")
)

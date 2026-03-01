// cache.go
//
// Copyright (C) 2023-2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultCache(t *testing.T) {
	context := "TestDefaultCache"
	key := "key"
	miss := Get0(context, key)
	require.Equal(t, "", miss)
	value := "value"
	Put0(context, key, value)
	hit := Get0(context, key)
	require.Equal(t, value, hit)
	err := Flush()
	require.NoError(t, err)
}

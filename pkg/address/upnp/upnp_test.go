// upnp_test.go
//
// Copyright (C) 2023-2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package upnp

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/ddns-updater/pkg/address"
)

func TestUPnPFinder(t *testing.T) {
	cfg := &UPnPFinderConfig{
		FinderConfig: address.FinderConfig{
			IPv4:    true,
			IPv6:    true,
			Private: true,
		},
	}
	finder := NewUPnPFinder(cfg)
	require.NotNil(t, finder)
	finder.Run()
}

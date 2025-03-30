// web_test.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package web

import (
	"testing"

	"github.com/tdrn-org/ddns-updater/pkg/address"
	"github.com/stretchr/testify/require"
)

func TestWebFinder(t *testing.T) {
	cfg := &WebFinderConfig{
		FinderConfig: address.FinderConfig{
			IPv4:    true,
			IPv6:    true,
			Private: false,
		},
		IPv4Specs:     [][2]string{{"https://ip4only.me/api/", "IPv4,([^,]*),.*"}},
		IPv6Specs:     [][2]string{{"https://ip6only.me/api/", "IPv6,([^,]*),.*"}},
		TLSSkipVerify: false,
	}
	finder := NewWebFinder(cfg)
	require.NotNil(t, finder)
	finder.Run()
}

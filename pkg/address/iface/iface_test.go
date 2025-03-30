// iface_test.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package iface

import (
	"net"
	"testing"

	"github.com/tdrn-org/ddns-updater/pkg/address"
	"github.com/stretchr/testify/require"
)

func TestIFaceFinder(t *testing.T) {
	testif := testInterface(t)
	cfg := &IFaceFinderConfig{
		FinderConfig: address.FinderConfig{
			IPv4:    true,
			IPv6:    true,
			Private: true,
		},
		IFace: testif.Name,
	}
	finder := NewIFaceFinder(cfg)
	require.NotNil(t, finder)
	finder.Run()
}

func testInterface(t *testing.T) net.Interface {
	ifaces, err := net.Interfaces()
	require.NoError(t, err)
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		require.NoError(t, err)
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if ok && ipnet.IP.IsGlobalUnicast() {
				return iface
			}
		}
	}
	panic("no suitable interface available")
}

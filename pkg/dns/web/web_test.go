// web_test.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package web

import (
	"net"
	"testing"

	"github.com/tdrn-org/ddns-updater/pkg/dns"
	"github.com/stretchr/testify/require"
)

func TestWebUpdater(t *testing.T) {
	// See https://help.dyn.com/remote-access-api/test-account/
	cfg := &WebUpdaterConfig{
		UpdaterConfig: dns.UpdaterConfig{
			Enabled: true,
			Host:    "test5.customtest.dyndns.org",
		},
		Username:      "test",
		Password:      "test",
		URL:           "https://test.dyndns.org/v3/update?hostname={hostname}&myip={myipv4},{myipv6}",
		TLSSkipVerify: true,
	}
	updater := NewWebUpdater(cfg)
	require.NotNil(t, updater)
	err := updater.Merge([]net.IP{net.ParseIP("10.1.1.1").To4(), net.ParseIP("fd10::1").To16()}, false, false)
	require.NoError(t, err)
}

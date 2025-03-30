// dummy.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dummy

import (
	"fmt"
	"net"

	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/tdrn-org/ddns-updater/pkg/dns"
	"github.com/rs/zerolog"
)

const Name = "dns_dummy"

type DummyUpdaterConfig struct {
	dns.UpdaterConfig
}

func NewDummyUpdater(cfg *DummyUpdaterConfig) dns.Updater {
	name := Name
	logger := logging.RootLogger().With().Str("updater", name).Logger()
	return &dummyUpdater{cfg: cfg, name: name, logger: logger}
}

type dummyUpdater struct {
	cfg    *DummyUpdaterConfig
	name   string
	logger zerolog.Logger
}

func (u *dummyUpdater) Name() string {
	return u.name
}

func (u *dummyUpdater) Merge(ips []net.IP, force bool, pretend bool) error {
	u.logger.Info().Msgf("Updating host '%s'...", u.cfg.Host)
	for _, ip := range ips {
		if len(ip) == 4 {
			fmt.Printf("DDNS update '%s' A %s\n", u.cfg.Host, ip.String())
		} else {
			fmt.Printf("DDNS update '%s' AAAA %s\n", u.cfg.Host, ip.String())
		}
	}
	if !pretend {
		u.logger.Info().Msg("DDNS update applied")
	} else {
		u.logger.Warn().Msg("DDNS update skipped due to pretend")
	}
	return nil
}

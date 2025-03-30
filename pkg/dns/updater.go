// updater.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dns

import (
	"net"

	"github.com/tdrn-org/ddns-updater/internal/logging"
)

type Updater interface {
	Name() string
	Merge(ips []net.IP, force bool, pretend bool) error
}

type UpdaterConfig struct {
	Enabled bool   `toml:"enabled"`
	Host    string `toml:"host"`
}

func (cfg *UpdaterConfig) IsEnabled(updaterName string) bool {
	if cfg.Enabled && cfg.Host == "" {
		logger := logging.RootLogger()
		logger.Warn().Msgf("Ignoring DNS updater '%s' due to missing host name", updaterName)
	}
	return cfg.Enabled && cfg.Host != ""
}

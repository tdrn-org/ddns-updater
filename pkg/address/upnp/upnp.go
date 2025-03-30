// upnp.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package upnp

import (
	"context"
	"fmt"
	"net"

	"github.com/tdrn-org/ddns-updater/internal/cache"
	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/tdrn-org/ddns-updater/pkg/address"
	"github.com/rs/zerolog"
	"gitlab.com/NebulousLabs/go-upnp"
)

const igdLocationCacheKey = "igd_location"

type UPnPFinderConfig struct {
	address.FinderConfig
}

func NewUPnPFinder(cfg *UPnPFinderConfig) address.Finder {
	name := "address_upnp"
	logger := logging.RootLogger().With().Str("finder", name).Logger()
	return &upnpFinder{cfg: cfg, name: name, logger: logger}
}

type upnpFinder struct {
	cfg    *UPnPFinderConfig
	name   string
	logger zerolog.Logger
}

func (f *upnpFinder) Name() string {
	return f.name
}

func (f *upnpFinder) Run() ([]net.IP, error) {
	f.logger.Info().Msg("Discovering UPnP devices...")
	found := make([]net.IP, 0)
	igd := f.lookupIGD()
	if igd == nil {
		return found, nil
	}
	f.logger.Debug().Msgf("Using IGD at '%s'", igd.Location())
	ipString, err := igd.ExternalIP()
	if err != nil {
		return nil, fmt.Errorf("failed to query external IP from IGD '%s'\n\tcause: %v", igd.Location(), err)
	}
	f.logger.Debug().Msgf("IGD returned external IP '%s'", ipString)
	ip := net.ParseIP(ipString)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address '%s' returned from IGD '%s'\n\tcause: %v", ipString, igd.Location(), err)
	}
	// Map IPv4-Mapped addresses back to IPv4
	ipv4 := ip.To4()
	if ipv4 != nil {
		ip = ipv4
	}
	if f.cfg.IsMatch(ip) {
		f.logger.Info().Msgf("Found address %s", ip)
		found = append(found, ip)
	}
	return found, nil
}

func (f *upnpFinder) lookupIGD() *upnp.IGD {
	location := cache.Get0(f.Name(), igdLocationCacheKey)
	if location != "" {
		igd, _ := upnp.Load(location)
		cache.Put0(f.Name(), igdLocationCacheKey, igd.Location())
		return igd
	}
	ctx := context.Background()
	igd, err := upnp.DiscoverCtx(ctx)
	if err != nil {
		f.logger.Warn().Msgf("No IGD device discovered")
		return nil
	}
	cache.Put0(f.Name(), igdLocationCacheKey, igd.Location())
	return igd
}

// iface.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package iface

import (
	"fmt"
	"net"
	"path"

	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/tdrn-org/ddns-updater/pkg/address"
	"github.com/rs/zerolog"
)

type IFaceFinderConfig struct {
	address.FinderConfig
	IFace string `toml:"interface"`
}

func NewIFaceFinder(cfg *IFaceFinderConfig) address.Finder {
	name := fmt.Sprintf("address_interface[%s]", cfg.IFace)
	logger := logging.RootLogger().With().Str("finder", name).Logger()
	return &ifaceFinder{cfg: cfg, name: name, logger: logger}
}

type ifaceFinder struct {
	cfg    *IFaceFinderConfig
	name   string
	logger zerolog.Logger
}

func (f *ifaceFinder) Name() string {
	return f.name
}

func (f *ifaceFinder) Run() ([]net.IP, error) {
	f.logger.Info().Msgf("Checking interfaces '%s'...", f.cfg.IFace)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("%s: Failed to retrieve interfaces\n\tcause: %v", f.Name(), err)
	}
	found := make([]net.IP, 0)
	for _, iface := range ifaces {
		match, err := path.Match(f.cfg.IFace, iface.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid interface match name '%s'\n\tcause: %v", f.cfg.IFace, err)
		}
		if match {
			addrs, err := iface.Addrs()
			if err != nil {
				return nil, fmt.Errorf("%s: Failed to get addresses of interface '%s'\n\tcause: %v", f.Name(), f.cfg.IFace, err)
			}
			for _, addr := range addrs {
				f.logger.Debug().Msgf("Considering address %s:%s...", iface.Name, addr)
				ipnet, ok := addr.(*net.IPNet)
				if ok {
					ip := ipnet.IP
					if f.cfg.IsMatch(ip) {
						f.logger.Info().Msgf("Found address %s", ip)
						found = append(found, ip)
					}
				}
			}
		}
	}
	return found, nil
}

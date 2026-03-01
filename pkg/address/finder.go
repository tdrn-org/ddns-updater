// finder.go
//
// Copyright (C) 2023-2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package address

import "net"

type Finder interface {
	Name() string
	Run() ([]net.IP, error)
}

type FinderConfig struct {
	IPv4    bool `toml:"ipv4"`
	IPv6    bool `toml:"ipv6"`
	Private bool `toml:"private"`
}

func (cfg *FinderConfig) IsEnabled() bool {
	return cfg.IPv4 || cfg.IPv6
}

func (cfg *FinderConfig) IsMatch(ip net.IP) bool {
	if !ip.IsGlobalUnicast() {
		return false
	}
	if !cfg.Private && ip.IsPrivate() {
		return false
	}
	switch len(ip) {
	case 4:
		return cfg.IPv4
	case 16:
		return cfg.IPv6
	}
	return false
}

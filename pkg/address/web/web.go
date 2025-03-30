// web.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package web

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"

	"github.com/tdrn-org/ddns-updater/internal/httpclient"
	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/tdrn-org/ddns-updater/pkg/address"
	"github.com/rs/zerolog"
)

type WebFinderConfig struct {
	address.FinderConfig
	IPv4Specs     [][2]string `toml:"ipv4_specs"`
	IPv6Specs     [][2]string `toml:"ipv6_specs"`
	TLSSkipVerify bool        `toml:"tls_skip_verify"`
}

func NewWebFinder(cfg *WebFinderConfig) address.Finder {
	name := "address.web"
	logger := logging.RootLogger().With().Str("finder", name).Logger()
	return &webFinder{cfg: cfg, name: name, logger: logger}
}

type webFinder struct {
	cfg    *WebFinderConfig
	name   string
	logger zerolog.Logger
}

func (f *webFinder) Name() string {
	return f.name
}

func (f *webFinder) Run() ([]net.IP, error) {
	found := make([]net.IP, 0)
	client := httpclient.PrepareClient(httpclient.DefaultTimeout, f.cfg.TLSSkipVerify)
	if f.cfg.FinderConfig.IPv4 {
		for _, spec := range f.cfg.IPv4Specs {
			ip, err := f.runIPv4Spec(client, spec)
			if err != nil {
				continue
			}
			if f.cfg.IsMatch(ip) {
				f.logger.Info().Msgf("Found address %s", ip)
				found = append(found, ip)
				break
			}
		}
	}
	if f.cfg.FinderConfig.IPv6 {
		for _, spec := range f.cfg.IPv6Specs {
			ip, err := f.runIPv6Spec(client, spec)
			if err != nil {
				continue
			}
			if f.cfg.IsMatch(ip) {
				f.logger.Info().Msgf("Found address %s", ip)
				found = append(found, ip)
				break
			}
		}
	}
	return found, nil
}

func (f *webFinder) runIPv4Spec(client *http.Client, spec [2]string) (net.IP, error) {
	ip, err := f.runSpec(client, spec)
	if err != nil {
		return ip, err
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil, fmt.Errorf("")
	}
	return ipv4, nil
}

func (f *webFinder) runIPv6Spec(client *http.Client, spec [2]string) (net.IP, error) {
	ip, err := f.runSpec(client, spec)
	if err != nil {
		return ip, err
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf("")
	}
	return ipv6, nil
}

func (f *webFinder) runSpec(client *http.Client, spec [2]string) (net.IP, error) {
	url, err := url.Parse(spec[0])
	if err != nil {
		return nil, fmt.Errorf("invalid url '%s'\n\tcause: %v", spec[0], err)
	}
	rspexp, err := regexp.Compile(spec[1])
	if err != nil {
		return nil, fmt.Errorf("invalid regexp '%s'\n\tcause: %v", spec[1], err)
	}
	f.logger.Info().Msgf("Querying address service '%s'...", url.String())
	rsp, err := client.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query url '%s'\n\tcause: %v", url.String(), err)
	}
	defer rsp.Body.Close()
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read url '%s'\n\tcause: %v", url.String(), err)
	}
	loc := rspexp.FindSubmatchIndex(data)
	if loc == nil {
		return nil, fmt.Errorf("failed to decode ip")
	}
	ip := f.parseIP(data, loc)
	if ip == nil {
		return nil, fmt.Errorf("failed to parse ip\n\tcause: %v", err)
	}
	return ip, nil
}

func (f *webFinder) parseIP(data []byte, loc []int) net.IP {
	if len(loc) == 4 {
		return net.ParseIP(string(data[loc[2]:loc[3]]))
	}
	return net.ParseIP(string(data[loc[0]:loc[1]]))
}

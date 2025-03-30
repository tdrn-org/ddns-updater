// web.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package web

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"

	"github.com/tdrn-org/ddns-updater/internal/buildinfo"
	"github.com/tdrn-org/ddns-updater/internal/httpclient"
	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/tdrn-org/ddns-updater/pkg/dns"
	"github.com/rs/zerolog"
)

const Name = "dns_web"

type WebUpdaterConfig struct {
	dns.UpdaterConfig
	Username      string `toml:"username"`
	Password      string `toml:"password"`
	URL           string `toml:"url"`
	TLSSkipVerify bool   `toml:"tls_skip_verify"`
}

func NewWebUpdater(cfg *WebUpdaterConfig) dns.Updater {
	name := Name
	logger := logging.RootLogger().With().Str("updater", name).Logger()
	return &webUpdater{cfg: cfg, name: name, logger: logger}
}

type webUpdater struct {
	cfg    *WebUpdaterConfig
	name   string
	logger zerolog.Logger
}

func (u *webUpdater) Name() string {
	return u.name
}

func (u *webUpdater) Merge(ips []net.IP, force bool, pretend bool) error {
	u.logger.Info().Msgf("Updating host '%s'...", u.cfg.Host)
	url := u.cfg.URL
	url = strings.ReplaceAll(url, "{hostname}", u.cfg.Host)
	url, ipv4UpdateMsg, err := u.replaceIP(url, "{myipv4}", ips, "A")
	if err != nil {
		return err
	}
	url, ipv6UpdateMsg, err := u.replaceIP(url, "{myipv6}", ips, "AAAA")
	if err != nil {
		return err
	}
	u.logger.Debug().Msgf("Using update URL '%s'", url)
	if !pretend {
		u.invokeURL(url)
	}
	if ipv4UpdateMsg != "" {
		fmt.Println(ipv4UpdateMsg)
	}
	if ipv6UpdateMsg != "" {
		fmt.Println(ipv6UpdateMsg)
	}
	return nil
}

func (u *webUpdater) replaceIP(url string, variable string, ips []net.IP, rtype string) (string, string, error) {
	if !strings.Contains(url, variable) {
		return url, "", nil
	}
	for _, ip := range ips {
		if len(ip) == len(rtype)*4 {
			ipString := ip.String()
			updateMsg := fmt.Sprintf("DDNS update '%s' %s %s", u.cfg.Host, rtype, ipString)
			return strings.ReplaceAll(url, variable, ipString), updateMsg, nil
		}
	}
	return "", "", fmt.Errorf("failed to substitute parameter %s due to missing IP address", variable)
}

func (u *webUpdater) invokeURL(url string) error {
	client := httpclient.PrepareClient(httpclient.DefaultTimeout, u.cfg.TLSSkipVerify)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to construct service request (cause: %v)", err)
	}
	userAgent := fmt.Sprintf("ddns-updater/%s (%s; %s) %s https://tdrn-org.github.io/ddns-updater/", buildinfo.Version(), runtime.GOOS, runtime.GOARCH, buildinfo.Timestamp())
	req = httpclient.ReqUserAgent(req, userAgent)
	req = httpclient.ReqAuthorization(req, u.cfg.Username, u.cfg.Password)
	rsp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute update request (cause: %v)", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("update request failed (status: %s)", rsp.Status)
	}
	return nil
}

// ddnsupdater.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package ddnsupdater

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/kong"
	"github.com/tdrn-org/ddns-updater/internal/buildinfo"
	"github.com/tdrn-org/ddns-updater/internal/cache"
	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/tdrn-org/ddns-updater/pkg/address"
	"github.com/tdrn-org/ddns-updater/pkg/address/iface"
	"github.com/tdrn-org/ddns-updater/pkg/address/upnp"
	addressweb "github.com/tdrn-org/ddns-updater/pkg/address/web"
	"github.com/tdrn-org/ddns-updater/pkg/dns"
	"github.com/tdrn-org/ddns-updater/pkg/dns/dummy"
	"github.com/tdrn-org/ddns-updater/pkg/dns/route53"
	dnsweb "github.com/tdrn-org/ddns-updater/pkg/dns/web"
	"github.com/rs/zerolog"
)

const defaultConfigFile = "/etc/ddns-updater/ddns-updater.toml"

const cacheContext = "global"
const lastUpdateCacheKey = "last_update"

type ddnsupdater struct {
	Config     string `help:"Use this config file"`
	Pretend    bool   `help:"Only show changes, but do not apply them"`
	Force      bool   `help:"Force DNS update"`
	Verbose    bool   `help:"Enable verbose output"`
	Debug      bool   `help:"Enable debug output"`
	Ansi       bool   `help:"Force ANSI colored output"`
	ResetCache bool   `help:"Reset cache"`
	logger     zerolog.Logger
}

func (cmd *ddnsupdater) Run() error {
	config, err := cmd.readConfig()
	if err != nil {
		return err
	}
	cmd.applyGlobalConfig(config)
	cmd.logger.Info().Msg(buildinfo.FullVersion())
	finders := cmd.evalFinderConfigs(config)
	updaters := cmd.evalUpdaterConfigs(config)
	if len(finders) == 0 {
		cmd.logger.Warn().Msg("No address finders configured")
	} else if len(updaters) == 0 {
		cmd.logger.Warn().Msg("No DNS updaters configured")
	} else {
		err := cmd.rundFindAndUpdate(finders, updaters)
		if err != nil {
			return err
		}
		// Only flush if DNS has changed
		if !cmd.Pretend {
			err = cache.Flush()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd *ddnsupdater) rundFindAndUpdate(finders []address.Finder, updaters []dns.Updater) error {
	ips, err := cmd.findIPs(finders)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		cmd.logger.Warn().Msg("No addresses found, nothing to update")
		return nil
	}
	ips = cmd.normalizeIPs(ips)
	if !cmd.updateRequired(ips) {
		return nil
	}
	return cmd.updateIPs(updaters, ips)
}

func (cmd *ddnsupdater) findIPs(finders []address.Finder) ([]net.IP, error) {
	cmd.logger.Info().Msg("Gathering addresses...")
	found := make([]net.IP, 0)
	for _, finder := range finders {
		ips0, err := finder.Run()
		if err != nil {
			return nil, err
		}
		found = append(found, ips0...)
	}
	return found, nil
}

func (cmd *ddnsupdater) updateIPs(updaters []dns.Updater, ips []net.IP) error {
	cmd.logger.Info().Msg("Updating DNS...")
	for _, updater := range updaters {
		err := updater.Merge(ips, cmd.Force, cmd.Pretend)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd *ddnsupdater) normalizeIPs(ips []net.IP) []net.IP {
	normalized := make([]net.IP, 0)
	for _, ip := range ips {
		insertAt := 0
		for _, current := range normalized {
			comparison := bytes.Compare(ip, current)
			if comparison == 0 {
				insertAt = -1
				break
			} else if comparison > 0 {
				break
			}
			insertAt = insertAt + 1
		}
		if insertAt >= 0 {
			normalized = append(normalized, nil)
			copy(normalized[insertAt+1:], normalized[insertAt:])
			normalized[insertAt] = ip
		}
	}
	return normalized
}

func (cmd *ddnsupdater) updateRequired(ips []net.IP) bool {
	update := make([]string, len(ips))
	for i, ip := range ips {
		update[i] = ip.String()
	}
	if cmd.Force {
		cmd.logger.Info().Msg("DNS update forced")
		cache.Put(cacheContext, lastUpdateCacheKey, update)
		return true
	}
	lastUpdate := cache.Get(cacheContext, lastUpdateCacheKey)
	updateRequired := len(update) != len(lastUpdate)
	for i, lastUpdateI := range lastUpdate {
		if updateRequired {
			break
		}
		updateRequired = lastUpdateI != update[i]
	}
	if updateRequired {
		cmd.logger.Info().Msg("DNS update required")
		cache.Put(cacheContext, lastUpdateCacheKey, update)
	} else {
		cmd.logger.Info().Msg("DNS still up-to-date, skipping update")
	}
	return updateRequired
}

func (cmd *ddnsupdater) readConfig() (*ddnsupdaterConfig, error) {
	configFile := cmd.Config
	if configFile == "" {
		configFile = defaultConfigFile
	}
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s'\n\tcause: %v", configFile, err)
	}
	var config ddnsupdaterConfig
	err = toml.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file '%s'\n\tcause: %v", configFile, err)
	}
	return &config, nil
}

func (cmd *ddnsupdater) applyGlobalConfig(config *ddnsupdaterConfig) {
	logger := logging.NewConsoleLogger(os.Stdout, cmd.Ansi || config.Global.Ansi)
	var level zerolog.Level
	if cmd.Debug || config.Global.Debug {
		level = zerolog.DebugLevel
	} else if cmd.Verbose || config.Global.Verbose {
		level = zerolog.InfoLevel
	} else {
		level = zerolog.WarnLevel
	}
	logging.UpdateRootLogger(logger, level)
	if config.Global.CacheEnabled {
		cache.EnableCaching(config.Global.CacheDuration, cmd.ResetCache)
	}
}

func (cmd *ddnsupdater) evalFinderConfigs(config *ddnsupdaterConfig) []address.Finder {
	finders := make([]address.Finder, 0)
	if config.IFaceAddressFinder.IsEnabled() {
		finder := iface.NewIFaceFinder(&config.IFaceAddressFinder)
		finders = append(finders, finder)
	}
	if config.UPnPAddressFinder.IsEnabled() {
		finder := upnp.NewUPnPFinder(&config.UPnPAddressFinder)
		finders = append(finders, finder)
	}
	if config.WebAddressFinder.IsEnabled() {
		finder := addressweb.NewWebFinder(&config.WebAddressFinder)
		finders = append(finders, finder)
	}
	return finders
}

func (cmd *ddnsupdater) evalUpdaterConfigs(config *ddnsupdaterConfig) []dns.Updater {
	updaters := make([]dns.Updater, 0)
	if config.DummyUpdaterConfig.IsEnabled(dummy.Name) {
		updater := dummy.NewDummyUpdater(&config.DummyUpdaterConfig)
		updaters = append(updaters, updater)
	}
	if config.Route53UpdaterConfig.IsEnabled(route53.Name) {
		updater := route53.NewRoute53Updater(&config.Route53UpdaterConfig)
		updaters = append(updaters, updater)
	}
	if config.WebUpdaterConfig.IsEnabled(dnsweb.Name) {
		updater := dnsweb.NewWebUpdater(&config.WebUpdaterConfig)
		updaters = append(updaters, updater)
	}
	return updaters
}

type ddnsupdaterConfig struct {
	Global               globalConfig                 `toml:"global"`
	IFaceAddressFinder   iface.IFaceFinderConfig      `toml:"address_interface"`
	UPnPAddressFinder    upnp.UPnPFinderConfig        `toml:"address_upnp"`
	WebAddressFinder     addressweb.WebFinderConfig   `toml:"address_web"`
	DummyUpdaterConfig   dummy.DummyUpdaterConfig     `toml:"dns_dummy"`
	Route53UpdaterConfig route53.Route53UpdaterConfig `toml:"dns_route53"`
	WebUpdaterConfig     dnsweb.WebUpdaterConfig      `toml:"dns_web"`
}

type globalConfig struct {
	Verbose       bool          `toml:"verbose"`
	Debug         bool          `toml:"debug"`
	Ansi          bool          `toml:"ansi"`
	CacheEnabled  bool          `toml:"cache_enabled"`
	CacheDuration time.Duration `toml:"cache_duration"`
}

var cmd ddnsupdater = ddnsupdater{logger: logging.RootLogger()}

func Run() error {
	err := kong.Parse(&cmd).Run()
	if err != nil {
		cmd.logger.Error().Msgf("ddns-updater command failed\n\tcause: %v", err)
	}
	return err
}

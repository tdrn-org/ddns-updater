// cache.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/rs/zerolog"
)

type Provider interface {
	Get(key string) []string
	Put(key string, values []string, validity time.Duration)
	Flush() error
}

var cacheProvider Provider = newDefaultProvider(false, false)
var cacheValidity time.Duration = time.Hour * 24

func EnableCaching(validatiy time.Duration, resetCache bool) {
	cacheProvider = newDefaultProvider(true, resetCache)
	cacheValidity = validatiy
}

func Get0(context string, key string) string {
	values := Get(context, key)
	if values == nil && len(values) != 1 {
		return ""
	}
	return values[0]
}

func Get(context string, key string) []string {
	return cacheProvider.Get(context + "." + key)
}

func Put0(context string, key string, value string) {
	Put(context, key, []string{value})
}

func Put(context string, key string, values []string) {
	cacheProvider.Put(context+"."+key, values, cacheValidity)
}

func Flush() error {
	return cacheProvider.Flush()
}

type defaultProvider struct {
	modified  bool
	cache     map[string]*cacheEntry
	cacheFile string
	logger    zerolog.Logger
}

type cacheEntry struct {
	ValidTill time.Time `toml:"valid_till"`
	Values    []string  `toml:"values"`
}

func newDefaultProvider(persistent bool, resetCache bool) Provider {
	cache := make(map[string]*cacheEntry)
	cacheFile := ""
	logger := logging.RootLogger().With().Str("cache", "default").Logger()
	if persistent {
		userCacheDir, err := os.UserCacheDir()
		if err == nil {
			cacheFile = filepath.Join(userCacheDir, "ddns-updater", "cache.toml")
		} else {
			logger.Warn().Err(err).Msgf("Failed to determine cache file path; caching disabled\n\tcause: %v", err)
		}
	}
	if cacheFile != "" {
		if !resetCache {
			_, err := toml.DecodeFile(cacheFile, &cache)
			if err != nil && !os.IsNotExist(err) {
				logger.Warn().Err(err).Msgf("Failed to read cache file '%s'\n\tcause: %v", cacheFile, err)
			}
		} else {
			err := os.Truncate(cacheFile, 0)
			if err != nil && !os.IsNotExist(err) {
				logger.Warn().Err(err).Msgf("Failed to reset cache file '%s'\n\tcause: %v", cacheFile, err)
			}
		}
	}
	return &defaultProvider{modified: false, cache: cache, cacheFile: cacheFile, logger: logger}
}

func (p *defaultProvider) Get(key string) []string {
	found := p.cache[key]
	if found == nil || time.Now().After(found.ValidTill) {
		return nil
	}
	return found.Values
}

func (p *defaultProvider) Put(key string, values []string, validity time.Duration) {
	p.modified = true
	p.cache[key] = &cacheEntry{ValidTill: time.Now().Add(validity), Values: values}
}

func (p *defaultProvider) Flush() error {
	if !p.modified {
		return nil
	}
	now := time.Now()
	for key, entry := range p.cache {
		if now.After(entry.ValidTill) {
			delete(p.cache, key)
		}
	}
	if p.cacheFile != "" {
		p.logger.Debug().Msgf("Writing cache file '%s'...", p.cacheFile)
		cacheFileDir := filepath.Dir(p.cacheFile)
		err := os.MkdirAll(cacheFileDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create cache file directory '%s'\n\tcause: %v", cacheFileDir, err)
		}
		fd, err := os.Create(p.cacheFile)
		if err != nil {
			return fmt.Errorf("failed to create cache file '%s'\n\tcause: %v", p.cacheFile, err)
		}
		defer p.closeCacheFile(fd)
		err = toml.NewEncoder(fd).Encode(&p.cache)
		if err != nil {
			return fmt.Errorf("failed to write cache file '%s'\n\tcause: %v", p.cacheFile, err)
		}
	}
	return nil
}

func (p *defaultProvider) closeCacheFile(fd *os.File) {
	err := fd.Close()
	if err != nil {
		p.logger.Warn().Err(err).Msgf("Failed to close cache file '%s'\n\t:cause: %v", p.cacheFile, err)
	}
}

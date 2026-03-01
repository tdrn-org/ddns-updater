// ddnsupdater_test.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package ddnsupdater

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	cmd.Config = "../../ddns-updater.toml"
	cfg, err := cmd.readConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

const ddnsUpdaterCmd string = "ddns-updater"
const ddnsUpdaterConfigFlag string = "--config=./testdata/ddns-updater.toml"
const ddnsUpdaterPretendFlag string = "--pretend"
const ddnsUpdaterResetCacheFlag string = "--reset-cache"

func TestRunUpdate(t *testing.T) {
	os.Args = []string{ddnsUpdaterCmd, ddnsUpdaterConfigFlag, ddnsUpdaterPretendFlag, ddnsUpdaterResetCacheFlag}
	err := Run()
	require.NoError(t, err)

	os.Args = []string{ddnsUpdaterCmd, ddnsUpdaterConfigFlag}
	err = Run()
	require.NoError(t, err)

	os.Args = []string{ddnsUpdaterCmd, ddnsUpdaterConfigFlag}
	err = Run()
	require.NoError(t, err)
}

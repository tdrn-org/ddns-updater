// main.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package main

import "github.com/tdrn-org/ddns-updater/internal/ddnsupdater"

func main() {
	_ = ddnsupdater.Run()
}

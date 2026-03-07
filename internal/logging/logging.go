// logging.go
//
// Copyright (C) 2023-2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package logging

import (
	"log"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
)

var rootLogger = NewConsoleLogger(os.Stdout, false)

func UpdateRootLogger(logger zerolog.Logger, level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	rootLogger = logger
	log.SetFlags(0)
	log.SetOutput(logger)
	rootLogger.Debug().Msg("root logger updated")
}

func RootLogger() zerolog.Logger {
	return rootLogger
}

func NewConsoleLogger(out *os.File, forceColor bool) zerolog.Logger {
	color := forceColor
	if !color {
		color = isatty.IsTerminal(out.Fd())
	}
	return zerolog.New(zerolog.ConsoleWriter{Out: out, NoColor: !color}).With().Timestamp().Logger()
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

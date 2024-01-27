package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	zerologging "github.com/fanyang89/zerologging/v1"

	"github.com/fanyang89/eaglexport/cmd"
)

func main() {
	zerologging.WithConsoleLog(zerolog.InfoLevel)
	app := cmd.NewApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Err(err).Msg("run app failed")
	}
}

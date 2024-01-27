package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/fanyang89/eagle-sync/cmd"
)

func main() {
	app := cmd.NewApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Err(err).Msg("run app failed")
	}
}

package main

import (
	"github.com/r-mol/ObsidianBot/internal/app"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := app.GetApp().Execute(); err != nil {
		log.Fatal(err)
	}
}

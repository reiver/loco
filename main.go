package main

import (
	"loco/srv/log"
)

func main() {
	log := logsrv.Begin()
	defer log.End()

	log.Highlightf("loco ⚡")
	defer log.Highlightf("loco 👻")
}

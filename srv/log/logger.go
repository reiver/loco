package logsrv

import (
	"io"
	"os"

	"codeberg.org/reiver/go-log"

	_ "loco/flg"
)

var writer io.Writer = os.Stdout

var logger log.Logger = log.CreateLogger(writer)

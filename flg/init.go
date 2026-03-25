package flg

import (
	"flag"

	"codeberg.org/reiver/go-log"
)

var Help bool

func init() {
	// Register the logger related flags:
	//
	//	• -v
	//	• -vv
	//	• -vvv
	//	• -vvvv
	//	• -vvvvv
	//	• -vvvvvv
	//	• -vvvvvvv
	//
	// The logger related flags will automatically be used by the logger.
	if err := log.DeclareFlags(); nil != err {
		panic(err)
	}

	flag.BoolVar(&Help, "help", false, "show help")

	flag.Parse()
}

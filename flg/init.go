package flg

import (
	"flag"

	"codeberg.org/reiver/go-log"
)

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

	flag.Parse()
}

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"loco/flg"
	"loco/srv/help"
	"loco/srv/log"
)

func main() {
	log := logsrv.Begin()
	defer log.End()

	// --help always runs built-in help
	if flg.Help {
		helpsrv.Run(os.Stdout)
		return
	}

	args := flag.Args()

	// No subcommand → built-in help
	if 0 == len(args) {
		helpsrv.Run(os.Stdout)
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	// "help" subcommand: try loco-help on $PATH first, fall back to built-in
	if "help" == subcommand {
		path, err := exec.LookPath("loco-help")
		if nil == err {
			dispatch(path, subArgs)
			return
		}
		helpsrv.Run(os.Stdout)
		return
	}

	// Normal dispatch: find loco-<subcommand> on $PATH
	binaryName := "loco-" + subcommand
	path, err := exec.LookPath(binaryName)
	if nil != err {
		fmt.Fprintf(os.Stderr, "loco: '%s' is not a loco command. See 'loco help'.\n", subcommand)
		os.Exit(1)
	}

	dispatch(path, subArgs)
}

func dispatch(path string, args []string) {
	argv := append([]string{path}, args...)
	err := syscall.Exec(path, argv, os.Environ())
	if nil != err {
		fmt.Fprintf(os.Stderr, "loco: failed to execute %s: %v\n", path, err)
		os.Exit(1)
	}
}

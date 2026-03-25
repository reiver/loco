package helpsrv

import (
	"fmt"
	"io"
	"os"
	"sort"

	"loco/lib/cmds"
	"loco/srv/builtins"
)

func Run(w io.Writer) {
	pathEnv := os.Getenv("PATH")
	found := libcmds.ScanPath("loco-", pathEnv)

	printHelp(w, found)
}

func printHelp(w io.Writer, found []libcmds.Command) {
	fmt.Fprintln(w, "loco — manage self-hosted apps")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage: loco <command> [args...]")

	printBuiltIn(w, found)
	printThirdParty(w, found)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run 'loco <command> --help' for more information on a specific command.")
}

func printBuiltIn(w io.Writer, found []libcmds.Command) {
	installedBuiltIns := map[string]bool{}
	for _, cmd := range found {
		if builtinssrv.IsBuiltIn(cmd.Name) {
			installedBuiltIns[cmd.Name] = true
		}
	}

	all := builtinssrv.Commands()

	maxLen := 0
	for _, cmd := range all {
		if len(cmd.Name) > maxLen {
			maxLen = len(cmd.Name)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Built-in commands:")
	for _, cmd := range all {
		marker := ""
		if !installedBuiltIns[cmd.Name] {
			marker = " (not installed)"
		}
		fmt.Fprintf(w, "    %-*s  %s%s\n", maxLen, cmd.Name, cmd.Description, marker)
	}
}

func printThirdParty(w io.Writer, found []libcmds.Command) {
	var thirdParty []libcmds.Command
	for _, cmd := range found {
		if !builtinssrv.IsBuiltIn(cmd.Name) {
			thirdParty = append(thirdParty, cmd)
		}
	}

	if 0 == len(thirdParty) {
		return
	}

	sort.Slice(thirdParty, func(i, j int) bool {
		return thirdParty[i].Name < thirdParty[j].Name
	})

	maxLen := 0
	for _, cmd := range thirdParty {
		if len(cmd.Name) > maxLen {
			maxLen = len(cmd.Name)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Third-party commands:")
	for _, cmd := range thirdParty {
		fmt.Fprintf(w, "    %-*s  %s\n", maxLen, cmd.Name, cmd.Path)
	}
}

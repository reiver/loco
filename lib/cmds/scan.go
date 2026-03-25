package libcmds

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ScanPath(prefix string, pathEnv string) []Command {
	seen := map[string]bool{}
	var commands []Command

	dirs := filepath.SplitList(pathEnv)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if nil != err {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			if !isExecutable(entry) {
				continue
			}

			subcommand := strings.TrimPrefix(name, prefix)
			if "" == subcommand {
				continue
			}
			if seen[subcommand] {
				continue
			}
			seen[subcommand] = true

			commands = append(commands, Command{
				Name: subcommand,
				Path: filepath.Join(dir, name),
			})
		}
	}

	return commands
}

func isExecutable(entry fs.DirEntry) bool {
	info, err := entry.Info()
	if nil != err {
		return false
	}
	return info.Mode()&0111 != 0
}

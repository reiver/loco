package builtinssrv

type Command struct {
	Name        string
	Description string
}

var commands = []Command{
	{"init", "Initialize a loco root"},
	{"install", "Install an app"},
	{"update", "Update an installed app"},
	{"start", "Start an app"},
	{"stop", "Stop an app"},
	{"remove", "Remove an app"},
	{"list", "List installed apps"},
	{"status", "Show app status"},
	{"logs", "Show app logs"},
	{"snapshot", "Create a data snapshot"},
	{"snapshots", "List snapshots"},
	{"restore", "Restore from a snapshot"},
	{"scheduler", "Manage scheduled tasks"},
}

func Commands() []Command {
	result := make([]Command, len(commands))
	copy(result, commands)
	return result
}

func IsBuiltIn(name string) bool {
	for _, cmd := range commands {
		if cmd.Name == name {
			return true
		}
	}
	return false
}

func Description(name string) string {
	for _, cmd := range commands {
		if cmd.Name == name {
			return cmd.Description
		}
	}
	return ""
}

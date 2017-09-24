package main

import (
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/cmd"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	if len(os.Args) <= 1 ||
		(len(os.Args) == 1 && os.Args[1] == "help") {
		showUsage()
		return 1
	}
	switch os.Args[1] {
	case "get":
		return cmd.Get(os.Args[2:])
	case "rm":
		return cmd.Rm(os.Args[2:])
	case "query":
		return cmd.Query(os.Args[2:])
	case "profile":
		return cmd.Profile(os.Args[2:])
	case "enable":
		return cmd.Enable(os.Args[2:])
	case "disable":
		return cmd.Disable(os.Args[2:])
	case "version":
		return cmd.Version(os.Args[2:])
	case "help":
		fmt.Printf("[ERROR] Run 'volt %s -help' to see its help\n", os.Args[2])
		return 2
	default:
		fmt.Fprintln(os.Stderr, "[ERROR] Unknown command '"+os.Args[1]+"'")
		return 3
	}
}

func showUsage() {
	fmt.Println(`
Usage
  volt COMMAND ARGS

Command
  get [-l] [-u] [-v] {repository}
    Install / Upgrade vim plugin, and system plugconf files from
    https://github.com/vim-volt/plugconf-templates

  rm {repository}
    Uninstall vim plugin and system plugconf files

  query [-j] [-l] [{repository}]
    Output queried vim plugin info

  profile [get]
    Get current profile name

  profile set {name}
    Set profile name

  profile show {name}
    Show profile info

  profile new {name}
    Create new profile

  profile destroy {name}
    Delete profile

  profile add {name} {repository} [{repository2} ...]
    Add one or more repositories to profile

  profile rm {name} {repository} [{repository2} ...]
    Remove one or more repositories to profile

  version
    Show volt command version
`)
}

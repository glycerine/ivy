package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(2)
	}

	switch flag.Arg(0) {
	case "migrate",
		"invite-user",
		"create-account",
		"add-account-user",
		"create-team",
		"add-team-user",
		"create-project",
		"grant-project",
		"list-users",
		"list-accounts",
		"list-teams",
		"list-projects":
		fmt.Fprintf(os.Stderr, "ivy-control-admin %s is planned but not implemented yet\n", flag.Arg(0))
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", flag.Arg(0))
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `usage: ivy-control-admin <command>

Commands:
  migrate
  invite-user
  create-account
  add-account-user
  create-team
  add-team-user
  create-project
  grant-project
  list-users
  list-accounts
  list-teams
  list-projects
`)
}

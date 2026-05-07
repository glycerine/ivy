package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/control"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "seed-alpha":
		seedAlpha(os.Args[2:])
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
		fmt.Fprintf(os.Stderr, "ivy-control-admin %s is planned; use seed-alpha for the current spike workflow\n", os.Args[1])
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func seedAlpha(args []string) {
	fs := flag.NewFlagSet("seed-alpha", flag.ExitOnError)
	dsn := fs.String("dsn", envDefault("IVY_CONTROL_DATABASE_DSN", "postgres://ivyvue_app:ivyvue_app_dev@127.0.0.1:5432/ivyvue?sslmode=disable"), "ivyvue PostgreSQL DSN")
	email := fs.String("email", "", "tester email address")
	displayName := fs.String("display-name", "", "tester display name")
	idpIssuer := fs.String("idp-issuer", "email", "identity-provider issuer to seed")
	idpSubject := fs.String("idp-subject", "", "identity-provider subject; defaults to email")
	accountSlug := fs.String("account", "alpha", "billing account slug")
	accountName := fs.String("account-name", "Alpha Testers", "billing account display name")
	teamSlug := fs.String("team", "core", "team slug")
	teamName := fs.String("team-name", "Core", "team display name")
	projectSlug := fs.String("project", "client-server", "project slug")
	projectName := fs.String("project-name", "Client/server example", "project display name")
	role := fs.String("role", string(control.ProjectRoleWrite), "project role: read, write, or admin")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "parse seed-alpha flags: %v\n", err)
		os.Exit(2)
	}
	store, err := control.OpenPostgresStore(*dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open postgres store: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()
	view, err := store.SeedAlphaTester(context.Background(), control.SeedAlphaTesterParams{
		Email:       *email,
		DisplayName: *displayName,
		IDPIssuer:   *idpIssuer,
		IDPSubject:  *idpSubject,
		AccountSlug: *accountSlug,
		AccountName: *accountName,
		TeamSlug:    *teamSlug,
		TeamName:    *teamName,
		ProjectSlug: *projectSlug,
		ProjectName: *projectName,
		ProjectRole: control.ProjectRole(*role),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "seed alpha tester: %v\n", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(view); err != nil {
		fmt.Fprintf(os.Stderr, "encode result: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `usage: ivy-control-admin <command>

Commands:
  seed-alpha
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

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

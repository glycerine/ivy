package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/glycerine/ivy/goivy/svk/server"
	"github.com/glycerine/ivy/goivy/svk/server/db"
	"github.com/jackc/pgx/v5"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "migrate":
		if err := migrate(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func migrate(args []string) error {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	dsn := flags.String("dsn", "", "Postgres DSN; defaults to IVYSVK_DB_DSN")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *dsn == "" {
		cfg, err := server.LoadConfigFromEnv()
		if err != nil && os.Getenv("IVYSVK_DB_DSN") == "" {
			return err
		}
		*dsn = cfg.DatabaseDSN
	}
	if *dsn == "" {
		return fmt.Errorf("missing Postgres DSN")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, *dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return db.DefaultMigrator().Apply(ctx, conn)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: ivysvk-admin migrate [-dsn postgres://...]\n")
}

package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"webplow/internal/auth"
	"webplow/internal/config"
)

func main() {
	cfg := config.Load()
	store, err := auth.NewStore(cfg.TokenFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: webplow-token add <name>")
			os.Exit(1)
		}
		t, err := store.Add(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Token created for %q:\n%s\n", t.Name, t.Key)

	case "list":
		tokens := store.List()
		if len(tokens) == 0 {
			fmt.Println("No tokens.")
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tKEY\tCREATED")
		for _, t := range tokens {
			fmt.Fprintf(w, "%s\t%s\t%s\n", t.Name, t.Key, t.CreatedAt.Format("2006-01-02 15:04"))
		}
		w.Flush()

	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: webplow-token delete <key>")
			os.Exit(1)
		}
		if err := store.Delete(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Deleted.")

	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: webplow-token <add|list|delete> [args]")
	os.Exit(1)
}

package main

import (
	"fmt"
	"os"

	"glance/internal/cli"
	licensecontent "glance/internal/license"
)

func main() {
	cfg, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if cfg.ShowLicense {
		fmt.Println(licensecontent.Text)
		return
	}

	if cfg.ShowHelp {
		fmt.Print(cli.HelpText())
		return
	}

	if err := cli.Run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

package main

import (
	"flag"
	"fmt"
	"os"

	"etf-insight/core"
	"etf-insight/utils"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	runOnce := flag.Bool("run-once", false, "run update once and exit")
	flag.Parse()

	app, err := core.New(*configPath)
	if err != nil {
		utils.Fatal("Failed to initialize application", err)
	}

	if *runOnce {
		app.RunOnce()
		return
	}

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}

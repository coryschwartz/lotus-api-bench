package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/coryschwartz/lotus-api-bench/commands"
)

func main() {
	app := cli.App{
		Name:                 "lotus-api-bench",
		Usage:                "benchmark lotus and the lotus gateway",
		EnableBashCompletion: true,
		ArgsUsage:            "benchmark [benchmark]...",
		Commands: cli.Commands{
			commands.ListCommand,
			commands.BenchCommand,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

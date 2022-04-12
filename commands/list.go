package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/coryschwartz/lotus-api-bench/benchmarks"
)

var ListCommand = &cli.Command{
	Name:   "list",
	Action: listBenchmarks,
}

func listBenchmarks(cctx *cli.Context) error {
	for k := range benchmarks.Map() {
		fmt.Println(k)
	}
	return nil
}

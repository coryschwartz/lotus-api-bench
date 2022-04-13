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
	var keys []string
	if cctx.Bool("gateway") {
		for k := range benchmarks.GwMap() {
			keys = append(keys, k)
		}
	} else {
		for k := range benchmarks.Map() {
			keys = append(keys, k)
		}
	}
	for _, k := range keys {
		fmt.Println(k)
	}
	return nil
}

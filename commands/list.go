package commands

import (
	"fmt"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/coryschwartz/lotus-api-bench/benchmarks"
)

var ListCommand = &cli.Command{
	Name:   "list",
	Action: listBenchmarks,
}

func listBenchmarks(cctx *cli.Context) error {
	var keys []string
	for k := range benchmarks.Map() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k)
	}
	return nil
}

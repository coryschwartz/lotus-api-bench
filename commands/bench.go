package commands

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/coryschwartz/lotus-api-bench/benchmarks"
	lapi "github.com/filecoin-project/lotus/api"
	cliutil "github.com/filecoin-project/lotus/cli/util"
	"github.com/urfave/cli/v2"
)

var BenchCommand = &cli.Command{
	Name:   "bench",
	Action: runBenchmarks,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "concurrency",
			Aliases: []string{"c"},
			Value:   1,
		},
		&cli.DurationFlag{
			Name:    "timeout",
			Usage:   "how long should each test be conducted",
			Aliases: []string{"t"},
			Value:   time.Minute,
		},
		&cli.IntFlag{
			Name:    "qps",
			Usage:   "queries per second to send",
			Aliases: []string{"q"},
			Value:   10,
		},
		&cli.DurationFlag{
			Name:    "sleep",
			Usage:   "how long to sleep between each benchmark",
			Aliases: []string{"s"},
			Value:   time.Second,
		},
	},
}

func runBenchmarks(cctx *cli.Context) error {
	concurrency := cctx.Int("concurrency")
	timeout := cctx.Duration("timeout")
	delay := time.Second / time.Duration(cctx.Int("qps"))
	sleep := cctx.Duration("sleep")
	var benches []string

	if cctx.NArg() > 0 {
		benches = cctx.Args().Slice()
	} else {
		for k := range benchmarks.Map() {
			benches = append(benches, k)
		}
	}
	sort.Strings(benches)

	bmap := benchmarks.Map()
	if len(benches) == 0 {
		for k := range bmap {
			benches = append(benches, k)
		}
	}

	for _, bench := range benches {
		bfunc := bmap[bench]
		ctx, cancel := context.WithTimeout(cctx.Context, timeout)
		defer cancel()
		f := func(api lapi.Gateway, results benchmarks.Results) error {
			return bfunc(ctx, delay, api, results)
		}
		results, err := runConcurrently(concurrency, cctx, f)
		if err != nil {
			return err
		}
		printResults(results, bench)
		time.Sleep(sleep)
	}
	return nil
}

func runConcurrently(concurrency int, cctx *cli.Context, f func(lapi.Gateway, benchmarks.Results) error) ([]benchmarks.Results, error) {
	errs := make(chan error)
	go func() {
		for {
			select {
			case err := <-errs:
				if err != nil {
					fmt.Fprintf(os.Stderr, "error while executing benchmark: %v", err)
				}
			}
		}
	}()

	var results []benchmarks.Results

	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		api, _, err := cliutil.GetGatewayAPI(cctx)
		if err != nil {
			return nil, err
		}
		wg.Add(1)
		res := make(benchmarks.Results)
		results = append(results, res)
		go func() {
			errs <- f(api, res)
			wg.Done()
		}()
	}
	wg.Wait()
	return results, nil
}

func printResults(results []benchmarks.Results, bench string) {
	writer := csv.NewWriter(os.Stdout)
	header := []string{"bench", "worker"}
	var keys []string
	for k := range results[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	header = append(header, keys...)
	writer.Write(header)

	row := make([]string, len(header))
	row[0] = bench
	for i, res := range results {
		row[1] = fmt.Sprintf("%d", i)
		for j, key := range keys {
			row[j+2] = fmt.Sprintf("%d", res[key])
		}
		writer.Write(row)
	}
	writer.Flush()
}

package commands

import (
	"context"
	"fmt"
	"os"
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
			Aliases: []string{"t"},
			Value:   time.Minute,
		},
		&cli.DurationFlag{
			Name:    "delay",
			Aliases: []string{"d"},
			Value:   time.Millisecond,
		},
	},
}

func runBenchmarks(cctx *cli.Context) error {
	concurrency := cctx.Int("concurrency")
	timeout := cctx.Duration("timeout")
	delay := cctx.Duration("delay")
	ctx, _ := context.WithTimeout(cctx.Context, timeout)
	var benches []string

	if cctx.NArg() > 0 {
		benches = cctx.Args().Slice()
	} else {
		for k := range benchmarks.Map() {
			benches = append(benches, k)
		}
	}
	if cctx.Bool("gateway") {
		api, _, err := cliutil.GetGatewayAPI(cctx)
		if err != nil {
			return nil
		}
		return runGatewayBenchmarks(ctx, concurrency, delay, benches, api)
	} else {
		api, _, err := cliutil.GetFullNodeAPIV1(cctx)
		if err != nil {
			return err
		}
		return runFullBenchmarks(ctx, concurrency, delay, benches, api)
	}
}

func runGatewayBenchmarks(ctx context.Context, concurrency int, delay time.Duration, benches []string, api lapi.Gateway) error {
	bmap := benchmarks.GwMap()
	if len(benches) == 0 {
		for k := range bmap {
			benches = append(benches, k)
		}
	}

	for _, bench := range benches {
		bfunc := bmap[bench]
		f := func(results benchmarks.Results) error {
			return bfunc(ctx, delay, api, results)
		}
		results := runConcurrently(concurrency, f)
		printResults(results)
	}
	return nil
}

func runFullBenchmarks(ctx context.Context, concurrency int, delay time.Duration, benches []string, api lapi.FullNode) error {
	bmap := benchmarks.Map()
	if len(benches) == 0 {
		for k := range bmap {
			benches = append(benches, k)
		}
	}

	for _, bench := range benches {
		bfunc := bmap[bench]
		f := func(results benchmarks.Results) error {
			return bfunc(ctx, delay, api, results)
		}
		results := runConcurrently(concurrency, f)
		printResults(results)
	}
	return nil
}

func runConcurrently(concurrency int, f func(benchmarks.Results) error) []benchmarks.Results {
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
		wg.Add(1)
		res := make(benchmarks.Results)
		results = append(results, res)
		go func() {
			errs <- f(res)
			wg.Done()
		}()
	}
	wg.Wait()
	return results
}
func printResults(results []benchmarks.Results) {
	avgResults := make(benchmarks.Results)
	for i, res := range results {
		fmt.Printf("\n\nRun %d\n", i)
		for k, v := range res {
			fmt.Println(k, v)
			avgv := v / float64(len(results))
			if c, ok := avgResults[k]; ok {
				avgResults[k] = c + avgv
			} else {
				avgResults[k] = avgv
			}
		}
	}
	fmt.Println("Average results:")
	for k, v := range avgResults {
		fmt.Println(k, v)
	}
}

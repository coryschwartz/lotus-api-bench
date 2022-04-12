package commands

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/coryschwartz/lotus-api-bench/benchmarks"
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
	bmap := benchmarks.Map()
	concurrency := cctx.Int("concurrency")
	timeout := cctx.Duration("timeout")
	delay := cctx.Duration("delay")
	var benches []string

	if cctx.NArg() > 0 {
		benches = cctx.Args().Slice()
	} else {
		for k := range benchmarks.Map() {
			benches = append(benches, k)
		}
	}

	api, closer, err := cliutil.GetFullNodeAPIV1(cctx)
	if err != nil {
		return err
	}
	defer closer()

	fmt.Println("Running benchmarks:", benches)

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

	ctx, cancel := context.WithTimeout(cctx.Context, timeout)
	defer cancel()

	for _, bench := range benches {
		fmt.Println(bench)
		bfunc := bmap[bench]
		var results []benchmarks.Results

		wg := sync.WaitGroup{}
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			res := make(benchmarks.Results)
			results = append(results, res)
			go func() {
				errs <- bfunc(ctx, delay, api, res)
				wg.Done()
			}()
		}
		wg.Wait()
		printResults(results)
	}

	return nil
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

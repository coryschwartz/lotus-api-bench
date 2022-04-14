package benchmarks

import (
	"context"
	"time"

	lapi "github.com/filecoin-project/lotus/api"
)

type Results map[string]int64

type BenchFunc func(context.Context, time.Duration, lapi.Gateway, Results) error

type BenchMap map[string]BenchFunc

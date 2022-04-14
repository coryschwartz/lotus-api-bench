package benchmarks

var (
	bmap = BenchMap{
		"genesis":       GetGenesisBench,
		"head":          HeadBench,
		"inspectminers": InspectMiners,
		"walkback":      WalkBack,
	}
)

func Map() BenchMap {
	return bmap
}

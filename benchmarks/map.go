package benchmarks

var (
	bmap = BenchMap{
		"head": HeadBench,
	}
	gwbmap = GwBenchMap{
		"head2": GwHeadBench,
	}
)

func Map() BenchMap {
	return bmap
}

func GwMap() GwBenchMap {
	return gwbmap
}

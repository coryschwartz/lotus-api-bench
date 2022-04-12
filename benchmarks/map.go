package benchmarks

var (
	bmap = BenchMap{
		"bogus": BogusBenchFunc,
		"head":  HeadBench,
	}
)

func Map() BenchMap {
	return bmap
}

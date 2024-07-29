package iter_test

import (
	"strconv"
	"testing"

	"github.com/sourcegraph/conc/iter"
)

func BenchmarkForEachSeq(b *testing.B) {
	for _, count := range []int{0, 1, 8, 100, 1000, 10000, 100000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			ints := make([]int, count)
			for i := 0; i < b.N; i++ {
				iter.ForEachSeq(iter.Slice(ints), func(i *int) bool {
					*i = 0
					return true
				})
			}
		})
	}
}

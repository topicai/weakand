package weakand

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func shuffledSlice(size int) []int {
	r := make([]int, size)
	for i := range r {
		r[i] = i
	}

	for i := range r {
		j := rand.Intn(i + 1)
		r[i], r[j] = r[j], r[i]
	}

	return r
}

func TestResultHeap(t *testing.T) {
	assert := assert.New(t)

	size := 10
	mh := NewResultHeap(size)

	for _, s := range shuffledSlice(1024 * 1024) {
		mh.Grow(Result{
			Posting: &Posting{DocId: DocId(s)},
			Score:   float64(s)})
	}
	assert.Equal(size, mh.Len())

	mh.Sort()
	for i := 0; i < mh.Len(); i++ {
		assert.Equal(float64(1024*1024-i-1), mh.rank[i].Score)
	}
}

package weakand

import (
	"bufio"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/huichen/sego"
	"github.com/stretchr/testify/assert"
)

var (
	testingCorpus = []string{
		"apple pie",
		"apple iphone",
		"iphone jailbreak"}

	sgmt *sego.Segmenter
)

func testBuildIndex() *SearchIndex {
	guaranteeSegmenter(&sgmt)

	ch := make(chan string)
	go func() {
		for _, d := range testingCorpus {
			ch <- d
		}
		close(ch)
	}()
	return NewIndex(NewVocab(nil), sgmt).BatchAdd(ch)
}

func TestBuildIndex(t *testing.T) {
	idx := testBuildIndex()

	assert := assert.New(t)

	assert.Equal(4, len(idx.Vocab.Terms))
	assert.Equal(4, len(idx.Vocab.TermIndex))

	assert.Equal(len(testingCorpus), len(idx.Fwd))
	assert.Equal(4, len(idx.Ivt))

	for i := range idx.Ivt {
		assert.True(sort.IsSorted(idx.Ivt[i]))
	}

	assert.Equal(2, len(idx.Ivt[idx.Vocab.IdOrAdd("apple")]))
	assert.Equal(1, len(idx.Ivt[idx.Vocab.IdOrAdd("pie")]))
	assert.Equal(2, len(idx.Ivt[idx.Vocab.IdOrAdd("iphone")]))
	assert.Equal(1, len(idx.Ivt[idx.Vocab.IdOrAdd("jailbreak")]))

	assert.Equal(2, idx.Fwd[documentHash(testingCorpus[0])].Len)
	assert.Equal(2, idx.Fwd[documentHash(testingCorpus[1])].Len)
	assert.Equal(2, idx.Fwd[documentHash(testingCorpus[2])].Len)
}

func TestDocumentHashCollision(t *testing.T) {
	WithFile(path.Join(gosrc(), "github.com/wangkuiyi/weakand/testdata/internet-zh.num"),
		func(f *os.File) {
			dict := make(map[DocId][]string)
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				fs := strings.Fields(scanner.Text())
				if len(fs) == 2 {
					content := fs[1]
					did := documentHash(content)
					if _, ok := dict[did]; ok {
						t.Errorf("Collision between %v and %v", content, dict[did])
					}
					dict[did] = append(dict[did], content)
				}
			}
			if e := scanner.Err(); e != nil {
				t.Errorf("Reading %s error: %v", f.Name(), e)
			}
		})
}

func gosrc() string {
	return path.Join(os.Getenv("GOPATH"), "src")
}

func guaranteeSegmenter(sgmt **sego.Segmenter) {
	if *sgmt == nil {
		s := new(sego.Segmenter)
		s.LoadDictionary(path.Join(gosrc(),
			"github.com/huichen/sego/data/dictionary.txt"))
		*sgmt = s
	}
}

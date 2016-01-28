package weakand

import (
	"flag"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	pretty       bool
	indexDumpDir string
)

func init() {
	flag.BoolVar(&pretty, "pretty", false, "Pretty print index and frontier when calling Search")
	flag.StringVar(&indexDumpDir, "indexDir", "/tmp", "Directory containing index dumps")
}

func TestSearch(t *testing.T) {
	guaranteeSegmenter(&sgmt)

	idx := testBuildIndex()
	rs := idx.Search(strings.Join(idx.Vocab.Terms, " "), 10, pretty) // Pretty print intermediate steps.
	assert.Equal(t, len(testingCorpus), len(rs))                     // All documents should be retrieved.
	for _, r := range rs {
		assert.Equal(t, 0.5, r.Score) // Jaccard coeffcient of all documents should be 1/2.
	}
}

func TestSearchWithAAAI14Titles(t *testing.T) {
	testWithBigData(t,
		"github.com/wangkuiyi/weakand/testdata/aaai14papers.txt",
		"incomplete ontologies",
		"aaai14titlesindex.csv")
}

func TestSearchWithZhWikiNews(t *testing.T) {
	testWithBigData(t,
		"github.com/wangkuiyi/weakand/testdata/zhwikinews.txt",
		"中药商",
		"zhwikinewsindex.csv")
}

func testWithBigData(t *testing.T, corpusFile string, query string, indexDumpFile string) {
	guaranteeSegmenter(&sgmt)

	idx := NewIndexFromFile(
		path.Join(gosrc(), corpusFile),
		sgmt,
		path.Join(indexDumpDir, indexDumpFile))

	q := NewQuery(query, idx.Vocab, sgmt)

	for _, r := range idx.Search(query, 10, pretty) {
		doc := strings.ToLower(r.Literal)

		contain := false
		for qterm := range q.Terms {
			contain = contain || strings.Contains(doc, idx.Vocab.Term(qterm))
		}
		assert.True(t, contain)
	}
}

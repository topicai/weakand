package weakand

import (
	"os"
	"strings"
	"testing"
)

func TestPrettyPrint(t *testing.T) {
	guaranteeSegmenter(&sgmt)

	idx := testBuildIndex()
	idx.Pretty(NewPlotTable(os.Stdout), nil, nil, 0)

	query := NewDocument(
		strings.Join(idx.Vocab.Terms, " "), // query includes all terms.
		idx.Vocab,
		sgmt)
	fr := newFrontier(query, idx)
	idx.Pretty(NewPlotTable(os.Stdout), fr.terms, fr.postings, fr.cur)
}

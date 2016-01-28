package weakand

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"

	"github.com/olekukonko/tablewriter"
)

// Print the forwar and inverted index using vocab.  terms, postings
// and currentDoc are supposed fields from type Frontier.  If they are
// not nil, also shows the fronteir on the plot of index.
func (idx *SearchIndex) Pretty(table Table, terms []TermId, postings []int, currentDoc DocId) {
	// Convert terms and postings into a map TermdId->DocId
	termDoc := make(map[TermId]DocId)
	for i, t := range terms {
		if postings[i] < len(idx.Ivt[t]) {
			termDoc[t] = idx.Ivt[t][postings[i]].DocId
		} else {
			termDoc[t] = EndOfPostingList
		}
	}

	// Construct a posting list containing all documents, and use
	// PostList's sortablility to sort them and get docId->index mapping.
	ps := make(PostingList, 0, len(idx.Fwd))
	for d, _ := range idx.Fwd {
		ps = append(ps, Posting{DocId: d, TF: 0})
	}
	ps = append(ps, Posting{DocId: EndOfPostingList, TF: 0})
	sort.Sort(ps)

	docIdx := make(map[DocId]int)
	for i, p := range ps {
		docIdx[p.DocId] = i
	}
	docIdx[EndOfPostingList] = len(ps)

	row := []string{"Term"}
	for _, p := range ps {
		s := fmt.Sprintf("%016x", p.DocId)
		if currentDoc == p.DocId {
			s = "●" + s
		} else {
			s = " " + s
		}
		row = append(row, s)
	}
	table.SetHeader(row)

	// NOTE: Do not range over idx.Ivt, which is a map and range is random.
	for termId, term := range idx.Vocab.Terms {
		pl := idx.Ivt[TermId(termId)]
		row = make([]string, len(idx.Fwd))
		for _, p := range pl {
			mark := "○"
			if p.DocId == termDoc[TermId(termId)] {
				mark = "●"
			}
			row[docIdx[p.DocId]] = mark
		}
		table.AddRow(append(append([]string{term}, row...), ""))
	}

	row = make([]string, len(idx.Fwd))
	for d, c := range idx.Fwd {
		row[docIdx[d]] = c.Pretty(idx.Vocab)
	}
	table.SetFooter(append(append([]string{" "}, row...), " "))

	table.Done() // Send output
}

// Interface Table has two implementations: plotTable and csvTable
type Table interface {
	SetHeader([]string)
	AddRow([]string)
	SetFooter([]string)
	Done()
}

func NewPlotTable(w io.Writer) Table {
	return &plotTable{tablewriter.NewWriter(w)}
}
func NewCSVTable(w io.Writer) Table {
	return &csvTable{csv.NewWriter(w)}
}

type plotTable struct {
	*tablewriter.Table
}

func (t *plotTable) SetHeader(header []string) {
	t.Table.SetHeader(header)
}
func (t *plotTable) AddRow(row []string) {
	t.Table.Append(row)
}
func (t *plotTable) SetFooter(footer []string) {
	t.Table.SetFooter(footer)
}
func (t *plotTable) Done() {
	t.Table.Render()
}

type csvTable struct {
	*csv.Writer
}

func (c *csvTable) write(row []string) {
	if e := c.Writer.Write(row); e != nil {
		log.Panic(e)
	}
}
func (c *csvTable) SetHeader(header []string) {
	c.write(header)
}
func (c *csvTable) AddRow(row []string) {
	c.write(row)
}
func (c *csvTable) SetFooter(footer []string) {
	c.write(footer)
}
func (c *csvTable) Done() {
	c.Writer.Flush()
}

func (d *Document) Pretty(vocab *Vocab) string {
	var buf bytes.Buffer
	for t, n := range d.Terms {
		if n > 1 {
			fmt.Fprintf(&buf, "%dx", n)
		}
		fmt.Fprintf(&buf, "%s ", vocab.Term(t))
	}
	return buf.String()
}

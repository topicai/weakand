package weakand

import (
	"bufio"
	"io"
	"log"
	"strings"
)

type Vocab struct {
	TermIndex map[string]int // term to term-Id.
	Terms     []string       // term-Id to term.
}

// If in is not nil, load terms from it and fill into the initialized Vocab.
func NewVocab(in io.Reader) *Vocab {
	v := &Vocab{
		TermIndex: make(map[string]int),
		Terms:     make([]string, 0),
	}

	if in != nil {
		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			fs := strings.Fields(scanner.Text())
			// Assumes that each line has multiple fields, and the last one is the term.
			if len(fs) > 0 {
				v.IdOrAdd(fs[len(fs)-1])
			}
		}
		if e := scanner.Err(); e != nil {
			log.Panicf("Parsing vocab error %v", e)
		}
	}
	return v
}

// IdOrAdd returns TermId of a term.  It adds the term into v, if it
// is not already in.  This mutation makes IdOrAdd not thread-safe.
// Since IdOrAdd is called by NewDocument, which is in turn called by
// SearchIndex.add, the index building process is not thread-safe.
//
// TODO(y): Make SearchIndex.Add/AddBatch thread-safe.
func (v *Vocab) IdOrAdd(term string) TermId {
	id, ok := v.TermIndex[term]
	if !ok {
		v.Terms = append(v.Terms, term)
		id = len(v.Terms) - 1
		v.TermIndex[term] = id
		return TermId(id)
	}
	return TermId(id)
}

func (v *Vocab) Term(id TermId) string {
	return v.Terms[int(id)]
}

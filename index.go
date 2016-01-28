package weakand

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"log"
	"os"
	"sort"

	"github.com/huichen/sego"
)

type DocId uint64 // MD5 hash of document content.
type TermId int   // depends on Vocab.

type InvertedIndex map[TermId]PostingList
type PostingList []Posting
type Posting struct {
	DocId
	TF int // The term frequency in Doc.
}

type ForwardIndex map[DocId]*Document
type Document struct {
	Terms   map[TermId]int // map makes it fast to compute Σt∈q∩d U_t.
	Len     int            // sum over Terms.
	Literal string
}

type SearchIndex struct {
	Fwd   ForwardIndex
	Ivt   InvertedIndex
	Vocab *Vocab
	Sgmt  *sego.Segmenter
}

// If a word in content is not in vocab, add it into vocab.
func NewDocument(literal string, vocab *Vocab, sgmt *sego.Segmenter) *Document {
	d := &Document{
		Terms:   make(map[TermId]int),
		Len:     0,
		Literal: literal}

	for _, term := range tokenize(literal, sgmt) {
		d.Terms[vocab.IdOrAdd(term)]++
		d.Len++
	}
	return d
}

// In InvertedIndex, posting lists are sorted by asceding order DocId.
func (p PostingList) Len() int           { return len(p) }
func (p PostingList) Less(i, j int) bool { return p[i].DocId < p[j].DocId }
func (p PostingList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func NewIndex(vocab *Vocab, sgmt *sego.Segmenter) *SearchIndex {
	if vocab == nil {
		vocab = NewVocab(nil)
	}
	return &SearchIndex{
		Ivt:   make(map[TermId]PostingList),
		Fwd:   make(map[DocId]*Document),
		Vocab: vocab,
		Sgmt:  sgmt}
}

// add Add a document into the index, but not sorting posting lists.
func (idx *SearchIndex) add(doc string, changed map[TermId]int) {
	did := documentHash(doc)
	if _, ok := idx.Fwd[did]; !ok { // De-duplicate
		d := NewDocument(doc, idx.Vocab, idx.Sgmt)
		idx.Fwd[did] = d
		for term, tf := range d.Terms {
			idx.Ivt[term] = append(idx.Ivt[term], Posting{DocId: did, TF: tf})
			changed[term]++
		}
	}
}

// Add a single document into index.
func (idx *SearchIndex) Add(doc string) *SearchIndex {
	changed := make(map[TermId]int)
	idx.add(doc, changed)

	for termId := range changed {
		sort.Sort(idx.Ivt[termId])
	}

	return idx
}

// BatchAdd adds all documents read from channel corpus into the index.
func (idx *SearchIndex) BatchAdd(corpus chan string) *SearchIndex {
	changed := make(map[TermId]int)
	for doc := range corpus {
		idx.add(doc, changed)
	}

	for termId := range changed {
		sort.Sort(idx.Ivt[termId])
	}

	return idx
}

func documentHash(literal string) DocId {
	md5Bytes := md5.Sum([]byte(literal))
	return DocId(binary.BigEndian.Uint64(md5Bytes[:]))
}

func NewIndexFromFile(corpusFile string, sgmt *sego.Segmenter, dump string) *SearchIndex {
	idx := NewIndex(NewVocab(nil), sgmt)

	if len(corpusFile) > 0 {
		ch := make(chan string)
		go func() {
			WithFile(corpusFile,
				func(f *os.File) {
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						ch <- scanner.Text()
					}
					if e := scanner.Err(); e != nil {
						log.Panicf("Scanning corpus error:%v", e)
					}
				})
			close(ch)
		}()

		idx.BatchAdd(ch)
	}

	if len(dump) > 0 {
		idx.Pretty(NewCSVTable(CreateOrDie(dump)), nil, nil, 0)
	}

	return idx
}

func tokenize(doc string, sgmt *sego.Segmenter) []string {
	var terms []string
	for _, seg := range sgmt.Segment([]byte(doc)) {
		term := seg.Token().Text()
		if !AllPunctOrSpace(term) {
			terms = append(terms, term)
		}
	}
	return terms
}

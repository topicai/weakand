package main

import (
	"flag"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"sync"

	"github.com/huichen/sego"
	"github.com/wangkuiyi/weakand"
)

type SearchServer struct {
	mutex sync.RWMutex
	index *weakand.SearchIndex
	sgmt  *sego.Segmenter
}

func NewSearchServer(corpusFile, sgmtDictFile string) *SearchServer {
	sgmt := new(sego.Segmenter)
	sgmt.LoadDictionary(sgmtDictFile)

	idx := weakand.NewIndexFromFile(corpusFile, sgmt, "") // no index dump
	return &SearchServer{index: idx, sgmt: sgmt}
}

func (s *SearchServer) Add(document string, _ *int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.index.Add(document)
	return nil
}

func (s *SearchServer) Search(query string, results *[]weakand.Result) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	*results = s.index.Search(query, 50, false) // cap=50 and no debug output
	return nil
}

func main() {
	addr := flag.String("addr", ":18082", "weak-and backserver listening address")
	corpusFile := flag.String("corpus", "", "A text file where each line is a document")
	sgmtDictFile := flag.String("sgmt", "", "Segmenter dictionary file")
	flag.Parse()

	if len(*sgmtDictFile) <= 0 {
		*sgmtDictFile = path.Join(
			os.Getenv("GOPATH"),
			"src/github.com/huichen/sego/data/dictionary.txt")
	}

	rpc.Register(NewSearchServer(*corpusFile, *sgmtDictFile))
	rpc.HandleHTTP()
	if e := http.ListenAndServe(*addr, nil); e != nil {
		log.Fatal(e)
	}
}

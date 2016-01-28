package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/rpc"

	"github.com/wangkuiyi/weakand"
)

var (
	bend *rpc.Client

	inputAndSubmit = `
<html>
  <body>
    <form action="/%s/">
      <input type="text" name="text">
      <input type="submit" value="OK">
    </form>
`
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("searchHandler", r.URL)

	fmt.Fprintf(w, inputAndSubmit, "search")

	if q := r.FormValue("text"); len(q) > 0 {
		log.Printf("Search query=%s", q)
		var rs []weakand.Result
		if e := bend.Call("SearchServer.Search", q, &rs); e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
		for _, r := range rs {
			fmt.Fprintf(w, "%s<br>\n", r.Literal) // TODO(y): Print r.Score
		}
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("addHandler", r.URL)

	fmt.Fprintf(w, inputAndSubmit, "add")

	if q := r.FormValue("text"); len(q) > 0 {
		log.Println("Add document: ", q)
		if e := bend.Call("SearchServer.Add", q, nil); e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		}
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/search", http.StatusFound)
}

func main() {
	backend := flag.String("backend", ":18082", "weak-and backserver listening address")
	addr := flag.String("addr", ":18081", "frontend server listening address")
	flag.Parse()

	var e error
	bend, e = rpc.DialHTTP("tcp", *backend)
	if e != nil {
		log.Fatalf("Cannot dial backend RPC server: %v", e)
	}

	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/add/", addHandler)
	http.HandleFunc("/", redirectHandler)

	http.ListenAndServe(*addr, nil)
}

package weakand

import (
	"log"
	"os"
	"unicode"
)

func OpenOrDie(file string) *os.File {
	f, e := os.Open(file)
	if e != nil {
		log.Panic(e)
	}
	return f
}

func WithFile(file string, fn func(f *os.File)) {
	f := OpenOrDie(file)
	defer f.Close()
	fn(f)
}

func CreateOrDie(file string) *os.File {
	f, e := os.Create(file)
	if e != nil {
		log.Panic(e)
	}
	return f
}

func AllPunctOrSpace(s string) bool {
	for _, u := range s {
		if !unicode.IsPunct(u) && !unicode.IsSpace(u) {
			return false
		}
	}
	return true
}

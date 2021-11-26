package main

/*
A few notes on my implementation:
* I time-boxed myself to 2 hours.
* No tests were written to try and get the end user experience as well functioning as possible.
* The code is not written in a particularly extensible way. Some special-case logic was implemented such as
  hard-coded document titles (in constants.go) to partitition the works.
* I initially planned to have another endpoint /read/{document_title}#line_number, which would be hyper-linked from the results
  but wasn't able to implement it in time.
* A few not-so-niceties remain such as not explaining why the query "for" results in nothing (it's filtered as a stop word)
* Snippet generation is imperfect (filtering duplicates, combining)
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type JSONResult struct {
	Title string   `json:"title"`
	Lines []string `json:"lines"`
}

type Document struct {
	Title string
	Lines []string
}

type Occurrence struct {
	Doc    *Document
	Token  string
	Lineno int
}

type Searcher struct {
	Documents []*Document
	Index     map[string][]*Occurrence
}

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}
	searcher.BuildIndex()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		byt, err := json.Marshal(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(byt)
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}

	// Dummy value that does not land in the result
	curDoc := &Document{}

	for _, line := range strings.Split(string(dat), "\r\n") {
		// Ignore extra non-content bit at the end
		if line == "* END CONTENT NOTE *" {
			break
		}
		// Assuming that every work starts with a single line containing only it's title
		if isTitle := workTitles[line]; isTitle {
			// Create a new document
			curDoc = &Document{
				line,
				[]string{},
			}
			s.Documents = append(s.Documents, curDoc)
		}
		curDoc.Lines = append(curDoc.Lines, line)
	}
	return nil
}

func (s *Searcher) BuildIndex() {
	s.Index = make(map[string][]*Occurrence)
	// Apologies for this somewhat wildly nested monstrosity
	// This walks through every document, figures out which tokens
	// it contains and adds it to the inverse index once.
	for _, doc := range s.Documents {
		for lineno, line := range doc.Lines {
			for _, token := range extractTokens(line) {
				occurences, ok := s.Index[token]
				// Ignore any token occurrences for a given document past the first one
				if ok && occurences[len(occurences)-1].Doc.Title == doc.Title {
					continue
				}
				s.Index[token] = append(occurences, &Occurrence{doc, token, lineno})
			}
		}
	}
}

func (s *Searcher) Search(query string) []JSONResult {
	// Figure out which tokens occur in which documents
	// and create a mapping of document:[query tokens that appear in it]
	docOccurrences := make(map[string][]*Occurrence)
	queryTokens := extractTokens(query)

	for _, token := range queryTokens {
		for _, occ := range s.Index[token] {
			docOccurrences[occ.Doc.Title] = append(docOccurrences[occ.Doc.Title], occ)
		}
	}

	// Only show documents that contain all tokens
	for k, v := range docOccurrences {
		if len(v) != len(queryTokens) {
			delete(docOccurrences, k)
		}
	}

	results := []JSONResult{}
	for docTitle, occurences := range docOccurrences {
		result := JSONResult{docTitle, []string{}}
		for _, occ := range occurences {
			matchingLine := occ.Doc.Lines[occ.Lineno]
			matchingLine = formatLine(matchingLine, occ.Token, occ.Lineno)
			result.Lines = append(result.Lines, matchingLine)
		}
		results = append(results, result)
	}

	return results
}

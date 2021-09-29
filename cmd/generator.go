package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	defaultInput       = "https://iso639-3.sil.org/sites/iso639-3/files/downloads/iso-639-3.tab"
	httpTimeout        = 60 * time.Second
	utf8BOM            = "\uFEFF"
	inputFileSeparator = '\t'

	sourceFilePrefix = `package iso639_3

// Languages lookup table. Keys are ISO 639-3 codes
var Languages = map[string]Language{`
	sourceFileSuffix     = `}`
	languageStructFormat = `"%s": {
	ID: "%s",
	Part2B: "%s",
	Part2T: "%s",
	Part1: "%s",
	Scope: "%s",
	LanguageType: "%s",
	Name: "%s",
	Comment: "%s",
}, `
)

func main() {
	inputFile := flag.String("i", defaultInput,
		fmt.Sprintf("Path or URL to input file in tab-separated iso639-3.sil.org format (default %s)", defaultInput))
	outfile := flag.String("o", "", "Output file (default - standard output)")
	flag.Parse()

	rd := getInput(*inputFile)
	tsvReader := csv.NewReader(rd)
	tsvReader.Comma = inputFileSeparator

	langInput, err := tsvReader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading input file '%s': %v", *inputFile, err)
	}

	langInput = langInput[1:] // skip header

	wr := os.Stdout
	if *outfile != "" {
		var err error
		wr, err = os.Create(*outfile)
		if err != nil {
			log.Fatalf("Can't create output file '%s': %v", *outfile, err)
		}
	}

	outputLookup(wr, langInput)
}

func getInput(uri string) io.Reader {
	parsedUrl, err := url.Parse(uri)
	if err != nil || parsedUrl.Scheme == "" {
		f, err := os.Open(uri)
		if err != nil {
			log.Fatalf("Can't open input file '%s': %v", uri, err)
		}
		return bufio.NewReader(f)
	}

	httpClient := &http.Client{
		Timeout: httpTimeout,
	}

	r, err := httpClient.Get(uri)
	if err != nil {
		log.Fatalf("Can't download input file '%s': %v", uri, err)
	}
	defer r.Body.Close()

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Error reading response from '%s': %v", uri, err)
	}

	return bytes.NewReader(bs)
}

func outputLookup(w io.Writer, records [][]string) {
	buf := bytes.Buffer{}

	_, err := fmt.Fprintf(&buf, sourceFilePrefix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	for _, record := range records {
		_, err = fmt.Fprintf(&buf, languageStructFormat,
			record[0],
			record[0],
			record[1],
			record[2],
			record[3],
			record[4],
			record[5],
			record[6],
			record[7],
		)
		if err != nil {
			log.Fatalf("Error generating: %v", err)
		}
	}

	_, err = fmt.Fprintf(&buf, sourceFileSuffix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	outBytes, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("Error formatting generated code: %v", err)
	}

	_, err = w.Write(outBytes)
	if err != nil {
		log.Fatalf("Error writing to output: %v", err)
	}
}

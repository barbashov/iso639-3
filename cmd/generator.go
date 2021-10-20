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
	"reflect"
	"time"
)

const (
	defaultInput       = "https://iso639-3.sil.org/sites/iso639-3/files/downloads/iso-639-3.tab"
	httpTimeout        = 60 * time.Second
	utf8BOM            = "\uFEFF"
	inputFileSeparator = '\t'

	sourceFilePrefix = `package iso639_3

`

	part3Prefix = `// Languages part 3 lookup table. Keys are ISO 639-3 codes
var LanguagesPart3 = map[string]Language{
`

	part2Prefix = `// Languages part 2 lookup table. Keys are ISO 639-2 codes
var LanguagesPart2 = map[string]Language{
`

	part1Prefix = `// Languages part 1 lookup table. Keys are ISO 639-1 codes
var LanguagesPart1 = map[string]Language{
`

	lookupSuffix = `}
`

	languageStructFormat = `"%s": {	Part3: "%s", Part2B: "%s", Part2T: "%s", Part1: "%s", 	Scope: "%s", LanguageType: "%s", Name: "%s", Comment: "%s", },
`
)

var (
	languageStructFields = []struct {
		name      string
		fieldType reflect.Kind
	}{
		{"Part3", reflect.String},
		{"Part2B", reflect.String},
		{"Part2T", reflect.String},
		{"Part1", reflect.String},
		{"Scope", reflect.Uint8}, // no rune kind :(
		{"LanguageType", reflect.Uint8},
		{"Name", reflect.String},
		{"Comment", reflect.String},
	}
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

func outputStruct(w io.Writer, key string, record []string) error {
	if len(record) != len(languageStructFields) {
		log.Fatalf("outputStruct got malformed record: %v", record)
	}

	_, err := fmt.Fprintf(w, `"%s": {`, key)
	if err != nil {
		return err
	}

	comma := false
	for i, value := range record {
		if value == "" {
			continue
		}

		if comma {
			_, err = fmt.Fprint(w, ", ")
			if err != nil {
				return err
			}
		}

		field := languageStructFields[i]

		if field.fieldType == reflect.String {
			_, err = fmt.Fprintf(w, `%s: "%s"`, field.name, value)
		} else if field.fieldType == reflect.Uint8 {
			_, err = fmt.Fprintf(w, `%s: '%s'`, field.name, value)
		} else {
			return fmt.Errorf("unknown field kind: %v", field)
		}

		if err != nil {
			return err
		}

		comma = true
	}

	_, err = fmt.Fprintln(w, "},")
	return err
}

func outputLookup(w io.Writer, records [][]string) {
	buf := bytes.Buffer{}

	_, err := fmt.Fprintf(&buf, sourceFilePrefix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	/* Part 3 lookup */

	_, err = fmt.Fprintf(&buf, part3Prefix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	for _, record := range records {
		key := record[0]
		err = outputStruct(&buf, key, record)
		if err != nil {
			log.Fatalf("Error generating: %v", err)
		}
	}

	_, err = fmt.Fprintf(&buf, lookupSuffix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	/* Part 2 lookup */

	_, err = fmt.Fprintf(&buf, part2Prefix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	for _, record := range records {
		key2b := record[1]
		key2t := record[2]
		if key2b == "" {
			continue
		}

		err = outputStruct(&buf, key2b, record)
		if err != nil {
			log.Fatalf("Error generating: %v", err)
		}

		// there are no conflicts between part2b and part2t identifiers so we're allowed to do that
		if key2b != key2t {
			err = outputStruct(&buf, key2t, record)
			if err != nil {
				log.Fatalf("Error generating: %v", err)
			}
		}
	}

	_, err = fmt.Fprintf(&buf, lookupSuffix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	/* Part 1 lookup */

	_, err = fmt.Fprintf(&buf, part1Prefix)
	if err != nil {
		log.Fatalf("Error generating: %v", err)
	}

	for _, record := range records {
		key := record[3]
		if key == "" {
			continue
		}

		err = outputStruct(&buf, key, record)
		if err != nil {
			log.Fatalf("Error generating: %v", err)
		}
	}

	_, err = fmt.Fprintf(&buf, lookupSuffix)
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

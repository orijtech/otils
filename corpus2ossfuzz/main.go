// nolint: gosec

// corpus2fuzz converts a corpus in native Go format to OSS-Fuzz format. It is
// intended as a stop-gap until OSS-Fuzz is able to read native corpus files.
//
// See https://github.com/AdaLogics/go-fuzz-headers/blob/main/consumer.go for the
// OSS-Fuzz data format.
package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	corpusDir := flag.String("corpus", "", `the directory containing the Go corpus`)
	output := flag.String("o", "", `the output zip file to write the OSS-Fuzz corpus`)
	flag.Parse()

	initCorpus(*corpusDir, *output)
}

func initCorpus(corpusDir, output string) {
	log.SetFlags(0)
	log.SetPrefix("corpus2ossfuzz: ")

	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("ignoring non-existing corpus directory %q", corpusDir)
			os.Exit(0)
		}
		log.Fatal(err)
	}

	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	z := zip.NewWriter(f)
	for _, entry := range entries {
		filename := filepath.Join(corpusDir, entry.Name())
		datum, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}

		ossDatum, err := convert(datum)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to convert %q: %w", filename, err))
		}
		w, err := z.Create(entry.Name())
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create zip entry %q to %q: %w", entry.Name(), output, err))
		}
		if _, err := io.Copy(w, bytes.NewReader(ossDatum)); err != nil {
			log.Fatalf("can't write zip entry %q in %q: %w", entry.Name(), output, err)
		}

		log.Printf("added %q", filename)
	}
	if err := z.Close(); err != nil {
		log.Fatal(fmt.Errorf("failed to close zip writer for %q: %w", output, err))
	}
	if err := f.Close(); err != nil {
		log.Fatal(fmt.Errorf("closing %q: %w", output, err))
	}
}

// convert a Go fuzz corpus file to OSS-Fuzz format.
func convert(fdata []byte) ([]byte, error) {
	v1header := []byte("go test fuzz v1\n")
	if !bytes.HasPrefix(fdata, v1header) {
		return nil, errors.New("missing or unsupported Go fuzz header")
	}
	fdata = fdata[len(v1header):]
	var result []byte
	for len(fdata) > 0 {
		arg, rest, _ := bytes.Cut(fdata, []byte("\n"))
		fdata = rest
		switch {
		case bytes.HasPrefix(arg, []byte(`[]byte(`)) && bytes.HasSuffix(arg, []byte(`)`)):
			enc := arg[len(`[]byte(`) : len(arg)-len(`)`)]
			data, err := strconv.Unquote(string(enc))
			if err != nil {
				return nil, fmt.Errorf("failed to unquote []byte: %s", arg)
			}
			// OSS-Fuzz byte slice length is encoded in a byte
			n := byte(len(data))
			result = append(result, n)
			result = append(result, data[:n]...)
		// TODO: add more data types.
		default:
			return nil, fmt.Errorf("unsupported argument type: %s", arg)
		}
	}
	return result, nil
}

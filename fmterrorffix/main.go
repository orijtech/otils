package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var reBadFmtErrorf = regexp.MustCompile("(?P<fmt>fmt.Errorf)\\((?P<args>(\"[^%]*\"|`[^%]*`))\\)")

func main() {
	dirPath := "."
	if len(os.Args) > 1 {
		dirPath = os.Args[1]
	}

	targetsCh := make(chan string, 10)
	go func() {
		defer close(targetsCh)

		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				panic(err)
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".go") {
				return nil
			}
			if strings.Contains(path, "vendor") {
				return nil
			}

			targetsCh <- path
			return nil
		})
		if err != nil {
			panic(err)
		}
	}()

	for path := range targetsCh {
		if err := searchAndReplace(path); err != nil {
			panic(err)
		}
	}
}

func searchAndReplace(fullPath string) error {
	f, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	blob, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	if !reBadFmtErrorf.Match(blob) {
		return nil
	}

	// Preserve the prior permissions.
	fi, err := f.Stat()
	f.Close()
	if err != nil {
		return nil
	}
	wf, err := os.OpenFile(fullPath, os.O_WRONLY, fi.Mode())
	if err != nil {
		return err
	}
	defer wf.Close()

	ml := reBadFmtErrorf.ReplaceAll(blob, []byte("errors.New(${args})"))
	if _, err := wf.Write(ml); err != nil {
		panic(err)
	}
	return nil
}

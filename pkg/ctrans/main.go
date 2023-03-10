package main

import (
	"bytes"
	"flag"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"

	gt "github.com/bas24/googletranslatefree"
)

func main() {
	var dir, sourceLang, targetLang string
	flag.StringVar(&dir, "d", "", "directory to translate comments in")
	flag.StringVar(&sourceLang, "s", "ru", "source language")
	flag.StringVar(&targetLang, "t", "en", "target language")
	flag.Parse()

	filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {

		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		log.Printf("Processing %v\n", info.Name())
		fset := token.NewFileSet()

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Printf("Error parsing: %v\n", info.Name())
			return err
		}

		for _, gr := range f.Comments {
			for _, c := range gr.List {
				result, err := gt.Translate(c.Text, sourceLang, targetLang)
				if err != nil {
					log.Printf("Error translating: %v %v\n", info.Name(), err)
					continue
				}
				c.Text = result
			}

		}
		os.Remove(path)
		sourceF, err := os.OpenFile(path, os.O_CREATE, 0777)
		if err != nil {
			log.Printf("Error opening file: %v %v\n", info.Name(), err)
			return err
		}
		var b bytes.Buffer
		printer.Fprint(&b, fset, f)
		formatted, err := format.Source(b.Bytes())
		if err != nil {
			return err
		}
		_, err = sourceF.Write(formatted)
		if err != nil {
			return err
		}
		return nil
	})

}

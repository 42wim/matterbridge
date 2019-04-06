// +build ignore

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
)

var tengoModFileRE = regexp.MustCompile(`^srcmod_(\w+).tengo$`)

func main() {
	modules := make(map[string]string)

	// enumerate all Tengo module files
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		m := tengoModFileRE.FindStringSubmatch(file.Name())
		if m != nil {
			modName := m[1]

			src, err := ioutil.ReadFile(file.Name())
			if err != nil {
				log.Fatalf("file '%s' read error: %s", file.Name(), err.Error())
			}

			modules[modName] = string(src)
		}
	}

	var out bytes.Buffer
	out.WriteString(`// Code generated using gensrcmods.go; DO NOT EDIT.

package stdlib

// SourceModules are source type standard library modules.
var SourceModules = map[string]string{` + "\n")
	for modName, modSrc := range modules {
		out.WriteString("\t\"" + modName + "\": " + strconv.Quote(modSrc) + ",\n")
	}
	out.WriteString("}\n")

	const target = "source_modules.go"
	if err := ioutil.WriteFile(target, out.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

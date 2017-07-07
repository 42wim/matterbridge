package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/charset/iconv"
	"io"
	"os"
	"strings"
)

var listFlag = flag.Bool("l", false, "list available character sets")
var verboseFlag = flag.Bool("v", false, "list more information")
var fromCharset = flag.String("f", "utf-8", "translate from this character set")
var toCharset = flag.String("t", "utf-8", "translate to this character set")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: tcs [-l] [-v] [charset]\n")
		fmt.Fprintf(os.Stderr, "\ttcs [-f charset] [-t charset] [file]\n")
	}
	flag.Parse()
	if *listFlag {
		cs := ""
		switch flag.NArg() {
		case 1:
			cs = flag.Arg(0)
		case 0:
		default:
			flag.Usage()
		}
		listCharsets(*verboseFlag, cs)
		return
	}
	var f *os.File
	switch flag.NArg() {
	case 0:
		f = os.Stdin
	case 1:
		var err error
		f, err = os.Open(flag.Arg(0))
		if err != nil {
			fatalf("cannot open %q: %v", err)
		}
	}
	r, err := charset.NewReader(*fromCharset, f)
	if err != nil {
		fatalf("cannot translate from %q: %v", *fromCharset, err)
	}
	w, err := charset.NewWriter(*toCharset, os.Stdout)
	if err != nil {
		fatalf("cannot translate to %q: ", err)
	}
	_, err = io.Copy(w, r)
	if err != nil {
		fatalf("%v", err)
	}
}

func listCharsets(verbose bool, csname string) {
	var buf bytes.Buffer
	if !verbose {
		if csname != "" {
			cs := charset.Info(csname)
			if cs == nil {
				fatalf("no such charset %q", csname)
			}
			fmt.Fprintf(&buf, "%s %s\n", cs.Name, strings.Join(cs.Aliases, " "))
		} else {
			fmt.Fprintf(&buf, "%v\n", strings.Join(charset.Names(), " "))
		}
	} else {
		var charsets []*charset.Charset
		if csname != "" {
			cs := charset.Info(csname)
			if cs == nil {
				fatalf("no such charset %q", csname)
			}
			charsets = []*charset.Charset{cs}
		} else {
			for _, name := range charset.Names() {
				if cs := charset.Info(name); cs != nil {
					charsets = append(charsets, cs)
				}
			}
		}
		for _, cs := range charsets {
			fmt.Fprintf(&buf, "%s %s\n", cs.Name, strings.Join(cs.Aliases, " "))
			if cs.Desc != "" {
				fmt.Fprintf(&buf, "\t%s\n", cs.Desc)
			}
		}
	}
	os.Stdout.Write(buf.Bytes())
}

func fatalf(f string, a ...interface{}) {
	s := fmt.Sprintf(f, a...)
	fmt.Fprintf(os.Stderr, "%s\n", s)
	os.Exit(2)
}

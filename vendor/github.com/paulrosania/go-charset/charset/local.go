package charset

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	readLocalCharsetsOnce sync.Once
	localCharsets         = make(map[string]*localCharset)
)

type localCharset struct {
	Charset
	arg string
	*class
}

// A class of character sets.
// Each class can be instantiated with an argument specified in the config file.
// Many character sets can use a single class.
type class struct {
	from, to func(arg string) (Translator, error)
}

// The set of classes, indexed by class name.
var classes = make(map[string]*class)

func registerClass(charset string, from, to func(arg string) (Translator, error)) {
	classes[charset] = &class{from, to}
}

type localFactory struct{}

func (f localFactory) TranslatorFrom(name string) (Translator, error) {
	f.init()
	name = NormalizedName(name)
	cs := localCharsets[name]
	if cs == nil {
		return nil, fmt.Errorf("character set %q not found", name)
	}
	if cs.from == nil {
		return nil, fmt.Errorf("cannot translate from %q", name)
	}
	return cs.from(cs.arg)
}

func (f localFactory) TranslatorTo(name string) (Translator, error) {
	f.init()
	name = NormalizedName(name)
	cs := localCharsets[name]
	if cs == nil {
		return nil, fmt.Errorf("character set %q not found", name)
	}
	if cs.to == nil {
		return nil, fmt.Errorf("cannot translate to %q", name)
	}
	return cs.to(cs.arg)
}

func (f localFactory) Names() []string {
	f.init()
	var names []string
	for name, cs := range localCharsets {
		// add names only for non-aliases.
		if localCharsets[cs.Name] == cs {
			names = append(names, name)
		}
	}
	return names
}

func (f localFactory) Info(name string) *Charset {
	f.init()
	lcs := localCharsets[NormalizedName(name)]
	if lcs == nil {
		return nil
	}
	// copy the charset info so that callers can't mess with it.
	cs := lcs.Charset
	return &cs
}

func (f localFactory) init() {
	readLocalCharsetsOnce.Do(readLocalCharsets)
}

// charsetEntry is the data structure for one entry in the JSON config file.
// If Alias is non-empty, it should be the canonical name of another
// character set; otherwise Class should be the name
// of an entry in classes, and Arg is the argument for
// instantiating it.
type charsetEntry struct {
	Aliases []string
	Desc    string
	Class   string
	Arg     string
}

// readCharsets reads the JSON config file.
// It's done once only, when first needed.
func readLocalCharsets() {
	csdata, err := readFile("charsets.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "charset: cannot open \"charsets.json\": %v\n", err)
		return
	}

	var entries map[string]charsetEntry
	err = json.Unmarshal(csdata, &entries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "charset: cannot decode config file: %v\n", err)
	}
	for name, e := range entries {
		class := classes[e.Class]
		if class == nil {
			continue
		}
		name = NormalizedName(name)
		for i, a := range e.Aliases {
			e.Aliases[i] = NormalizedName(a)
		}
		cs := &localCharset{
			Charset: Charset{
				Name:    name,
				Aliases: e.Aliases,
				Desc:    e.Desc,
				NoFrom:  class.from == nil,
				NoTo:    class.to == nil,
			},
			arg:   e.Arg,
			class: class,
		}
		localCharsets[cs.Name] = cs
		for _, a := range cs.Aliases {
			localCharsets[a] = cs
		}
	}
}

// A general cache store that local character set translators
// can use for persistent storage of data.
var (
	cacheMutex sync.Mutex
	cacheStore = make(map[interface{}]interface{})
)

func cache(key interface{}, f func() (interface{}, error)) (interface{}, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if x := cacheStore[key]; x != nil {
		return x, nil
	}
	x, err := f()
	if err != nil {
		return nil, err
	}
	cacheStore[key] = x
	return x, err
}

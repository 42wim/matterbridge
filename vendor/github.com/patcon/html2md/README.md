
Html2md is a Go library for converting html to markdown.

# Installation

If you have [gopm](https://github.com/gpmgo/gopm) installed, 

	gopm get github.com/lunny/html2md
	
Or

	go get github.com/lunny/html2md

# Usage

* Html2md already has some built-in html tag rules. For basic use:

```Go
    md := html2md.Convert(html)
```

* If you want to add your own rules, you can

```Go
   html2md.AddRule(&html2md.Rule{
       patterns: []string{"hr"},
	   tp:       Void,
	   replacement: func(innerHTML string, attrs []string) string {
			return "\n\n* * *\n"
		},
   })
```

or

```Go
html2md.AddConvert(func(content string) string {
    return strings.ToLower(content)
})
```

# Docs

* [GoDoc](http://godoc.org/github.com/lunny/html2md)

* [GoWalker](http://gowalker.org/github.com/lunny/html2md)

# LICENSE

 BSD License
 [http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)

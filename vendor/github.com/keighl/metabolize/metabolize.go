package metabolize

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	TagName    = `meta`
	htmlRegion = `head`
	htmlTag    = `meta`
)

var (
	NotStructError = fmt.Errorf(`Destination is not a struct`)
)

type MetaData map[string]string

type Meta struct {
	Title string `meta:og:title`
	Desc  string `meta:og:image`
}

func Metabolize(doc io.Reader, obj interface{}) error {
	data, err := ParseDocument(doc)
	if err != nil {
		return err
	}
	return Decode(data, obj)
}

func Decode(data MetaData, obj interface{}) error {
	elem := reflect.ValueOf(obj).Elem()
	if elem.Kind() != reflect.Struct {
		return NotStructError
	}

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)

		fieldValue := elem.FieldByName(field.Name)
		if !fieldValue.IsValid() {
			continue
		}
		if !fieldValue.CanSet() {
			continue
		}

		tag := field.Tag.Get(TagName)
		if tag == "" {
			continue
		}

		tags := strings.Split(tag, ",")
		for _, tagItem := range tags {
			if data[tagItem] == "" {
				continue
			}

			if fieldValue.Kind() == reflect.String {
				val := string(data[tagItem])
				fieldValue.SetString(val)
			}

			if fieldValue.Kind() == reflect.Bool {
				val, err := strconv.ParseBool(data[tagItem])
				if err != nil {
					continue
				}
				fieldValue.SetBool(val)
			}

			if fieldValue.Kind() == reflect.Float64 {
				val, err := strconv.ParseFloat(data[tagItem], 64)
				if err != nil {
					continue
				}
				fieldValue.SetFloat(val)
			}

			if fieldValue.Kind() == reflect.Int64 {
				val, err := strconv.ParseInt(data[tagItem], 0, 64)
				if err != nil {
					continue
				}
				fieldValue.SetInt(val)
			}

			if field.Type.Name() == "URL" {
				val, err := url.Parse(data[tagItem])
				if err != nil {
					continue
				}
				fieldValue.Set(reflect.ValueOf(*val))
			}

			if field.Type.Name() == "Time" {
				val, err := time.Parse(time.RFC3339, data[tagItem])
				if err != nil {
					continue
				}
				fieldValue.Set(reflect.ValueOf(val))
			}
		}
	}

	return nil
}

func ParseDocument(doc io.Reader) (MetaData, error) {
	data := MetaData{}
	tokenizer := html.NewTokenizer(doc)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				return data, nil
			}
			return nil, tokenizer.Err()
		}

		token := tokenizer.Token()

		if token.Type == html.EndTagToken && token.Data == htmlRegion {
			return data, nil
		}

		if token.Data == htmlTag {
			var property, content string
			for _, attr := range token.Attr {
				switch attr.Key {
				case "property", "name":
					property = strings.ToLower(attr.Val)
				case "content":
					content = attr.Val
				}
			}

			if property != "" {
				data[strings.TrimSpace(property)] = content
			}

		}
	}
	return data, nil
}

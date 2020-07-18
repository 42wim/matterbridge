package msauth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ReadLocation reads data from file with path or URL
func ReadLocation(loc string) ([]byte, error) {
	u, err := url.Parse(loc)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "", "file":
		return ioutil.ReadFile(u.Path)
	case "http", "https":
		res, err := http.Get(loc)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%s", res.Status)
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, fmt.Errorf("Unsupported location to load: %s", loc)
}

// WriteLocation writes data to file with path or URL
func WriteLocation(loc string, b []byte, m os.FileMode) error {
	u, err := url.Parse(loc)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "", "file":
		return ioutil.WriteFile(u.Path, b, m)
	case "http", "https":
		if strings.HasSuffix(u.Host, ".blob.core.windows.net") {
			// Azure Blob Storage URL with SAS assumed here
			cli := &http.Client{}
			req, err := http.NewRequest(http.MethodPut, loc, bytes.NewBuffer(b))
			if err != nil {
				return err
			}
			req.Header.Set("x-ms-blob-type", "BlockBlob")
			res, err := cli.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusCreated {
				return fmt.Errorf("%s", res.Status)
			}
			return nil
		}
	}
	return fmt.Errorf("Unsupported location to save: %s", loc)
}

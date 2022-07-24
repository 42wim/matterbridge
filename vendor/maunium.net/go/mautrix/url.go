// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mautrix

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func ParseAndNormalizeBaseURL(homeserverURL string) (*url.URL, error) {
	hsURL, err := url.Parse(homeserverURL)
	if err != nil {
		return nil, err
	}
	if hsURL.Scheme == "" {
		hsURL.Scheme = "https"
		fixedURL := hsURL.String()
		hsURL, err = url.Parse(fixedURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse fixed URL '%s': %v", fixedURL, err)
		}
	}
	hsURL.RawPath = hsURL.EscapedPath()
	return hsURL, nil
}

// BuildURL builds a URL with the given path parts
func BuildURL(baseURL *url.URL, path ...interface{}) *url.URL {
	createdURL := *baseURL
	rawParts := make([]string, len(path)+1)
	rawParts[0] = strings.TrimSuffix(createdURL.RawPath, "/")
	parts := make([]string, len(path)+1)
	parts[0] = strings.TrimSuffix(createdURL.Path, "/")
	for i, part := range path {
		switch casted := part.(type) {
		case string:
			parts[i+1] = casted
		case int:
			parts[i+1] = strconv.Itoa(casted)
		case fmt.Stringer:
			parts[i+1] = casted.String()
		default:
			parts[i+1] = fmt.Sprint(casted)
		}
		rawParts[i+1] = url.PathEscape(parts[i+1])
	}
	createdURL.Path = strings.Join(parts, "/")
	createdURL.RawPath = strings.Join(rawParts, "/")
	return &createdURL
}

// BuildURL builds a URL with the Client's homeserver and appservice user ID set already.
func (cli *Client) BuildURL(urlPath PrefixableURLPath) string {
	return cli.BuildURLWithQuery(urlPath, nil)
}

// BuildClientURL builds a URL with the Client's homeserver and appservice user ID set already.
// This method also automatically prepends the client API prefix (/_matrix/client).
func (cli *Client) BuildClientURL(urlPath ...interface{}) string {
	return cli.BuildURLWithQuery(ClientURLPath(urlPath), nil)
}

type PrefixableURLPath interface {
	FullPath() []interface{}
}

type BaseURLPath []interface{}

func (bup BaseURLPath) FullPath() []interface{} {
	return bup
}

type ClientURLPath []interface{}

func (cup ClientURLPath) FullPath() []interface{} {
	return append([]interface{}{"_matrix", "client"}, []interface{}(cup)...)
}

type MediaURLPath []interface{}

func (mup MediaURLPath) FullPath() []interface{} {
	return append([]interface{}{"_matrix", "media"}, []interface{}(mup)...)
}

// BuildURLWithQuery builds a URL with query parameters in addition to the Client's homeserver
// and appservice user ID set already.
func (cli *Client) BuildURLWithQuery(urlPath PrefixableURLPath, urlQuery map[string]string) string {
	hsURL := *BuildURL(cli.HomeserverURL, urlPath.FullPath()...)
	query := hsURL.Query()
	if cli.SetAppServiceUserID {
		query.Set("user_id", string(cli.UserID))
	}
	if urlQuery != nil {
		for k, v := range urlQuery {
			query.Set(k, v)
		}
	}
	hsURL.RawQuery = query.Encode()
	return hsURL.String()
}

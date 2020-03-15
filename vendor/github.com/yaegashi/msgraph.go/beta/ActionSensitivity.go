// Code generated by msgraph-generate.go DO NOT EDIT.

package msgraph

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/yaegashi/msgraph.go/jsonx"
)

// SensitivityLabelCollectionEvaluateRequestParameter undocumented
type SensitivityLabelCollectionEvaluateRequestParameter struct {
	// DiscoveredSensitiveTypes undocumented
	DiscoveredSensitiveTypes []DiscoveredSensitiveType `json:"discoveredSensitiveTypes,omitempty"`
	// CurrentLabel undocumented
	CurrentLabel *CurrentLabel `json:"currentLabel,omitempty"`
}

// Sublabels returns request builder for SensitivityLabel collection
func (b *SensitivityLabelRequestBuilder) Sublabels() *SensitivityLabelSublabelsCollectionRequestBuilder {
	bb := &SensitivityLabelSublabelsCollectionRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.baseURL += "/sublabels"
	return bb
}

// SensitivityLabelSublabelsCollectionRequestBuilder is request builder for SensitivityLabel collection
type SensitivityLabelSublabelsCollectionRequestBuilder struct{ BaseRequestBuilder }

// Request returns request for SensitivityLabel collection
func (b *SensitivityLabelSublabelsCollectionRequestBuilder) Request() *SensitivityLabelSublabelsCollectionRequest {
	return &SensitivityLabelSublabelsCollectionRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client},
	}
}

// ID returns request builder for SensitivityLabel item
func (b *SensitivityLabelSublabelsCollectionRequestBuilder) ID(id string) *SensitivityLabelRequestBuilder {
	bb := &SensitivityLabelRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.baseURL += "/" + id
	return bb
}

// SensitivityLabelSublabelsCollectionRequest is request for SensitivityLabel collection
type SensitivityLabelSublabelsCollectionRequest struct{ BaseRequest }

// Paging perfoms paging operation for SensitivityLabel collection
func (r *SensitivityLabelSublabelsCollectionRequest) Paging(ctx context.Context, method, path string, obj interface{}, n int) ([]SensitivityLabel, error) {
	req, err := r.NewJSONRequest(method, path, obj)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	var values []SensitivityLabel
	for {
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			b, _ := ioutil.ReadAll(res.Body)
			errRes := &ErrorResponse{Response: res}
			err := jsonx.Unmarshal(b, errRes)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", res.Status, string(b))
			}
			return nil, errRes
		}
		var (
			paging Paging
			value  []SensitivityLabel
		)
		err := jsonx.NewDecoder(res.Body).Decode(&paging)
		if err != nil {
			return nil, err
		}
		err = jsonx.Unmarshal(paging.Value, &value)
		if err != nil {
			return nil, err
		}
		values = append(values, value...)
		if n >= 0 {
			n--
		}
		if n == 0 || len(paging.NextLink) == 0 {
			return values, nil
		}
		req, err = http.NewRequest("GET", paging.NextLink, nil)
		if ctx != nil {
			req = req.WithContext(ctx)
		}
		res, err = r.client.Do(req)
		if err != nil {
			return nil, err
		}
	}
}

// GetN performs GET request for SensitivityLabel collection, max N pages
func (r *SensitivityLabelSublabelsCollectionRequest) GetN(ctx context.Context, n int) ([]SensitivityLabel, error) {
	var query string
	if r.query != nil {
		query = "?" + r.query.Encode()
	}
	return r.Paging(ctx, "GET", query, nil, n)
}

// Get performs GET request for SensitivityLabel collection
func (r *SensitivityLabelSublabelsCollectionRequest) Get(ctx context.Context) ([]SensitivityLabel, error) {
	return r.GetN(ctx, 0)
}

// Add performs POST request for SensitivityLabel collection
func (r *SensitivityLabelSublabelsCollectionRequest) Add(ctx context.Context, reqObj *SensitivityLabel) (resObj *SensitivityLabel, err error) {
	err = r.JSONRequest(ctx, "POST", "", reqObj, &resObj)
	return
}

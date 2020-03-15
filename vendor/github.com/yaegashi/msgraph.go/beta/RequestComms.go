// Code generated by msgraph-generate.go DO NOT EDIT.

package msgraph

import "context"

// CommsApplicationRequestBuilder is request builder for CommsApplication
type CommsApplicationRequestBuilder struct{ BaseRequestBuilder }

// Request returns CommsApplicationRequest
func (b *CommsApplicationRequestBuilder) Request() *CommsApplicationRequest {
	return &CommsApplicationRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client},
	}
}

// CommsApplicationRequest is request for CommsApplication
type CommsApplicationRequest struct{ BaseRequest }

// Get performs GET request for CommsApplication
func (r *CommsApplicationRequest) Get(ctx context.Context) (resObj *CommsApplication, err error) {
	var query string
	if r.query != nil {
		query = "?" + r.query.Encode()
	}
	err = r.JSONRequest(ctx, "GET", query, nil, &resObj)
	return
}

// Update performs PATCH request for CommsApplication
func (r *CommsApplicationRequest) Update(ctx context.Context, reqObj *CommsApplication) error {
	return r.JSONRequest(ctx, "PATCH", "", reqObj, nil)
}

// Delete performs DELETE request for CommsApplication
func (r *CommsApplicationRequest) Delete(ctx context.Context) error {
	return r.JSONRequest(ctx, "DELETE", "", nil, nil)
}

// CommsOperationRequestBuilder is request builder for CommsOperation
type CommsOperationRequestBuilder struct{ BaseRequestBuilder }

// Request returns CommsOperationRequest
func (b *CommsOperationRequestBuilder) Request() *CommsOperationRequest {
	return &CommsOperationRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client},
	}
}

// CommsOperationRequest is request for CommsOperation
type CommsOperationRequest struct{ BaseRequest }

// Get performs GET request for CommsOperation
func (r *CommsOperationRequest) Get(ctx context.Context) (resObj *CommsOperation, err error) {
	var query string
	if r.query != nil {
		query = "?" + r.query.Encode()
	}
	err = r.JSONRequest(ctx, "GET", query, nil, &resObj)
	return
}

// Update performs PATCH request for CommsOperation
func (r *CommsOperationRequest) Update(ctx context.Context, reqObj *CommsOperation) error {
	return r.JSONRequest(ctx, "PATCH", "", reqObj, nil)
}

// Delete performs DELETE request for CommsOperation
func (r *CommsOperationRequest) Delete(ctx context.Context) error {
	return r.JSONRequest(ctx, "DELETE", "", nil, nil)
}

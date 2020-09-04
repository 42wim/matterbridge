// Code generated by msgraph.go/gen DO NOT EDIT.

package msgraph

import "context"

//
type WorkbookFunctionsMidbRequestBuilder struct{ BaseRequestBuilder }

// Midb action undocumented
func (b *WorkbookFunctionsRequestBuilder) Midb(reqObj *WorkbookFunctionsMidbRequestParameter) *WorkbookFunctionsMidbRequestBuilder {
	bb := &WorkbookFunctionsMidbRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.BaseRequestBuilder.baseURL += "/midb"
	bb.BaseRequestBuilder.requestObject = reqObj
	return bb
}

//
type WorkbookFunctionsMidbRequest struct{ BaseRequest }

//
func (b *WorkbookFunctionsMidbRequestBuilder) Request() *WorkbookFunctionsMidbRequest {
	return &WorkbookFunctionsMidbRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client, requestObject: b.requestObject},
	}
}

//
func (r *WorkbookFunctionsMidbRequest) Post(ctx context.Context) (resObj *WorkbookFunctionResult, err error) {
	err = r.JSONRequest(ctx, "POST", "", r.requestObject, &resObj)
	return
}

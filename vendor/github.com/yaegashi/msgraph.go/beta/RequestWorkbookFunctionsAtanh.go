// Code generated by msgraph.go/gen DO NOT EDIT.

package msgraph

import "context"

//
type WorkbookFunctionsAtanhRequestBuilder struct{ BaseRequestBuilder }

// Atanh action undocumented
func (b *WorkbookFunctionsRequestBuilder) Atanh(reqObj *WorkbookFunctionsAtanhRequestParameter) *WorkbookFunctionsAtanhRequestBuilder {
	bb := &WorkbookFunctionsAtanhRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.BaseRequestBuilder.baseURL += "/atanh"
	bb.BaseRequestBuilder.requestObject = reqObj
	return bb
}

//
type WorkbookFunctionsAtanhRequest struct{ BaseRequest }

//
func (b *WorkbookFunctionsAtanhRequestBuilder) Request() *WorkbookFunctionsAtanhRequest {
	return &WorkbookFunctionsAtanhRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client, requestObject: b.requestObject},
	}
}

//
func (r *WorkbookFunctionsAtanhRequest) Post(ctx context.Context) (resObj *WorkbookFunctionResult, err error) {
	err = r.JSONRequest(ctx, "POST", "", r.requestObject, &resObj)
	return
}

// Code generated by msgraph-generate.go DO NOT EDIT.

package msgraph

import "context"

//
type WorkbookFunctionsTextRequestBuilder struct{ BaseRequestBuilder }

// Text action undocumented
func (b *WorkbookFunctionsRequestBuilder) Text(reqObj *WorkbookFunctionsTextRequestParameter) *WorkbookFunctionsTextRequestBuilder {
	bb := &WorkbookFunctionsTextRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.BaseRequestBuilder.baseURL += "/text"
	bb.BaseRequestBuilder.requestObject = reqObj
	return bb
}

//
type WorkbookFunctionsTextRequest struct{ BaseRequest }

//
func (b *WorkbookFunctionsTextRequestBuilder) Request() *WorkbookFunctionsTextRequest {
	return &WorkbookFunctionsTextRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client, requestObject: b.requestObject},
	}
}

//
func (r *WorkbookFunctionsTextRequest) Post(ctx context.Context) (resObj *WorkbookFunctionResult, err error) {
	err = r.JSONRequest(ctx, "POST", "", r.requestObject, &resObj)
	return
}

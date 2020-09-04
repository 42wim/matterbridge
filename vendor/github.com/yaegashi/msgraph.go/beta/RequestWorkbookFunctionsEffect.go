// Code generated by msgraph.go/gen DO NOT EDIT.

package msgraph

import "context"

//
type WorkbookFunctionsEffectRequestBuilder struct{ BaseRequestBuilder }

// Effect action undocumented
func (b *WorkbookFunctionsRequestBuilder) Effect(reqObj *WorkbookFunctionsEffectRequestParameter) *WorkbookFunctionsEffectRequestBuilder {
	bb := &WorkbookFunctionsEffectRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.BaseRequestBuilder.baseURL += "/effect"
	bb.BaseRequestBuilder.requestObject = reqObj
	return bb
}

//
type WorkbookFunctionsEffectRequest struct{ BaseRequest }

//
func (b *WorkbookFunctionsEffectRequestBuilder) Request() *WorkbookFunctionsEffectRequest {
	return &WorkbookFunctionsEffectRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client, requestObject: b.requestObject},
	}
}

//
func (r *WorkbookFunctionsEffectRequest) Post(ctx context.Context) (resObj *WorkbookFunctionResult, err error) {
	err = r.JSONRequest(ctx, "POST", "", r.requestObject, &resObj)
	return
}

// Code generated by msgraph.go/gen DO NOT EDIT.

package msgraph

import "context"

//
type WorkbookFunctionsDec2BinRequestBuilder struct{ BaseRequestBuilder }

// Dec2Bin action undocumented
func (b *WorkbookFunctionsRequestBuilder) Dec2Bin(reqObj *WorkbookFunctionsDec2BinRequestParameter) *WorkbookFunctionsDec2BinRequestBuilder {
	bb := &WorkbookFunctionsDec2BinRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.BaseRequestBuilder.baseURL += "/dec2Bin"
	bb.BaseRequestBuilder.requestObject = reqObj
	return bb
}

//
type WorkbookFunctionsDec2BinRequest struct{ BaseRequest }

//
func (b *WorkbookFunctionsDec2BinRequestBuilder) Request() *WorkbookFunctionsDec2BinRequest {
	return &WorkbookFunctionsDec2BinRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client, requestObject: b.requestObject},
	}
}

//
func (r *WorkbookFunctionsDec2BinRequest) Post(ctx context.Context) (resObj *WorkbookFunctionResult, err error) {
	err = r.JSONRequest(ctx, "POST", "", r.requestObject, &resObj)
	return
}

//
type WorkbookFunctionsDec2HexRequestBuilder struct{ BaseRequestBuilder }

// Dec2Hex action undocumented
func (b *WorkbookFunctionsRequestBuilder) Dec2Hex(reqObj *WorkbookFunctionsDec2HexRequestParameter) *WorkbookFunctionsDec2HexRequestBuilder {
	bb := &WorkbookFunctionsDec2HexRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.BaseRequestBuilder.baseURL += "/dec2Hex"
	bb.BaseRequestBuilder.requestObject = reqObj
	return bb
}

//
type WorkbookFunctionsDec2HexRequest struct{ BaseRequest }

//
func (b *WorkbookFunctionsDec2HexRequestBuilder) Request() *WorkbookFunctionsDec2HexRequest {
	return &WorkbookFunctionsDec2HexRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client, requestObject: b.requestObject},
	}
}

//
func (r *WorkbookFunctionsDec2HexRequest) Post(ctx context.Context) (resObj *WorkbookFunctionResult, err error) {
	err = r.JSONRequest(ctx, "POST", "", r.requestObject, &resObj)
	return
}

//
type WorkbookFunctionsDec2OctRequestBuilder struct{ BaseRequestBuilder }

// Dec2Oct action undocumented
func (b *WorkbookFunctionsRequestBuilder) Dec2Oct(reqObj *WorkbookFunctionsDec2OctRequestParameter) *WorkbookFunctionsDec2OctRequestBuilder {
	bb := &WorkbookFunctionsDec2OctRequestBuilder{BaseRequestBuilder: b.BaseRequestBuilder}
	bb.BaseRequestBuilder.baseURL += "/dec2Oct"
	bb.BaseRequestBuilder.requestObject = reqObj
	return bb
}

//
type WorkbookFunctionsDec2OctRequest struct{ BaseRequest }

//
func (b *WorkbookFunctionsDec2OctRequestBuilder) Request() *WorkbookFunctionsDec2OctRequest {
	return &WorkbookFunctionsDec2OctRequest{
		BaseRequest: BaseRequest{baseURL: b.baseURL, client: b.client, requestObject: b.requestObject},
	}
}

//
func (r *WorkbookFunctionsDec2OctRequest) Post(ctx context.Context) (resObj *WorkbookFunctionResult, err error) {
	err = r.JSONRequest(ctx, "POST", "", r.requestObject, &resObj)
	return
}

package request

const (
	// ApplicationJSON application/json
	ApplicationJSON ContentType = "application/json"

	// ApplicationXWwwFormURLEncoded application/x-www-form-urlencoded
	ApplicationXWwwFormURLEncoded ContentType = "application/x-www-form-urlencoded"

	// MultipartFormData multipart/form-data
	MultipartFormData ContentType = "multipart/form-data"
)

const (
	// OPTIONS http options
	OPTIONS = "OPTIONS"

	// GET http get
	GET = "GET"

	// HEAD http head
	HEAD = "HEAD"

	// POST http post
	POST = "POST"

	// PUT http put
	PUT = "PUT"

	// DELETE http delete
	DELETE = "DELETE"

	// TRACE http trace
	TRACE = "TRACE"

	// CONNECT http connect
	CONNECT = "CONNECT"

	// PATCH http patch
	PATCH = "PATCH"
)

const (
	emptyString = ""
	contentType = "Content-Type"
)
